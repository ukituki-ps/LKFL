# ADR-038: Переезд staging с serverDev на serverAI

| Поле     | Значение |
|----------|----------|
| Status   | Accepted |
| Date     | 2026-05-29 |
| Веха     | M22 (F1 Hardening) |
| Авторы   | architect-lkfl |

## Context

serverDev (arm64, Orange Pi 4 Pro) использовался как staging-хост для Docker-контейнеров LKFL.
serverAI (amd64, AMD Ryzen 7 7700X) использовался как build-сервер с self-hosted GitHub Actions runners.

Проблемы serverDev:
- **arm64** — не мог использовать amd64-образы из GHCR напрямую (QEMU emulation недоступна)
- **Медленный** — Docker compose operations занимали больше времени
- **Два сервера** — усложнял поддержку: build на одном, staging на другом

Решение: переехать staging на serverAI → один сервер для build + staging.

## Decision

**serverAI — единый сервер для build + staging.**

### serverAI (ai)

| Параметр | Значение |
|----------|----------|
| CPU | AMD Ryzen 7 7700X 8-Core |
| RAM | 30 GB |
| Disk | 221 GB (184 GB used, 27 GB free) |
| Архитектура | amd64 (x86_64) |
| SSH alias | `serverAi` |

**Роли:**
- GitHub Actions self-hosted runners (label: `lkfl`)
- Staging Docker-контейнеры (LKFL, Keycloak, Postgres, Redis, nginx, deploy-worker)
- Docker compose: `/home/ukituki/LKFL-staging/docker-compose.staging.yml`

### serverDev (orangepi4pro) — ⛔ отключён

| Параметр | Значение |
|----------|----------|
| CPU | Orange Pi 4 Pro (arm64) |
| Статус | Staging-контейнеры остановлены |

Staging-контейнеры на serverDev можно оставить как «холодный бэкап» или полностью отключить.

## Consequences

### Позитивные
- **Один сервер** — build + staging на serverAI
- **amd64** — нативные образы из GHCR, без эмуляции
- **Быстрее** — Docker compose operations быстрее на x86_64
- **Меньше инфраструктуры** — не нужно поддерживать два сервера

### Негативные
- **Ресурсы** — staging контейнеры + runners + build конкурируют за CPU/RAM на одном сервере
- **Одна точка отказа** — если serverAI упал, нет staging-резерва

## Изменения

| Файл | Изменение |
|------|-----------|
| `doc/архитектура/adr/036-ci-cd-deploy-worker.md` | serverDev → serverAI |
| `doc/деплой.md` | serverDev → serverAI, обновлена архитектура |
| `scripts/deploy.sh` | serverDev → serverAI |
| `scripts/predeploy.sh` | serverDev → serverAI |
| `doc/план/вехи.md` | serverDev → serverAI |
| Все T2213/T2215 doc | serverDev → serverAI |
