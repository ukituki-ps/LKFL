-- Ref: T2001 — Migrations: Engagement (каталог льгот/активностей)
-- Description: Таблицы категорий, типов и офферов энгейджментов.

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

-- Unique: tenant + slug (slug уникален в рамках tenant'а)
CREATE UNIQUE INDEX idx_cat_tenant_slug ON lkfl_platform.engagement_categories (tenant_id, slug);
-- Index: tenant_id (isolation)
CREATE INDEX idx_cat_tenant ON lkfl_platform.engagement_categories (tenant_id);

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

-- Unique: tenant + slug (slug уникален в рамках tenant'а)
CREATE UNIQUE INDEX idx_eng_type_tenant_slug ON lkfl_platform.engagement_types (tenant_id, slug);
-- Index: category_id (поиск по категории)
CREATE INDEX idx_eng_type_category ON lkfl_platform.engagement_types (category_id);
-- Index: status (фильтрация по статусу)
CREATE INDEX idx_eng_type_status ON lkfl_platform.engagement_types (status);
-- Index: type (фильтрация benefit/activity)
CREATE INDEX idx_eng_type_type ON lkfl_platform.engagement_types (type);
-- Index: tenant_id (isolation)
CREATE INDEX idx_eng_type_tenant ON lkfl_platform.engagement_types (tenant_id);

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

-- Index: engagement_type_id (поиск офферов по типу)
CREATE INDEX idx_offer_eng_type ON lkfl_platform.engagement_offers (engagement_type_id);
-- Index: tenant_id (isolation)
CREATE INDEX idx_offer_tenant ON lkfl_platform.engagement_offers (tenant_id);
