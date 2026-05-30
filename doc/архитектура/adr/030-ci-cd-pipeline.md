# ADR-030: CI/CD Pipeline + Observability

**Статус:** Accepted
**Дата:** 2026-05-30 (обновлено с учётом решений по M17)
**Контекст:** Платформа гибких льгот (LKFL) — monorepo Go + React. CI/CD pipeline для автоматизации build, test, lint, deploy + Observability в dev.

---

## Context and Problem Statement

После перехода на modular monolith (ADR-024) — один бинарник `lkfl-server`, один `go.mod`, один `go build ./...`.
M16 добавил второй бинарник `lkfl-integration-proxy` (ADR-035).
Frontend — React SPA (Vite + Nginx).

**Решения (зафиксированы на архитектурном совещании):**
1. CI — GitHub Actions
2. Integration tests — testcontainers-go (Keycloak + PG + Redis)
3. Dockerfile — 3 штуки (server, proxy, frontend) в M17
4. Registry — ghcr.io
5. Secrets dev — `.env` (Vault отложен до production)
6. Observability — сразу в dev docker-compose (Prometheus + Grafana + Loki + дашборды)
7. OpenAPI spec — в коде

---

## Decision

### Архитектура CI/CD

```
developer
    │
    ├── git push feature/xxx
    │       ↓
    │   [CI Pipeline]
    │       ├── lint: golangci-lint + ESLint
    │       ├── test: go test -race -cover (unit) + vitest (frontend)
    │       ├── build: go build ./...
    │       ├── e2e: Playwright (browser → docker-compose: PG + Redis + Keycloak + server + nginx)
    │       └── coverage: codecov upload
    │
    ├── git push → main (merge)
    │       ↓
    │   [CD Pipeline]
    │       ├── Docker build (multi-stage × 3)
    │       ├── Docker push → ghcr.io
    │       ├── integration: testcontainers (PG + Redis + Keycloak)
    │       └── deploy: dev (auto) → staging (manual) → production (manual)
    │
    └── docker-compose up (dev local)
            ├── lkfl-server
            ├── lkfl-integration-proxy
            ├── postgres:17
            ├── redis:7
            ├── keycloak:26.x
            ├── nginx
            ├── frontend (Vite dev)
            ├── prometheus
            ├── grafana
            └── loki
```

### Registry: ghcr.io

| Artifact | Image |
|----------|-------|
| lkfl-server | `ghcr.io/lkfl/lkfl-server:{sha}` |
| lkfl-integration-proxy | `ghcr.io/lkfl/lkfl-integration-proxy:{sha}` |
| lkfl-frontend | `ghcr.io/lkfl/lkfl-frontend:{sha}` |

Tag strategy: `main` → `{short-sha}`, release → `v{major}.{minor}.{patch}`, dev branch → `{branch}-{sha}`.

### Dockerfiles (multi-stage)

```
├── Dockerfile.server        # lkfl-server
│   stage 1: golang:1.24-alpine → go build -ldflags="-s -w" -o /lkfl-server
│   stage 2: alpine:3.20     → cp binary + migrations + entrypoint
├── Dockerfile.proxy         # lkfl-integration-proxy
│   stage 1: golang:1.24-alpine → go build -ldflags="-s -w" -o /lkfl-integration-proxy
│   stage 2: alpine:3.20     → cp binary + adapters config + entrypoint
└── Dockerfile.frontend      # lkfl-frontend
    stage 1: node:20-alpine  → npm ci && npm run build
    stage 2: nginx:1.27-alpine → cp dist/ + nginx.conf
```

### E2E тесты: Playwright

```yaml
# .github/workflows/ci.yml — e2e job
e2e-test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with: { node-version: '20' }
    - run: npm ci
    - run: npx playwright install --with-deps
    # Поднимаем backend + frontend + Keycloak через docker-compose
    - run: docker-compose -f docker-compose.yml up -d postgres redis keycloak lkfl-server nginx
    - run: npm run test:e2e
    - uses: actions/upload-artifact@v4
      if: always()
      with:
        name: playwright-report
        path: playwright-report/
    - run: docker-compose down
      if: always()
```

**Что тестирует Playwright (M17):**
- `tests/e2e/login.spec.ts` — guest → Keycloak login → Dashboard
- `tests/e2e/forbidden.spec.ts` — 403 screen для неавторизованного доступа

**Как работает:** Playwright запускает реальный браузер (Chromium), подключается к docker-compose среде (lkfl-server + frontend + Keycloak + nginx), проходит login flow через Keycloak UI и проверяет что пользователь попал на Dashboard.

### Integration tests: testcontainers-go

```go
// internal/testutil/containers.go
func SetupIntegrationEnv(ctx context.Context) (*PostgresContainer, *RedisContainer, *KeycloakContainer, error) {
    pg := postgres.Run(ctx, "postgres:17-alpine", ... )
    redis := redis.Run(ctx, "redis:7-alpine", ... )
    kc := keycloak.Run(ctx, "quay.io/keycloak/keycloak:26.0", ... )
    // Wait for Keycloak ready
    kc.WaitForHTTPReady(ctx, "/admin/")
    // Create test realm
    CreateTestRealm(kc.AdminURL(), "test-realm")
    return pg, redis, kc, nil
}
```

CI: `go test -tags=integration -count=1 ./...` — testcontainers поднимают PG + Redis + Keycloak в Docker.
После теста — автоматическая очистка.

**Почему testcontainers, а не GH Actions services:**
- Keycloak не работает как GH Actions service (требует DB init + complex startup)
- Полная изоляция: каждый тест запускает свой set контейнеров
- Детерминированная среда: версия image фиксирована в коде

### Secrets

| Этап | Хранение | Передача |
|------|----------|----------|
| **Dev (M17)** | `.env` + `docker-compose.yml` env vars | Direct env injection |
| **CI (GitHub Actions)** | GitHub Actions secrets | `${{ secrets.X }}` в workflow |
| **Staging** | GitHub Actions secrets | Runtime env injection |
| **Production** | HashiCorp Vault (deferred) | Vault agent inject (ADR-030-v2) |

**M17: `.env` file.** Vault отложен до production-фазы.

### OpenAPI spec в коде

```
├── openapi/
│   ├── openapi.yaml          # Master spec (auto-generated from Go types + manual extensions)
│   └── generated/            # Auto-generated from Go types
```

Генерация из Go-кода с помощью `swaggo/swag` или `oapi-codegen`.
Валидация в CI: `redocly lint openapi.yaml`.
Frontend types генерируются из OpenAPI spec (ADR-032): `openapi-typescript openapi.yaml > src/types/api.ts`.

### Observability в docker-compose (dev)

```yaml
# docker-compose.yml — dev environment
services:
  # ... app services ...

  prometheus:
    image: prom/prometheus:latest
    ports: ["9090:9090"]
    volumes:
      - ./infra/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    command: --web.enable-lifecycle

  grafana:
    image: grafana/grafana:latest
    ports: ["3000:3000"]
    volumes:
      - ./infra/grafana/provisioning:/etc/grafana/provisioning
      - grafana-data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin

  loki:
    image: grafana/loki:latest
    ports: ["3100:3100"]
    command: -config.file=/etc/loki/local-config.yaml

volumes:
  grafana-data:
```

### Grafana дашборды (auto-provisioned)

| Дашборд | Источник | Показатели |
|---------|----------|-----------|
| Platform Overview | Prometheus | RPS, latency, errors, uptime |
| Backend Metrics | Prometheus | http_requests_total, duration, goroutines, memstats |
| Billing Operations | Prometheus | billing_transactions_total, balance_query_total |
| CEL + LLM | Prometheus | cel_generation, cel_evaluation, llm_requests, cost |
| Provider Health | Prometheus | provider latency, errors, circuit breaker state |
| User Activity | Prometheus | active users, catalog views, activations |
| Security | Prometheus | failed_auth, rate_limit_hits, consent_revocations |
| Infrastructure | Prometheus | container_cpu, container_memory, disk_io, network |
| Application Logs | Loki → Grafana | JSON logs с filter по tenant_id, level, svc |

---

## Consequences

### Positive
- Один monorepo pipeline — один CI, один CD
- testcontainers — детерминированные integration tests с Keycloak
- ghcr.io — бесплатный, интегрирован с GitHub, immutable tags
- Observability из коробки в dev — сразу видим метрики
- Multi-stage Dockerfiles — минимальный image size
- OpenAPI в коде — автотипы для frontend, валидация в CI

### Negative
- CI time: ~8 min (Go lint+test+build + Frontend lint+test + integration testcontainers)
- testcontainers требуют Docker daemon в runner (ubuntu-latest — OK)
- 3 Dockerfile = 3 build job в CD pipeline
- Grafana provisioning files нужно поддерживать в репозитории
- `.env` в dev — нет секрет-ротации (Vault deferred)

---

## Related ADR

- ADR-001: Go + React выбор
- ADR-002: PostgreSQL
- ADR-024: Modular Monolith (один бинарник → один pipeline)
- ADR-032: API Types — OpenAPI codegen
- ADR-035: Integration Proxy (второй бинарник → второй Dockerfile)
