# T1801 — Migrations: Tenants

## Веха

M18-multi-tenancy

## Тип

code

## Контекст

Первые бизнес-таблицы: `tenants` + `tenant_brand_config`.
Schema: `lkfl_platform` (из T1705).

Исходник: `doc/архитектура/schema.md` — таблица TENANTS (строка ~106).

## Что сделать

### `migrations/20260526110000_tenants.sql`

```sql
-- +goose Up
-- +goose StatementBegin

-- Tenants — основа multi-tenancy
CREATE TABLE lkfl_platform.tenants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          VARCHAR(64) NOT NULL UNIQUE,          -- уникальный идентификатор tenant'а (sdek, yandex, etc.)
    name          VARCHAR(255) NOT NULL,                -- отображаемое название
    status        VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'terminated')),
    settings      JSONB NOT NULL DEFAULT '{}',          -- tenant-specific settings
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index для быстрого поиска по slug (используется в tenant resolver middleware)
CREATE INDEX idx_tenants_slug ON lkfl_platform.tenants (slug);

-- Tenant brand config — white-label customization
CREATE TABLE lkfl_platform.tenant_brand_config (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    primary_color VARCHAR(7) NOT NULL DEFAULT '#000000', -- HEX
    secondary_color VARCHAR(7) DEFAULT '#FFFFFF',
    logo_url      TEXT,
    favicon_url   TEXT,
    brand_name    VARCHAR(255),                         -- override tenant.name для display
    css_variables JSONB DEFAULT '{}',                   -- custom CSS variables override
    meta_title    VARCHAR(255),
    meta_description TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One brand config per tenant
CREATE UNIQUE INDEX idx_brand_config_tenant ON lkfl_platform.tenant_brand_config (tenant_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS lkfl_platform.tenant_brand_config;
DROP TABLE IF EXISTS lkfl_platform.tenants;

-- +goose StatementEnd
```

## Требования

- PK: UUID (`gen_random_uuid()`)
- `slug` — UNIQUE index (tenant resolver использует slug → tenant_id)
- `status` — CHECK constraint (active/suspended/terminated)
- `settings` — JSONB (гибкие настройки без миграций)
- `tenant_brand_config` — 1:1 с tenant (UNIQUE index на tenant_id)
- ON DELETE CASCADE для brand_config
- Timestamps: `created_at`, `updated_at`

## Критерии приёмки

- [ ] `tenants` таблица создана
- [ ] `tenant_brand_config` таблица создана
- [ ] UNIQUE index на `slug`
- [ ] CHECK constraint на `status`
- [ ] UNIQUE index на `tenant_brand_config.tenant_id`
- [ ] ON DELETE CASCADE работает
- [ ] Migration apply + rollback OK
