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

# Deploy worker
DEPLOY_WORKER_URL ?= http://serverDev:9091

# ============================================================================
# Build
# ============================================================================

build: ## Собрать lkfl-server
	cd $(BACKEND_DIR) && go build -o ../bin/lkfl-server ./cmd/server/

build-proxy: ## Собрать lkfl-integration-proxy
	cd $(BACKEND_DIR) && go build -o ../bin/lkfl-integration-proxy ./cmd/integration-proxy/

build-all: build build-proxy ## Собрать всё

docker-build: ## Собрать Docker images
	docker compose -f docker-compose.dev.yml build

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
# Docker — Dev (docker-compose.dev.yml)
# ============================================================================

dev-up: ## Локальная разработка (docker-compose.dev.yml)
	docker compose -f docker-compose.dev.yml up -d

dev-down: ## Остановить локальные сервисы
	docker compose -f docker-compose.dev.yml down

dev-logs: ## Логи локальных сервисов
	docker compose -f docker-compose.dev.yml logs -f

# ============================================================================
# Docker — Legacy (up/down без -f, для staging)
# ============================================================================

up: ## Поднять все контейнеры (staging)
	docker compose -f docker-compose.staging.yml up -d

down: ## Остановить все контейнеры (staging)
	docker compose -f docker-compose.staging.yml down

down-v: ## Остановить + удалить volumes (staging)
	docker compose -f docker-compose.staging.yml down -v

logs: ## Логи всех контейнеров (staging)
	docker compose -f docker-compose.staging.yml logs -f

logs-server: ## Логи lkfl-server (staging)
	docker compose -f docker-compose.staging.yml logs -f lkfl-server

logs-proxy: ## Логи lkfl-integration-proxy (staging)
	docker compose -f docker-compose.staging.yml logs -f lkfl-integration-proxy

# ============================================================================
# Deploy (через deploy-worker API)
# ============================================================================

deploy: ## Деплой на staging (через deploy-worker API)
	curl -X POST $(DEPLOY_WORKER_URL)/deploy \
		-H "Content-Type: application/json" \
		-d '{"branch":"main","imageTag":"main-latest"}'

deploy-health: ## Healthcheck через deploy-worker
	curl -s $(DEPLOY_WORKER_URL)/status | jq .

deploy-rollback: ## Роллбэк через deploy-worker
	curl -X POST $(DEPLOY_WORKER_URL)/rollback

predeploy: ## Пред-деплой валидация (локально + deploy-worker)
	./scripts/predeploy.sh

predeploy-quick: ## Быстрая пред-деплой валидация
	./scripts/predeploy.sh --quick

# ============================================================================
# Docker Build — Образы
# ============================================================================

docker-build-server: ## Собрать lkfl-server
	docker build -f Dockerfile.server -t lkfl-server:latest .

docker-build-proxy: ## Собрать lkfl-integration-proxy
	docker build -f Dockerfile.proxy -t lkfl-integration-proxy:latest .

docker-build-frontend: ## Собрать lkfl-frontend
	docker build -f Dockerfile.frontend -t lkfl-frontend:latest .

docker-build-worker: ## Собрать deploy-worker
	docker build -f Dockerfile.deploy-worker -t deploy-worker:latest .

docker-build-all: docker-build-server docker-build-proxy docker-build-frontend docker-build-worker ## Собрать все образы

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
	dev-up dev-down dev-logs \
	up down down-v logs logs-server logs-proxy \
	dev dev-frontend seed clean help \
	deploy deploy-health deploy-rollback predeploy predeploy-quick \
	docker-build-server docker-build-proxy docker-build-frontend docker-build-worker docker-build-all
