# T2001 — Migrations: Engagement

## Веха

M20-catalog

## Тип

code

## Контекст

Таблицы каталога: категории, типы энгейджментов, офферы.
Исходник: `doc/архитектура/schema.md` — таблицы ENGAGEMENT_CATEGORIES, ENGAGEMENT_TYPES, ENGAGEMENT_OFFERS.

## Что сделать

### `migrations/20260526130000_engagement.sql`

```sql
-- +goose Up
-- +goose StatementBegin

-- Категории энгейджментов (ДМС, фитнес, питание, образование, мерч)
CREATE TABLE lkfl_platform.engagement_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    slug       VARCHAR(64) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    icon       VARCHAR(50),                          -- icon name (AprilIcon)
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_cat_tenant_slug ON lkfl_platform.engagement_categories (tenant_id, slug);

-- Типы энгейджментов (конкретные льготы/активности)
CREATE TABLE lkfl_platform.engagement_types (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    category_id     UUID NOT NULL REFERENCES lkfl_platform.engagement_categories(id) ON DELETE CASCADE,
    slug            VARCHAR(64) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    type            VARCHAR(20) NOT NULL DEFAULT 'benefit' CHECK (type IN ('benefit', 'activity')),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'promo', 'hidden', 'completed')),
    cost_cents      BIGINT,                          -- стоимость в минимальных единицах (NULL для activity)
    provider_name   VARCHAR(255),                    -- название провайдера (для display)
    image_url       TEXT,
    metadata        JSONB DEFAULT '{}',              -- provider_id, external_id, etc.
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_eng_type_tenant_slug ON lkfl_platform.engagement_types (tenant_id, slug);
CREATE INDEX idx_eng_type_category ON lkfl_platform.engagement_types (category_id);
CREATE INDEX idx_eng_type_status ON lkfl_platform.engagement_types (status);
CREATE INDEX idx_eng_type_type ON lkfl_platform.engagement_types (type);

-- Офферы (тарифы/планы внутри типа)
CREATE TABLE lkfl_platform.engagement_offers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    engagement_type_id UUID NOT NULL REFERENCES lkfl_platform.engagement_types(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    cost_cents      BIGINT NOT NULL DEFAULT 0,
    billing_rule_id UUID,                            -- ссылка на billing rule (F2)
    metadata        JSONB DEFAULT '{}',
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_offer_eng_type ON lkfl_platform.engagement_offers (engagement_type_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS lkfl_platform.engagement_offers;
DROP TABLE IF EXISTS lkfl_platform.engagement_types;
DROP TABLE IF EXISTS lkfl_platform.engagement_categories;

-- +goose StatementEnd
```

## Требования

- `engagement_categories` — категории (ДМС, фитнес, питание...)
- `engagement_types` — типы (конкретные льготы/активности)
- `engagement_offers` — офферы (тарифы внутри типа)
- `type` — CHECK constraint (benefit/activity)
- `status` — CHECK constraint (draft/active/promo/hidden/completed)
- `cost_cents` — BIGINT (не float)
- Unique indexes на tenant + slug
- Indexes для фильтрации (category, status, type)

## Критерии приёмки

- [ ] `engagement_categories` таблица
- [ ] `engagement_types` таблица
- [ ] `engagement_offers` таблица
- [ ] CHECK constraints на type и status
- [ ] Unique indexes на tenant + slug
- [ ] Indexes для фильтрации
- [ ] Migration apply + rollback OK
