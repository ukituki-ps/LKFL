# T3501-T3508 — Контент + Collections

## Веха

M35-content-collections

## T3501 — Migrations: Content
```sql
CREATE TABLE lkfl_platform.content_items (
    id UUID PK, tenant_id UUID FK, type VARCHAR(20) CHECK (type IN ('faq','banner','description')),
    title VARCHAR(255), body TEXT, metadata JSONB,
    status VARCHAR(20) CHECK (status IN ('draft','published','archived')),
    sort_order INT, created_at, updated_at
);
```

## T3502 — internal/content/ (Engine)
- FAQ CRUD, banners CRUD, card description override
- Redis cache: `content:` TTL 10min

## T3503 — API: Content (public)
```
GET /api/v1/faq         — FAQ список
GET /api/v1/banners     — баннеры (tenant-aware)
```

## T3504 — Admin API: Content
```
CRUD FAQ, banners, card descriptions
```

## T3505 — Redis cache
- Key prefix: `content:`
- TTL: 10min
- Invalidate при admin изменении

## T3506 — Migrations: Collections
```sql
CREATE TABLE lkfl_platform.engagement_collections (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255), description TEXT,
    start_date DATE, end_date DATE, banner_image_url TEXT,
    status VARCHAR(20), created_at, updated_at
);
CREATE TABLE lkfl_platform.collection_items (
    id UUID PK, collection_id UUID FK, engagement_type_id UUID FK,
    sort_order INT, created_at
);
```

## T3507 — Collections engine
- Collection CRUD
- Activation all → batch debit
- GMV tracking

## T3508 — Admin API: Collections
```
CRUD collections
GET /admin/collections/:id/metrics — конверсия, GMV
```

## Критерии приёмки
- [ ] Все 8 задач
- [ ] Content + cache
- [ ] Collections + batch debit
