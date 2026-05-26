# T2203 — Integration тесты

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Integration тесты с testcontainers (PostgreSQL + Redis).
Полные пути через HTTP handlers.

## Что сделать

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

### Integration тесты

```go
// integration/tenant_test.go
func TestTenantCRUD(t *testing.T) {
    db, cleanup := testutil.SetupTestDB()
    defer cleanup()

    // Seed tenant
    // Create → Get → List → Update → Delete
    // Multi-tenant isolation test
}

// integration/auth_test.go
func TestAuthFlow(t *testing.T) {
    server, cleanup := testutil.SetupTestServer(db, redis)
    defer cleanup()

    // Login redirect → callback → token → /users/me
}

// integration/catalog_test.go
func TestCatalogFlow(t *testing.T) {
    // Seed data → List → Get → Filter → Pagination
    // Admin create → cache invalidation → List updated
}
```

## Требования

- testcontainers-go для PG + Redis
- `go test -tags=integration`
- Real DB + Redis (не mock)
- HTTP handler integration (httptest.Server)
- Multi-tenant isolation тест

## Критерии приёмки

- [ ] testutil package
- [ ] Tenant CRUD integration test
- [ ] Auth flow integration test
- [ ] Catalog flow integration test
- [ ] Multi-tenant isolation test
- [ ] `go test -tags=integration` — 0 failures
