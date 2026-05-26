# ADR-010: Nginx как API Gateway и Reverse Proxy

**Статус:** Accepted
**Дата:** 2026-05-22, обновлено 2026-05-25 (M12)
**Контекст:** М01-создание-описания

## Обновление M12

> **M12:** Один upstream — `lkfl-server:8080`. `/billing/` → `/api/v1/billing/`, `/payments/` → `/api/v1/payments/`, `/webhooks/` → lkfl-server. Один health endpoint `/api/healthz`.

## Контекст

С 3+ сервисами (platform, billing, integrations) + frontend SPA нужен единый entry point. April Profile/Worker используют Nginx.

## Решение

**Nginx 1.27-alpine** как единый reverse proxy:

```nginx
# M12: один upstream для lkfl-server
upstream lkfl-server  { server lkfl-server:8080; }
upstream frontend    { server frontend:3000; }

server {
    listen 80;

    location /               { proxy_pass http://frontend; }
    location /api/           { proxy_pass http://lkfl-server; }
    location /webhooks/      { proxy_pass http://lkfl-server; }
    location /healthz        { proxy_pass http://lkfl-server; }
}
```

**Security headers:** CSP, HSTS, X-Frame-Options, X-Content-Type-Options.

## Альтернативы

| Вариант | Плюсы | Минусы |
|-----|--|-|
| Traefik | Auto-discovery, Docker native | Сложнее config, меньше знаком команды |
| Caddy | Auto HTTPS | Меньше mature для multi-upstream |
| API Gateway (Kong, Apigee) | Rate limiting, analytics | Overkill для 1 сервиса, сложная infra |

## Вердикт

**Nginx.** Знаком команде (April Profile/Worker), простой config, достаточно для lkfl-server.

## Следствия

- `infra/nginx/` — config для dev и prod
- Security headers из ADR «Безопасность»
- **M12:** Один upstream, все routes через lkfl-server:8080
- Production: Nginx + certbot для HTTPS (Caddy убран из стека)