-- Ref: T1901 — Migrations: Users
-- Description: Таблица users для хранения профилей сотрудников.

CREATE TABLE lkfl_platform.users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    keycloak_user_id    VARCHAR(255) NOT NULL,
    first_name          VARCHAR(100) NOT NULL,
    last_name           VARCHAR(100) NOT NULL,
    email               VARCHAR(255) NOT NULL,
    date_of_birth       DATE,
    phone               VARCHAR(50),
    grade               VARCHAR(50),
    years_of_service    DECIMAL(5,1),
    department          VARCHAR(200),
    position            VARCHAR(200),
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
    has_children        BOOLEAN NOT NULL DEFAULT FALSE,
    location            VARCHAR(20) NOT NULL DEFAULT 'office',
    deactivated_at      TIMESTAMPTZ,
    deactivated_reason  TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_users_status CHECK (status IN ('active', 'deactivated', 'pending')),
    CONSTRAINT chk_users_location CHECK (location IN ('office', 'remote'))
);

CREATE UNIQUE INDEX idx_users_tenant_keycloak ON lkfl_platform.users(tenant_id, keycloak_user_id);
CREATE INDEX idx_users_department ON lkfl_platform.users(department);
CREATE INDEX idx_users_grade ON lkfl_platform.users(grade);
