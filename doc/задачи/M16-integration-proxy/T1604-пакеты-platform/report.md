# T1604 — Отчёт: `архитектура/пакеты-platform.md` update

## Веха

M16-integration-proxy

## Выполнено

### Удалено
- Секция `### Пакет internal/integrations/` — полностью удалена из монолита (была ~70 строк)

### Добавлено

1. **M16 секция в истории** — после M12, описывает переход: `integrations/` → `integrationclient/` + `integration-proxy/`. Включает tree-структуру нового layout с двумя бинарниками.

2. **`### Пакет internal/integrationclient/`** — новый пакет монолита:
   - `IntegrationClient` struct — typed gRPC client к `lkfl-integration-proxy`
   - `IntegrationService` interface — для mock в тестах (9 методов: Activate, Deactivate, GetProviderStatus, GetCatalog, HealthCheck, ListProviders, GetProvider, UpdateProvider, TriggerSync, GetSyncLogs)
   - Зависимости: `google.golang.org/grpc`, `proto/integration/v1/`

3. **`### integration-proxy/`** — структура proxy бинарника:
   - `adapters/` — 11 провайдеров (9 YAML + 2 hard-coded)
   - `circuitbreaker/` — circuit breaker per provider
   - `webhook/` — webhook receiver + verifier
   - `grpc/` — gRPC server + generated code
   - `config/` — YAML config загрузка
   - `cmd/integration-proxy/main.go` — entry point
   - Circuit breaker параметры: threshold 10/60s, recovery 30s

4. **Фаза 4 миграционного плана** — M16 Integration Proxy с чеклистом

### Обновлено

1. **TL;DR** — 17→16 пакетов, ссылка `integrations/` → `integrationclient/`, добавлен M16 note
2. **Содержание** — обновлены строковые ссылки, добавлен `integration-proxy/`
3. **PublicHandlerDeps** — `Integrations *integrations.ProviderGateway` → `IntegrationClient integrationclient.IntegrationService`
4. **DI граф** — ASCII диаграмма обновлена: показывает Nginx → lkfl-server (:8080) + lkfl-integration-proxy (:8090 gRPC, :8091 HTTP webhook), gRPC client связь
5. **Зависимости между пакетами:**
   - `engagement/flow/`: `integrations/` (direct call) → `integrationclient/` (gRPC → proxy)
   - `integrations/` удалён → `integrationclient/` добавлен (grpc + proto)
   - `api/` (Public): `integrations` → `integrationclient`
6. **Asynq workers** — добавлена заметка M16: catalog-sync вынесен в proxy worker pool
7. **Что НЕ меняется** — добавлен столбец M16, обновлены строки: бинарников (3), PostgreSQL schema (2), Nginx routes, go.mod, Integrations сервис
8. **Почему пакеты, а не gRPC** — обновлено обоснование: integrations вынесен как proxy (I/O boundary exception), добавлена M16 note
9. **Сводная таблица** — разбита на две части:
   - Монолит: 16 пакетов (добавлен `integrationclient/`, удалён `integrations/`)
   - Proxy: 5 пакетов (adapters, circuitbreaker, webhook, grpc, config)

## Консистентность

- Все изменения согласованы с ADR-035 (`doc/архитектура/adr/035-integration-proxy.md`)
- Сохранён один `go.mod` (два бинарника — два `cmd/`)
- Format Go code blocks, таблиц сохранён
- Не затронуты другие пакеты

## Время

~40 минут

## Замечания

Нет.
