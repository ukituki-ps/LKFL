# T1804 — Tenant Isolation (query builder)

## Веха

M18-multi-tenancy

## Тип

code

## Контекст

Tenant isolation — автоматическое добавление `WHERE tenant_id = :tid` ко всем бизнес-query.
Реализуется через query builder pattern в repository layer.

## Что сделать

### `internal/tenant/isolation.go`

```go
package tenant

import "context"

// WithTenantID — добавляет tenant_id WHERE clause к query
func WithTenantID(ctx context.Context, query string) (string, []interface{}) {
    tid := TenantIDFromContext(ctx)
    if tid == uuid.Nil {
        return query, nil
    }

    // Добавляем WHERE или AND в зависимости от наличия WHERE в query
    if strings.Contains(strings.ToUpper(query), " WHERE ") {
        return query + " AND tenant_id = $1", []interface{}{tid}
    }
    return query + " WHERE tenant_id = $1", []interface{}{tid}
}

// TenantContext — context с tenant_id (для тестов)
func TenantContext(ctx context.Context, tid uuid.UUID) context.Context {
    return context.WithValue(ctx, TenantIDKey, tid)
}
```

### Usage в business repository

```go
// internal/engagement/catalog/repository.go
func (r *pgRepository) List(ctx context.Context, filter CatalogFilter) ([]Engagement, int64, error) {
    baseQuery := `
        SELECT e.id, e.tenant_id, e.name, e.description, e.type, e.status
        FROM lkfl_platform.engagement_types e
    `

    query, args := tenant.WithTenantID(ctx, baseQuery)

    // Добавляем фильтры
    query += " AND e.status = $1"
    args = append(args, filter.Status)

    // Pagination
    query += " LIMIT $2 OFFSET $3"
    args = append(args, filter.PerPage, (filter.Page-1)*filter.PerPage)

    // Выполняем
    rows, err := r.pool.Query(ctx, query, args...)
    // ...
}
```

### Prepared statements pattern

Для часто используемых query — prepared statements с tenant_id:

```go
type PreparedQueries struct {
    ListEngagements *pgx.PreparedStatement
    GetEngagement   *pgx.PreparedStatement
    // ...
}
```

## Требования

- `WithTenantID()` — функция для добавления WHERE clause
- Работает с существующим WHERE (AND vs WHERE)
- Если tenant_id = uuid.Nil — не добавляет (admin queries без tenant)
- Все бизнес-repository используют `WithTenantID()`
- Admin queries — без tenant isolation (отдельные repository)
- Не использовать в `tenants` таблице (самой таблице не нужен tenant_id filter)

## Критерии приёмки

- [ ] `isolation.go` — WithTenantID() функция
- [ ] Работает с существующим WHERE
- [ ] uuid.Nil → без фильтра
- [ ] Документация usage pattern в README пакета
- [ ] Unit tests: с WHERE, без WHERE, nil tenant_id
