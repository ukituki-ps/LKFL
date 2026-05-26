# T1703 — Makefile

## Веха

M17-bootstrap

## Тип

code

## Контекст

`Makefile` — единая точка входа для всех developer команд.
Заменяет запоминание `go build`, `docker compose`, `go test` флагов.

## Что сделать

Создать `Makefile` со следующими targets:

### Build

| Target | Команда | Описание |
|--------|---------|----------|
| `build` | `go build -o bin/lkfl-server ./cmd/server/` | Собрать lkfl-server |
| `build-proxy` | `go build -o bin/lkfl-integration-proxy ./cmd/integration-proxy/` | Собрать proxy |
| `build-all` | `make build && make build-proxy` | Собрать всё |
| `docker-build` | `docker compose build` | Собрать Docker images |

### Test

| Target | Команда | Описание |
|--------|---------|----------|
| `test` | `go test ./... -count=1 -race` | Все тесты + race detector |
| `test-coverage` | `go test ./... -coverprofile=coverage.out -count=1` | Тесты + coverage |
| `test-coverage-html` | `go tool cover -html=coverage.out -o coverage.html` | HTML отчёт |
| `test-integration` | `go test ./... -tags=integration -count=1` | Integration тесты (testcontainers) |
| `test-e2e` | `cd frontend && npx playwright test` | E2E тесты (Playwright) |

### Lint

| Target | Команда | Описание |
|--------|---------|----------|
| `lint` | `golangci-lint run ./...` | Go linter |
| `lint-fix` | `golangci-lint run ./... --fix` | Go linter auto-fix |
| `lint-frontend` | `cd frontend && npm run lint` | ESLint |
| `fmt` | `go fmt ./...` | Go formatting |
| `vet` | `go vet ./...` | Go vet |

### Migrations

| Target | Команда | Описание |
|--------|---------|----------|
| `migrate-up` | `atlas migrate apply --url "$(echo $$DB_DSN)" --dir file://migrations` | Применить миграции |
| `migrate-down` | `atlas migrate undo --url "$(echo $$DB_DSN)" --dir file://migrations --count 1` | Откатить 1 миграцию |
| `migrate-status` | `atlas migrate status --url "$(echo $$DB_DSN)" --dir file://migrations` | Статус миграций |

### Docker

| Target | Команда | Описание |
|--------|---------|----------|
| `up` | `docker compose up -d` | Поднять все контейнеры |
| `down` | `docker compose down` | Остановить все контейнеры |
| `down-v` | `docker compose down -v` | Остановить + удалить volumes |
| `logs` | `docker compose logs -f` | Логи всех контейнеров |
| `logs-server` | `docker compose logs -f lkfl-server` | Логи server |
| `logs-proxy` | `docker compose logs -f lkfl-integration-proxy` | Логи proxy |

### Dev

| Target | Команда | Описание |
|--------|---------|----------|
| `dev` | `air -c .air.toml` | Hot reload (air) |
| `dev-frontend` | `cd frontend && npm run dev` | Vite dev server |
| `seed` | `go run ./cmd/seed/` | Загрузить seed данные |
| `clean` | `rm -rf bin/ coverage.out coverage.html` | Очистка артефактов |

### Default

```makefile
.DEFAULT_GOAL := build
.PHONY: build build-proxy build-all test lint up down docker-build migrate-up
```

## Требования

- `make` без аргументов = `make build`
- Все targets с `.PHONY`
- DB_DSN берётся из `.env` (через `$(shell grep DB_DSN .env | cut -d= -f2)`)
- Не использовать `cd` в targets — использовать `workdir` pattern или `$(MAKE) -C`

## Критерии приёмки

- [ ] `make build` собирает lkfl-server
- [ ] `make test` запускает все тесты
- [ ] `make up` поднимает docker-compose
- [ ] `make lint` запускает golangci-lint
- [ ] `make migrate-up` применяет миграции
- [ ] Все targets с `.PHONY`
- [ ] `make` без аргументов = build
