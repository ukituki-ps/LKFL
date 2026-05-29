# ADR-036: CI/CD — serverAI self-hosted runners + Deploy Worker

| Поле     | Значение |
|----------|----------|
| Status   | Accepted |
| Date     | 2026-05-28 |
| Задача   | T2213 |
| Веха     | M22 (F1 Hardening) |
| Авторы   | architect-lkfl |

## Context

Текущий пайплайн деплоя (deploy.sh через SSH) ненадёжен. При аудите выявлено 35 проблем:

### Критические (🔴)

1. `ci.yml` конфликтует с `build.yml` — два параллельных push в GHCR
2. `KC_DB_URL` hostname `postgres` ≠ `lkfl-postgres` (источник проблем)
3. Фронтенд — volume mount `./frontend/dist/` вместо отдельного образа
4. Seed и миграции не работают (Go нет на сервере, distroless без shell)
5. `.gitignore` не покрывает `.env.staging` — пароли могут попасть в репозиторий
6. `/nginx-health` endpoint отсутствует — nginx forever unhealthy
7. Фронтенд upstream не добавлен в nginx — `location /` serve static, а не proxy на `lkfl-frontend`
8. Loki ломает compose (несовместимый конфиг)
9. Порт 9090 конфликт: prometheus и deploy-worker оба хотят 9090

### Серьёзные (🟡)

10. Go version mismatch: `ci.yml` использует 1.22, `Dockerfile` — 1.24
11. Один `docker-compose.yml` на всё (dev + staging) — `build:` вместо `image:`
12. Grafana/Promtail избыточны на staging
13. Hardcoded `X-Tenant-ID sdek` в nginx
14. Порт 8080 маппится на nginx (дублирование)
15. Нет `stop_grace_period` — graceful shutdown 30s в main.go требует 35s
16. Нет GHCR login step до build — cache pull не работает
17. `/callback` location serve static → в staging volume нет → Keycloak callback сломается
18. `docker compose pull` без `docker login GHCR` → 401 Unauthorized
19. Port mapping 8080:8080 → должно быть 8080:80

## Decision

### Архитектура

```
GitHub (push / PR)
  │
  │ .github/workflows/build.yml
  │   ├── lint + test          → runs-on: ubuntu-latest (public)
  │   ├── docker buildx        → runs-on: lkfl (serverAI, self-hosted)
  │   │   ├── docker login GHCR (до build — для cache pull!)
  │   │   ├── server   → ghcr.io/ukituki-ps/lkfl/server:{tag}
  │   │   ├── proxy    → ghcr.io/ukituki-ps/lkfl/proxy:{tag}
  │   │   ├── frontend → ghcr.io/ukituki-ps/lkfl/frontend:{tag}
  │   │   └── deploy-worker → ghcr.io/ukituki-ps/lkfl/deploy-worker:latest
  │   │
  │   .github/workflows/deploy.yml (только main)
  │     └── POST → https://dev.april.ukituki.tech/deploy-webhook/deploy
  │
  ▼
serverDev — Deploy Worker (docker-compose.staging.yml)
  ├── deploy-worker (:9091) ← webhook receiver
  ├── lkfl-server           ← pull ghcr.io/.../server:{tag}
  ├── lkfl-integration-proxy← pull ghcr.io/.../proxy:{tag}
  ├── lkfl-frontend         ← pull ghcr.io/.../frontend:{tag}
  ├── lkfl-migrate          ← one-shot из образа server
  ├── lkfl-seed             ← one-shot из образа server
  ├── lkfl-postgres, redis, keycloak, nginx
  └── prometheus (profile: monitoring)
```

### Ключевые решения

1. **serverAI** — 7 self-hosted GitHub Actions runner'ов (Debian 13, 30GB RAM, 16 CPU)
2. **4 Dockerfile** — server, proxy, frontend, deploy-worker (разделены из monolithic)
3. **2 compose файла** — dev (build:) и staging (image: из GHCR)
4. **Docker profiles** — monitoring не на staging по умолчанию
5. **Deploy Worker** — Go HTTP API (порт 9091), webhook validation, serial deploy
6. **CLI subcommands** — `server migrate` + `server seed` в одном бинарнике
7. **Nginx** — upstream lkfl_frontend, /nginx-health, rate limiting
8. **Secrets** — .env.staging в .gitignore, .env.staging.example как шаблон

### Тегирование

| Событие | Тег | Пример |
|---------|-----|--------|
| Push в `main` | `main-{short-sha}` | `main-a1b2c3d` |
| PR | `pr-{number}-{short-sha}` | `pr-42-a1b2c3d` |
| Релиз | `v{version}` | `v0.1.0` |

### Deploy Worker API

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/deploy` | Деплой: `{branch, sha, pr, imageTag}` |
| `POST` | `/deploy/pr` | PR preview на отдельный порт |
| `POST` | `/rollback` | Откат к предыдущему IMAGE_TAG |
| `GET` | `/status` | Текущий деплой, очередь, результат |
| `GET` | `/logs` | Логи последнего деплоя |

## Consequences

### Положительные

- Сборка на мощном сервере (serverAI) с кэшем Docker layers
- Repeatable: любая ветка деплоится через webhook
- Secrets вне репозитория (.env.staging в .gitignore)
- Мониторинг опционален (Docker profile)
- CI pipeline заменён (ci.yml → build.yml)
- Фронтенд — отдельный образ (не volume mount)
- Migrate/seed — one-shot контейнеры из образа server

### Отрицательные

- Зависимость от self-hosted runners (serverAI down = no builds)
- Deploy Worker требует Docker socket mount (безопасность)
- X-Tenant-ID sdek остаётся hardcoded (staging-only limitation, M23+)

## Operational decisions (2026-05-28)

Решения, принятые при деплое staging на serverAI.

### max-parallel: 2 в build-push matrix

| Поле | Значение |
|------|----------|
| Проблема | serverAI имеет 1 Docker daemon, 4 параллельных buildx перегружают его → buildx connection error |
| Решение | `max-parallel: 2` в `strategy.matrix` build.yml → не более 2 одновременных buildx |
| Файл | `.github/workflows/build.yml` |

```yaml
strategy:
  fail-fast: false
  max-parallel: 2  # serverAI: 1 Docker daemon, не перегружать buildx
  matrix:
    service: [server, proxy, frontend, deploy-worker]
```

### Перенос staging: serverDev (arm64) → serverAi (amd64)

| Поле | Значение |
|------|----------|
| Причина | serverDev (arm64) не мог использовать amd64-образы из GHCR, QEMU emulation недоступна |
| Решение | Staging переведён на serverAi (192.168.1.46, amd64, 30GB RAM), где уже работают CI runners |
| Внешний nginx | serverPr01 → upstream 192.168.1.46:8888 |

### Порт deploy-worker: 9091 → 9092

| Поле | Значение |
|------|----------|
| Причина | Port 9091 конфликтовал с legacy prometheus конфигурацией |
| Решение | Deploy Worker переведён на порт 9092 |
| Файлы | `Dockerfile.deploy-worker`, `config.go`, `docker-compose.staging.yml` |

### docker-compose plugin v2.38.0 в образе

| Поле | Значение |
|------|----------|
| Причина | Образ deploy-worker без docker-compose плагина → `docker compose: command not found` |
| Решение | `Dockerfile.deploy-worker`: apk docker-cli + compose plugin v2.38.0 + верификация `docker compose version` при сборке |

### GHCR login через Docker daemon host

| Поле | Значение |
|------|----------|
| Причина | Deploy-worker не может залогиниться в GHCR из контейнера (Docker daemon на хосте) |
| Решение | `gh auth login` на serverAI → Docker daemon auth → deploy-worker pull без явного login |

## Зависимости

- T2209 (Docker Production) — multi-stage Dockerfile
- T2208 (CI Pipeline) — CI workflow
- T2210 (Деплой на стенд) — staging docker-compose
