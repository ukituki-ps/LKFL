-- Ref: T1711 — Rename keycloak_user_id → keycloak_sub + cleanup seed data
-- Description: Align staging DB with code: column name fix + drop stale seed users.

-- 1. Drop old index (will recreate with new name)
DROP INDEX IF EXISTS lkfl_platform.idx_users_tenant_keycloak;

-- 2. Rename column
ALTER TABLE lkfl_platform.users RENAME COLUMN keycloak_user_id TO keycloak_sub;

-- 3. Make keycloak_sub globally unique (was tenant+keycloak before)
CREATE UNIQUE INDEX idx_users_keycloak_sub ON lkfl_platform.users (keycloak_sub);
CREATE INDEX idx_users_tenant_id ON lkfl_platform.users (tenant_id);

-- 4. Create unique on tenant+email (was missing)
CREATE UNIQUE INDEX idx_users_tenant_email ON lkfl_platform.users (tenant_id, email);

-- 5. Drop stale seed users (kc-*)
DELETE FROM lkfl_platform.user_roles WHERE user_id IN (
    SELECT id FROM lkfl_platform.users WHERE keycloak_sub LIKE 'kc-%'
);
DELETE FROM lkfl_platform.accounts WHERE user_id IN (
    SELECT id FROM lkfl_platform.users WHERE keycloak_sub LIKE 'kc-%'
);
DELETE FROM lkfl_platform.users WHERE keycloak_sub LIKE 'kc-%';

-- 6. Also drop the admin-user created by old realm config
DELETE FROM lkfl_platform.user_roles WHERE user_id IN (
    SELECT id FROM lkfl_platform.users WHERE keycloak_sub = 'admin-user'
);
DELETE FROM lkfl_platform.accounts WHERE user_id IN (
    SELECT id FROM lkfl_platform.users WHERE keycloak_sub = 'admin-user'
);
DELETE FROM lkfl_platform.users WHERE keycloak_sub = 'admin-user';
