# T1702 — Backend: auth core

## Контекст

Реализация ядра аутентификации и авторизации на Go backend. Включает shared/pkg/auth (переиспользуемый), internal пакеты (auth, tenant, api), миграции БД и entry points.

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1701 (инфраструктура)
**ADR:** ADR-036 (Authorization System), ADR-003 (Keycloak), ADR-009 (Multi-tenancy)

## Что включено

### shared/pkg/auth (переиспользуемый)
- `tenantresolver.go` — TenantResolver: subdomain→slug, cache, DB lookup, context
- `verifier.go` — JWT Verifier: RS256, JWKS per realm, iss/azp/exp validation
- `middleware.go` — JWT middleware wrapper (extracts Bearer, calls verifier, stores in context)
- `rbac.go` — RBACGuard middleware (realm roles check)
- `claims.go` — Claims struct, context helpers
- `errors.go` — Auth errors (401/403), writeError helper
- `cache.go` — JWKS cache per realm with TTL, Redis integration

### internal пакеты
- `internal/auth/` — OIDCVerifier (thin wrapper), tenant config builder
- `internal/tenant/` — Tenant CRUD, brand config, repository, service
- `internal/api/` — router с middleware chain, route registration, health endpoints

### Миграции и seed
- `migrations/` — Atlas init, только auth-таблицы: tenants, tenant_brand, users (согласно schema.md §lkfl_platform). Остальные 44 таблицы — в будущих вехах.
- Seed data — демо tenant (slug: `demo`), демо users с ролями
- Seed Keycloak — `infra/keycloak/seed-demo.sh`:
  - Создание realm "demo" из `tenant-template.json`
  - Создание пользователя `admin` (роль `admin`, пароль `admin`)
  - Создание пользователя `employee` (роль `employee`, пароль `employee`)
  - Создание client `lkfl-frontend` + `lkfl-server`

### Entry points
- `cmd/server/main.go` — заменить stub (из T1701) на рабочий: DI, middleware chain, route registration, graceful shutdown
- `cmd/worker/main.go` — stub (Asynq worker, нужен для `go build ./...` но не используется в M17)

### Unit-тесты (пишутся параллельно с кодом)
- `shared/pkg/auth/*_test.go` — verifier, middleware, rbac, tenant resolver
- `internal/tenant/*_test.go` — CRUD, brand config
- `internal/api/*_test.go` — router middleware chain

## Результат

- `lkfl-server` компилируется: `go build ./...`
- `/api/v1/me` возвращает профиль с ролями (JWT valid)
- `/api/v1/healthz` → 200
- Tenant resolution работает по subdomain
- JWT валидация работает с realm-specific JWKS
- RBAC guard блокирует endpoints без нужных ролей
- Migrations применяются: `make migrate`
- Unit-тесты зелёные: `go test ./...`
