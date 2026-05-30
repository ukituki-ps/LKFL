# T2006 — Redis Cache: Каталог

## Веха

M20-catalog

## Тип

code

## Контекст

Redis cache для каталога. Кэшируем ListTypes query (самый частый запрос).
Key prefix: `catalog:` (из `doc/архитектура/schema.md` строка 1146).

**После M19 — ⚠️ Tenant в cache key:**
- Redis общий для всех tenant'ов (один контейнер, docker-compose.yml)
- Cache key **обязательно** включает tenant_id для изоляции:
  - `catalog:list:{tenant_id}:{type}:{status}:{page}`
  - `catalog:type:{tenant_id}:{type_id}`
  - `catalog:categories:{tenant_id}`
- Без tenant_id в key — cross-tenant cache pollution (пользователи tenant A увидят каталог tenant B)
- Invalidate по tenant_id (не глобально)

## Что сделать

### `internal/engagement/catalog/cache.go`

```go
package catalog

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

const (
    cachePrefix      = "catalog:"
    cacheTTL         = 5 * time.Minute
    cacheListKey     = cachePrefix + "list:%s:%s:%d" // type:status:page
    cacheTypeKey     = cachePrefix + "type:%s"       // type_id
    cacheCategoriesKey = cachePrefix + "categories:%s" // tenant_id
)

type Cache struct {
    client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
    return &Cache{client: client}
}

// GetList — получить каталог из кэша
func (c *Cache) GetList(ctx context.Context, filter CatalogFilter) ([]byte, bool) {
    key := fmt.Sprintf(cacheListKey, filter.Type, filter.Status, filter.Page)
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, false
    }
    if err != nil {
        return nil, false
    }
    return data, true
}

// SetList — сохранить каталог в кэш
func (c *Cache) SetList(ctx context.Context, filter CatalogFilter, data []byte) error {
    key := fmt.Sprintf(cacheListKey, filter.Type, filter.Status, filter.Page)
    return c.client.Set(ctx, key, data, cacheTTL).Err()
}

// Invalidate — инвалидация кэша (при admin изменении)
func (c *Cache) Invalidate(ctx context.Context, tenantID string) error {
    // Pattern: catalog:list:*
    cursor := uint64(0)
    for {
        keys, nextCursor, err := c.client.Scan(ctx, cursor, fmt.Sprintf("catalog:list:*"), 100).Result()
        if err != nil {
            return err
        }
        if len(keys) > 0 {
            c.client.Del(ctx, keys...)
        }
        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }
    return nil
}

// GetType — получить тип из кэша
func (c *Cache) GetType(ctx context.Context, id string) ([]byte, bool) {
    key := fmt.Sprintf(cacheTypeKey, id)
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, false
    }
    if err != nil {
        return nil, false
    }
    return data, true
}

// SetType — сохранить тип в кэш
func (c *Cache) SetType(ctx context.Context, id string, data []byte) error {
    key := fmt.Sprintf(cacheTypeKey, id)
    return c.client.Set(ctx, key, data, cacheTTL).Err()
}
```

### Integration в service

```go
func (s *Service) ListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error) {
    // Try cache
    if cached, ok := s.cache.GetList(ctx, filter); ok {
        var result ListResponse
        if err := json.Unmarshal(cached, &result); err == nil {
            return result.Data, result.Total, nil
        }
    }

    // DB query
    types, total, err := s.repo.ListTypes(ctx, filter)
    if err != nil {
        return nil, 0, err
    }

    // Cache set
    data, _ := json.Marshal(ListResponse{Data: types, Total: total})
    s.cache.SetList(ctx, filter, data)

    return types, total, nil
}
```

### Инвалидация при admin изменении

```go
// В admin handler:
func (h *AdminHandler) CreateType(w http.ResponseWriter, r *http.Request) {
    // ... create ...
    h.cache.Invalidate(r.Context(), tenantID)
}
```

## Требования

- Key prefix: `catalog:`
- TTL: 5 минут
- Cache ListTypes (самый частый query)
- Cache GetTypeByID
- Cache Categories
- Invalidate при admin изменении (pattern scan + del)
- Cache miss → DB fallback (не error)
- JSON serialization для cache values

## Критерии приёмки

- [ ] `cache.go` — Cache struct + methods
- [ ] GetList/SetList с TTL 5min
- [ ] GetType/SetType
- [ ] Invalidate при admin изменении
- [ ] Cache miss → DB fallback
- [ ] Integration в service
- [ ] Unit tests (mock Redis)
