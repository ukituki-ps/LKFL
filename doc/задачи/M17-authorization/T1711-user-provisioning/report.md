# T1711 — Report

## Дата

2026-05-31

## Что сделано

### 1. `user/repository.go` — EnsureAccount

- Добавлен метод `EnsureAccount(ctx, userID) error` в интерфейс `Repository`
- Реализация: `INSERT ... ON CONFLICT (user_id) DO NOTHING` — idempotent
- **Поправка:** исправлены все SQL-запросы `keycloak_user_id` → `keycloak_sub` (DB column mismatch)

### 2. `auth/service.go` — EnsureAccount в CreateOrUpdateUser

- В create-ветке после `s.userRepo.Create()` вызывается `s.userRepo.EnsureAccount()`
- Non-blocking: ошибка логируется через `slog.Warn`, не прерывает provisioning

### 3. `auth/handler.go` — provisioning в Me

- `Me` больше не делает `GetUserByKeycloakSub` → 404
- Вместо этого: извлекает `Claims` + `Roles` из context и вызывает `CreateOrUpdateUser`
- Создаёт пользователя при первом входе, обновляет данные при повторном
- Аккаунт гарантирован через `EnsureAccount`

### 4. `cmd/server/main.go` — удаление seed-пользователей

- Удалён `seedUser` struct
- Удалена `seedUsersDB()` — 8 фейковых пользователей с `kc-*` UID
- Удалены `seedBalanceForUser()`, `seedContains()`
- В `runSeed()` оставлен комментарий T1711: пользователи создаются через Keycloak provisioning

### 5. `infra/keycloak/realm-lkfl-sdek.json` — realm config

- 8 seed-пользователей с email `@sdek.local`, правильными ФИО
- Фиксированные UUID: `aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa` → `...aag1`
- Realm roles: admin, hr, catalog_manager, employee
- Пароль: `dev-password` (temporary=false)
- Protocol mapper: «realm roles (ID token)» → `realm_access.roles` в ID token + access token

### 6. Очистка БД

- Удалены 8 seed-пользователей с `keycloak_sub LIKE 'kc-%'`
- Каскадно удалены: 8 user_roles, 8 accounts

### 7. Тесты

- Добавлен `EnsureAccount` в `mockUserRepository` (auth/service_test.go)
- Добавлен `EnsureAccount` в `mockRepository` (user/service_test.go)
- Все существующие тесты прошли без изменений

## Изменённые файлы

| Файл | Действие |
|------|---------|
| `backend/internal/user/repository.go` | +EnsureAccount, keycloak_user_id → keycloak_sub |
| `backend/internal/auth/service.go` | +EnsureAccount в CreateOrUpdateUser |
| `backend/internal/auth/handler.go` | Me → provisioning |
| `backend/cmd/server/main.go` | -seedUsersDB, -seedUser, -seedBalanceForUser, -seedContains |
| `backend/internal/auth/service_test.go` | +EnsureAccount в mock |
| `backend/internal/user/service_test.go` | +EnsureAccount в mock |
| `infra/keycloak/realm-lkfl-sdek.json` | 8 пользователей + protocol mapper |
| `doc/задачи/.../T1711/plan.yaml` | progress update |

## Статус

✅ Все критерии приёмки выполнены:
- [x] EnsureAccount — idempotent
- [x] Me provisioning — create + update + roles
- [x] Seed удалён из main.go
- [x] Realm config: 8 пользователей + mapper
- [x] БД очищена от stale-записей
- [x] Тесты зелёные
- [x] go build ./... OK
