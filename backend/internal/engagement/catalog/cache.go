// Package catalog — каталог энгейджментов (льготы/активности).
//
// cache.go — Redis кэш для каталога с tenant isolation.
//
// Key pattern (tenant_id обязателен для multi-tenant isolation):
//
//	catalog:list:{tenant_id}:{type}:{status}:{search}:{page}
//	catalog:type:{tenant_id}:{type_id}
//	catalog:categories:{tenant_id}
//
// При nil client кэширование отключено (silently).
package catalog

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"lkfl/internal/metrics"
)

const (
	// cachePrefix — префикс всех ключей кэша каталога.
	cachePrefix = "catalog:"

	// cacheTTL — время жизни записи в кэше.
	cacheTTL = 5 * time.Minute

	// cacheListKeyFmt — шаблон ключа для списка типов.
	// catalog:list:{tenant_id}:{type}:{status}:{search}:{page}
	cacheListKeyFmt = cachePrefix + "list:%s:%s:%s:%s:%d"

	// cacheTypeKeyFmt — шаблон ключа для отдельного типа.
	// catalog:type:{tenant_id}:{type_id}
	cacheTypeKeyFmt = cachePrefix + "type:%s:%s"

	// cacheCategoriesKeyFmt — шаблон ключа для категорий.
	// catalog:categories:{tenant_id}
	cacheCategoriesKeyFmt = cachePrefix + "categories:%s"
)

// Cache — Redis кэш для каталога.
//
// Все методы nil-safe: если client == nil, кэширование пропускается.
// Cache miss и ошибки Redis не прерывают работу — fallback к DB.
type Cache struct {
	client  *redis.Client
	metrics *metrics.Metrics
}

// NewCache создаёт Cache.
// Если client == nil, кэширование отключено.
// Если m == nil, метрики не собираются.
func NewCache(client *redis.Client, m *metrics.Metrics) *Cache {
	return &Cache{client: client, metrics: m}
}

// GetList — получить каталог из кэша.
// Возвращает (data, true) при cache hit, (nil, false) при miss или ошибке.
func (c *Cache) GetList(ctx context.Context, filter CatalogFilter) ([]byte, bool) {
	if c.client == nil {
		return nil, false
	}

	key := fmt.Sprintf(cacheListKeyFmt,
		filter.TenantID.String(),
		filter.Type,
		filter.Status,
		filter.Search,
		filter.Page,
	)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		if c.metrics != nil {
			c.metrics.RedisCacheMisses.Inc()
		}
		return nil, false
	}
	if err != nil {
		// Ошибка Redis — silent fallback к DB.
		return nil, false
	}
	if c.metrics != nil {
		c.metrics.RedisCacheHits.Inc()
	}
	return data, true
}

// SetList — сохранить каталог в кэш.
// Ошибки игнорируются (fire-and-forget).
func (c *Cache) SetList(ctx context.Context, filter CatalogFilter, data []byte) error {
	if c.client == nil {
		return nil
	}

	key := fmt.Sprintf(cacheListKeyFmt,
		filter.TenantID.String(),
		filter.Type,
		filter.Status,
		filter.Search,
		filter.Page,
	)

	return c.client.Set(ctx, key, data, cacheTTL).Err()
}

// GetType — получить тип из кэша.
// Возвращает (data, true) при cache hit, (nil, false) при miss или ошибке.
func (c *Cache) GetType(ctx context.Context, tenantID string, typeID string) ([]byte, bool) {
	if c.client == nil {
		return nil, false
	}

	key := fmt.Sprintf(cacheTypeKeyFmt, tenantID, typeID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		if c.metrics != nil {
			c.metrics.RedisCacheMisses.Inc()
		}
		return nil, false
	}
	if err != nil {
		return nil, false
	}
	if c.metrics != nil {
		c.metrics.RedisCacheHits.Inc()
	}
	return data, true
}

// SetType — сохранить тип в кэш.
func (c *Cache) SetType(ctx context.Context, tenantID string, typeID string, data []byte) error {
	if c.client == nil {
		return nil
	}

	key := fmt.Sprintf(cacheTypeKeyFmt, tenantID, typeID)
	return c.client.Set(ctx, key, data, cacheTTL).Err()
}

// GetCategories — получить категории из кэша.
// Возвращает (data, true) при cache hit, (nil, false) при miss или ошибке.
func (c *Cache) GetCategories(ctx context.Context, tenantID string) ([]byte, bool) {
	if c.client == nil {
		return nil, false
	}

	key := fmt.Sprintf(cacheCategoriesKeyFmt, tenantID)
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		if c.metrics != nil {
			c.metrics.RedisCacheMisses.Inc()
		}
		return nil, false
	}
	if err != nil {
		return nil, false
	}
	if c.metrics != nil {
		c.metrics.RedisCacheHits.Inc()
	}
	return data, true
}

// SetCategories — сохранить категории в кэш.
func (c *Cache) SetCategories(ctx context.Context, tenantID string, data []byte) error {
	if c.client == nil {
		return nil
	}

	key := fmt.Sprintf(cacheCategoriesKeyFmt, tenantID)
	return c.client.Set(ctx, key, data, cacheTTL).Err()
}

// Invalidate — инвалидация всех ключей кэша для tenant'а.
// Используется при admin изменении (create/update/delete).
//
// Сканирует ключи по паттерну catalog:*:{tenantID}:* и удаляет их.
// Это обеспечивает tenant isolation — ключи других tenant'ов не затрагиваются.
func (c *Cache) Invalidate(ctx context.Context, tenantID string) error {
	if c.client == nil {
		return nil
	}

	// Pattern: catalog:*:{tenantID}:*
	// Но для категорий ключ: catalog:categories:{tenantID} (без двоеточия в конце)
	// Поэтому удаляем оба паттерна.
	patterns := []string{
		cachePrefix + "*" + ":" + tenantID + ":*",
		cachePrefix + "*" + tenantID, // для catalog:categories:{tenantID}
	}

	for _, pattern := range patterns {
		cursor := uint64(0)
		for {
			keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				return err
			}
			if len(keys) > 0 {
				if c.metrics != nil {
					c.metrics.RedisCacheEvictions.Add(float64(len(keys)))
				}
				c.client.Del(ctx, keys...)
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}
	return nil
}
