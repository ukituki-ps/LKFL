# ADR-011: Monorepo (M12: один go.mod)

**Статус:** ⚠️ Note: M12 — single go.mod
**Дата:** 2026-05-22, обновлено 2026-05-25 (M12)
**Контекст:** М01-создание-описания

## Обновление M12

> **M12:** Четыре go.mod (platform, billing, integrations, payment-gateway) → один `go.mod` в корне backend/ ([ADR-024](./024-modular-monolith.md)). `shared/` остаётся — common пакеты (auth, celcontext). Один `go build ./...`, один Dockerfile.

## Контекст (исторический)

Три Go-сервиса (platform, billing, integrations) + React frontend. Нужно решить: monorepo или multi-repo.

## Решение

**Monorepo** с `shared` Go packages (`shared/pkg/*`). Один Git repo, один CI pipeline. **M10 T1002:** `llm-proxy` слит в platform. **M10 T1003+T1005:** `shared/` добавlen. **M12:** один `go.mod`.

```
lkfl/
├── backend/          (один go.mod, один Dockerfile)
│   ├── cmd/server/   (lkfl-server)
│   ├── cmd/worker/   (asynq worker)
│   ├── internal/     (13 пакетов)
│   └── shared/pkg/   (auth, celcontext)
├── frontend/          (package.json)
├── docker-compose.yml
└── openapi/
```

> **M12:** `billing/`, `integrations/`, `payment-gateway/` → удалены. Всё в `backend/`.

## Аргументы «за monorepo»

- Общие типы (User, Engagement, Transaction) — без vendor между репозиториями
- Один PR меняет API контракт + consumer одновременно
- Единая CI pipeline
- Одна команда — не нужна git submodule синхронизация
- **M10:** shared packages (`shared/pkg/*`) — zero-copy тип safety между платформой и billing

## Аргументы «против»

- Repo растёт → медленнее `git clone`
- Нет independent deploy без additional CI config

## Вердикт

**Monorepo.** LKFL — одна команда, один проект. April Profile/Worker — multi-repo, потому что разные команды и разные продукты. Для LKFL monorepo — правильный выбор.

## Следствия

- **M12:** `go build ./...` — один билд, один результат
- **M12:** Один Docker build, один Docker image
- OpenAPI specs в корневом `openapi/` (accessible из всех сервисов)
- CI: один workflow, один build
