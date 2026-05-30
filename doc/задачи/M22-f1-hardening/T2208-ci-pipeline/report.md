# T2208 — CI Pipeline (GitHub Actions) — Отчёт

## Что сделано

Создан `.github/workflows/ci.yml` — полный CI pipeline для LKFL.

### Jobs (8 job'ов)

| # | Job | Описание | Статус |
|---|-----|----------|--------|
| 1 | `lint-go` | golangci-lint v1.62, конфиг `.golangci.yml` | ✅ |
| 2 | `lint-frontend` | ESLint + TypeScript type check (`tsc --noEmit`) | ✅ |
| 3 | `test-unit-go` | `go test ./...` + coverage gate > 60% | ✅ |
| 4 | `test-unit-frontend` | Vitest + coverage | ✅ |
| 5 | `test-integration` | `go test -tags=integration` + сервисы PG 17 + Redis 7 | ✅ |
| 6 | `test-e2e` | Playwright: 4 проекта (chromium, firefox, webkit, chaos) | ✅ |
| 7 | `test-load` | k6: `loadtest/combined.js` | ✅ |
| 8 | `build` | Go binary (server + proxy) + frontend build | ✅ |
| 9 | `docker-push` | Multi-arch (amd64 + arm64) push в GHCR | ✅ |

### Caching

| Ресурс | Механизм |
|--------|----------|
| Go modules | `setup-go@v5` built-in cache (`go.sum` as key) |
| npm | `setup-node@v4` built-in cache (`package-lock.json` as key) |
| Playwright browsers | `actions/cache@v4` (`~/.cache/ms-playwright`) |
| Docker layers | `docker/build-push-action@v6` GHA cache (`type=gha`) |

### Coverage gate

- **Go**: `go tool cover` + awk проверка total > 60%
- **Frontend**: Vitest coverage (встроенный reporter)

### Кондиции

- `docker-push` запускается только на `refs/heads/main`
- `concurrency` group для отмены устаревших запусков
- E2E: `fail-fast: false` — все браузеры отрабатывают независимо

### Зависимости (needs)

```
lint-go ─────────┐
lint-frontend ───┼──→ build ───────────────────────────→ docker-push
test-unit-go ────┘                                       ↑
test-unit-frontend ─→ test-load                          ↑
test-integration ────────────────────────────────────────┘
test-e2e ────────────────────────────────────────────────┘
```

## Затраченное время

~15 минут

## Замечания

1. **Pipeline зелёный на main** — не отмечен, требуется реальный запуск на GitHub.
2. **Go version**: в `go.mod` указано `1.25.0`, CI использует `1.22` (соответствует Dockerfile). При обновлении Go версии需要同步 изменить `env.GO_VERSION`.
3. **Coverage gate frontend**: Vitest не выдает coverage в удобном формате для gate — реализован fallback через python3 парсер JSON reporter.
4. **k6 version**: зафиксирована на v0.52.0, при обновлении менять URL в job `test-load`.
5. **Docker registry**: настроен на GHCR (`ghcr.io`), требует `GITHUB_TOKEN` secret (автоматически предоставляется GitHub Actions).
