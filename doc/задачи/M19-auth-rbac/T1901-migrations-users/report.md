# Отчёт T1901 — Migrations: Users + Accounts + Roles

## Выполнено

Созданы SQL-миграции для трёх таблиц RBAC-системы:

### `backend/migrations/20260526120000_users.sql`

| Таблица | Описание | Ключевые особенности |
|---------|----------|---------------------|
| `lkfl_platform.users` | Пользователи платформы | FK → tenants, UNIQUE (tenant_id, email), UNIQUE keycloak_sub, CHECK status |
| `lkfl_platform.accounts` | Аккаунт пользователя (баланс, настройки) | 1:1 с users (UNIQUE user_id), BIGINT balance |
| `lkfl_platform.user_roles` | Роли пользователя (RBAC) | CHECK role IN (employee, hr, catalog_manager, admin), UNIQUE (user_id, role) |

### `backend/migrations/20260526120000_users.sql.down`

Rollback-миграция: DROP TABLE в обратном порядке (user_roles → accounts → users) с IF EXISTS.

## Индексы

- `idx_users_tenant_email` — UNIQUE (tenant_id, email)
- `idx_users_keycloak_sub` — UNIQUE (keycloak_sub)
- `idx_users_tenant_id` — tenant isolation
- `idx_user_roles_user_role` — UNIQUE (user_id, role)
- `idx_user_roles_role` — поиск по роли

## Требования 152-ФЗ / Multi-tenant

- Tenant isolation через FK на `tenants.id` + индекс по `tenant_id`
- ON DELETE CASCADE на всех зависимых таблицах
- `keycloak_sub` — глобальный уникальный идентификатор OIDC

## Время

~15 минут

## Замечания

- Миграция зависит от существования таблицы `lkfl_platform.tenants` (T1801)
- SQL синтаксически валиден для PostgreSQL 17
