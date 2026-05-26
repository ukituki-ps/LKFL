# LKFL — Makefile
# Единая точка входа для developer-команд.

.DEFAULT_GOAL := build

# ============================================================================
# Переменные
# ============================================================================

BACKEND_DIR  := backend
FRONTEND_DIR := frontend

# DB_DSN из .env (если файл существует), иначе пустая строка
DB_DSN := $(shell if [ -f .env ]; then grep -m1 '^DB_DSN=' .env | cut -d= -f2-; fi)

# ============================================================================
# Build
# ============================================================================

build: ## Собрать lkfl-server
	cd $(BACKEND_DIR) && go build -o ../bin/lkfl-server ./cmd/server/

build-proxy: ## Собрать lkfl-integration-proxy
	cd $(BACKEND_DIR) && go build -o ../bin/lkfl-integration-proxy ./cmd/integration-proxy/

build-all: build build-proxy ## Собрать всё

docker-build: ## Собрать Docker images
	docker compose build

# ============================================================================
# Test
# ============================================================================

test: ## Все тесты + race detector
	cd $(BACKEND_DIR) && go test ./... -count=1 -race

test-coverage: ## Тесты + coverage
	cd $(BACKEND_DIR) && go test ./... -coverprofile=../coverage.out -count=1

test-coverage-html: ## HTML отчёт coverage
	cd $(BACKEND_DIR) && go tool cover -html=../coverage.out -o ../coverage.html

test-integration: ## Integration тесты (testcontainers)
	cd $(BACKEND_DIR) && go test ./... -tags=integration -count=1

test-e2e: ## E2E тесты (Playwright)
	cd $(FRONTEND_DIR) && npx playwright test

# ============================================================================
# Lint
# ============================================================================

lint: ## Go linter
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint-fix: ## Go linter auto-fix
	cd $(BACKEND_DIR) && golangci-lint run ./... --fix

lint-frontend: ## ESLint
	cd $(FRONTEND_DIR) && npm run lint

fmt: ## Go formatting
	cd $(BACKEND_DIR) && go fmt ./...

vet: ## Go vet
	cd $(BACKEND_DIR) && go vet ./...

# ============================================================================
# Migrations (Atlas)
# ============================================================================

migrate-up: ## Применить миграции
	atlas migrate apply --url "$(DB_DSN)" --dir file://migrations

migrate-down: ## Откатить 1 миграцию
	atlas migrate undo --url "$(DB_DSN)" --dir file://migrations --count 1

migrate-status: ## Статус миграций
	atlas migrate status --url "$(DB_DSN)" --dir file://migrations

# ============================================================================
# Docker
# ============================================================================

up: ## Поднять все контейнеры
	docker compose up -d

down: ## Остановить все контейнеры
	docker compose down

down-v: ## Остановить + удалить volumes
	docker compose down -v

logs: ## Логи всех контейнеров
	docker compose logs -f

logs-server: ## Логи lkfl-server
	docker compose logs -f lkfl-server

logs-proxy: ## Логи lkfl-integration-proxy
	docker compose logs -f lkfl-integration-proxy

# ============================================================================
# Dev
# ============================================================================

dev: ## Hot reload (air)
	air -c .air.toml

dev-frontend: ## Vite dev server
	cd $(FRONTEND_DIR) && npm run dev

seed: ## Загрузить seed данные
	cd $(BACKEND_DIR) && go run ./cmd/seed/

clean: ## Очистка артефактов
	rm -rf bin/ coverage.out coverage.html

# ============================================================================
# Help
# ============================================================================

help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# .PHONY
# ============================================================================

.PHONY: build build-proxy build-all docker-build \
	test test-coverage test-coverage-html test-integration test-e2e \
	lint lint-fix lint-frontend fmt vet \
	migrate-up migrate-down migrate-status \
	up down down-v logs logs-server logs-proxy \
	dev dev-frontend seed clean help
