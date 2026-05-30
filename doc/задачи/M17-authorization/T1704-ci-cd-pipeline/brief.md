# T1704 — CI/CD pipeline

## Контекст

Настраиваем CI/CD pipeline: GitHub Actions для проверки кода на PR, Docker build + push на merge, testcontainers для integration tests.

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1702 (backend auth), T1703 (frontend auth)
**ADR:** ADR-030 (CI/CD Pipeline)

**Внутреннее разбиение:**
- **Фаза A (3 дня):** Dockerfiles + GitHub Actions workflows + docker-compose.prod.yml
- **Фаза B (2 дня):** testcontainers-go + integration tests + OpenAPI spec

## Что включено

### GitHub Actions workflows
> **Заменяет** `.github/workflows/go.yml` + `frontend.yml` из T1701 (smoke check).
- `.github/workflows/ci.yml` — lint + test + build + Playwright E2E на PR
- `.github/workflows/cd.yml` — docker build + push ghcr.io + integration test на merge
- `.github/workflows/deploy.yml` — dev deploy (auto), staging (manual approve)

### Dockerfiles (multi-stage — заменяют stub из T1701)
- `Dockerfile.server` — lkfl-server (Go build → alpine). Заменяет stub single-stage из T1701.
- `Dockerfile.proxy` — lkfl-integration-proxy (Go build → alpine). Заменяет stub single-stage из T1701.
- `Dockerfile.frontend` — lkfl-frontend (Vite build → nginx)

### Integration tests (testcontainers-go)
- `internal/testutil/containers.go` — PG + Redis + Keycloak контейнеры
- `*_integration_test.go` — integration тесты с реальными сервисами

### OpenAPI spec
- `openapi/openapi.yaml` — master spec (auto-generated from Go types)
- CI step: `redocly lint openapi.yaml`

### Docker Compose production
- `docker-compose.prod.yml` — production конфигурация (без observability, ghcr.io images)

## Результат

- `git push` → CI запускается автоматически (lint + test + build + Playwright E2E)
- `git push main` → Docker build + push в ghcr.io
- Integration tests проходят через testcontainers (PG + Redis + Keycloak)
- OpenAPI spec валидируется в CI
