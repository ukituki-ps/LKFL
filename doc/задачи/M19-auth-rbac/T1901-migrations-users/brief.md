# T1901 — Migrations: Users + Accounts + Roles

## Веха

M19-auth-rbac

## Тип

code

## Контекст

Таблицы пользователей, аккаунтов и ролей.
Исходник: `doc/архитектура/schema.md` — таблицы USERS, ACCOUNTS.

## Что сделать

### `migrations/20260526120000_users.sql`

```sql
-- +goose Up
-- +goose StatementBegin

-- Пользователи платформы
CREATE TABLE lkfl_platform.users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    email         VARCHAR(255) NOT NULL,
    first_name    VARCHAR(100),
    last_name     VARCHAR(100),
    phone         VARCHAR(20),
    status        VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deactivated', 'deleted')),
    keycloak_sub  VARCHAR(255) NOT NULL UNIQUE,           -- Keycloak subject ID (OIDC sub claim)
    metadata      JSONB DEFAULT '{}',                     -- HR data: greid, department, hire_date
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique: tenant + email
CREATE UNIQUE INDEX idx_users_tenant_email ON lkfl_platform.users (tenant_id, email);
-- Index: keycloak_sub (OIDC lookup)
CREATE INDEX idx_users_keycloak_sub ON lkfl_platform.users (keycloak_sub);
-- Index: tenant_id (isolation)
CREATE INDEX idx_users_tenant_id ON lkfl_platform.users (tenant_id);

-- Аккаунт пользователя (баланс, настройки)
CREATE TABLE lkfl_platform.accounts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL UNIQUE REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    total_balance BIGINT NOT NULL DEFAULT 0,              -- общий баланс (копейки/минимальная единица)
    settings      JSONB DEFAULT '{}',                     -- notification preferences, etc.
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Роли пользователя (RBAC)
CREATE TABLE lkfl_platform.user_roles (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    role          VARCHAR(50) NOT NULL CHECK (role IN ('employee', 'hr', 'catalog_manager', 'admin')),
    granted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    granted_by    UUID REFERENCES lkfl_platform.users(id), -- кто назначил (admin)
    expires_at    TIMESTAMPTZ                              -- опциональный срок действия роли
);

-- Unique: user + role
CREATE UNIQUE INDEX idx_user_roles_user_role ON lkfl_platform.user_roles (user_id, role);
-- Index: role (поиск по роли)
CREATE INDEX idx_user_roles_role ON lkfl_platform.user_roles (role);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS lkfl_platform.user_roles;
DROP TABLE IF EXISTS lkfl_platform.accounts;
DROP TABLE IF EXISTS lkfl_platform.users;

-- +goose StatementEnd
```

## Требования

- `keycloak_sub` — UNIQUE (OIDC subject, связывает Keycloak user с нашей БД)
- `tenant_id + email` — UNIQUE (email уникален в рамках tenant'а)
- `status` — CHECK constraint (active/deactivated/deleted)
- `metadata` — JSONB (HR данные: грейд, отдел, дата найма — для eligibility CEL)
- `accounts` — 1:1 с user (UNIQUE на user_id)
- `total_balance` — BIGINT (копейки, не float)
- `user_roles` — role CHECK constraint (employee/hr/catalog_manager/admin)
- ON DELETE CASCADE для accounts и user_roles

## Критерии приёмки

- [ ] `users` таблица (id, tenant_id, email, name, phone, status, keycloak_sub, metadata)
- [ ] `accounts` таблица (id, user_id, total_balance, settings)
- [ ] `user_roles` таблица (id, user_id, role, granted_at, granted_by, expires_at)
- [ ] UNIQUE на `keycloak_sub`
- [ ] UNIQUE на `tenant_id + email`
- [ ] CHECK на `status` и `role`
- [ ] Indexes для tenant isolation и OIDC lookup
- [ ] Migration apply + rollback OK
