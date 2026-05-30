-- Ref: T1801 — Migrations: Tenants
-- Description: Таблицы tenants и tenant_brand_config для multi-tenancy.

-- Tenants — основа multi-tenancy
CREATE TABLE lkfl_platform.tenants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          VARCHAR(64) NOT NULL UNIQUE,
    name          VARCHAR(255) NOT NULL,
    status        VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'terminated')),
    settings      JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON lkfl_platform.tenants (slug);

-- Tenant brand config — white-label customization
CREATE TABLE lkfl_platform.tenant_brand_config (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    primary_color   VARCHAR(7) NOT NULL DEFAULT '#000000',
    secondary_color VARCHAR(7) DEFAULT '#FFFFFF',
    logo_url        TEXT,
    favicon_url     TEXT,
    brand_name      VARCHAR(255),
    css_variables   JSONB DEFAULT '{}',
    meta_title      VARCHAR(255),
    meta_description TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_brand_config_tenant ON lkfl_platform.tenant_brand_config (tenant_id);
