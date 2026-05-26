# T2002 — internal/engagement/catalog/ (Repository + Service)

## Веха

M20-catalog

## Тип

code

## Контекст

`internal/engagement/catalog/` — repository и service для каталога энгейджментов.
Исходник: `doc/архитектура/engagement.md` (строка 8 — Концепция).

## Что сделать

### Структура

```
internal/engagement/catalog/
├── model.go       # EngagementType, EngagementCategory, EngagementOffer structs
├── repository.go  # Repository interface + pgx implementation
├── service.go     # Service (business logic)
```

### `model.go`

```go
package catalog

type EngagementType struct {
    ID            uuid.UUID `json:"id"`
    TenantID      uuid.UUID `json:"tenant_id"`
    CategoryID    uuid.UUID `json:"category_id"`
    Slug          string    `json:"slug"`
    Name          string    `json:"name"`
    Description   string    `json:"description,omitempty"`
    Type          string    `json:"type"` // benefit, activity
    Status        string    `json:"status"` // draft, active, promo, hidden, completed
    CostCents     *int64    `json:"cost_cents,omitempty"`
    ProviderName  string    `json:"provider_name,omitempty"`
    ImageURL      *string   `json:"image_url,omitempty"`
    Metadata      JSONB     `json:"metadata,omitempty"`
    Category      *EngagementCategory `json:"category,omitempty"`
    Offers        []EngagementOffer   `json:"offers,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

type EngagementCategory struct {
    ID        uuid.UUID `json:"id"`
    TenantID  uuid.UUID `json:"tenant_id"`
    Slug      string    `json:"slug"`
    Name      string    `json:"name"`
    Icon      string    `json:"icon,omitempty"`
    SortOrder int       `json:"sort_order"`
}

type EngagementOffer struct {
    ID              uuid.UUID `json:"id"`
    TenantID        uuid.UUID `json:"tenant_id"`
    EngagementTypeID uuid.UUID `json:"engagement_type_id"`
    Name            string    `json:"name"`
    Description     string    `json:"description,omitempty"`
    CostCents       int64     `json:"cost_cents"`
    BillingRuleID   *uuid.UUID `json:"billing_rule_id,omitempty"`
    Metadata        JSONB     `json:"metadata,omitempty"`
    SortOrder       int       `json:"sort_order"`
}
```

### `repository.go`

```go
type CatalogFilter struct {
    TenantID  uuid.UUID
    Type      string    // benefit, activity (optional)
    Status    string    // active, promo (optional)
    Category  string    // category slug (optional)
    Search    string    // ILIKE name/description (optional)
    Page      int
    PerPage   int
}

type Repository interface {
    // Public
    ListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error)
    GetTypeByID(ctx context.Context, id uuid.UUID) (EngagementType, error)
    GetCategories(ctx context.Context, tenantID uuid.UUID) ([]EngagementCategory, error)

    // Admin
    AdminCreateType(ctx context.Context, t EngagementType) (EngagementType, error)
    AdminUpdateType(ctx context.Context, t EngagementType) (EngagementType, error)
    AdminDeleteType(ctx context.Context, id uuid.UUID) error
    AdminCreateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error)
    AdminUpdateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error)

    // Offers
    GetOffersByType(ctx context.Context, typeID uuid.UUID) ([]EngagementOffer, error)
    AdminCreateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error)
    AdminUpdateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error)
}
```

### Queries (пример ListTypes)

```sql
SELECT
    et.id, et.tenant_id, et.category_id, et.slug, et.name, et.description,
    et.type, et.status, et.cost_cents, et.provider_name, et.image_url,
    et.metadata, et.created_at, et.updated_at,
    ec.id as cat_id, ec.slug as cat_slug, ec.name as cat_name, ec.icon as cat_icon
FROM lkfl_platform.engagement_types et
LEFT JOIN lkfl_platform.engagement_categories ec ON et.category_id = ec.id
WHERE et.tenant_id = $1
  AND et.status IN ('active', 'promo')
  -- Optional filters applied dynamically
ORDER BY
    CASE et.status WHEN 'promo' THEN 0 ELSE 1 END,
    ec.sort_order,
    et.name
LIMIT $2 OFFSET $3
```

## Требования

- Repository interface (не concrete type в service)
- Tenant isolation через `tenant.WithTenantID()` или явный tenant_id
- N-фильтры: type, status, category, search (ILIKE)
- Pagination: page/per_page (default 20, max 100)
- Promo items first в ORDER BY
- Category join (LEFT JOIN для display)
- Offers loading (separate query или JSON aggregation)

## Критерии приёмки

- [ ] `model.go` — structs
- [ ] `repository.go` — interface + pgx impl
- [ ] `service.go` — business logic
- [ ] ListTypes с N-фильтрами
- [ ] GetTypeByID с offers + category
- [ ] GetCategories
- [ ] Admin CRUD
- [ ] Pagination
- [ ] Promo first ordering
- [ ] Unit tests
