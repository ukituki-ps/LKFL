-- Engagement categories
CREATE TABLE IF NOT EXISTS lkfl_platform.engagement_categories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    slug            VARCHAR(64) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    icon            VARCHAR(50),
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, slug)
);

-- Engagement types
CREATE TABLE IF NOT EXISTS lkfl_platform.engagement_types (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    category_id     UUID NOT NULL REFERENCES lkfl_platform.engagement_categories(id) ON DELETE CASCADE,
    slug            VARCHAR(64) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    type            VARCHAR(30) NOT NULL DEFAULT 'benefit',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    cost_cents      BIGINT,
    provider_name   VARCHAR(255),
    image_url       TEXT,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, slug)
);

-- Engagement offers
CREATE TABLE IF NOT EXISTS lkfl_platform.engagement_offers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    engagement_type_id UUID NOT NULL REFERENCES lkfl_platform.engagement_types(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    cost_cents      BIGINT,
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Accounts
CREATE TABLE IF NOT EXISTS lkfl_platform.accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    total_balance   BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id)
);

-- User roles
CREATE TABLE IF NOT EXISTS lkfl_platform.user_roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    role            VARCHAR(50) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, role)
);
