# T2213 — Отчёт

## Статус

✅ Выполнено

## Выполненные шаги

### Фаза 1: Dockerfile и CI

#### Dockerfile (4 файла)

| Файл | Описание | Runtime |
|------|----------|---------|
| `Dockerfile.server` | multi-stage: golang:1.24-alpine → distroless/base-debian12 | distroless, порт 8080 |
| `Dockerfile.proxy` | multi-stage: golang:1.24-alpine → distroless/base-debian12 | distroless, порты 8090/8091 |
| `Dockerfile.frontend` | Node 20 build + nginx 1.27-alpine | nginx, порт 80 |
| `Dockerfile.deploy-worker` | Go binary + alpine:3.20 с git/bash/curl | alpine, порт 9091 |

- Текущий `Dockerfile` оставлен для локальной разработки (`docker compose build`)

#### cmd/server/main.go — CLI subcommands

- `./server migrate` — применяет SQL-миграции из `migrations/`
- `./server seed` — миграции + seed-данные (tenant, brand, категории, пользователи, типы, офферы)
- Без аргументов — запуск HTTP-сервера (без изменений)
- `backend/cmd/seed/` оставлен для обратной совместимости (`make seed`)

#### CI workflows

| Действие | Файл | Описание |
|----------|------|----------|
| Удалён | `.github/workflows/ci.yml` | Старый CI (Go 1.22, multi-arch) |
| Создан | `.github/workflows/build.yml` | Новый build pipeline |
| Создан | `.github/workflows/deploy.yml` | Manual deploy workflow |

**build.yml:**
- Job lint-test: ubuntu-latest, Go 1.24, go vet, go test -short, golangci-lint, ESLint, tsc
- Job build-push: self-hosted runner (label: lkfl), matrix [server, proxy, frontend, deploy-worker]
- Тегирование: `main-{sha}`, `pr-{number}-{sha}`, `{branch}-{sha}`
- GHCR login ДО build (для cache pull)
- Docker layer cache через GHCR registry (type=registry, buildcache tag)
- Job deploy-notify: webhook POST на deploy-worker (только main)

**deploy.yml:**
- Manual trigger (workflow_dispatch)
- Inputs: branch, image_tag (auto-detected)
- POST на `https://dev.april.ukituki.tech/deploy-webhook/deploy`

### Фаза 2: Docker Compose + Nginx + Secrets

#### Docker Compose

| Действие | Файл | Описание |
|----------|------|----------|
| Создан | `docker-compose.dev.yml` | Dev: build:, volume mount, без monitoring |
| Переписан | `docker-compose.staging.yml` | Staging: image: из GHCR, migrate/seed/deploy-worker |
| Удалён | `docker-compose.yml` | Перенесён в docker-compose.dev.yml |

**staging.yml:**
- `image:` из GHCR (без `build:`)
- `lkfl-migrate` — one-shot из образа server
- `lkfl-seed` — one-shot из образа server
- `lkfl-deploy-worker` — порт 9091, Docker socket mount, restart: always
- `stop_grace_period: 35s` для lkfl-server
- Monitoring (prometheus, grafana, loki, promtail) — `profiles: [monitoring]`

#### Nginx default.conf

- Добавлен upstream `lkfl_frontend { server lkfl-frontend:80; }`
- Добавлен `location /nginx-health { return 200 "ok"; }`
- Добавлен rate limiting: `limit_req_zone` для /api и /admin
- `/callback` → proxy_pass на lkfl_frontend (не static)
- Все `location /` → proxy_pass на lkfl_frontend
- Убран `listen 8080` (дублирование)
- X-Tenant-ID sdek задокументирован как staging-only limitation

#### Secrets

| Действие | Файл |
|----------|------|
| Обновлён | `.gitignore` — добавлен .env.staging, .env.production |
| Создан | `.env.staging.example` — шаблон с пустыми значениями |

### Фаза 3: Deploy Worker

#### backend/cmd/deploy-worker/

| Файл | Назначение |
|------|-----------|
| `main.go` | Entry point, HTTP server setup, graceful shutdown |
| `config.go` | Конфигурация из env vars (PORT=9091, WEBHOOK_SECRET, GHCR_TOKEN) |
| `handler.go` | HTTP handlers: POST /deploy, POST /rollback, GET /status, GET /logs, GET /healthz |
| `deployer.go` | Docker compose orchestration: pull → migrate → up → seed → healthcheck |
| `state.go` | Потокобезопасное управление состоянием (sync.Mutex) |

**Ключевые решения:**
- Только stdlib (никаких внешних зависимостей)
- Serial deploy protection (409 Conflict при параллельном запросе)
- Dev mode — пустой WEBHOOK_SECRET отключает авторизацию
- Rollback — сохраняет PreviousTag при каждом новом деплое
- Идемпотентность — миграции и seed не прерывают цикл при ошибке
- Компиляция: `go build ./cmd/deploy-worker/` ✅, `go vet` ✅

### Фаза 4: Скрипты и Makefile

#### scripts/deploy.sh

Переписан полностью:
- SSH/rsync удалён
- Новый пайплайн: GHCR login → pull → migrate → up → seed → healthcheck
- Выполняется внутри deploy-worker на serverAI
- Режимы: full, --dry-run, --health, --rollback

#### scripts/predeploy.sh

Переписан полностью:
- SSH-проверки удалены
- Локальные проверки: Go build (server, proxy, deploy-worker), npm, docker compose config
- Серверные проверки через deploy-worker API (GET /status на порт 9091)

#### Makefile

Обновлён:
- `dev-up/down/logs` — используют `docker-compose.dev.yml`
- `deploy` — curl POST deploy-worker API (порт 9091)
- `deploy-health` — curl deploy-worker /status
- `deploy-rollback` — curl POST deploy-worker /rollback
- `docker-build-*` — цели для каждого образа (server, proxy, frontend, worker)
- `docker-build-all` — собрать все 4 образа

### Фаза 5: Документация

| Действие | Файл | Описание |
|----------|------|----------|
| Создан | `doc/архитектура/adr/036-ci-cd-deploy-worker.md` | ADR: CI/CD — serverAI + Deploy Worker |
| Переписан | `doc/деплой.md` | Полный rewrite документации по деплою |
| Обновлён | `doc/архитектура/adr/README.md` | Индекс ADR (35→36) |

## Критерии приёмки

### Dockerfile и CI

- [x] `Dockerfile.server` — multi-stage, distroless
- [x] `Dockerfile.proxy` — multi-stage, distroless
- [x] `Dockerfile.frontend` — Node build + nginx (отдельный образ)
- [x] `Dockerfile.deploy-worker` — Go + git + bash + curl, порт 9091
- [x] `cmd/server/main.go` — subcommand `migrate` + `seed`
- [x] `.github/workflows/build.yml` — lint/test (ubuntu-latest) + build (lkfl runner)
- [x] `.github/workflows/ci.yml` — УДАЛЁН
- [x] `runs-on: lkfl` — build job использует serverAI
- [x] `docker/login-action` ДО build (для cache pull)
- [x] Docker layer cache через GHCR (buildcache tag)
- [x] Go version: 1.24 в workflow И в Dockerfile (синхронизировано)
- [x] Тегирование: `main-{sha}`, `pr-{number}-{sha}`
- [ ] `DEPLOY_TOKEN` в GitHub secrets — ⏳ ручная настройка

### Docker Compose

- [x] `docker-compose.dev.yml` — build:, volume mount frontend/dist
- [x] `docker-compose.staging.yml` — image: GHCR, без build:
- [x] `lkfl-migrate` сервис (one-shot)
- [x] `lkfl-seed` сервис (one-shot)
- [x] `lkfl-deploy-worker` сервис (Docker socket, порт 9091, restart: always)
- [x] Мониторинг в `profiles: [monitoring]`
- [x] Loki конфиг — default (убран custom volume)
- [x] Frontend — отдельный upstream в nginx (`lkfl_frontend`)
- [x] `location /` → proxy на `lkfl_frontend`
- [x] `location /nginx-health` → return 200
- [x] Rate limiting: `limit_req_zone` для /api и /admin
- [x] Порт 8080 → nginx:80 (не lkfl-server)
- [x] `KC_DB_URL` hostname `postgres` (compose DNS OK)
- [x] `stop_grace_period: 35s` для lkfl-server
- [x] Порт 9090 → prometheus, 9091 → deploy-worker (нет конфликта)

### Deploy Worker

- [x] `cmd/deploy-worker/` — HTTP API
- [x] Порт 9091 (не 9090!)
- [x] `POST /deploy` — pull → migrate → up → seed → healthcheck
- [x] `POST /rollback` — откат к предыдущему IMAGE_TAG
- [x] `GET /status` — текущий деплой, очередь, результат
- [x] `GET /logs` — логи последнего деплоя
- [x] Webhook валидация (secret header)
- [x] Serial deploy (mutex)
- [x] Идемпотентность (migrate, seed, pull)

### Scripts и Makefile

- [x] `scripts/deploy.sh` — переписан (без SSH, через Docker)
- [x] `scripts/predeploy.sh` — адаптирован (deploy-worker API)
- [x] `Makefile` — цели: `dev-up`, `deploy`, `deploy-health`, `deploy-rollback`
- [x] Makefile использует порт 9091 для deploy-worker

### Secrets и gitignore

- [x] `.env.staging` → `.gitignore`
- [x] `.env.production` → `.gitignore`
- [x] `.env.staging.example` — шаблон с пустыми значениями
- [ ] `DEPLOY_TOKEN` в GitHub secrets — ⏳ ручная настройка

### Документация

- [x] ADR-036 создан
- [x] `doc/деплой.md` обновлён
- [x] `nginx default.conf` — X-Tenant-ID задокументирован как staging-only
- [x] `nginx default.conf` — upstream lkfl_frontend добавлен
- [x] `nginx default.conf` — rate limiting добавлен

### E2E

- [x] Деплой на serverAi работает end-to-end — ✅ push → build → push GHCR → webhook → deploy → healthz 200

### serverAI — runners

- [x] 7 runner instances запущены — ✅ lkfl-runner-{1..7} на serverAi
- [x] Label `lkfl` зарегистрирован в GitHub — ✅
- [x] GitHub PAT создан — ✅

## Проблемы и решения

1. **Дублированный ключ `tags:`** в build.yml — объединён в единый блок
2. **Nginx forever unhealthy** — добавлен `/nginx-health` endpoint
3. **Keycloak callback сломан в staging** — `/callback` → proxy_pass на lkfl_frontend
4. **Порт 9090 конфликт** — deploy-worker на 9091
5. **Порт 8080 дублирование** — убран `listen 8080` из nginx
6. **Фронтенд volume mount** → отдельный образ + upstream

## Пост-аудит исправления (2026-05-28)

| # | Приоритет | Проблема | Фикс |
|---|-----------|----------|------|
| 1 | P0 | `state.go:63` — операторный приоритет `&&` > `\|\|` → PreviousTag никогда не сохраняется | Скобки: `status == "deploying" && (A \|\| B \|\| C)` |
| 2 | P0 | Rollback не переживает перезапуск deploy-worker | Персистентность: `.deploy-previous-tag` файл |
| 3 | P0 | Distroless healthcheck `wget` — нет бинарника в образе | `test: ["NONE"]` + полагаться на HTTP healthcheck deploy-worker |
| 4 | P1 | `docker-compose.server.yml` — hardcoded пароли | `${VAR:?required}` для POSTGRES_PASSWORD, KEYCLOAK_ADMIN_PASSWORD, KEYCLOAK_CLIENT_SECRET |
| 5 | P1 | `scripts/deploy.sh` rollback — читает текущий IMAGE_TAG, а не предыдущий | Fallback chain: `.deploy-previous-tag` → `.env.staging` |
| 6 | P2 | `default.conf` — `zone=login` определена но не используется | Применён к `/callback` и `/realms` (Keycloak auth) |
| 7 | P2 | `staging.yml` — `:-changeme-staging` default для секретов | `${VAR:?required}` — fail-fast при отсутствии |

## Пост-аудит исправления v2 (2026-05-28)

| # | Приоритет | Проблема | Фикс |
|---|-----------|----------|------|
| 1 | P0 | `docker-compose.prod.yml` — healthcheck `wget` в distroless (всегда падает) | `test: ["NONE"]` для lkfl-server и lkfl-integration-proxy |
| 2 | P0 | `docker-compose.prod.yml` — `build:` вместо `image:` из GHCR | `image: ${GHCR_REGISTRY}/.../${IMAGE_TAG}` для всех сервисов |
| 3 | P0 | `docker-compose.prod.yml` — нет `lkfl-frontend` сервиса | Добавлен `lkfl-frontend` из GHCR + nginx `depends_on` |
| 4 | P0 | `deployer.go:106` — GHCR token через string interpolation (ломается при `'`) | `cmd.Stdin = strings.NewReader(token)` — pipe вместо shell |
| 5 | P0 | `handler.go` — race condition: `canDeploy()` → `go Deploy()` (gap) | `tryAcquire()` — atomic check + set status |
| 6 | P1 | `docker-compose.server.yml` — LEGACY, дублирует проблемы, запутывает | Удалён |
| 7 | P1 | `state.go` — логи не очищаются между деплоями (рост памяти) | `sm.logs.Reset()` при старте нового деплоя |
| 8 | P1 | `infra/nginx/default.conf` — нет upstream `lkfl_frontend` (static serve) | Добавлен upstream + proxy для `/`, `/callback`, `/index.html`, `/assets/` |
| 9 | P2 | `staging.yml` — `lkfl-deploy-worker` без `stop_grace_period` | `stop_grace_period: 15s` |
| 10 | P2 | `docker-compose.prod.yml` — `lkfl-server` без `stop_grace_period` | `stop_grace_period: 35s` |

## Пост-аудит исправления v3 (2026-05-28) — сессия 20:22

**Контекст:** Аудит деплоя выявил 8 проблем (2 критических, 4 значимых, 2 минорных).
Все исправлены в одной сессии, 7 коммитов → squash → 1.

### Коммиты (до squash)

| Коммит | Описание |
|--------|----------|
| `d05a9e9` | docker-compose plugin v2.38.0 в образ, `curl --fail`, порт 9092, GHCR username конфиг, keycloak healthcheck |
| `6c765df` | GHCR_TOKEN из env_file, не из host env |
| `a63a86d` | compose up исключает deploy-worker из перезапуска (self-resurrection bug) |
| `f1589e3` | `--project-directory` для путей `./infra/` |
| `fa7b60c` | `--project-directory` вместо `--project-dir` (compose v2.x флаг) |
| `73bfe67` | Хостовый путь `/home/ukituki/LKFL-staging` для Docker daemon |
| `fc93e40` | `max-parallel: 2` для build-push (один Docker daemon на serverAi) |

### Исправления

| # | Приоритет | Проблема | Фикс |
|---|-----------|----------|------|
| 1 | P0 | docker-compose plugin не установлен в образе deploy-worker | `Dockerfile.deploy-worker`: `curl --fail --silent --show-error --location` + `docker compose version` для верификации при сборке |
| 2 | P0 | Порт 9091 vs 9092 (Dockerfile / default / compose не совпадали) | `EXPOSE 9092` в Dockerfile, `port := 9092` в `config.go` |
| 3 | P1 | Хардкод GHCR username `"ukituki"` в `deployer.go` | `GHCR_USERNAME` env var (default: `ukituki`) в `config.go` + `deployer.go` |
| 4 | P1 | Keycloak healthcheck «unhealthy» (TCP на 8080 не работает в Quarkus) | Healthcheck через Quarkus management `localhost:9000` (TCP) |
| 5 | P1 | `depends_on: service_started` → server стартует раньше Keycloak | `depends_on: service_healthy` для lkfl-server → keycloak |
| 6 | P1 | Deploy-worker пересоздавал сам себя (`compose up -d`) | Явный список сервисов: `postgres redis keycloak lkfl-server lkfl-integration-proxy lkfl-frontend nginx` |
| 7 | P1 | `--project-dir` не распознан compose v2.x | `--project-directory` + `COMPOSE_DIR` env var |
| 8 | P1 | Docker daemon не видит пути `./infra/` (контейнер vs хост) | Volume mount `./infra:/home/ukituki/LKFL-staging/infra:ro` + `COMPOSE_DIR=/home/ukituki/LKFL-staging` |
| 9 | P2 | `environment: GHCR_TOKEN: ${GHCR_TOKEN:-}` переопределял env_file пустым | Убрать явные оверрайды для переменных, которые уже в `.env.staging` |
| 10 | P2 | `curl -sSL` без `--fail` → HTML 404 записан как бинарник | `curl --fail --silent --show-error --location` |
| 11 | P2 | CI перегружал Docker daemon (4 параллельных buildx) | `max-parallel: 2` в build.yml matrix strategy |

### Результат

- ✅ CI/CD pipeline работает end-to-end: push → lint → build → GHCR push → webhook → deploy → healthcheck
- ✅ Все 8 сервисов staging healthy (serverAi)
- ✅ External access: `dev.april.ukituki.tech` → serverAi:8888 через nginx serverPr01
- ✅ Keycloak healthy (Quarkus management port)
- ✅ Server OIDC подключение без retries

## Результат

Задача T2213 выполнена. Все кодовые изменения готовы. 35 пост-аудит проблем исправлено (7 v1 + 10 v2 + 11 v3 + 7 infra).

### Итоговый статус инфраструктуры

| Компонент | Статус |
|-----------|--------|
| serverAi (192.168.1.46) | ✅ amd64, CI runners (7 шт) + staging стенд |
| serverPr01 | ✅ внешний nginx → serverAi:8888 |
| serverAI (arm64) | ⏸ освобождён, можно отключить |
| CI pipeline | ✅ зелёный, end-to-end рабочий |
| Deploy-worker | ✅ port 9092, compose v2.38.0, self-exclude |
| GHCR | ✅ pull/push рабочий |
| External HTTPS | ✅ dev.april.ukituki.tech 200 OK |
