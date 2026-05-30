# T1802 — Отчёт о выполнении

## Статус

✅ выполнено

## Что сделано

### `backend/internal/tenant/model.go`

- **Tenant** struct: id, slug, name, status, settings, created_at, updated_at
- **BrandConfig** struct: id, tenant_id, primary_color, secondary_color, logo_url, favicon_url, brand_name, css_variables, meta_title, meta_description, created_at, updated_at
- **JSONB** type: реализует sql.Scanner + driver.Valuer для PostgreSQL JSONB

### `backend/internal/tenant/repository.go`

- **Repository** interface: Create, GetByID, GetBySlug, List (pagination), Update, Delete, GetBrandConfig, UpsertBrandConfig
- **pgRepository** — pgx реализация с *pgxpool.Pool
- **TenantFilter** — filter по status, page, per_page
- **NewRepository(pool)** — конструктор

### `backend/internal/tenant/service.go`

- **Service** struct с Repository dependency
- **CreateTenant** — валидация slug (`^[a-z0-9]+(-[a-z0-9]+)*$`), проверка уникальности, дефолты
- **GetBySlug** — проверка status != suspended
- **ListTenants** — пагинация (default 20, max 100)
- **UpdateTenant** — валидация slug при изменении
- **DeleteTenant** — удаление
- **GetBrandConfig** / **UpsertBrandConfig** — brand config CRUD
- Custom errors: ErrNotFound, ErrSlugExists, ErrTenantSuspended, ErrInvalidSlug, ErrBrandNotFound

### `backend/internal/tenant/handler.go`

HTTP handlers (admin only, без JWT — будет в M19):

| Method | Path | Handler | Описание |
|--------|------|---------|----------|
| POST | /admin/tenants | Create | Создать tenant |
| GET | /admin/tenants | List | Список (pagination: page, per_page, status) |
| GET | /admin/tenants/:id | GetByID | Детали tenant |
| PUT | /admin/tenants/:id | Update | Обновить tenant |
| DELETE | /admin/tenants/:id | Delete | Удалить tenant |
| GET | /admin/tenants/:id/brand | GetBrandConfig | Brand config |
| PUT | /admin/tenants/:id/brand | UpsertBrandConfig | Upsert brand config |

### Unit тесты (`service_test.go`)

- CreateTenant: valid slug, duplicate slug, invalid slug (7 subtests), default values
- GetBySlug: active, suspended, not found
- GetByID
- ListTenants: default pagination, filter by status
- DeleteTenant
- UpsertBrandConfig

## Критерии приёмки

- [x] `model.go` — Tenant + BrandConfig structs
- [x] `repository.go` — Repository interface + pgx implementation
- [x] `service.go` — business logic (slug validation, status check)
- [x] `handler.go` — admin HTTP handlers (CRUD + brand config)
- [x] Slug validation regex
- [x] Slug uniqueness
- [x] Pagination (page/per_page)
- [x] Custom error types
- [x] Unit tests: Create, GetBySlug (active/suspended), List (pagination)

## Время

~40 мин
