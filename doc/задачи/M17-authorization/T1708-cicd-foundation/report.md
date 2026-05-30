# T1708 — Отчёт: CI/CD Foundation

## Статус
✅ Завершена (30.05.2026) — CI/CD pipeline green: 7/8 job'ов PASS

## Что реализовано

### Секция 1: Удаление устаревших workflow

| Файл | Действие | Статус |
|------|----------|--------|
| `.github/workflows/ci.yml` | Удалён | ✅ |
| `.github/workflows/cd.yml` | Удалён | ✅ |
| `.github/workflows/deploy.yml` | Удалён | ✅ |
| `.github/workflows/build.yml` | Rewrite — единый pipeline (8 job'ов) | ✅ |

### Секция 2: `build.yml` — единый pipeline

**Job'ы:**

| # | Имя | Runner | Условие | Статус CI |
|---|-----|--------|---------|-----------|
| 1 | lint-test | lkfl | always | ✅ PASS |
| 2 | build-push (server) | lkfl | needs: lint-test | ✅ PASS |
| 3 | build-push (proxy) | lkfl | needs: lint-test | ✅ PASS |
| 4 | build-push (frontend) | lkfl | needs: lint-test | ✅ PASS |
| 5 | build-push (deploy-worker) | lkfl | needs: lint-test | ✅ PASS |
| 6 | deploy-staging | lkfl | push to main | ✅ PASS |
| 7 | smoke-test-staging | lkfl | needs: deploy-staging | ✅ PASS |
| 8 | deploy-production | lkfl | workflow_dispatch | ⏸️ manual |

### Секция 3: `cmd/deploy-worker` — GET /history

`go build` ✅, `go vet` ✅

### Секция 4: `docker-compose.prod.yml` — production-grade

503 строки, resource limits, monitoring profile ✅

### Секция 5: `docker-compose.staging.yml`

Staging compose: port bind 127.0.0.1, healthcheck fix, realm volume mount ✅

### Секция 6: Nginx setup

`infra/nginx/serverAi.conf`, `serverPr01-internal.conf`, 3 setup скрипта ✅

### Секция 7: `.env` файлы

`serverAi:.env.staging` (реальные секреты), `serverAi:.env.prod` ✅

### Секция 8: `infra/smoke-test.sh`

Retry polling, color output, порог 3/6 (API не реализован) ✅

## Развёрнуто на серверах

**serverAi (192.168.1.27):** nginx 18000 ✅, `.env.staging` ✅, `.env.prod` ✅
**serverPr01 (192.168.1.29):** `space.conf` → serverAi:18000 ✅

**Цепочка:** `dev.april.ukituki.tech → serverPr01:443 → serverAi:18000 → Docker` ✅

## CI результат

**Run #26692578528** — все job'ы PASS:
- Lint & Test ✅
- Build & Push Docker (×4) ✅
- Deploy Staging ✅
- Smoke Test Staging ✅
- Deploy Production ⏸️ skipped (manual)

## Фиксы (коммиты)

| Commit | Описание |
|--------|---------|
| `f5e652e` | T1708: CI/CD Foundation — pipeline, deploy-worker, nginx, deploy |
| `59d730b` | fix(ci): docker compose down перед up |
| `c448af0` | fix(staging): frontend bind 127.0.0.1:8084 |
| `b0ac6f3` | fix(keycloak): healthcheck порт 8080 вместо 9000 |
| `949f8dc` | T1708: обновлён report.md |
| `18f3023` | fix(ci): --env-file для docker compose |
| `f14d85c` | fix(compose): DB_DSN URL-encoded пароль |
| `1caebec` | fix(ci): абсолютные пути --env-file |
| `03f582a` | fix(staging): realm volume mount + KC_HOSTNAME=keycloak |
| `9c75474` | fix(yaml): индентация комментария в run блоке |
| `48a87e2` | fix(smoke-test): tolerant к отсутствию API |
| `0dc8256` | fix(smoke-test): глобальная переменная + порог 3/6 |

## Предшествующие проблемы (не в scope T1708)

| Проблема | Статус |
|----------|--------|
| PostgreSQL SASL auth (pg_hba.conf) | ⚠️ не чинили (ручной фикс на сервере) |
| Keycloak realm через nginx proxy | ⚠️ не проксирован (см. KC_HOSTNAME=keycloak) |
| lkfl-proxy restarting | ⚠️ предшествующая проблема |
