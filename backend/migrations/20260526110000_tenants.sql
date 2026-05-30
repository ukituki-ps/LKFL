-- Ref: T2203 — Migrations: Tenants + Brand Config
-- Description: Таблицы tenants и бренд-конфигурации для multi-tenancy.

-- Tenants (мультитенантность)
CREATE TABLE lkfl_platform.tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       VARCHAR(64) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended')),
    settings   JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique: slug (глобальная уникальность)
CREATE UNIQUE INDEX idx_tenants_slug ON lkfl_platform.tenants (slug);

-- White-label брендирование tenant'а
CREATE TABLE lkfl_platform.tenant_brand_config (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id        UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    primary_color    VARCHAR(20) NOT NULL DEFAULT '#000000',
    secondary_color  VARCHAR(20) NOT NULL DEFAULT '#FFFFFF',
    logo_url         TEXT,
    favicon_url      TEXT,
    brand_name       TEXT,
    css_variables    JSONB DEFAULT '{}',
    meta_title       TEXT,
    meta_description TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique: один brand config на tenant
CREATE UNIQUE INDEX idx_brand_config_tenant ON lkfl_platform.tenant_brand_config (tenant_id);
