# ADR-030: CI/CD Pipeline

**Статус:** Accepted
**Дата:** 2026-05-25
**Контекст:** Платформа гибких льгот (LKFL) — monorepo Go + React. Нужен CI/CD pipeline для автоматизации build, test, lint, deploy.

---

## Context and Problem Statement

После перехода на modular monolith (ADR-024) — один бинарник `lkfl-server`, один `go.mod`, один `go build ./...`. Нужен единый pipeline для всего monorepo.

**Требования:**
- Автоматическая проверка кода при PR
- Build + test на каждом push
- Docker image build + push в registry
- Deploy в dev/staging/production
- Migration execution перед deploy

---

## Decision

### Stack

| Stage | Tool | Назначение |
|-------|------|-----------|
| CI Runner | GitHub Actions | Pipeline orchestration |
| Go Lint | golangci-lint | Static analysis |
| Go Test | `go test -race -cover` | Unit + race detection |
| Go Build | `go build ./...` | Compilation check |
| TS Lint | ESLint + Prettier | Frontend linting |
| TS Test | Vitest | Frontend unit tests |
| Docker | Buildx | Multi-arch image build |
| DB Migration | Atlas | SQL-first migrations |
| Deploy | docker-compose / k8s | Dev: compose, Prod: k8s |

### Pipeline

```yaml
# .github/workflows/ci.yml

name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  # 1. Go checks
  go-checks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: golangci-lint run --timeout=5m ./...
      - run: go test -race -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  # 2. Frontend checks
  frontend-checks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npm run lint
      - run: npm run test

  # 3. Docker build
  docker-build:
    needs: [go-checks, frontend-checks]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v6
        with:
          push: ${{ github.event_name != 'pull_request' }}
          tags: lkfl-server:${{ github.sha }}

  # 4. Integration tests (dev env)
  integration:
    needs: [docker-build]
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:17
      redis:
        image: redis:7
    steps:
      - uses: actions/checkout@v4
      - run: atlas migrate apply
      - run: go test -tags=integration ./...
```

### Environment Promotion

```
PR → CI (lint + test) → merge → main → Docker build → dev deploy
                                                          ↓ (manual approve)
                                                     staging deploy
                                                          ↓ (manual approve)
                                                     production deploy
```

### Secrets Management

| Secret | Где хранится | Как передаётся |
|--------|-------------|---------------|
| Keycloak URL | GitHub Actions secrets | `KEYCLOAK_URL` env |
| DB credentials | HashiCorp Vault | Vault agent inject |
| Provider API keys | HashiCorp Vault | Vault agent inject |
| OpenAI API key | HashiCorp Vault | Vault agent inject |
| Sentry DSN | GitHub Actions secrets | env var |

### Migration Strategy

```
1. Docker build → 2. Atlas migrate apply → 3. Health check → 4. Traffic switch
```

- Migrations выполняются **до** запуска приложения
- Rollback: `atlas migrate down` + предыдущий Docker image tag

---

## Consequences

### Positive
- Один pipeline для всего monorepo
- Автоматическая проверка качества кода
- Zero-downtime deploy (blue-green через docker-compose)
- Reproducible builds (Docker multi-stage)

### Negative
- CI time: ~5 min (Go + Frontend + Integration)
- Нужен Docker registry (GitHub Packages / Harbor)
- Нужно настраивать Vault для production

---

## Related ADR

- ADR-001: Go + React выбор
- ADR-002: PostgreSQL
- ADR-024: Modular Monolith (один бинарник → один pipeline)
