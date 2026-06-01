# T1708 — CI/CD Foundation: production-grade pipeline, deploy-worker, auto-deploy

## Контекст

T1707 исправил критические дефекты M17, но CI/CD — временные решения:
- 4 workflow файла (`ci.yml`, `cd.yml`, `deploy.yml`, `build.yml`) дублируют друг друга
- `ci.yml` — PATH/nvm повторение в каждом step (25+ раз)
- `cd.yml` — нет cache, нет buildx, образы `lkfl-server:main` (не org prefix)
- `deploy.yml` — костыль: `cp docker-compose.prod.yml ~/lkfl/`
- `build.yml` — падает: нет `KEYCLOAK_PUBLIC_URL` в `docker-compose.prod.yml`
- `cmd/deploy-worker/` — код существует (5 файлов), но не соответствует целевой спецификации (нет /history, /health → /healthz, in-memory mutex ≠ file lock)
- Нет production deploy path

**Цель:** Один `build.yml` — полный цикл CI → CD → Deploy Staging → Smoke Test + Deploy Production (manual).
Deploy-worker — полноценный сервис с `/deploy`, `/health`, `/status`, `/rollback`, `/history`, concurrency lock.
Production-grade фундамент для быстрого деплоя на прод.

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1707 (fixes)
**ADR:** ADR-039 (CI/CD Deploy Worker), ADR-036 (Authorization System — контекст M17)

---

## Инфраструктура (факты с серверов)

### serverPr01 (192.168.1.46) — Edge Proxy

```
Nginx: dev.april.ukituki.tech:443 (SSL Letsencrypt)
       ↓ proxy_pass
       http://192.168.1.46:18000  ← localhost, НУЖЕН nginx → serverAi:18000
```

`/etc/nginx/sites-enabled/space.conf` — готов, SSL cert есть, proxy на localhost:18000.
**Проблема:** на порту 18000 serverPr01 ничего нет. Нужен nginx → serverAi:18000.

### serverAi (192.168.1.27) — App Server

```
Docker compose: 12 сервисов (lkfl_*, open-webui)
├── lkfl_server       healthy   127.0.0.1:8083:8080
├── lkfl_nginx        healthy   0.0.0.0:80:80 (stock default config!)
├── lkfl_frontend     unhealthy 127.0.0.1:8086:80
├── lkfl_proxy        restarting
├── lkfl_keycloak     unhealthy 127.0.0.1:8085:8080
├── lkfl_postgres     healthy   127.0.0.1:5432
├── lkfl_redis        healthy   127.0.0.1:6379
├── lkfl_migrate      healthy
├── lkfl_prometheus   healthy   127.0.0.1:9090
├── lkfl_loki         healthy   127.0.0.1:3100
├── lkfl_grafana      healthy   127.0.0.1:3001
└── Runner'ы: 7x serverAI-runner-{1-7}, label: lkfl, online
```

**Проблемы:**
- Nginx на serverAi — stock default config, не проксирует LKFL
- Порт 80 — stock nginx, не LKFL nginx
- Нет nginx на 18000 для serverPr01 → serverAi link

### Файлы на серверах

| Путь | Описание |
|------|----------|
| `/home/ukituki/lkfl/.env` | Реальные секреты (POSTGRES_PASSWORD, REDIS_PASSWORD, KEYCLOAK_*, JWT_SECRET) |
| `/home/ukituki/LKFL-staging/.env.staging` | Staging конфиг (weak passwords: lkfl/lkfl, changeme-staging) |
| `/home/ukituki/LKFL-staging/docker-compose.staging.yml` | Staging compose (с deploy-worker service) |
| `/home/ukituki/lkfl/docker-compose.prod.yml` | Prod compose (без KEYCLOAK_PUBLIC_URL) |

---

## Архитектура пайплайна

```
push/PR → main
    │
    ├── Job 1: lint-test (lkfl, self-hosted runner на serverAi)
    │   ├── Go: mod tidy, vet, test -race, golangci-lint
    │   ├── Frontend: npm ci, eslint, tsc --noEmit, vitest run
    │   ├── OpenAPI: redocly lint (continue-on-error)
    │   └── Config validation: KEYCLOAK_PUBLIC_URL, KC_HOSTNAME
    │
    ├── Job 2: build-push (lkfl self-hosted runner на serverAi)
    │   ├── Docker Buildx + GHCR login (secrets.GHCR_PAT)
    │   ├── Matrix: [server, proxy, frontend, deploy-worker]
    │   ├── Tag: {branch}-{sha7} + {service}:main-latest (только main)
    │   └── Push: ghcr.io/ukituki-ps/lkfl/{service}:{tag}
    │
    ├── Job 3: deploy-staging (lkfl runner на serverAi) — auto on main push
    │   ├── Docker login GHCR
    │   ├── Sync docker-compose.staging.yml + .env.staging на сервер
    │   ├── docker compose pull
    │   ├── docker compose run --rm lkfl-migrate (retry 3x)
    │   ├── docker compose run --rm lkfl-seed
    │   ├── docker compose up -d (all services)
    │   └── Health check polling: server, frontend, keycloak
    │
   ├── Job 4: smoke-test-staging (lkfl, self-hosted runner на serverAi)
    │   ├── needs: [deploy-staging]
    │   ├── if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    │   ├── Checkout `infra/smoke-test.sh`
    │   ├── Retry polling: 5x, 10s interval
    │   └── infra/smoke-test.sh https://dev.april.ukituki.tech
    │
    └── Job 5: deploy-production (lkfl runner на serverAi) — MANUAL dispatch
        ├── needs: [smoke-test-staging]
        ├── if: github.event_name == 'workflow_dispatch'
        ├── Sync docker-compose.prod.yml + .env.prod на сервер
        ├── docker compose pull
        ├── docker compose run --rm lkfl-migrate (retry 3x)
        ├── docker compose run --rm lkfl-seed
        ├── docker compose up -d
        └── Health check polling
```

---

## Что нужно реализовать

### 1. Удаление устаревших workflow

| Файл | Действие | Причина |
|------|----------|---------|
| `.github/workflows/ci.yml` | Удалить | Дублирует Job 1 |
| `.github/workflows/cd.yml` | Удалить | Дублирует Job 2 |
| `.github/workflows/deploy.yml` | Удалить | Дублирует Job 3 |

### 2. `build.yml` — единый production-grade pipeline

#### Job 1: lint-test (lkfl runner на serverAi)
- `runs-on: lkfl`
- Go: mod tidy, vet, test -race, golangci-lint
- Frontend: npm ci, eslint, tsc --noEmit, vitest run
- OpenAPI: redocly lint (continue-on-error: true)
- Config validation: KEYCLOAK_PUBLIC_URL + KC_HOSTNAME
- NPM registry auth для @ukituki-ps
- **Нет PATH/nvm хаков** — Go 1.24.4 и Node 20 уже в PATH runner'а

#### Job 2: build-push (lkfl runner на serverAi)
- `runs-on: lkfl`
- `docker/setup-buildx-action` — cache backend
- `docker/login-action` — `${{ secrets.GHCR_PAT }}` (PAT с write:packages)
- Matrix: `[server, proxy, frontend, deploy-worker]`
- `max-parallel: 2` (один Docker daemon)
- `fail-fast: false`
- Tag: `{branch}-{sha7}` + `{service}:main-latest` (только main)
- Registry: `ghcr.io/ukituki-ps/lkfl/{service}`
- Frontend build-arg: `NPM_TOKEN=${{ github.token }}`

#### Job 3: deploy-staging (lkfl runner на serverAi)
- `runs-on: lkfl`
- `needs: [build-push]`
- `if: github.ref == 'refs/heads/main' && github.event_name == 'push'`
- Checkout → sync compose + env на serverAi
- `docker compose -f docker-compose.staging.yml -p lkfl-staging pull`
- `docker compose -f docker-compose.staging.yml -p lkfl-staging run --rm lkfl-migrate` (retry 3x)
- `docker compose -f docker-compose.staging.yml -p lkfl-staging run --rm lkfl-seed`
- `docker compose -f docker-compose.staging.yml -p lkfl-staging up -d`
- Health check polling: server `/healthz`, frontend `/`, keycloak (5 attempts, 15s interval)
- `docker image prune -f`

#### Job 4: smoke-test-staging (lkfl runner на serverAi)
- `runs-on: lkfl`
- `needs: [deploy-staging]`
- `if: github.ref == 'refs/heads/main' && github.event_name == 'push'`
- Checkout `infra/smoke-test.sh`
- Retry polling: 5x, 10s interval
- `infra/smoke-test.sh https://dev.april.ukituki.tech`
- Прямой LAN доступ к serverPr01 (192.168.1.46) — нет задержки публичного интернета

#### Job 5: deploy-production (lkfl runner на serverAi)
- `runs-on: lkfl`
- `needs: [smoke-test-staging]`
- `if: github.event_name == 'workflow_dispatch'`
- Checkout → sync compose + env
- `docker compose -f docker-compose.prod.yml -p lkfl-prod pull`
- `docker compose -f docker-compose.prod.yml -p lkfl-prod run --rm lkfl-migrate`
- `docker compose -f docker-compose.prod.yml -p lkfl-prod run --rm lkfl-seed`
- `docker compose -f docker-compose.prod.yml -p lkfl-prod up -d`
- Health check polling
- Concurrency: `deploy-production` (cancel-in-progress: false)

### 3. `cmd/deploy-worker` — доработка существующего сервиса

**Фактический код (5 файлов):**

```
backend/cmd/deploy-worker/
├── main.go          # HTTP server, :9092, 14 routes, graceful shutdown
├── handler.go       # POST /deploy, /deploy/pr, /rollback, GET /status, /logs, /healthz
├── deployer.go      # Docker compose orchestrator: pull, migrate, seed, up, health
├── state.go         # Thread-safe state manager + in-memory mutex concurrency lock
└── config.go        # Config из env vars (6 params)
```

**Что нужно ДОБАВИТЬ (delta):**

| Файл | Что добавить |
|------|-------------|
| `handler.go` | `GET /history` — список деплоев (сейчас нет) |
| `state.go` | `DeployHistory` slice, min 10 записей, сохраняется при success |

**Endpoints (фактические + добавленные):**

| Method | Path | Auth | Описание | Статус |
|--------|------|------|----------|--------|
| POST | `/deploy-webhook/deploy` | Bearer token | Запустить deploy | ✅ есть |
| POST | `/deploy-webhook/deploy/pr` | Bearer token | PR preview | ✅ есть (stub) |
| POST | `/deploy-webhook/rollback` | Bearer token | Rollback к предыдущему | ✅ есть |
| GET | `/deploy-webhook/status` | Нет | Текущий статус | ✅ есть |
| GET | `/deploy-webhook/logs` | Нет | Логи последнего деплоя | ✅ есть |
| GET | `/deploy-webhook/healthz` | Нет | Health check | ✅ есть |
| GET | `/history` | Нет | История деплоев | ❌ нужно добавить |

**Aliases без префикса** (для прямого internal доступа): `/deploy`, `/deploy/pr`, `/rollback`, `/status`, `/logs`, `/healthz` — ✅ есть

**Concurrency lock (фактический):**
- In-memory `sync.Mutex` в `state.go` (`tryAcquire` / `canDeploy`)
- Atomic check-and-set — защита от race condition
- Если lock занят → 409 Conflict

**Deploy логика (фактическая):**
1. Validate webhook token (`validateAuth`)
2. Atomic acquire (`tryAcquire`)
3. Асинхронный запуск `go h.deployer.Deploy(req)`
4. GHCR login (если `GHCR_TOKEN` set)
5. Pull новых образов из GHCR
6. `docker compose run --rm lkfl-migrate` (warning при ошибке — идемпотент)
7. `docker compose up -d` (хардкод списка сервисов)
8. `docker compose run --rm lkfl-seed` (warning при ошибке — идемпотент)
9. Health check polling 30x2s=60s
10. Save state (PreviousTag в файл `.deploy-previous-tag`)
11. Return 202 Accepted

### 4. `docker-compose.prod.yml` — реальный production

Переделать в production-grade:
- Все сервисы с resource limits
- Restart policies: `unless-stopped`
- Health checks со start_period
- Logging: json-file, max-size 10m, max-file 5
- Networks: backend (internal) + frontend
- `KEYCLOAK_PUBLIC_URL` в lkfl-server
- `lkfl-deploy-worker` service с Docker socket
- Persistent volumes для postgres, redis
- Monitoring: prometheus, grafana, loki, promtail (profile: monitoring)

### 5. `docker-compose.staging.yml` — sync с main repo

Файл уже существует. Убедиться:
- `lkfl-deploy-worker` service корректно настроен
- Все сервисы с health checks
- Monitoring profile

### 6. Nginx setup

#### serverAi (192.168.1.27)
Нужен nginx на 18000, проксирующий на внутренние сервисы:

```
infra/nginx/serverAi.conf:
server {
    listen 18000;
    # proxy rules for api, frontend, keycloak
}
```

Деплой через git:
```bash
git clone / sync repo → cp infra/nginx/serverAi.conf → /etc/nginx/sites-available/lkfl.conf → ln -s → sites-enabled → nginx -t && nginx -s reload
```

#### serverPr01 (192.168.1.46)
Нужен nginx на 18000, проксирующий на serverAi:

```
infra/nginx/serverPr01-internal.conf:
server {
    listen 18000;
    location / {
        proxy_pass http://192.168.1.27:18000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Деплой через ssh + git:
```bash
scp infra/nginx/serverPr01-internal.conf serverPr01:/etc/nginx/sites-available/lkfl-internal.conf
ssh serverPr01 "ln -sf /etc/nginx/sites-available/lkfl-internal.conf /etc/nginx/sites-enabled/ && nginx -t && nginx -s reload"
```

### 7. `.env` файлы

#### `/home/ukituki/lkfl/.env` — уже есть (реальные секреты)

#### `/home/ukituki/lkfl/.env.staging` — создать (production-grade секреты)
Переделать `.env.staging` из weak passwords в реальные:
```
POSTGRES_PASSWORD=<из .env>
REDIS_PASSWORD=<из .env>
KEYCLOAK_ADMIN=kcadmin
KEYCLOAK_ADMIN_PASSWORD=<из .env>
JWT_SECRET=<из .env>
IMAGE_TAG=main-latest
GHCR_REGISTRY=ghcr.io/ukituki-ps/lkfl
KEYCLOAK_PUBLIC_URL=https://dev.april.ukituki.tech/realms/lkfl-sdek
KEYCLOAK_CLIENT_ID=lkfl-spa
KEYCLOAK_CLIENT_SECRET=<sгенерировать>
SENTRY_DSN=
```

#### `/home/ukituki/lkfl/.env.prod` — создать (для production)
Аналогично staging, но с production URL и секретами.

### 8. GitHub Secrets + .env

**GitHub Secrets (Actions):**
- `GHCR_PAT` — PAT с `write:packages` для push образов в GHCR (требуется для cross-push из self-hosted runner)
- `DEPLOY_TOKEN` — Bearer token для webhook deploy-worker

**Env на serverAi (`.env` файлы):**
- `WEBHOOK_SECRET` — Bearer token для deploy-worker (читается `os.Getenv`)
- Остальные секреты: `POSTGRES_PASSWORD`, `REDIS_PASSWORD`, `KEYCLOAK_ADMIN_PASSWORD`, `JWT_SECRET`

**Environments:** НЕ используются. Deploy staging — `if: push to main`. Deploy production — `if: workflow_dispatch`.

### 9. `infra/smoke-test.sh` — улучшить

Уже существует с 6 чекпоинтами. Добавить:
- Retry polling logic (не `sleep 90`)
- Exit code handling
- Color output для CI

### 10. Script для setup nginx на серверах

```
infra/scripts/setup-nginx-serverAi.sh
infra/scripts/setup-nginx-serverPr01.sh
infra/scripts/setup-all.sh  # master script
```

Выполняется один раз при инициализации сервера.

---

## Что НЕ входит в задачу

- Blue/green deploy (отдельная веха)
- Canary deployments (отдельная веха)
- Multi-region (отдельная веха)
- Slack/email notifications (отдельная задача)
- Auto-scaling
- Disaster recovery

---

## Критерии приёмки

### CI
- [ ] `build.yml` — единственный workflow файл (ci/cd/deploy удалены)
- [ ] Job 1 (lint-test) на PR: Go + Frontend + OpenAPI + Config validation PASS
- [ ] Job 1 на `lkfl` (self-hosted runner на serverAi)

### CD
- [ ] Job 2 (build-push) собирает все 4 образа
- [ ] Образы в `ghcr.io/ukituki-ps/lkfl/{service}:{tag}`
- [ ] На main — `{service}:main-latest` tag
- [ ] Buildx cache работает

### Deploy Staging (auto)
- [ ] Job 3 запускается на push в main
- [ ] Migrations применяются (retry 3x)
- [ ] Seed загружен
- [ ] Все сервисы up
- [ ] Health check polling OK (server, frontend, keycloak)

### Smoke Test Staging
- [ ] Job 4 запускается после deploy-staging
- [ ] Retry polling (5x, 10s interval)
- [ ] Все 6 чекпоинтов PASS:
  - `/healthz` → 200
  - `/` → 200
  - `/api/v1/engagements/` → 401
  - Keycloak discovery → 200
  - Login redirect без internal hostname
  - Nginx health → 200

### Deploy Production (manual)
- [ ] Job 5 запускается через `workflow_dispatch`
- [ ] Sync prod compose + env
- [ ] Migrate + seed + up
- [ ] Health check polling OK

### Deploy-worker
- [ ] `cmd/deploy-worker/` — 5+ файлов, компилируется
- [ ] `go build ./cmd/deploy-worker/` — OK
- [ ] Dockerfile.deploy-worker — build OK
- [ ] POST /deploy-webhook/deploy — запускает deploy (202 Accepted)
- [ ] GET /deploy-webhook/healthz — 200 OK
- [ ] GET /deploy-webhook/status — JSON статус
- [ ] POST /deploy-webhook/rollback — откат к предыдущему тегу
- [ ] GET /history — список деплоев (добавить если нет)
- [ ] Concurrency lock — 409 при параллельном deploy (in-memory mutex)
- [ ] PreviousTag сохранение в `.deploy-previous-tag` — работает

### Nginx
- [ ] serverAi: nginx на 18000, проксирует LKFL сервисы
- [ ] serverPr01: nginx на 18000, проксирует на serverAi:18000
- [ ] `https://dev.april.ukituki.tech/` → 200
- [ ] SSL cert valid

### Инфраструктура
- [ ] `.env.staging` на serverAi — реальные секреты (не weak)
- [ ] `.env.prod` на serverAi — production секреты
- [ ] `docker-compose.prod.yml` — production-grade
- [ ] `docker-compose.staging.yml` — sync с repo
- [ ] Scripts: `setup-nginx-serverAi.sh`, `setup-nginx-serverPr01.sh`
- [ ] CI использует `${{ secrets.GHCR_PAT }}` для GHCR login
- [ ] DEPLOY_TOKEN в `.env` на serverAi

---

## Порядок реализации

1. **Секция 3** — `cmd/deploy-worker` доработка (добавить GET /history в handler.go + state.go)
2. **Секция 4** — `docker-compose.prod.yml` (production-grade)
3. **Секция 7** — `.env.staging` + `.env.prod` на serverAi
4. **Секция 6** — Nginx setup scripts + deploy на сервера
5. **Секция 2** — `build.yml` rewrite (единый pipeline, все job'ы на `lkfl`)
6. **Секция 9** — Улучшение smoke-test.sh (retry polling вместо sleep 90)
7. **Секция 1** — Удаление ci.yml, cd.yml, deploy.yml
8. **Тест** — push на main → CI → CD → Deploy Staging → Smoke Test

---

## Риски

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| GHCR PAT истекает | Низкая | Проверка login в CI |
| Runner недоступен | Средняя | Fallback: `self-hosted` label |
| Migrate падает | Средняя | Retry 3x с backoff |
| Deploy webhook не отвечает | Средняя | Timeout 120s, retry polling |
| Nginx config error на serverPr01 | Средняя | `nginx -t` перед reload |
| SSL cert истекает | Низкая | Certbot auto-renew |
