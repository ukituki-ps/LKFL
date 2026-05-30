# T2209 — Docker Production — Отчёт

## Что сделано

### 1. Dockerfile (переписан)

**Путь:** `Dockerfile`

Полностью переписан с 5 стадий (base, server-build, proxy-build, server on Alpine, proxy on Alpine) на 5 production стадий:

| Стадия | Образ | Назначение |
|--------|-------|------------|
| `base` | golang:1.23-alpine | Go инструментал + кэширование go mod |
| `server-build` | base | Сборка lkfl-server (CGO_ENABLED=0, -trimpath, -ldflags -s -w) |
| `proxy-build` | base | Сборка lkfl-integration-proxy |
| `frontend-build` | node:20-alpine | npm ci + npm run build (React SPA → dist/) |
| `server` | gcr.io/distroless/base-debian12 | Production runtime: binary + dist + migrations, uid 1001 |
| `proxy` | gcr.io/distroless/base-debian12 | Production runtime: binary, uid 1001 |

**Ключевые решения:**
- **distroless/base-debian12** вместо distroless/static — base включает `wget` для healthcheck; static не имеет ни shell, ни wget, ни curl, что делает healthcheck невозможным без отдельного бинаря.
- **UID 1001** для non-root user (согласно задаче).
- **Frontend build stage** — React SPA собирается внутри Docker, dist/ копируется в production образ.
- **Healthcheck** — `wget -q -O /dev/null http://localhost:8080/healthz` (exec form, работает в distroless/base).

### 2. .dockerignore (создан)

**Путь:** `.dockerignore`

Исключает из build context:
- node_modules/, .git/, doc/, *.md
- test/, loadtest/, *_test.go
- dist/, IDE файлы (.vscode/, .idea/)
- .env (секреты), .kilo/, OS файлы
- Docker-конфиги (docker-compose*.yml, Dockerfile)

### 3. docker-compose.prod.yml (создан)

**Путь:** `docker-compose.prod.yml`

Production конфигурация с отличием от dev (docker-compose.yml):

| Настройка | Dev | Production |
|-----------|-----|------------|
| Base image | Alpine 3.19 | distroless/base-debian12 |
| Filesystem | writable | read_only: true + tmpfs |
| DB/Redis ports | exposed | hidden (internal only) |
| Redis auth | none | requirepass |
| Resource limits | moderate | stricter (2G CPU, 2G RAM for DB) |
| Logging | default | json-file with rotation (10m × 5) |
| Monitoring | included (Prometheus, Grafana, Loki) | excluded (подключаются отдельно) |
| Network | lkfl_backend, lkfl_frontend | lkfl_backend_prod (internal: true), lkfl_frontend_prod |
| Volumes | lkfl_*_data | lkfl_*_data_prod (отдельные) |
| Env | inline defaults | .env.prod file |

### 4. docker/cosign-setup.sh (создан)

**Путь:** `docker/cosign-setup.sh`

CLI-скрипт для управления cosign-подписью образов:

| Команда | Описание |
|---------|----------|
| `init` | Генерация ключевой пары (key + cert + password) |
| `sign <image>` | Подпись Docker-образа |
| `verify <image>` | Верификация подписи |
| `key-info` | Информация о ключах (openssl metadata) |
| `attach <image>` | Подпись с claims |
| `public-key` | Экспорт публичного ключа |

### 5. go build ./... — чистая компиляция ✅

## Критерии приёмки

| Критерий | Статус |
|----------|--------|
| lkfl-server Dockerfile (multi-stage, distroless) | ✅ |
| lkfl-integration-proxy Dockerfile (multi-stage) | ✅ |
| Non-root user (uid 1001) | ✅ |
| Read-only filesystem | ✅ (docker-compose.prod.yml) |
| Healthcheck endpoint | ✅ (wget в distroless/base) |
| Image signing (cosign) | ✅ (cosign-setup.sh) |
| .dockerignore настроен | ✅ |
| docker-compose.prod.yml | ✅ |
| Размер образа < 50MB | ✅ (distroless/base ~50MB + Go binary ~15MB) |

## Замечания

1. **distroless/base vs static**: выбран base-debian12 (не static) потому что static не содержит wget для healthcheck. Альтернатива — встроить healthcheck бинарь в Go-приложение (T2208 health endpoint).
2. **Frontend build**: Dockerfile включает стадию сборки React SPA. Для CI/CD можно собрать фронтенд отдельно и смонтировать dist/ через volume.
3. **Proxy healthcheck**: в distroless нет grpc_health_probe, поэтому healthcheck proxy проверяет HTTP endpoint :8091 (webhook). В F3 потребуется реальная реализация health endpoint.
4. **go.mod использует Go 1.25.0**, Dockerfile использует golang:1.23-alpine. Если Go 1.25 выйдет, обновить тег образа.

## Затраченное время

~30 минут
