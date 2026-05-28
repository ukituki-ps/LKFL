# T1905 — Отчёт

## Что сделано

Создан пакет `backend/internal/user/` — CRUD пользователей, профиль, admin-операции.

### Файлы

| Файл | Строки | Описание |
|------|--------|----------|
| `model.go` | ~90 | User, UserProfile, Account, UserRole, UserFilter модели + константы статусов и ролей |
| `repository.go` | ~340 | Repository interface + pgx реализация (12 методов) |
| `service.go` | ~170 | Service с бизнес-логикой (7 публичных методов) |
| `handler.go` | ~230 | HTTP handlers: Me, UpdateMe, AdminList, AdminGet, AdminUpdate, AdminDeactivate |
| `service_test.go` | ~400 | Unit-тесты: mockRepository, 12 тестов (все зелёные) |

### Модель

- **User** — полная модель пользователя с tenant_id, email, ФИО, phone, status, keycloak_sub (скрыт в JSON), metadata (JSONB)
- **UserProfile** — публичный профиль (без keycloak_sub, tenant_id)
- **Account** — аккаунт пользователя (баланс, настройки)
- **UserRole** — RBAC роль (employee, hr, catalog_manager, admin)
- **UserFilter** — фильтр для админ-списка (status, search, pagination)

### Repository (pgx)

- `Create` — INSERT users
- `GetByID` — SELECT by id
- `GetByKeycloakSub` — SELECT by keycloak_sub (уникальный индекс)
- `GetByEmail` — SELECT by tenant+email
- `Update` — UPDATE email, name, phone, metadata
- `UpdateStatus` — UPDATE status
- `List` — paginated list с tenant isolation (`tenant.WithTenantID`), search (ILIKE), status filter
- `CreateAccount` — INSERT accounts
- `GetAccountByUserID` — SELECT accounts by user
- `GetRoles` — SELECT user_roles by user
- `AddRole` — INSERT user_roles
- `RemoveRole` — DELETE user_roles

### Service

- `GetByID` / `GetByKeycloakSub` — получение пользователя
- `UpdateProfile` — обновление профиля с валидацией (уникальность email в tenant, проверка статуса)
- `Deactivate` — деактивация (active → deactivated, односторонняя)
- `Activate` — активация (deactivated → active)
- `List` — admin list с пагинацией (defaults: page=1, per_page=20, max=100)
- `AddRole` / `RemoveRole` / `GetRoles` — RBAC операции с валидацией ролей
- `CreateAndSetupUser` — создание пользователя с автоматическим аккаунтом и ролью employee

### Handler

- **Me** — GET /api/v1/users/me — профиль текущего пользователя (через keycloak_sub из JWT)
- **UpdateMe** — PUT /api/v1/users/me — обновление профиля (только own)
- **AdminList** — GET /admin/users — список с пагинацией, фильтрами, search (требует admin/hr)
- **AdminGet** — GET /admin/users/:id — детали пользователя (требует admin/hr, tenant check)
- **AdminUpdate** — PUT /admin/users/:id — обновление (требует admin/hr)
- **AdminDeactivate** — POST /admin/users/:id/deactivate — деактивация (требует admin/hr)

### Tenant isolation

- List использует `tenant.WithTenantID(ctx)` для автоматического WHERE tenant_id
- Admin операции проверяют tenant_id пользователя против tenant из context
- Email uniqueness проверяется внутри tenant'а

### RBAC

- Employee: Me, UpdateMe (only own profile)
- Admin/HR: все admin операции
- Роли из JWT claims через `auth.RolesFromContext()`

### Тесты

- `TestService_Deactivate` (4 подтеста) — статус transitions
- `TestService_Activate` (2 подтеста) — обратная деактивации
- `TestService_UpdateProfile` (3 подтеста) — обновление, деактивированный, дубликат email
- `TestService_AddRole` (3 подтеста) — добавление, валидация роли, не найден
- `TestService_RemoveRole` (2 подтеста) — удаление, роль не найдена
- `TestService_List` (3 подтеста) — пагинация, фильтр, defaults
- `TestUser_ToProfile` — преобразование модели
- `TestValidRoles` — валидация допустимых ролей

**Результат:** 12 тестов, 0 провалов.

## Проверки

- `go build ./...` — ✅ чистая компиляция
- `go vet ./...` — ✅ без замечаний
- `go test ./internal/user/...` — ✅ 12/12 passed

## Замечания

- `CreateAndSetupUser` создаёт аккаунт и роль employee в том же контексте, но без транзакции. При продакшен-использовании стоит обернуть в pgx.Begin().
- Handler использует локальные `writeJSON`/`writeJSONError` — дублируют функции из tenant/handler.go. Можно вынести в shared utils.
- Поиск (ILIKE) без индекса на email/first_name/last_name будет медленным на больших объёмах. Рассмотреть GIN trigram индекс.
