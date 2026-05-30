# T2203 — Integration тесты

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Integration тесты с testcontainers (PostgreSQL + Redis).
Полные пути через HTTP handlers.

## Что сделать

> **🔴 Критическое требование:** 100% покрытие всех интеграционных путей F1. Каждый endpoint, каждая комбинация middleware, каждый user journey — покрыт integration тестом. Тестов должно быть супермного.

### `internal/testutil/testcontainers.go`

```go
package testutil

// SetupTestDB — запускает PostgreSQL container
func SetupTestDB() (*pgxpool.Pool, func(), error)

// SetupTestRedis — запускает Redis container
func SetupTestRedis() (*redis.Client, func(), error)

// SetupTestServer — запускает тестовый HTTP server с real dependencies
func SetupTestServer(db *pgxpool.Pool, redis *redis.Client) (*httptest.Server, func(), error)
```

### Integration тесты — tenant/

- Tenant CRUD полный цикл: Create → Get → List → Update → Delete
- Multi-tenant isolation: tenant A данные ≠ tenant B данные
- Tenant isolation bypass для admin routes
- Brand config CRUD: create, update, get, delete
- Concurrent tenant operations (race condition test)
- Tenant suspension: create → suspend → verify blocked → activate

### Integration тесты — auth/

- Login flow: redirect → state generation → Keycloak → callback → token → session
- Logout flow: POST logout → session delete → redirect
- Auth/me: valid token → user profile, expired token → 401, missing token → 401
- RBAC integration: admin route with employee role → 403, admin route with admin role → 200
- Concurrent auth: multiple login attempts, session collision

### Integration тесты — user/

- User CRUD: create → get → update → list → deactivate
- User profile update: own profile only, admin update any
- User list with filters: search, pagination, role filter
- User deactivation cascade: deactivate → verify engagements blocked

### Integration тесты — catalog/

- Public catalog flow: seed data → list → filter → search → pagination → get by id
- Admin CRUD: create category → create type → create offer → update status → delete
- Cache integration: admin create → cache invalidation → list updated
- Multi-tenant catalog: tenant A catalog ≠ tenant B catalog
- Status transitions: draft → active → promo → completed
- Admin delete protection: delete with active engagements → error

### Integration тесты — frontend API layer

- API client integration: auth header injection, 401 redirect, 403 handling, 204 response, 5xx retry
- OpenAPI types: generated types match backend response
- React Query: cache behavior, staleTime, refetch on mutation

## Требования

- testcontainers-go для PG + Redis
- `go test -tags=integration`
- Real DB + Redis (не mock)
- HTTP handler integration (httptest.Server)
- **100% покрытие всех интеграционных путей F1**
- **Каждый endpoint покрыт минимум 3 тестами:** happy path, error path, edge case
- **Все user journeys покрыты** — от login до catalog browsing
- **Тестов должно быть супермного** — минимум 20 integration тестов

## Критерии приёмки

- [ ] testutil package
- [ ] Tenant CRUD integration test (5+ тестов)
- [ ] Auth flow integration test (5+ тестов)
- [ ] User CRUD integration test (5+ тестов)
- [ ] Catalog flow integration test (5+ тестов)
- [ ] Multi-tenant isolation test (3+ тестов)
- [ ] Cache invalidation test (2+ тестов)
- [ ] **100% покрытие всех интеграционных путей F1**
- [ ] **Каждый endpoint покрыт 3+ тестами**
- [ ] **Все user journeys покрыты**
- [ ] `go test -tags=integration` — 0 failures
