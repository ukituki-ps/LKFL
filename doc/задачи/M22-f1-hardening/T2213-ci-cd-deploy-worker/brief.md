# T2213 — CI/CD: GitHub Actions + Deploy Worker

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Текущий `deploy.sh` выполняется локально и лезет на сервер по SSH. Это ненадёжно:
- Локальная машина — не триггер, требует ручного запуска
- Go-сборка на сервере (медленно, требует Go в Docker multi-stage)
- Seed и миграции не работают (Go нет на сервере, distroless без shell)
- Hardcoded `serverDev` — нельзя масштабировать
- Один `docker-compose.yml` на всё (dev + staging) — `build:` вместо `image:`
- Фронтенд — volume mount `./frontend/dist/` вместо отдельного образа
- Loki ломает compose (несовместимый конфиг)
- Grafana/Promtail избыточны на staging
- Secrets (`.env.staging`) в репозитории
- Нет `docker-compose.staging.yml`
- `KC_DB_URL` hostname `postgres` ≠ `lkfl-postgres` (источник проблем)
- Hardcoded `X-Tenant-ID sdek` в nginx
- Порт 8080 маппится на nginx (дублирование)

**Дополнительно (аудит 28.05):**
- `.github/workflows/ci.yml` конфликтует с новым `build.yml` — два параллельных push в GHCR
- `/nginx-health` endpoint отсутствует в `default.conf` — nginx forever unhealthy
- Фронтенд upstream не добавлен в nginx — `location /` serve static, а не proxy на `lkfl-frontend`
- `.gitignore` не покрывает `.env.staging` — пароли могут попасть в репозиторий
- Go version mismatch: `ci.yml` использует 1.22, `Dockerfile` — 1.24
- **Порт 9090 конфликт**: prometheus и deploy-worker оба хотят 9090
- Нет GHCR login step до build — cache pull не работает
- Нет `stop_grace_period` — graceful shutdown 30s в main.go требует 35s
- **Webhook URL** — настроен через nginx serverPr01: `https://dev.april.ukituki.tech/deploy-webhook` → serverDev:9091
- **`/callback` location** serve static из volume → в staging volume нет → Keycloak callback сломается
- **`docker compose pull`** без `docker login GHCR` → 401 Unauthorized для private repo
- **Port mapping 8080:8080** → должно быть 8080:80 (nginx внутри слушает 80, не 8080)

Новая архитектура:
- **serverAI** — мощный сервер (Debian 13, 30GB RAM, 16 CPU, Docker 29.4.3) с 7 self-hosted GitHub Actions runner'ами → сборка Docker-образов
- **GitHub Actions** оркестрирует пайплайн: lint/test на public runner, build на serverAI
- **GHCR** — registry для образов
- **Deploy Worker** на serverDev получает webhook, пуллит образы, запускает миграции/seed
- Каждая ветка и каждый PR триггерит пересборку
- Отдельные compose файлы: `dev` (build) и `staging` (pull из GHCR)
- Docker profiles для мониторинга (не на staging по умолчанию)

## Зависимости

- T2209 (Docker Production) — нужны multi-stage Dockerfile
- T2208 (CI Pipeline) — CI workflow расширяется deploy-шагом
- T2210 (Деплой на стенд) — staging docker-compose адаптируется под pull из GHCR

## Архитектура

```
GitHub (push / PR)
  │
  │ .github/workflows/build.yml (ЗАМЕНЯЕТ ci.yml)
  │   ├── lint + test          → runs-on: ubuntu-latest (public, лёгкие)
  │   │
  │   ├── docker buildx        → runs-on: lkfl (serverAI, self-hosted, 7 runner'ов)
  │   │   ├── docker login GHCR (до build — для cache pull!)
  │   │   ├── server   → ghcr.io/ukituki-ps/lkfl/server:{tag}
  │   │   ├── proxy    → ghcr.io/ukituki-ps/lkfl/proxy:{tag}
  │   │   ├── frontend → ghcr.io/ukituki-ps/lkfl/frontend:{tag}
  │   │   └── deploy-worker → ghcr.io/ukituki-ps/lkfl/deploy-worker:latest
  │   │
  │   .github/workflows/deploy.yml (только main)
  │     └── POST → https://dev.april.ukituki.tech/deploy-webhook/deploy
  │         ← через nginx serverPr01 → serverDev:9091
  │
  ▼
serverDev — Deploy Worker (docker-compose.staging.yml)
  ├── deploy-worker (:9091) ← PORT 9091! (9090 занят prometheus)
  ├── lkfl-server           ← pull ghcr.io/.../server:{tag}
  ├── lkfl-integration-proxy← pull ghcr.io/.../proxy:{tag}
  ├── lkfl-frontend         ← pull ghcr.io/.../frontend:{tag} (nginx:alpine + dist/)
  ├── lkfl-migrate          ← one-shot из образа server
  ├── lkfl-seed             ← one-shot из образа server
  ├── lkfl-postgres
  ├── lkfl-redis
  ├── lkfl-keycloak
  ├── lkfl-nginx            ← volume: default.conf (без frontend/dist)
  │                          ← upstream lkfl_frontend:80 для location /
  │                          ← location /nginx-health return 200
  │                          ← rate limiting zone
  └── lkfl-prometheus       ← порт 9090 (оставлен)

  Docker profiles:
    [default]    — все сервисы выше
    [monitoring] — grafana, loki, promtail (не на staging по умолчанию)

serverAI — Build Server (self-hosted runners)
  ├── lkfl-runners/runner-{1..7} — GitHub Actions runner instances
  ├── label: lkfl
  ├── Docker 29.4.3 + Buildx v0.33.0
  ├── Cache: Docker layers через GHCR (type=registry)
  └── Изоляция: отдельная директория, systemd services lkfl-runner-{1..7}
```

### Тегирование

| Событие | Тег | Пример |
|---------|-----|--------|
| Push в `main` | `main-{short-sha}` | `main-a1b2c3d` |
| PR | `pr-{number}-{short-sha}` | `pr-42-a1b2c3d` |
| Релиз | `v{version}` | `v0.1.0` |

### Deploy Worker API

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/deploy` | Деплой ветки: `{branch, sha, pr, imageTag}` |
| `POST` | `/deploy/pr` | PR preview на отдельный порт |
| `POST` | `/rollback` | Откат к предыдущему IMAGE_TAG |
| `GET` | `/status` | Текущий деплой, очередь, результат |
| `GET` | `/logs` | Логи последнего деплоя |

### Migration & Seed

Отдельные subcommand в `cmd/server/main.go`:
- `server migrate` — apply migrations через Atlas
- `server seed` — загружает seed данные (перенос из `cmd/seed/`)

Запускаются как one-shot контейнеры через `docker compose run --rm`.

---

## Что сделать

### Фаза 0: serverAI — self-hosted runners

#### 0.1 Инфраструктура serverAI

Проверено:
- Debian 13 (trixie), Docker 29.4.3 active, Buildx v0.33.0
- 30GB RAM, 16 CPU, 54GB free
- GitHub ✅, GHCR ✅
- 7 runner instances создано: `/home/ukituki/lkfl-runners/runner-{1..7}`
- Setup скрипт: `/home/ukituki/lkfl-runners/setup-runners.sh`

Что нужно:
1. Создать GitHub PAT (scope: `repo`) — `https://github.com/settings/tokens`
2. На serverAI: `./setup-runners.sh <PAT>`
3. Проверить: `systemctl status lkfl-runner-{1..7}`
4. Проверить на GitHub: `https://github.com/ukituki/LKFL/settings/actions`

#### 0.2 Изоляция

- Директория: `/home/ukituki/lkfl-runners/` (только LKFL)
- Label: `lkfl` (не пересекается с другими проектами)
- Systemd: `lkfl-runner-{1..7}.service` (NoNewPrivileges, ProtectSystem)
- User: `ukituki` (не root)

### Фаза 1: Dockerfile и CI

#### 1.1 Dockerfile — рефакторинг текущего

Текущий `Dockerfile` содержит все stage в одном файле. Разделить:

- **`Dockerfile.server`** — multi-stage build lkfl-server (golang → distroless)
- **`Dockerfile.proxy`** — multi-stage build lkfl-integration-proxy
- **`Dockerfile.frontend`** — Node build + nginx static (отдельный образ!)
- **`Dockerfile.deploy-worker`** — Go binary + git + bash

Текущий `Dockerfile` оставить для локальной разработки (`docker compose -f docker-compose.dev.yml build`).

#### 1.2 Фронтенд — отдельный образ

Сейчас фронтенд — volume mount `./frontend/dist/`. С GHCR-архитектурой нужен отдельный образ:

```dockerfile
FROM node:20-alpine AS build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY infra/nginx/frontend.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

nginx upstream в `default.conf` указывает на `lkfl-frontend:80`.

#### 1.3 cmd/server/main.go — CLI subcommands

```go
func main() {
    switch len(os.Args) > 1 {
    case os.Args[1] == "migrate":
        runMigrate()  // Atlas migrate apply
    case os.Args[1] == "seed":
        runSeed()     // перенос из cmd/seed/
    default:
        runServer()   // текущий код (HTTP сервер)
    }
}
```

#### 1.4 GitHub Actions — ЗАМЕНА ci.yml на build.yml

**КРИТИЧНО:** существующий `.github/workflows/ci.yml` нужно **удалить/заменить** на новый `build.yml`. Иначе будут два параллельных push в GHCR.

**`.github/workflows/build.yml`** — всегда запускается:
```yaml
name: Build
on:
  push:
    branches: [main, 'feature/*']
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  # Лёгкие шаги — public runner
  lint-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: {go-version: '1.24'}  # ← синхронизировать с Dockerfile!
      - run: cd backend && go vet ./...
      - run: cd backend && go test ./... -short
      - run: cd frontend && npm ci && npm run lint

  # Тяжёлые шаги — serverAI (self-hosted, label: lkfl)
  build-push:
    needs: lint-test
    runs-on: lkfl  # ← self-hosted runner на serverAI
    strategy:
      matrix:
        service: [server, proxy, frontend, deploy-worker]
    steps:
      - uses: actions/checkout@v4
      # КРИТИЧНО: login ДО build — для cache pull
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          file: Dockerfile.${{ matrix.service }}
          push: true
          tags: ghcr.io/ukituki-ps/lkfl/${{ matrix.service }}:${{ env.IMAGE_TAG }}
          # Cache через GHCR registry (требует login ДО build)
          cache-from: type=registry,ref=ghcr.io/ukituki-ps/lkfl/${{ matrix.service }}:buildcache
          cache-to: type=registry,ref=ghcr.io/ukituki-ps/lkfl/${{ matrix.service }}:buildcache,mode=max

  # Deploy — только на push в main
  deploy:
    needs: build-push
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    runs-on: ubuntu-latest
    steps:
      - name: Notify deploy worker
        run: |
          curl -X POST https://dev.april.ukituki.tech/deploy-webhook/deploy \
            -H "Authorization: Bearer ${{ secrets.DEPLOY_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{"branch":"${{ github.ref_name }}","sha":"${{ github.sha }}",
                 "imageTag":"main-${{ github.sha }}"}'
```

#### 1.5 GitHub secrets

| Secret | Назначение |
|--------|-----------|
| `GITHUB_TOKEN` | Автоматический, для GHCR push |
| `DEPLOY_TOKEN` | Авторизация webhook на deploy-worker |

### Фаза 2: Docker Compose — разделить dev и staging

#### 2.1 docker-compose.dev.yml

Локальная разработка:
```yaml
services:
  lkfl-server:
    build:
      context: .
      dockerfile: Dockerfile
      target: server
    volumes:
      - ./backend:/src  # hot reload
  lkfl-integration-proxy:
    build:
      context: .
      dockerfile: Dockerfile
      target: proxy
  nginx:
    volumes:
      - ./frontend/dist:/usr/share/nginx/html:ro  # локальный mount
```

#### 2.2 docker-compose.staging.yml

Staging (pull из GHCR, profiles):
```yaml
services:
  lkfl-server:
    image: ${GHCR_REGISTRY}/lkfl/server:${IMAGE_TAG}
    stop_grace_period: 35s  # ← graceful shutdown 30s в main.go
    environment:
      DB_DSN: ${DB_DSN}
      REDIS_URL: ${REDIS_URL}
      KEYCLOAK_ISSUER: ${KEYCLOAK_ISSUER}
      # ...

  lkfl-integration-proxy:
    image: ${GHCR_REGISTRY}/lkfl/proxy:${IMAGE_TAG}

  lkfl-frontend:
    image: ${GHCR_REGISTRY}/lkfl/frontend:${IMAGE_TAG}
    # Нет volume mount — dist/ внутри образа

  nginx:
    image: nginx:1.27-alpine
    volumes:
      - ./infra/nginx/server/default.conf:/etc/nginx/conf.d/default.conf:ro
      # УБРАТЬ ./frontend/dist — фронтенд теперь отдельный upstream
    # upstream: lkfl-frontend:80 вместо /usr/share/nginx/html

  lkfl-migrate:
    image: ${GHCR_REGISTRY}/lkfl/server:${IMAGE_TAG}
    command: ["/app/server", "migrate"]
    environment:
      DB_DSN: ${DB_DSN}
    networks: [lkfl_backend]

  lkfl-seed:
    image: ${GHCR_REGISTRY}/lkfl/server:${IMAGE_TAG}
    command: ["/app/server", "seed"]
    environment:
      DB_DSN: ${DB_DSN}
    networks: [lkfl_backend]

  lkfl-deploy-worker:
    image: ${GHCR_REGISTRY}/lkfl/deploy-worker:latest
    ports: ["9091:9091"]  # ← PORT 9091! (9090 занят prometheus)
    restart: always  # ← always, не unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./.env.staging:/.env.staging:ro
    environment:
      WEBHOOK_SECRET: ${WEBHOOK_SECRET}
      GHCR_TOKEN: ${GHCR_TOKEN}
      COMPOSE_FILE: docker-compose.staging.yml

  # Мониторинг — profile [monitoring], не на staging по умолчанию
  prometheus:
    profiles: [monitoring]
    ports: ["9090:9090"]  # ← 9090 OK (deploy-worker на 9091)
  grafana:
    profiles: [monitoring]
  loki:
    profiles: [monitoring]
  promtail:
    profiles: [monitoring]
```

#### 2.3 Испорченные конфиги — исправить

| Проблема | Фикс |
|----------|------|
| `KC_DB_URL: jdbc:postgresql://postgres:5432/keycloak` | Оставить `postgres` — compose DNS резолвит по имени service key, не container_name |
| Loki config несовместим | Убрать из staging (profile monitoring) + починить `infra/loki/loki.yml` |
| Hardcoded `X-Tenant-ID sdek` в nginx | Задокументировать как staging-only limitation (M23+ — динамическое разрешение) |
| Порт 8080 → nginx (дублирование) | 8080 маппить на nginx:80, не на lkfl-server |
| `.env.staging` в репе с паролями | `.env.staging` → `.gitignore`, `.env.staging.example` → шаблон |
| **Порт 9090 конфликт** (prometheus vs deploy-worker) | **deploy-worker → 9091** |
| `/nginx-health` endpoint отсутствует | Добавить `location /nginx-health { return 200 "ok"; }` |
| Фронтенд upstream отсутствует | Добавить `upstream lkfl_frontend { server lkfl-frontend:80; }` + изменить `location /` |
| Нет `stop_grace_period` | Добавить `stop_grace_period: 35s` для lkfl-server |
| `.gitignore` не покрывает `.env.staging` | Добавить `.env.staging`, `.env.production` |

#### 2.4 Nginx default.conf — исправления

```nginx
# ДОБАВИТЬ upstream для фронтенда (было: serve static из volume)
upstream lkfl_server {
    server lkfl-server:8080;
}

upstream lkfl_frontend {  # ← НОВЫЙ
    server lkfl-frontend:80;
}

upstream keycloak {
    server keycloak:8080;
}

# ДОБАВИТЬ rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;

server {
    listen 80;
    # listen 8080;  # ← УБРАТЬ (порт 8080 маппится извне на 80)
    server_name dev.april.ukituki.tech localhost;

    # ДОБАВИТЬ health endpoint (было: отсутствует → nginx forever unhealthy)
    location /nginx-health {
        return 200 "ok";
        add_header Content-Type text/plain;
    }

    # CORS preflight for API
    location /api {
        limit_req zone=api burst=20 nodelay;  # ← rate limiting
        # ... existing proxy config
    }

    # Admin API
    location /admin {
        limit_req zone=api burst=20 nodelay;  # ← rate limiting
        # ... existing proxy config
    }

    # Frontend (было: serve static, стало: proxy на lkfl-frontend)
    location = /index.html {
        proxy_pass http://lkfl_frontend;
        add_header Cache-Control "no-cache, no-store, must-revalidate" always;
    }

    # Hashed assets
    location ~ ^/assets/.*\.[a-zA-Z0-9]+\.(js|css|map)$ {
        proxy_pass http://lkfl_frontend;
        add_header Cache-Control "public, max-age=31536000, immutable" always;
    }

    # Other static files
    location / {
        proxy_pass http://lkfl_frontend;  # ← НОВЫЙ: proxy вместо try_files
    }

    # ... existing healthz, keycloak routes
}
```

### Фаза 3: Deploy Worker

#### 3.1 cmd/deploy-worker/

Go-сервис:
- HTTP API (`/deploy`, `/rollback`, `/status`, `/logs`)
- Порт: **9091** (не 9090 — конфликт с prometheus)
- Webhook валидация (secret header `Authorization: Bearer ${DEPLOY_TOKEN}`)
- Queue: однопоточная (serial deploy, mutex на деплой)
- Docker compose orchestration:
  1. `docker compose -f docker-compose.staging.yml --env-file .env.staging pull`
  2. `docker compose run --rm lkfl-migrate`
  3. `docker compose up -d` (без migrate/seed)
  4. `docker compose run --rm lkfl-seed`
  5. Healthcheck: `curl http://lkfl-server:8080/healthz`
  6. Сохранить предыдущий IMAGE_TAG для rollback
- Идемпотентность:
  - Миграции: Atlas idempotent по умолчанию
  - Seed: INSERT ON CONFLICT DO NOTHING
  - Pull: `docker compose pull --ignore-buildable`

#### 3.2 Dockerfile.deploy-worker

```dockerfile
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o /deploy-worker ./cmd/deploy-worker/

FROM alpine:3.20
RUN apk add --no-cache git bash curl
COPY --from=build /deploy-worker /usr/local/bin/deploy-worker
EXPOSE 9091
ENTRYPOINT ["deploy-worker"]
```

### Фаза 4: Переписать deploy.sh

#### 4.1 scripts/deploy.sh

Переписать полностью — выполняется внутри deploy-worker на сервере:
```bash
#!/bin/bash
# Выполняется внутри deploy-worker на serverDev
# Доступ: Docker socket, .env.staging, docker-compose.staging.yml

set -euo pipefail

COMPOSE="docker compose -f docker-compose.staging.yml --env-file .env.staging"

# 1. Pull образов из GHCR
$COMPOSE pull --ignore-buildable

# 2. Миграции (idempotent)
$COMPOSE run --rm lkfl-migrate || log_warn "migrations skipped (already applied)"

# 3. Запуск сервисов
$COMPOSE up -d

# 4. Seed данных (idempotent)
$COMPOSE run --rm lkfl-seed || log_warn "seed skipped (data exists)"

# 5. Healthcheck
wait_for_http http://lkfl-server:8080/healthz
```

#### 4.2 scripts/predeploy.sh

Адаптировать под новую архитектуру:
- Локальные проверки: Go build, npm build, `docker compose -f docker-compose.dev.yml config`
- SSH проверки → проверки deploy-worker API (`GET /status`)

#### 4.3 Makefile

Обновить deploy цели:
```makefile
deploy:        ## Деплой на staging (через deploy-worker API)
  curl -X POST http://serverDev:9091/deploy ...

deploy-health: ## Healthcheck через deploy-worker
  curl http://serverDev:9091/status

deploy-rollback: ## Роллбэк через deploy-worker
  curl -X POST http://serverDev:9091/rollback

dev-up:        ## Локальная разработка
  docker compose -f docker-compose.dev.yml up -d

dev-down:      ## Остановить локальные сервисы
  docker compose -f docker-compose.dev.yml down
```

### Фаза 5: Документация

#### 5.1 ADR-036

`doc/архитектура/adr/036-ci-cd-deploy-worker.md`:
- Status: Accepted
- Context: текущий deploy.sh через SSH ненадёжен, 35 проблем найдено при аудите
- Decision: serverAI (self-hosted runners) → GHCR → Deploy Worker на serverDev
- Consequences: сборка на мощном сервере с кэшем Docker layers, repeatable, любая ветка деплоится

#### 5.2 doc/деплой.md

Полный rewrite:
- Новая архитектура (serverAI → GHCR → Deploy Worker)
- serverAI: 7 self-hosted runners, label `lkfl`, изоляция
- Два compose файла (dev vs staging)
- Docker profiles (default, monitoring)
- API deploy-worker
- Тегирование образов
- Secrets management (.env.staging.example)
- Troubleshooting (обновлённый)
- Чек-лист деплоя

---

## Критерии приёмки

### serverAI — runners

- [ ] 7 runner instances запущены: `lkfl-runner-{1..7}`
- [ ] Label `lkfl` зарегистрирован в GitHub
- [ ] GitHub PAT создан (scope: repo)
- [ ] Runners видны в Settings → Actions → Runners
- [ ] Изоляция: отдельная директория, systemd services

### Dockerfile и CI

- [ ] `Dockerfile.server` — multi-stage, distroless
- [ ] `Dockerfile.proxy` — multi-stage, distroless
- [ ] `Dockerfile.frontend` — Node build + nginx (отдельный образ)
- [ ] `Dockerfile.deploy-worker` — Go + git + bash + curl, порт 9091
- [ ] `cmd/server/main.go` — subcommand `migrate` + `seed`
- [ ] `.github/workflows/build.yml` — lint/test (ubuntu-latest) + build (lkfl runner)
- [ ] `.github/workflows/ci.yml` — УДАЛЁН/ЗАМЕНЁН на build.yml
- [ ] `runs-on: lkfl` — build job использует serverAI
- [ ] `docker/login-action` ДО build (для cache pull)
- [ ] Docker layer cache через GHCR (buildcache tag)
- [ ] Go version: 1.24 в workflow И в Dockerfile (синхронизировано)
- [ ] Тегирование: `main-{sha}`, `pr-{number}-{sha}`
- [ ] `DEPLOY_TOKEN` в GitHub secrets

### Docker Compose

- [ ] `docker-compose.dev.yml` — build:, volume mount frontend/dist
- [ ] `docker-compose.staging.yml` — image: GHCR, без build:
- [ ] `lkfl-migrate` сервис (one-shot, profile default)
- [ ] `lkfl-seed` сервис (one-shot, profile default)
- [ ] `lkfl-deploy-worker` сервис (Docker socket mount, порт 9091, restart: always)
- [ ] Мониторинг в `profiles: [monitoring]` (grafana, loki, promtail)
- [ ] Loki конфиг исправлен или убран из default
- [ ] Frontend — отдельный upstream в nginx (`lkfl_frontend`)
- [ ] `location /` → proxy на `lkfl_frontend` (не static)
- [ ] `location /nginx-health` → return 200
- [ ] Rate limiting: `limit_req_zone` для /api и /admin
- [ ] Порт 8080 → nginx:80 (не lkfl-server)
- [ ] `KC_DB_URL` hostname консистентен (`postgres` — compose DNS OK)
- [ ] `stop_grace_period: 35s` для lkfl-server
- [ ] Порт 9090 → prometheus, 9091 → deploy-worker (нет конфликта)

### Deploy Worker

- [ ] `cmd/deploy-worker/` — HTTP API
- [ ] Порт 9091 (не 9090!)
- [ ] `POST /deploy` — pull → migrate → up → seed → healthcheck
- [ ] `POST /rollback` — откат к предыдущему IMAGE_TAG
- [ ] `GET /status` — текущий деплой, очередь, результат
- [ ] `GET /logs` — логи последнего деплоя
- [ ] Webhook валидация (secret header)
- [ ] Serial deploy (mutex)
- [ ] Идемпотентность (migrate, seed, pull)

### Scripts и Makefile

- [ ] `scripts/deploy.sh` — переписан (без SSH, через Docker)
- [ ] `scripts/predeploy.sh` — адаптирован (deploy-worker API)
- [ ] `Makefile` — цели: `dev-up`, `deploy`, `deploy-health`, `deploy-rollback`
- [ ] Makefile использует порт 9091 для deploy-worker

### Secrets и gitignore

- [ ] `.env.staging` → `.gitignore`
- [ ] `.env.production` → `.gitignore`
- [ ] `.env.staging.example` — шаблон с пустыми значениями
- [ ] `DEPLOY_TOKEN` в GitHub secrets

### Документация

- [ ] ADR-036 создан
- [ ] `doc/деплой.md` обновлён (serverAI runners, архитектура)
- [ ] `nginx default.conf` — X-Tenant-ID задокументирован как staging-only
- [ ] `nginx default.conf` — upstream lkfl_frontend добавлен
- [ ] `nginx default.conf` — rate limiting добавлен

### E2E

- [ ] Деплой на serverDev работает end-to-end: push → serverAI build → GHCR → deploy → healthz 200

---

## Идеи на будущее (не делать сейчас)

- Очередь джоб — Asynq/BullMQ (параллельные PR не деплоятся одновременно)
- Blue-green — два набора контейнеров + переключение nginx без даунтайма
- Telegram/Slack нотификации — результат деплоя в чат
- Health dashboard — `/status` → страница с метриками всех деплоев
- Auto-rollback — если healthcheck провалился через N минут → откат
- Автоматический бэкап БД — cron pg_dump + retention policy
- Multi-tenant nginx config — динамический X-Tenant-ID по hostname
- Несколько self-hosted runner'ов на serverAI для разных проектов (label isolation)
