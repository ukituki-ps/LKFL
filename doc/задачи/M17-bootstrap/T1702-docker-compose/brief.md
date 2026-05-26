# T1702 — Docker Compose (production-ready)

## Веха

M17-bootstrap

## Тип

code

## Контекст

`docker-compose.yml` — единственная точка запуска development и staging окружения.
Описан в `doc/архитектура/стек.md` строка 80 (docker-compose v2, 6 контейнеров).

## Что сделать

Создать `docker-compose.yml` с 6 контейнерами:

### 1. PostgreSQL 17

- Image: `postgres:17-alpine`
- Port: `5432:5432`
- Volume: `lkfl_pg_data:/var/lib/postgresql/data` (persistent)
- Env: `POSTGRES_DB=lkfl_platform`, `POSTGRES_USER=lkfl`, `POSTGRES_PASSWORD` (из `.env`)
- Healthcheck: `pg_isready -U lkfl -d lkfl_platform`
- Limits: `deploy.resources.limits: {memory: 1G, cpus: '1'}`

### 2. Redis 7

- Image: `redis:7-alpine`
- Port: `6379:6379`
- Volume: `lkfl_redis_data:/data` (persistent)
- Command: `redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru`
- Healthcheck: `redis-cli ping`
- Limits: `deploy.resources.limits: {memory: 512M, cpus: '0.5'}`

### 3. Keycloak 25.0

- Image: `quay.io/keycloak/keycloak:25.0`
- Ports: `8081:8080` (HTTP для dev, TLS через Nginx в production)
- Env: `KEYCLOAK_ADMIN=admin`, `KEYCLOAK_ADMIN_PASSWORD` (из `.env`), `KC_DB=postgres`, `KC_DB_URL=jdbc:postgresql://postgres:5432/keycloak`
- Command: `start-dev` (dev) / `start` (prod)
- Volume: `lkfl_keycloak_data:/opt/keycloak/data`
- Healthcheck: HTTP GET `/admin/master/console/` на порту 8080

### 4. Nginx

- Image: `nginx:1.27-alpine`
- Ports: `80:80`, `443:443`
- Volume: `./infra/nginx/default.conf:/etc/nginx/conf.d/default.conf:ro`
- Depends on: `lkfl-server` (condition: service_healthy)
- Config: reverse proxy `/api/v1/*` → `lkfl-server:8080`, SPA serve `/` → `frontend dist`

### 5. lkfl-server

- Build: `.` (multi-stage Dockerfile)
- Ports: `8080:8080` (internal, через Nginx)
- Env: из `.env` (DB_DSN, REDIS_URL, KEYCLOAK_ISSUER, KEYCLOAK_CLIENT_ID, KEYCLOAK_CLIENT_SECRET)
- Healthcheck: `curl -f http://localhost:8080/healthz`
- Depends on: `postgres` (condition: service_healthy), `redis` (condition: service_healthy)
- Limits: `deploy.resources.limits: {memory: 512M, cpus: '1'}`

### 6. lkfl-integration-proxy (stub для F1, реальный в F3)

- Build: `.` (target: proxy)
- Ports: `8090:8090` (gRPC), `8091:8091` (webhooks)
- Healthcheck: gRPC health check на `:8090`
- Depends on: `redis` (condition: service_healthy)

### Nginx config (`infra/nginx/default.conf`)

```nginx
upstream lkfl_server {
    server lkfl-server:8080;
}

upstream lkfl_proxy {
    server lkfl-integration-proxy:8091;
}

server {
    listen 80;
    server_name _;

    # API
    location /api/v1/ {
        proxy_pass http://lkfl_server/api/v1/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 30s;
    }

    # Admin API
    location /admin/ {
        proxy_pass http://lkfl_server/admin/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Webhooks (proxy)
    location /webhooks/ {
        proxy_pass http://lkfl_proxy/webhooks/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # SPA
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }
}
```

### `.env.example`

```
# Database
DB_DSN=postgresql://lkfl:lkfl-dev@postgres:5432/lkfl_platform?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379

# Keycloak
KEYCLOAK_ISSUER=http://localhost:8081/realms/lkfl-sdek
KEYCLOAK_CLIENT_ID=lkfl-spa
KEYCLOAK_CLIENT_SECRET=changeme-dev

# Server
SERVER_PORT=8080
LOG_LEVEL=info

# Sentry
SENTRY_DSN=

# JWT
JWT_SECRET=changeme-dev
```

## Требования

- Все контейнеры с healthcheck
- Все volumes persistent (не anonymous)
- `.env` не коммитится (`.gitignore`)
- `.env.example` коммитится (без реальных секретов)
- Resource limits на всех контейнерах
- Nginx config — production-ready (proxy headers, timeouts)

## Критерии приёмки

- [ ] `docker-compose.yml` создан с 6 контейнерами
- [ ] `docker compose up -d` поднимает все контейнеры
- [ ] Все healthcheck проходят (green)
- [ ] `docker compose exec postgres pg_isready` — OK
- [ ] `docker compose exec redis redis-cli ping` — PONG
- [ ] Keycloak доступен на `http://localhost:8081`
- [ ] Nginx reverse proxy работает (502 на server — OK, proxy подключён)
- [ ] `.env.example` создан, `.env` в `.gitignore`
- [ ] Resource limits заданы на всех контейнерах
