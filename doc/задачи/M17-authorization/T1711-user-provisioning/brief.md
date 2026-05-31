# T1711 — User Provisioning: синхронизация Keycloak ↔ БД

## Веха

M17-authorization

## Тип

code

## Проблема

`GET /api/v1/auth/me` делает `GetUserByKeycloakSub(ctx, keycloakSub)`, где `keycloakSub` — реальный UID из Keycloak JWT-токена (например, `fbfb9581-b18d-...`). В БД `keycloak_sub` содержит фейковые строки из seed-данных (`kc-admin-001`), которые никогда не совпадут с реальным UID. `/me` всегда возвращает 404.

**Следствия:**
- Пользователь может авторизоваться в Keycloak, получить токен — но не видит профиль на платформе
- Seed-данные в `main.go` (8 пользователей) — мёртвый вес: они есть в БД, но недоступны по OIDC flow
- Keycloak realm config (`realm-lkfl-sdek.json`) содержит только 1 пользователя, 4 `test.*` добавлены вручную — рассинхронизация с БД
- Роли не передаются в ID token → provisioning через `Me` не может назначить роли

## Что делать

### 1. Добавить provisioning в `Me` handler

**Файл:** `backend/internal/auth/handler.go`

Заменить `GetUserByKeycloakSub` на `CreateOrUpdateUser` в методе `Me`. Извлечь `Claims` и `Roles` из context (уже установлены middleware).

**Обоснование:** `Me` — первый endpoint при загрузке страницы. Если пользователь вошёл через Keycloak, но его записи в БД нет (первый вход, refresh без callback, рестарт сервера) — он должен быть создан автоматически. Callback уже provisioning'ит, но `Me` подстраховывает при:
- Refresh страницы (токен в cookie, без повторного callback)
- API-запросах с access token напрямую
- Восстановлении сессии после рестарта

### 2. Добавить `EnsureAccount` в user repository

**Файл:** `backend/internal/user/repository.go`

Метод `EnsureAccount(ctx, userID)` — idempotent `INSERT ... ON CONFLICT DO NOTHING`.

**Обоснование:** При provisioning'е пользователя через `Me` нужно гарантировать существование аккаунта (баланс = 0). Без этого пользователь не увидит Dashboard. `ON CONFLICT` защищает от повторных вызовов.

### 3. Вызывать `EnsureAccount` в `CreateOrUpdateUser`

**Файл:** `backend/internal/auth/service.go`

В `create`-ветке после `s.userRepo.Create(ctx, newUser)` вызвать `s.userRepo.EnsureAccount(ctx, created.ID)`.

**Обоснование:** Пользователь без аккаунта = сломанный Dashboard. Баланс начислит CEL/billing.

### 4. Убрать seed-пользователей из `main.go`

**Файл:** `backend/cmd/server/main.go`

Удалить `seedUsersDB` (строки 268–358), `seedBalanceForUser`, `seedContains`. Удалить вызов из seed-поточка.

**Обоснование:** Seed с фейковыми `keycloak_sub` не работает (никогда не найдётся по реальному UID). При Plan B пользователи создаются автоматически при первом входе.

### 5. Обновить realm-конфиг Keycloak

**Файл:** `infra/keycloak/realm-lkfl-sdek.json`

Заменить существующих пользователей на 8 seed-пользователей с фиксированными UUID, email `@sdek.local`, realm roles.

Добавить `protocolMappers` в клиент `lkfl-spa`:
- **Mapper:** «Realm Roles» → `realm_access.roles` в ID Token (типа `realm-roless`)
- Без этого `Me` не получит роли из context → provisioning без ролей

**Маппинг пользователей:**

| Username | Email | Roles |
|----------|-------|-------|
| `admin` | `admin@sdek.local` | `admin` |
| `hr` | `hr@sdek.local` | `hr` |
| `catalog` | `catalog@sdek.local` | `catalog_manager` |
| `ivanov` | `ivanov@sdek.local` | `employee` |
| `petrova` | `petrova@sdek.local` | `employee` |
| `sidorov` | `sidorov@sdek.local` | `employee` |
| `kozlova` | `kozlova@sdek.local` | `employee` |
| `novikov` | `novikov@sdek.local` | `employee` |

### 6. Очистить БД от stale-записей

SQL: удалить пользователей с `keycloak_user_id LIKE 'kc-%'` (каскадно: roles, accounts, users).

**Обоснование:** После удаления seed из `main.go` в БД останутся записи с фейковыми UID. Они не будут найдены, но будут путать при отладке.

### 7. Написать тесты

- `TestCreateOrUpdateUser_EnsureAccount` — новый пользователь → created + account exists
- `TestMe_ProvisionNewUser` — `Me` с валидным токеном, пользователя нет → created 200
- `TestMe_ExistingUser` — `Me` с валидным токеном, пользователь есть → 200, данные updated

## Требования

- `Me` должен создавать пользователя при первом входе (provisioning)
- `Me` должен обновлять данные (email, first_name, last_name) при повторных входах
- Роли из Keycloak должны попадать в БД (через `user_roles`)
- Аккаунт с балансом 0 должен создаваться автоматически
- Seed-данные в `main.go` больше не создают пользователей
- Realm config — единственный источник правды для dev-пользователей Keycloak
- Realm roles должны передаваться в ID token (mapper)

## Критерии приёмки

- [ ] `Me` создаёт пользователя при первом входе (200 OK, не 404)
- [ ] `Me` обновляет email/имя при повторном входе
- [ ] `Me` назначает роли из JWT-токена
- [ ] `EnsureAccount` — idempotent (повторные вызовы без ошибки)
- [ ] `CreateOrUpdateUser` создаёт аккаунт с балансом 0
- [ ] Seed-пользователи удалены из `main.go`
- [ ] Realm config: 8 пользователей с правильными email/ролями
- [ ] Protocol mapper: realm roles → ID token
- [ ] БД очищена от stale-записей (kc-*)
- [ ] Тесты: provisioning create + update + ensure account
- [ ] `go build ./...` без ошибок
- [ ] E2E: login через Keycloak → `/me` → 200 с правильными данными

## Зависимости

- **depends_on:** T1901 (migrations users), T1902 (auth backend), T1903 (auth handlers)
- **touches:** `handler.go`, `service.go`, `repository.go`, `main.go`, `realm-lkfl-sdek.json`
- **risk:** Изменение behavior `Me` — требует проверки callback flow
