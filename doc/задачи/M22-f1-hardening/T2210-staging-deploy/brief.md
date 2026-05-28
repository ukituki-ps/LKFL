# T2210 — Деплой на стенд

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Деплой F1 на staging стенд.
Полный стек: backend, frontend, PostgreSQL, Redis, Keycloak, Nginx.

## Что сделать

### Staging docker-compose

`docker-compose.staging.yml`:

- **lkfl-server** — backend (:8080)
- **lkfl-integration-proxy** — proxy (:8090 gRPC, :8091 HTTP)
- **frontend** — nginx serving React SPA (:80)
- **postgres** — PostgreSQL 17 (:5432)
- **redis** — Redis 7 (:6379)
- **keycloak** — Keycloak 25 (:8081)
- **nginx** — reverse proxy + TLS (:443, :80)
- **prometheus** — metrics (:9090)
- **grafana** — dashboards (:3000)
- **loki** — logs (:3100)
- **promtail** — log collector

### Nginx config

- Reverse proxy: `/api/v1/` → lkfl-server, `/` → frontend
- TLS self-signed certificate для staging
- HTTP → HTTPS redirect
- Security headers: HSTS, X-Frame-Options, CSP
- Static files caching

### Environment variables

`.env.staging` — все переменные окружения:
- DB connection string
- Redis connection string
- Keycloak URL, realm, client ID/secret
- JWT signing key
- Tenant default slug
- Log level
- Feature flags

### Healthcheck verification

- Backend: `curl http://staging/healthz` → 200
- Frontend: `curl http://staging/` → 200, HTML loaded
- DB: `pg_isready -h postgres` → accepting connections
- Redis: `redis-cli ping` → PONG
- Keycloak: `curl http://staging:8081/realms/master` → 200

### Seed data

- `make seed` → sdek tenant + brand config + categories + types + offers

## Критерии приёмки

- [ ] docker-compose.staging.yml
- [ ] Nginx config (reverse proxy + TLS)
- [ ] .env.staging
- [ ] Все сервисы запущены и healthy
- [ ] Backend healthz → 200
- [ ] Frontend загружается
- [ ] Seed data загружена
- [ ] Login через Keycloak работает
- [ ] Каталог доступен
