# T1707 — Отчёт

## Статус

✅ Завершено

## Что сделано

### 1. SQL миграции
- ✅ `migrations/001_create_tenants.sql` — tenants table (UUID PK, slug UNIQUE, keycloak_realm, currency_config JSONB, rate_limits JSONB, status CHECK)
- ✅ `migrations/002_create_tenant_brand.sql` — tenant_brand (SERIAL PK, FK tenants ON DELETE CASCADE)
- ✅ `migrations/003_create_users.sql` — users (UUID PK, FK tenants, keycloak_user_id, HR fields, status/location CHECK)
- ✅ `migrations/atlas.hcl` — Atlas config dev environment

### 2. RSA JWKS верификация
- ✅ go.mod: добавлен `github.com/MicahParks/keyfunc/v3`
- ✅ `shared/pkg/auth/verifier.go`: полноценная RSA JWKS через keyfunc.NewDefault
- ✅ `shared/pkg/auth/verifier_test.go`: адаптированы тесты

### 3. migrate/seed subcommands
- ✅ `cmd/server/main.go`: `lkfl-server migrate` — dry-run миграций
- ✅ `cmd/server/main.go`: `lkfl-server seed` — dry-run seed
- ✅ `Makefile`: migrate target → `./bin/lkfl-server migrate`
- ✅ `Dockerfile.server`: COPY migrations/ в контейнер

### 4. Thread-safe + config fixes
- ✅ `internal/tenant/repository.go`: sync.RWMutex, копирование Tenant при возврате
- ✅ `shared/pkg/auth/tenantresolver.go`: cacheEntry с ExpiresAt, TTL-based eviction
- ✅ `shared/pkg/auth/middleware.go`: KeycloakURL из os.Getenv("KEYCLOAK_URL")

### 5. Docker
- ✅ `docker-compose.yml`: lkfl-frontend запускает `npm install && npm run dev -- --host`
- ✅ `docker-compose.yml`: убран dependency cycle (lkfl-frontend ↔ nginx)

### 6. CI/CD — переделка на self-hosted runner'ы
- ✅ `.github/workflows/ci.yml`: все jobs → `runs-on: self-hosted` (Go 1.24 + Node 20 на serverAi)
- ✅ `.github/workflows/cd.yml`: docker build+push на self-hosted, GHCR через GITHUB_TOKEN
- ✅ `.github/workflows/deploy.yml`: deploy на self-hosted, docker compose pull+up напрямую на сервере
- ✅ `serverAi`: Go 1.24.4 установлен, Node 20.20.2 (nvm), 7 runner'ов работают

## Проблемы

- go mod tidy может сломать build из-за testcontainers → opentelemetry → Go 1.25. Pinned зависимости сохраняются.
- migrate/seed — dry-run. Реальная миграция требует pgx (будет в будущих вехах).
- ~/lkfl на serverAi на ветке infra/clean-slate — deploy.yml sync'ит docker-compose.prod.yml из checkout.

## Проверки

| Команда | Результат |
|---------|-----------|
| `go build ./...` | ✅ OK (3 бинарника: server, integration-proxy, worker) |
| `go vet ./...` | ✅ OK |
| `go test -race ./...` | ✅ OK (53/53 PASS, 0 FAIL, 3 пакета) |
| `docker compose config` | ✅ OK |

## Метрики

| Метрика | Значение |
|---------|----------|
| Файлов создано | 4 (migrations/*.sql, migrations/atlas.hcl) |
| Файлов изменено | 9 (verifier.go, verifier_test.go, repository.go, tenantresolver.go, middleware.go, cmd/server/main.go, docker-compose.yml, Makefile, Dockerfile.server) |
| Файлов удалено | 1 (migrations/.gitkeep) |
| go.mod | +1 зависимость (keyfunc/v3) |
| Unit-тестов | 53 PASS (33 auth + 13 tenant + 7 api) |

## Дефекты, закрытые в T1707

| Код | Дефект | Приоритет | Статус |
|-----|--------|-----------|--------|
| K1 | Миграции БД отсутствуют | 🔴 HIGH | ✅ исправлено |
| K2 | JWT без RSA подписи | 🔴 HIGH | ✅ исправлено |
| K13 | migrate/seed команды отсутствуют | 🟡 MEDIUM | ✅ исправлено (dry-run) |
| K4 | InMemoryRepository не thread-safe | 🟡 MEDIUM | ✅ исправлено |
| K5 | TenantResolver cache без expiry | 🟡 MEDIUM | ✅ исправлено |
| K6 | Hard-coded KeycloakURL | 🟢 LOW | ✅ исправлено |
| T1 | Frontend docker-compose placeholder | 🟢 LOW | ✅ исправлено |
