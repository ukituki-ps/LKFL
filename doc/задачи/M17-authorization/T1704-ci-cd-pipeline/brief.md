# T1704 — CI/CD pipeline + Staging deploy

## Контекст

Настраиваем CI/CD pipeline: GitHub Actions для проверки кода на PR, Docker build + push на merge,
testcontainers для integration tests, **автоматический деплой на staging (serverAi)**.

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1702 (backend auth), T1703 (frontend auth)
**ADR:** ADR-030 (CI/CD Pipeline)

**Staging сервер:** serverAi (192.168.1.27) — Debian 13, 30GB RAM, 16 CPU, Docker 29.4, Compose 5.1.
Деплой: GitHub Actions → SSH → docker compose pull → docker compose up -d.
На сервере делается только то, что нельзя через git: volumes + .env секреты (разовый provision).

**Внутреннее разбиение:**
- **Фаза A (3 дня):** Dockerfiles + GitHub Actions workflows + docker-compose.prod.yml + deploy.yml
- **Фаза B (2 дня):** testcontainers-go + integration tests + OpenAPI spec

## Что включено

### GitHub Actions workflows
> **Заменяет** `.github/workflows/go.yml` + `frontend.yml` из T1701 (smoke check).
- `.github/workflows/ci.yml` — lint + test + build + Playwright E2E на PR
- `.github/workflows/cd.yml` — docker build + push ghcr.io + integration test на merge
- `.github/workflows/deploy.yml` — **auto deploy на serverAi при merge в main**

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
- `docker-compose.prod.yml` — staging конфигурация (ghcr.io images, persistent volumes)

### Staging deployment
- `infra/deploy/provision-server.sh` — один раз: volumes, .env, подготовка serverAi
- `infra/deploy/deploy-on-server.sh` — deploy script на сервере
- `.env` на serverAi — секреты (POSTGRES_PASSWORD, KEYCLOAK_*, JWT_SECRET)

## Результат

- `git push` → CI запускается автоматически (lint + test + build + Playwright E2E)
- `git push main` → Docker build + push в ghcr.io → **auto deploy на serverAi**
- Integration tests проходят через testcontainers (PG + Redis + Keycloak)
- OpenAPI spec валидируется в CI
- **Staging доступен: http://192.168.1.27 → lkfl-frontend через Nginx**
