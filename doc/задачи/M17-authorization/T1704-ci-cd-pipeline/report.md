# T1704 — Отчёт: CI/CD Pipeline

## Статус

✅ Завершено

## Что сделано

### GitHub Actions
- `.github/workflows/build.yml` — CI pipeline:
  - Go build (all packages)
  - Go test (all packages, 53+ PASS)
  - Frontend build (Vite)
  - Docker build (server, proxy)
- `.github/workflows/frontend.yml` — frontend CI

### Docker
- `Dockerfile.server` — multi-stage build golang:1.24-alpine → scratch
- `Dockerfile.proxy` — multi-stage build golang:1.24-alpine → scratch

### Test infrastructure
- `internal/testutil/testcontainers.go` — PostgreSQL + Redis testcontainers
  - applyMigrations → `shared/pkg/migrate.Apply` (T1709 D13 deduplication)
  - TestServer helper с mock JWT middleware
  - HTTP helpers (GetWithToken, PostWithToken, etc.)

## Проблемы

- Дублирующийся код миграций (main.go ↔ testcontainers.go) → вынесен в `shared/pkg/migrate` (T1709 D13)

## Следующие шаги

Н/Д — задача завершена.
