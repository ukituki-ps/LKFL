// Package tenant — multi-tenant изоляция и middleware.
//
// WithTenantID автоматически добавляет WHERE tenant_id clause к SQL query,
// обеспечивая multi-tenant isolation на уровне базы данных.
//
// Пример использования:
//
//	func (r *pgRepository) List(ctx context.Context, filter CatalogFilter) ([]Engagement, int64, error) {
//	    baseQuery := `
//	        SELECT e.id, e.tenant_id, e.name, e.description, e.type, e.status
//	        FROM lkfl_platform.engagement_types e
//	    `
//
//	    query, args := tenant.WithTenantID(ctx, baseQuery)
//
//	    // Добавляем фильтры (продолжаем нумерацию $1 от tenant_id)
//	    query += " AND e.status = $2"
//	    args = append(args, filter.Status)
//
//	    // Pagination
//	    query += " LIMIT $3 OFFSET $4"
//	    args = append(args, filter.PerPage, (filter.Page-1)*filter.PerPage)
//
//	    rows, err := r.pool.Query(ctx, query, args...)
//	    // ...
//	}
//
// Важно:
//   - Если tenant_id = uuid.Nil — фильтр не добавляется (для admin queries)
//   - Нумерация параметров: tenant_id всегда $1, остальные параметры сдвигаются
//   - Работает с существующим WHERE (добавляет AND) и без WHERE (добавляет WHERE)
//   - Не использовать в таблице tenants (самой таблице не нужен tenant_id filter)
package tenant

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"lkfl/internal/metrics"
)

// WithTenantID добавляет tenant_id WHERE clause к query.
//
// Если tenant_id = uuid.Nil в context — не добавляет фильтр (admin queries).
// Если query уже содержит WHERE — добавляет AND tenant_id = $1.
// Если WHERE отсутствует — добавляет WHERE tenant_id = $1.
//
// Возвращает модифицированный query и slice параметров ([]interface{}{tenantID}).
// Вызывающий должен продолжить нумерацию параметров с $2.
func WithTenantID(ctx context.Context, query string) (string, []any) {
	tid := TenantIDFromContext(ctx)
	if tid == uuid.Nil {
		return query, nil
	}

	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)

	// Проверяем наличие WHERE (case-insensitive)
	// Ищем " WHERE " чтобы не спутать с WHERE в подзапросах
	hasWhere := strings.Contains(upper, " WHERE ")

	// Также проверяем, заканчивается ли query на "WHERE" (без пробела после)
	if !hasWhere && strings.HasSuffix(upper, "WHERE") {
		hasWhere = true
	}

	var clause string
	if hasWhere {
		clause = " AND tenant_id = $1"
	} else {
		clause = " WHERE tenant_id = $1"
	}

	return trimmed + clause, []any{tid}
}

// TenantContext создаёт context с заданным tenant ID.
// Используется для тестов и админ-запросов с явным tenant.
func TenantContext(ctx context.Context, tid uuid.UUID) context.Context {
	return context.WithValue(ctx, TenantIDKey, tid)
}

// WithAdminTenant создаёт context без tenant фильтрации (uuid.Nil).
// Используется для admin queries, которые должны видеть данные всех tenants.
func WithAdminTenant(ctx context.Context) context.Context {
	return context.WithValue(ctx, TenantIDKey, uuid.Nil)
}

// AdminTenantMiddleware — лёгкий middleware для admin routes.
// Извлекает tenant из X-Tenant-ID header через Service (без subdomain resolution).
// Если header отсутствует — устанавливает uuid.Nil (глобальные admin запросы).
//
// Используется для admin routes, где tenant нужен для scoped операций
// (например, HR видит только пользователей своего tenant'а), но не для
// глобальных admin операций (tenant CRUD).
// Если m == nil, метрики не собираются.
func AdminTenantMiddleware(service *Service, redisClient *redis.Client, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := r.Header.Get("X-Tenant-ID")
			ctx := r.Context()

			if slug != "" {
				resolver := NewHostResolver(service, redisClient, m)
				tenant, err := resolver.ResolveBySlug(r.Context(), slug)
				if err == nil {
					ctx = context.WithValue(ctx, TenantIDKey, tenant.ID)
					ctx = context.WithValue(ctx, TenantSlugKey, tenant.Slug)
				}
				// Если ошибка — продолжаем с nil tenant; handler проверит
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
