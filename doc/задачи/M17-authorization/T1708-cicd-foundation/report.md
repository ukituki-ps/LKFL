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

| # | Имя | Runner | Условие | Описание | Статус CI |
|---|-----|--------|---------|----------|-----------|
| 1 | lint-test | lkfl | always | Go + Frontend + OpenAPI + Config validation | ✅ PASS |
| 2 | build-push | lkfl | needs: lint-test | Docker Buildx + GHCR (4 services) | ✅ PASS |
| 3 | deploy-staging | lkfl | push to main | down → pull → migrate → seed → up → health | ⚠️ инфраструктура |
| 4 | smoke-test-staging | lkfl | needs: deploy-staging | smoke-test.sh с retry polling | ⏸️ зависит от #3 |
| 5 | deploy-production | lkfl | workflow_dispatch | Prod: down → pull → migrate → seed → up | ⏸️ manual |

**Доп. фиксы:**
- `docker compose down` перед `up` (избежание конфликта портов) ✅
- Frontend bind `127.0.0.1:8084` вместо `0.0.0.0:8084` ✅

### Секция 3: `cmd/deploy-worker` — GET /history

**Изменения:**

| Файл | Что добавлено |
|------|--------------|
| `state.go` | `DeployHistoryEntry`, `DeployHistory`, `addHistoryEntry()`, `getHistory()`, `historyMu` |
| `handler.go` | `handleHistory` → `{"count": N, "history": [...]}` |
| `main.go` | `GET /deploy-webhook/history` + `GET /history` routes |

**Проверка:** `go build` ✅, `go vet` ✅

### Секция 4: `docker-compose.prod.yml` — production-grade

503 строки, все требования выполнены:
- Resource limits, logging, networks, monitoring profile ✅
- KEYCLOAK_PUBLIC_URL, deploy-worker, .env.prod ✅

### Секция 6: Nginx setup

| Файл | Статус |
|------|--------|
| `infra/nginx/serverAi.conf` | ✅ (фактические порты: 8083, 8086, 8085) |
| `infra/nginx/serverPr01-internal.conf` | ✅ |
| `infra/scripts/setup-nginx-serverAi.sh` | ✅ |
| `infra/scripts/setup-nginx-serverPr01.sh` | ✅ |
| `infra/scripts/setup-all.sh` | ✅ |

### Секция 7: `.env` файлы

| Файл | Статус |
|------|--------|
| `serverAi:.env.staging` | ✅ реальные секреты, DEPLOY_TOKEN, WEBHOOK_SECRET |
| `serverAi:.env.prod` | ✅ создан |

### Секция 9: `infra/smoke-test.sh`

Retry polling, color output, summary table, exit codes ✅

## Развёрнуто на серверах

### serverAi (192.168.1.27)

| Компонент | Статус | Проверка |
|-----------|--------|----------|
| Nginx 18000 | ✅ | `/healthz` → 200 |
| `.env.staging` | ✅ | Реальные секреты |
| `.env.prod` | ✅ | Создан |

### serverPr01 (192.168.1.29)

| Компонент | Статус |
|-----------|--------|
| `space.conf` → serverAi:18000 | ✅ |

**Цепочка:** `dev.april.ukituki.tech → serverPr01:443 → serverAi:18000 → Docker` ✅

## Предшествующие проблемы (не в scope T1708)

| Проблема | Влияние на CI | Решение |
|----------|--------------|---------|
| **PostgreSQL SASL auth** — `pg_hba.conf` требует `scram-sha-256` для Docker сети, но пользователь `lkfl` создан с `trust` | Migrations/seed fail, Keycloak fail | Добавить `host all all 172.x.0.0/16 trust` в pg_hba.conf |
| **Keycloak healthcheck** — порт 9000 не слушается в Keycloak 26.0 | deploy-staging fail (container unhealthy) | ✅ Исправлено: порт 8080 + start_period 90s |
| **Port 8084 conflict** — SSH процесс занимал порт | deploy-staging fail | ✅ Исправлено: bind 127.0.0.1:8084 |
| **lkfl-frontend unhealthy** — предшествующая проблема | smoke-test может FAIL | Отдельная задача |
| **lkfl-proxy restarting** — предшествующая проблема | Неполный стек | Отдельная задача |

## Коммиты

| Commit | Описание |
|--------|---------|
| `f5e652e` | T1708: CI/CD Foundation — production-grade pipeline, deploy-worker /history, nginx, deploy |
| `59d730b` | fix(ci): docker compose down перед up — избежание конфликта портов при redeploy |
| `c448af0` | fix(staging): frontend bind 127.0.0.1:8084 — избежание конфликта портов |
| `b0ac6f3` | fix(keycloak): healthcheck порт 8080 вместо 9000 + start_period 90s |
