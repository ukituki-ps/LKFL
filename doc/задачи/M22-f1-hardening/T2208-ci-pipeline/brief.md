# T2208 — CI Pipeline (GitHub Actions)

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Настройка CI pipeline на GitHub Actions.
Автоматический запуск lint, тестов, сборки и публикации Docker image.

## Что сделать

### Workflow

`.github/workflows/ci.yml`:

```yaml
name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    # golangci-lint + ESLint
  test-unit:
    # go test ./... (без integration)
    # npm test (frontend)
  test-integration:
    # go test -tags=integration ./...
    # testcontainers (PG + Redis)
  test-e2e:
    # playwright test
    # Chromium + Firefox + Webkit
  test-chaos:
    # playwright test --project=chaos
    # 100 chaos tests
  test-load:
    # k6 run loadtest/combined.js
  build:
    # go build ./...
    # npm run build
  docker-push:
    # docker buildx build --push
    # multi-arch (amd64, arm64)
```

### Stages

1. **Lint** — golangci-lint (Go) + ESLint (frontend)
2. **Unit test** — `go test ./...` + `npm test`
3. **Integration test** — `go test -tags=integration ./...` (testcontainers)
4. **E2E test** — Playwright (3 browsers)
5. **Chaos test** — Playwright chaos (100 тестов)
6. **Load test** — k6 (thresholds)
7. **Build** — Go binary + frontend build
8. **Docker push** — multi-arch image в registry

### Coverage gate

- Go: > 60% coverage (fail if below)
- Frontend: > 60% coverage (fail if below)
- Coverage report как artifact

### Caching

- Go modules: `~/.cache/go-build`, `go mod download` cache
- npm dependencies: `node_modules` cache
- Docker layers: buildx cache
- Playwright browsers: cache

### Secrets

- `DOCKER_REGISTRY_USERNAME` / `DOCKER_REGISTRY_PASSWORD`
- `GITHUB_TOKEN` (для package registry)

## Критерии приёмки

- [ ] `.github/workflows/ci.yml` создан
- [ ] Lint stage (golangci-lint + ESLint)
- [ ] Unit test stage (Go + frontend)
- [ ] Integration test stage (testcontainers)
- [ ] E2E test stage (3 browsers)
- [ ] Chaos test stage (100 тестов)
- [ ] Load test stage (k6 thresholds)
- [ ] Build stage
- [ ] Docker push stage (multi-arch)
- [ ] Coverage gate (> 60%)
- [ ] Caching настроен
- [ ] Pipeline зелёный на main
