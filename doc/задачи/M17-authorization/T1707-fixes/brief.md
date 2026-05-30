# T1707 — Исправление критических дефектов M17

## Контекст

Аудит реализации M17 (T1701–T1706) выявил 8 критических и средних дефектов. T1707 исправляет все выявленные проблемы в рамках одной задачи.

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1702 (backend auth), T1704 (CI/CD)
**ADR:** ADR-036 (Authorization System)

## Что включено

### 1. SQL миграции
- `migrations/001_create_tenants.sql` — tenants table (согласно schema.md)
- `migrations/002_create_tenant_brand.sql` — tenant_brand table
- `migrations/003_create_users.sql` — users table
- `migrations/atlas.hcl` — Atlas config для dev

### 2. RSA JWKS верификация
- `shared/pkg/auth/verifier.go` — заменить dev-mode decoder на полноценную RSA верификацию через `github.com/MicahParks/keyfunc/v3`
- `shared/pkg/auth/verifier_test.go` — тесты с RSA

### 3. migrate/seed subcommands
- `cmd/server/main.go` — добавить команды `migrate` (Atlas) и `seed` (Keycloak realm + users)

### 4. Thread-safe InMemoryRepository
- `internal/tenant/repository.go` — sync.RWMutex для InMemoryRepository

### 5. TenantResolver cache expiry
- `shared/pkg/auth/tenantresolver.go` — TTL-based expiry в resolve()

### 6. KeycloakURL из .env
- `shared/pkg/auth/middleware.go` — KeycloakURL из config, не hard-coded
- `shared/pkg/auth/verifier.go` — KeycloakURL из TenantInfo

### 7. Frontend docker-compose
- `docker-compose.yml` — lkfl-frontend запускает `npm run dev` реально

### 8. Конфигурация через .env
- Секреты (KeycloakURL, KeycloakPassword, JWT_SECRET) — через `.env` и `os.Getenv`, НЕ через GitHub Secrets

## Результат

- `go build ./...` ✅
- `go test ./...` ✅
- Миграции применяются: `make migrate`
- JWT валидация с RSA подписью
- InMemoryRepository thread-safe
- TenantResolver cache с TTL
- KeycloakURL из .env
- Frontend в docker-compose работает реально
