# T1708 — Отчёт: CI/CD Foundation

## Статус
✅ Реализовано (30.05.2026)

## Что реализовано

### Секция 1: Удаление устаревших workflow

| Файл | Действие | Статус |
|------|----------|--------|
| `.github/workflows/ci.yml` | Удалён | ✅ |
| `.github/workflows/cd.yml` | Удалён | ✅ |
| `.github/workflows/deploy.yml` | Удалён | ✅ |
| `.github/workflows/build.yml` | Rewrite — единый pipeline (5 job'ов) | ✅ |

### Секция 2: `build.yml` — единый pipeline

**Job'ы:**

| # | Имя | Runner | Условие | Описание |
|---|-----|--------|---------|----------|
| 1 | lint-test | lkfl | always | Go (mod tidy, vet, test, golangci-lint) + Frontend (npm ci, eslint, tsc, vitest) + OpenAPI lint + Config validation |
| 2 | build-push | lkfl | needs: lint-test | Docker Buildx + GHCR (matrix: server, proxy, frontend, deploy-worker) |
| 3 | deploy-staging | lkfl | push to main | Pull → migrate (retry 3x) → seed → up → health check → prune |
| 4 | smoke-test-staging | lkfl | needs: deploy-staging | smoke-test.sh с retry polling (5x, 10s) |
| 5 | deploy-production | lkfl | workflow_dispatch | Prod: pull → migrate → seed → up → health check |

### Секция 3: `cmd/deploy-worker` — GET /history

**Изменения:**

| Файл | Что добавлено |
|------|--------------|
| `state.go` | `DeployHistoryEntry` struct, `DeployHistory` type, `addHistoryEntry()`, `getHistory()`, `historyMu` |
| `handler.go` | `handleHistory` — GET /history → `{"count": N, "history": [...]}` |
| `main.go` | `GET /deploy-webhook/history` + `GET /history` routes |

**История деплоев:**
- In-memory хранение (до 10 записей)
- Auto-save при success/failed
- Duration вычисление (StartedAt → FinishedAt)
- Отдельный mutex (`historyMu`) для thread safety

**Проверка:**
- `go build ./cmd/deploy-worker/` ✅
- `go vet ./cmd/deploy-worker/` ✅

### Секция 4: `docker-compose.prod.yml` — production-grade

| Требование | Статус |
|------------|--------|
| Resource limits для всех сервисов | ✅ |
| Logging json-file (10m, 5 files) для Go | ✅ |
| Networks: lkfl_backend (internal) + lkfl_frontend | ✅ |
| KEYCLOAK_PUBLIC_URL в lkfl-server | ✅ |
| KEYCLOAK_ISSUER: lkfl-sdek | ✅ |
| deploy-worker service | ✅ |
| Monitoring profile (prometheus, grafana, loki, promtail) | ✅ |
| start_period у healthcheck | ✅ |
| Registry: ghcr.io/ukituki-ps/lkfl | ✅ |
| Volumes: lkfl_prod_ prefix | ✅ |
| env_file: .env.prod | ✅ |

### Секция 5: `docker-compose.staging.yml` — sync

Файл уже production-ready (486 строк), не требовал изменений. ✅

### Секция 6: Nginx setup

**serverAi (192.168.1.27):**

| Файл | Статус |
|------|--------|
| `infra/nginx/serverAi.conf` | ✅ создан (фактические порты: 8083, 8086, 8085) |
| Развёрнут на сервере | ✅ порт 18000, прокси на Docker-сервисы |

**serverPr01 (192.168.1.29):**

| Файл | Статус |
|------|--------|
| `infra/nginx/serverPr01-internal.conf` | ✅ создан (прокси на 192.168.1.27:18000) |
| `space.conf` обновлён | ✅ proxy_pass → 192.168.1.27:18000 |

**Setup скрипты:**

| Файл | Статус |
|------|--------|
| `infra/scripts/setup-nginx-serverAi.sh` | ✅ создан |
| `infra/scripts/setup-nginx-serverPr01.sh` | ✅ создан |
| `infra/scripts/setup-all.sh` | ✅ создан (master script) |

### Секция 7: `.env` файлы

**`.env.staging` (serverAi: `/home/ukituki/LKFL-staging/.env.staging`):**

| Переменная | Статус |
|------------|--------|
| POSTGRES_PASSWORD | ✅ реальный секрет (из .env) |
| REDIS_PASSWORD | ✅ реальный секрет |
| KEYCLOAK_ADMIN_PASSWORD | ✅ реальный секрет |
| JWT_SECRET | ✅ реальный секрет |
| DEPLOY_TOKEN | ✅ 64-char hex |
| WEBHOOK_SECRET | ✅ 64-char hex |
| IMAGE_TAG | ✅ main-latest |
| GHCR_REGISTRY | ✅ ghcr.io/ukituki-ps/lkfl |
| KEYCLOAK_PUBLIC_URL | ✅ https://dev.april.ukituki.tech/realms/lkfl-sdek |

**`.env.prod` (serverAi: `/home/ukituki/LKFL-prod/.env.prod`):** ✅ создан

### Секция 9: `infra/smoke-test.sh` — улучшение

| Улучшение | Статус |
|-----------|--------|
| Retry polling (--retry N, дефолт 1) | ✅ |
| 10s интервал между попытками | ✅ |
| Exit code: 0=PASS, 1=FAIL, 2=network error | ✅ |
| Color output (автоматически отключается в CI) | ✅ |
| Summary table по каждой попытке | ✅ |
| Timeout protection (10s connect, 15s max) | ✅ |
| 6 чекпоинтов | ✅ |

## Развёрнуто на серверах

### serverAi (192.168.1.27)

| Компонент | Статус | Проверка |
|-----------|--------|----------|
| Nginx на 18000 | ✅ | `http://127.0.0.1:18000/healthz` → 200 |
| `.env.staging` | ✅ | Реальные секреты |
| `.env.prod` | ✅ | Создан |
| `docker-compose.prod.yml` | ✅ | Production-grade |

### serverPr01 (192.168.1.29)

| Компонент | Статус | Проверка |
|-----------|--------|----------|
| space.conf → 192.168.1.27:18000 | ✅ | `https://dev.april.ukituki.tech/healthz` → 200 |

## Цепочка проверена

```
https://dev.april.ukituki.tech/
  → serverPr01:443 (nginx, SSL)
  → serverAi:18000 (nginx, прокси)
  → Docker-сервисы (lkfl-server:8083, lkfl-frontend:8086, keycloak:8085)
```

| Endpoint | Код | Статус |
|----------|-----|--------|
| `/healthz` → lkfl-server | 200 | ✅ |
| `/` → lkfl-frontend | 200 | ✅ |
| `/api/v1/engagements/` → lkfl-server | 404 | ⚠️ ожидается (Go код 0%) |
| Keycloak discovery | 404 | ⚠️ keycloak unhealthy (предшествующая проблема) |

### Предшествующие проблемы (не исправлены в этой задаче)
- Keycloak unhealthy (dev mode, health check port 9000)
- lkfl-frontend unhealthy (предшествующая проблема)
- lkfl-proxy restarting (предшествующая проблема)

## Метрики

| Метрика | Значение |
|---------|----------|
| Workflow файлов | 1 (build.yml) |
| Job'ов в pipeline | 5 (lint-test, build-push, deploy-staging, smoke-test-staging, deploy-production) |
| Deploy-worker endpoints | 8 (7 старых + GET /history) |
| Deploy-worker history | до 10 записей in-memory |
| Nginx конфигов | 2 (serverAi.conf, serverPr01-internal.conf) |
| Setup скриптов | 3 (serverAi, serverPr01, setup-all) |
| Compose файлов | 4 (dev, staging, prod, prod rewrite) |
| .env файлов | 2 (staging с реальными секретами, prod) |
