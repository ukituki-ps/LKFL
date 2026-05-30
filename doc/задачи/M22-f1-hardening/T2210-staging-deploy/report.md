# T2210 — Отчёт

## Что сделано

### docker-compose.staging.yml
Создан staging-профиль, объединяющий все сервисы F1:
- **11 сервисов**: postgres, redis, keycloak, lkfl-server, lkfl-integration-proxy, frontend, nginx, prometheus, grafana, loki, promtail
- Контейнеры с префиксом `staging-`
- Volumes с префиксом `staging_`
- Две сети: `lkfl_backend_staging` и `lkfl_frontend_staging`
- Keycloak в `start-dev` режиме для быстрого запуска
- Порты DB/Redis открыты для отладки (в отличие от prod)
- Logging driver: json-file с ротацией (10m × 5 файлов)
- env_file: `.env.staging`

### infra/nginx/staging.conf
Nginx reverse proxy конфигурация для staging:
- HTTP → HTTPS redirect (порт 80 → 443)
- Self-signed TLS (TLSv1.2 + TLSv1.3)
- Security headers: HSTS, X-Frame-Options, X-Content-Type-Options, XSS-Protection, CSP
- Upstream: `/api/v1/` → lkfl-server, `/admin/` → lkfl-server, `/webhooks/` → integration-proxy
- Health check: `/healthz` → lkfl-server
- Metrics: `/metrics` → lkfl-server (для Prometheus)
- Frontend SPA: proxy к контейнеру frontend с try_files fallback
- Rate limiting: 10 req/s per IP
- Static asset caching (30d, immutable)

### infra/nginx/frontend.conf
Конфигурация для контейнера frontend (nginx serving React SPA):
- Static file serving из `/usr/share/nginx/html`
- SPA fallback: `try_files $uri $uri/ /index.html`
- Gzip compression
- Asset caching (30d, immutable)

### .env.staging
Переменные окружения для staging:
- DB, Redis, Keycloak (localhost)
- JWT secret, Sentry (пустой)
- Grafana credentials
- Feature flags (FEATURE_CATALOG=true)

### scripts/healthcheck.sh
Скрипт проверки здоровья всех сервисов:
- Backend: healthz endpoint
- Frontend: HTTP 80 + HTTPS 443
- Nginx: health endpoint
- PostgreSQL: pg_isready + query
- Redis: ping
- Keycloak: admin console + realms
- Integration Proxy: healthz HTTP 8091
- Monitoring: Prometheus, Grafana, Loki
- Цветная консольная выдача с итоговой статистикой

## Созданные файлы

| Файл | Описание |
|------|----------|
| `docker-compose.staging.yml` | Staging compose (11 сервисов) |
| `infra/nginx/staging.conf` | Nginx reverse proxy + TLS |
| `infra/nginx/frontend.conf` | Frontend SPA serving |
| `.env.staging` | Переменные окружения staging |
| `scripts/healthcheck.sh` | Healthcheck скрипт |

## Компиляция

`go build ./...` — чистая компиляция ✅

## Замечания

- Staging использует `start-dev` режим Keycloak (не `start` как в prod) — быстрее запуск, проще отладка
- Self-signed TLS сертификаты требуют генерации: `openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout infra/nginx/ssl/server.key -out infra/nginx/ssl/server.crt -subj "/CN=staging.lkfl.local"`
- Frontend dist должен быть собран заранее (`npm run build` в frontend/)
- Seed data загружается отдельно через `make seed` (не автоматизирован в compose)
