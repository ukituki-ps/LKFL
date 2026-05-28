# T2203 — Integration тесты: Отчёт

## Что сделано

### 1. Миграция tenants и tenant_brand_config
- Создан `backend/migrations/20260526110000_tenants.sql` — таблицы `tenants` и `tenant_brand_config`
- Down migration: `backend/migrations/20260526110000_tenants.sql.down`

### 2. Пакет testutil (`backend/internal/testutil/testcontainers.go`)
- `SetupTestDB()` — запускает PostgreSQL 17 testcontainer, применяет все миграции
- `SetupTestRedis()` — запускает Redis 7 testcontainer
- `SetupTestServer()` — создаёт httptest.Server с реальными зависимостями (DB + Redis)
- `SetupAllWithServer()` — запускает всё сразу (DB + Redis + HTTP server)
- `TestMiddleware()` — mock JWT middleware для тестов (формат: `test:{subject}:{roles}`)
- `TestToken()`, `TestTokenAdmin()`, `TestTokenEmployee()` и др. — генерация тестовых токенов
- `TestServer` — обёртка над httptest.Server с helper-методами:
  - `GetWithToken()`, `PostWithToken()`, `PutWithToken()`, `PatchWithToken()`, `DeleteWithToken()`
  - `GetWithTokenAndTenant()`, `PostWithTokenAndTenant()`, `PutWithTokenAndTenant()`, `PatchWithTokenAndTenant()`, `DeleteWithTokenAndTenant()` — с X-Tenant-ID header
  - `DoJSON()`, `DoJSONWithTenant()` — JSON запросы
- DB seed helpers: `CreateTenant()`, `CreateUser()`, `CreateCategory()`, `CreateEngagementType()`, `AddUserRole()`
- `ReadBody()`, `ReadJSONBody()` — утилиты для чтения ответов

### 3. Integration тесты — tenant/ (`backend/internal/tenant/integration_test.go`)
13 тестов, все проходят:
- `TestTenantCRUD_FullCycle` — Create → Get → List → Update → Delete
- `TestTenantCreate_DuplicateSlug` — дублирующийся slug → 409
- `TestTenantCreate_InvalidSlug` — невалидные slug → 400
- `TestTenantCreate_UppercaseSlugLowercased` — UPPER slug lowercased handler'ом
- `TestTenantSuspension` — create → suspend → verify → activate
- `TestTenant_RBAC_EmployeeForbidden` — employee → 403
- `TestTenantBrandConfig_CRUD` — brand config create/update/get
- `TestTenantList_Pagination` — пагинация
- `TestTenantGet_InvalidID` — невалидный UUID → 400
- `TestTenantCreate_EmptyBody` — пустое тело → 400
- `TestTenantList_NoAuth` — без авторизации → 401
- `TestTenantConcurrent_Create` — конкурентное создание
- `TestTenantCreate_Defaults` — значения по умолчанию

### 4. Integration тесты — auth/ (`backend/internal/auth/integration_test.go`)
12 тестов, все проходят:
- `TestAuthMe_ValidToken` — валидный токен → 200 с профилем
- `TestAuthMe_MissingToken` — без токена → 401
- `TestAuthMe_InvalidTokenFormat` — невалидный формат → 401
- `TestAuthMe_UserNotFound` — пользователь не найден → 404
- `TestRBAC_AdminRoute_EmployeeForbidden` — employee → 403
- `TestRBAC_AdminRoute_AdminAllowed` — admin → 200
- `TestRBAC_HRRoute_HRAllowed` — hr → 200 (с tenant header)
- `TestLogout` — logout → 302/200
- `TestEmployeeRoutes_AccessControl` — employee → /users/me
- `TestAuthMe_MultipleRoles` — несколько ролей
- `TestRBAC_CatalogManagerRole` — catalog_manager → catalog admin
- `TestRBAC_EmployeeCannotAccessAdmin` — employee → admin/users → 403

### 5. Integration тесты — user/ (`backend/internal/user/integration_test.go`)
13 тестов, все проходят:
- `TestUserCRUD_FullCycle` — Create → Get → Update → Deactivate
- `TestUserMe_Profile` — GET /api/v1/users/me
- `TestUserUpdateMe` — PUT /api/v1/users/me (с tenant header)
- `TestUserList_WithFilters` — GET /admin/users с фильтрами (с tenant header)
- `TestUserList_Pagination` — пагинация (с tenant header)
- `TestUserDeactivate_AlreadyDeactivated` — повторная деактивация → 409
- `TestUserDeactivate_NoAuth` — без авторизации → 401
- `TestUserMe_NoAuth` — без авторизации → 401
- `TestUserAdminGet_InvalidID` — невалидный UUID → 400
- `TestUserAdminGet_NotFound` — пользователь не найден → 404
- `TestUserUpdateDeactivated` — обновление деактивированного → 403
- `TestUserMe_TenantIsolation` — tenant isolation
- `TestUserUpdateMe_NoAuth` — без авторизации → 401

### 6. Integration тесты — catalog/ (`backend/internal/engagement/catalog/integration_test.go`)
13 тестов, все проходят:
- `TestCatalogPublic_List` — GET /api/v1/engagements
- `TestCatalogPublic_GetByID` — GET /api/v1/engagements/{id}
- `TestCatalogPublic_GetNotFound` — не найден → 404
- `TestCatalogPublic_GetInvalidID` — невалидный UUID → 400
- `TestCatalogAdmin_CRUD` — category → type → offer → status (все с tenant header)
- `TestCatalogMultiTenant_Isolation` — tenant A ≠ tenant B
- `TestCatalogStatusTransitions` — draft → active → promo → completed (с tenant header)
- `TestCatalogCache_Invalidation` — admin create via API → cache invalidation → list updated
- `TestCatalogCategories_Public` — GET /api/v1/engagements/categories
- `TestCatalogAdmin_RBAC` — employee → 403
- `TestCatalogAdmin_DuplicateSlug` — дублирующийся slug → 409 (с tenant header)
- `TestCatalogAdmin_DeleteType` — soft delete → status=hidden (с tenant header)
- `TestCatalogList_NoAuth` — без авторизации → 401

### 7. Integration тесты — multi-tenant isolation (`backend/internal/testutil/isolation_test.go`)
6 тестов, все проходят:
- `TestMultiTenantIsolation_DataSeparation` — данные tenant A ≠ данные tenant B
- `TestMultiTenantIsolation_UserIsolation` — пользователи изолированы
- `TestMultiTenantIsolation_AdminSeesAll` — admin видит все tenants
- `TestMultiTenantIsolation_BrandConfigIsolation` — brand config изолирован
- `TestMultiTenantIsolation_CascadeDelete` — удаление tenant → CASCADE
- `TestMultiTenantIsolation_TenantHeader` — X-Tenant-ID header

### 8. Исправления в коде (T2203 фиксы)

#### Фикс #1: "tenant not found" — employee routes возвращают 401
- `backend/internal/tenant/middleware.go` — HostResolver fallback на X-Tenant-ID header теперь включает `127` (для 127.0.0.1 из httptest)
- `backend/internal/testutil/testcontainers.go` — добавлены `PutWithTokenAndTenant()`, `PatchWithTokenAndTenant()`, `DeleteWithTokenAndTenant()`

#### Фикс #2: "cannot scan NULL into *string" — catalog repository
- `backend/internal/engagement/catalog/model.go`:
  - `EngagementType.Description` — `string` → `*string` (nullable TEXT в БД)
  - `EngagementType.ProviderName` — `string` → `*string` (nullable VARCHAR в БД)
  - `EngagementCategory.Icon` — `string` → `*string` (nullable VARCHAR в БД)
  - `EngagementOffer.Description` — `string` → `*string` (nullable TEXT в БД)
- `backend/internal/engagement/catalog/repository.go` — все scan методы обновлены для nullable полей (GetTypeByID, ListTypes, AdminListTypes, AdminCreateType, AdminUpdateType, GetTypeBySlug, GetCategories, AdminCreateCategory, AdminUpdateCategory, GetOffersByType, AdminCreateOffer, AdminUpdateOffer)
- `backend/internal/engagement/catalog/handler.go` — `ToResponse()` обрабатывает `*string` поля
- `backend/internal/engagement/catalog/admin_handler.go` — request structs обрабатывают `*string` поля

#### Фикс #3: "tenant context missing" — admin routes
- Все integration тесты admin routes обновлены для использования tenant-aware методов (`PostWithTokenAndTenant`, `GetWithTokenAndTenant`, `PatchWithTokenAndTenant`, `DeleteWithTokenAndTenant`)

#### Фикс #4: ReadBody-before-decode паттерн в тестах
- Исправлены тесты, которые вызывали `ReadBody()` (закрывающий body) перед `json.NewDecoder(resp.Body).Decode()`:
  - `TestUserMe_Profile`, `TestUserMe_TenantIsolation`, `TestUserList_WithFilters`, `TestUserList_Pagination`
  - `TestCatalogPublic_List`, `TestCatalogPublic_GetByID`, `TestCatalogAdmin_CRUD`, `TestCatalogCategories_Public`, `TestCatalogCache_Invalidation`

#### Фикс #5: Unit тесты
- Добавлен `strPtr()` helper в `handler_test.go` для создания `*string` значений
- Обновлены все тестовые структуры с `Icon`, `Description`, `ProviderName` полями

### 9. Зависимости
- Добавлены `github.com/testcontainers/testcontainers-go` и модули postgres/redis

## Результаты тестов

```
go build ./... — чистая компиляция ✅
go test ./... — все unit тесты проходят ✅
go test -tags=integration -timeout 300s -count=1 ./...
ok   lkfl/internal/auth              — все 12 тестов проходят
ok   lkfl/internal/engagement/catalog — все 13 тестов проходят
ok   lkfl/internal/tenant            — все 13 тестов проходят
ok   lkfl/internal/testutil          — все 6 тестов проходят
ok   lkfl/internal/user              — все 13 тестов проходят
```

**Итого: 57 integration тестов** (13 tenant + 12 auth + 13 user + 13 catalog + 6 isolation), **все проходят** ✅

## Замечания

1. Employee routes (`/api/v1/`) требуют X-Tenant-ID header для tenant resolution на localhost/127.0.0.1
2. Auth/me endpoint исключён из tenant middleware (не требует tenant)
3. Login redirect тесты не включены (требуют реального Keycloak)
4. Все admin routes используют tenant-aware методы в тестах

## Время

~6 часов (4 часа оригинальная реализация + 2 часа фиксы)
