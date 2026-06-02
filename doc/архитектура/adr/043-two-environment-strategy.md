# ADR-043: Two-Environment Strategy — Dev (manual) + Staging (CI/CD)

| Статус | Date | Автор |
|--------|------|-------|
| Accepted | 2026-06-02 | devops-lkfl |

## Контекст

Проект использует один сервер (serverAi: 192.168.1.27) с двумя стендами.
Ранее существовал только staging (`dev.april.ukituki.tech`) с автоматическим деплоем
через CI/CD (GitHub Actions). Разработчикам негде было тестировать feature-ветки
без влияния на staging.

## Решение

Два стенда на одном сервере (serverAi), два домена:

| Стенд | URL | Деплой | БД | Volume prefix | Network prefix |
|-------|-----|--------|-----|---------------|----------------|
| **Dev** | `https://project.ukituki.tech` | Ручной (`dev-deploy.sh`) | `lkfl_platform_dev` | `dev_` | `lkfl_*_dev` |
| **Staging** | `https://dev.april.ukituki.tech` | Auto (push → main) | `lkfl_platform` | `staging_` | `lkfl_backend_staging` |

### Сетевая архитектура

```
Client → serverPr01 (443 ssl, Let's Encrypt *.ukituki.tech)
  ├── project.ukituki.tech → serverAi:18002 → Docker dev services
  └── dev.april.ukituki.tech → serverAi:18000 → Docker staging services
```

### Порт-маппинг на serverAi

| Сервис | Dev (port) | Staging (port) |
|--------|-----------|----------------|
| Backend (lkfl-server) | 18082 | 18080 |
| Frontend | 8088 | 8086 |
| Keycloak | 19083 | 19081 |
| Integration Proxy gRPC | 18094 | 18090 |
| Integration Proxy HTTP | 18095 | 18091 |
| Postgres | 15434 | 15432 |
| Redis | 16380 | 16379 |
| Nginx internal | 18002 | 18000 |
| Prometheus (monitoring) | 19092 | 19090 |
| Grafana (monitoring) | 13002 | 13000 |
| Loki (monitoring) | 13102 | 13100 |

### Docker Compose файлы

| Файл | Назначение |
|------|-----------|
| `docker-compose.dev.yml` | Локальная разработка (ноутбук, без HTTPS) |
| `docker-compose.dev-server.yml` | Dev стенд на serverAi (ручной деплой) |
| `docker-compose.staging.yml` | Staging стенд на serverAi (CI/CD) |
| `docker-compose.prod.yml` | Production (рассматривается) |

### Деплой

**Dev (ручной):**
```bash
# На serverAi
infra/deploy/dev-deploy.sh main-a1b2c3d
infra/deploy/dev-deploy.sh feature-x-login-abc1234
```

Скрипт делает: pull → migration → seed → down → up → healthcheck.

**Staging (автоматический):**
Push на `main` → GitHub Actions → build → deploy-staging → smoke-test → e2e.

### Базы данных

| Стенд | Schema | Keycloak DB |
|-------|--------|-------------|
| Dev | `lkfl_platform_dev` | `keycloak_dev` |
| Staging | `lkfl_platform` | `keycloak` |

Dev-стенд можно полностью сбросить: `docker compose -p lkfl-dev down -v && up -d`.

### Keycloak

Оба стенда используют одинаковый realm config (`infra/keycloak/realm-lkfl-sdek.json`),
но разные базы данных и разные `KC_HOSTNAME`:
- Dev: `KC_HOSTNAME=project.ukituki.tech`
- Staging: `KC_HOSTNAME=dev.april.ukituki.tech`

### Nginx

Два уровня:
1. **serverPr01** — TLS termination (Let's Encrypt), routing по server_name
2. **serverAi** — routing на Docker-сервисы по портам

Конфиги:
- `infra/nginx/serverPr01-dev.conf` — serverPr01 vhost для project.ukituki.tech
- `infra/nginx/serverAi-dev.conf` — serverAi internal nginx port 18002
- `infra/nginx/serverAi.conf` — serverAi internal nginx port 18000 (staging)

### Ресурсы

serverAi: 30GB RAM, 16 CPU, 221GB disk.

Оценка потребления двух стендов:

| Сервис | RAM (dev) | RAM (staging) | Итого |
|--------|-----------|---------------|-------|
| PostgreSQL | ~1G | ~1G | ~2G |
| Keycloak | ~1G | ~1G | ~2G |
| Redis | ~64M | ~64M | ~128M |
| lkfl-server | ~256M | ~256M | ~512M |
| lkfl-proxy | ~128M | ~128M | ~256M |
| Frontend nginx | ~32M | ~32M | ~64M |
| Monitoring (dev) | ~500M | ~500M | ~1G |
| **Итого (без monitoring)** | ~2.5G | ~2.5G | ~5G |
| **Итого (с monitoring)** | ~3G | ~3G | ~6G |

Всё влезает в 30GB.

## Последствия

### Положительные
- Разработчики могут тестировать feature-ветки не ломая staging
- Dev легко сбрасывать (volume `dev_pg_data`)
- CI/CD pipeline staging чист — не ломается экспериментальными коммитами
- Оба стенда делят одну PostgreSQL instance, один Redis instance экономит ресурсы

### Отрицательные
- Два nginx (staging на 18000, dev на 18002) — нужно поддерживать оба конфига
- Ручной деплой dev — можно забыть обновить IMAGE_TAG
- Если dev и staging используют один realm JSON, нужно следить за redirect_uri
- Дублирование ресурсов (Keycloak × 2, PostgreSQL × 2 volumes)

### Mitigations
- Realm JSON обновлять централизованно (один файл, используется обоими стендами)
- `dev-deploy.sh` idempotent, можно запускать сколько угодно
- Dev без deploy-worker (нет auto-deploy, только ручной)
- Monitoring опционален через profile

## Связанные ADR

- [ADR-037](./037-keycloak-reverse-proxy.md) — Keycloak reverse proxy через nginx
- [ADR-039](./039-deploy-rollback-strategy.md) — Rollback strategy
- [ADR-040](./040-backup-disaster-recovery.md) — Backup & DR
- [ADR-042](./042-zero-downtime-deployment.md) — Zero-downtime deployment

## Чеклист установки

- [ ] Создать `/home/ukituki/LKFL-dev/` на serverAi
- [ ] Копировать `docker-compose.dev-server.yml`, `.env.dev-server`, `infra/` в dev dir
- [ ] Установить nginx на serverAi: `serverAi-dev.conf` → port 18002
- [ ] Установить nginx на serverPr01: `serverPr01-dev.conf` → vhost project.ukituki.tech
- [ ] Проверить Let's Encrypt cert для `project.ukituki.tech` (wildcard `*.ukituki.tech` покрывает)
- [ ] Запустить первый деплой: `dev-deploy.sh main-latest`
- [ ] Проверить health: `curl https://project.ukituki.tech/healthz`
