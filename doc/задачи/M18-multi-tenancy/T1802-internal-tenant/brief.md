# T1802 — internal/tenant/ (CRUD + Brand Config)

## Веха

M18-multi-tenancy

## Тип

code

## Контекст

`internal/tenant/` — системный пакет для управления tenants и brand config.
Исходник: `doc/архитектура/пакеты-platform.md` (строка ~1 — tenant/).

## Что сделать

### Структура пакета

```
internal/tenant/
├── model.go        # Tenant, BrandConfig structs
├── repository.go   # DB operations (pgx)
├── service.go      # Business logic
└── handler.go      # HTTP handlers (admin only)
```

### `model.go`

```go
package tenant

import "time"

type Tenant struct {
    ID        uuid.UUID `json:"id"`
    Slug      string    `json:"slug"`
    Name      string    `json:"name"`
    Status    string    `json:"status"` // active, suspended, terminated
    Settings  JSONB     `json:"settings"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type BrandConfig struct {
    ID             uuid.UUID `json:"id"`
    TenantID       uuid.UUID `json:"tenant_id"`
    PrimaryColor   string    `json:"primary_color"`
    SecondaryColor string    `json:"secondary_color"`
    LogoURL        *string   `json:"logo_url,omitempty"`
    FaviconURL     *string   `json:"favicon_url,omitempty"`
    BrandName      *string   `json:"brand_name,omitempty"`
    CSSVariables   JSONB     `json:"css_variables"`
    MetaTitle      *string   `json:"meta_title,omitempty"`
    MetaDescription *string  `json:"meta_description,omitempty"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
```

### `repository.go`

```go
type Repository interface {
    Create(ctx context.Context, t Tenant) (Tenant, error)
    GetByID(ctx context.Context, id uuid.UUID) (Tenant, error)
    GetBySlug(ctx context.Context, slug string) (Tenant, error)
    List(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error)
    Update(ctx context.Context, t Tenant) (Tenant, error)
    Delete(ctx context.Context, id uuid.UUID) error

    GetBrandConfig(ctx context.Context, tenantID uuid.UUID) (BrandConfig, error)
    UpsertBrandConfig(ctx context.Context, bc BrandConfig) (BrandConfig, error)
}

type TenantFilter struct {
    Status  string
    Page    int
    PerPage int
}

type pgRepository struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
    return &pgRepository{pool: pool}
}
```

### `service.go`

```go
type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateTenant(ctx context.Context, t Tenant) (Tenant, error) {
    // Validate slug format: lowercase, alphanumeric, hyphens
    if !isValidSlug(t.Slug) {
        return Tenant{}, ErrInvalidSlug
    }

    // Check slug uniqueness
    if err := s.repo.GetBySlug(ctx, t.Slug); err == nil {
        return Tenant{}, ErrSlugExists
    }

    return s.repo.Create(ctx, t)
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (Tenant, error) {
    t, err := s.repo.GetBySlug(ctx, slug)
    if err != nil {
        return Tenant{}, err
    }
    if t.Status != "active" {
        return Tenant{}, ErrTenantSuspended
    }
    return t, nil
}
```

### `handler.go` (Admin only)

```go
type Handler struct {
    service *Service
}

// POST /admin/tenants — создать tenant (admin only)
// GET /admin/tenants — список tenants (admin only)
// GET /admin/tenants/:id — детали tenant (admin only)
// PUT /admin/tenants/:id — обновить tenant (admin only)
// DELETE /admin/tenants/:id — удалить tenant (admin only, только при 0 users)
// GET /admin/tenants/:id/brand — brand config (admin only)
// PUT /admin/tenants/:id/brand — upsert brand config (admin only)
```

## Требования

- Repository interface (не concrete type в service/handler — для тестов)
- Slug validation: `^[a-z0-9]+(-[a-z0-9]+)*$`
- Slug uniqueness check
- Status check (suspended tenant не проходит GetBySlug)
- Pagination для List (page/per_page, default 20)
- JSONB для settings и css_variables (pgx native support)
- Errors: custom error types (ErrNotFound, ErrSlugExists, ErrTenantSuspended)

## Критерии приёмки

- [ ] `model.go` — Tenant + BrandConfig structs
- [ ] `repository.go` — Repository interface + pgx implementation
- [ ] `service.go` — business logic (slug validation, status check)
- [ ] `handler.go` — admin HTTP handlers (CRUD + brand config)
- [ ] Slug validation regex
- [ ] Slug uniqueness
- [ ] Pagination (page/per_page)
- [ ] Custom error types
- [ ] Unit tests: Create, GetBySlug (active/suspended), List (pagination)
