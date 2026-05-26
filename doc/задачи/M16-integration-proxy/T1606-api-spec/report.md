# T1606 — Отчёт: спецификация/api.md — admin endpoints → proxy

## Веха
M16-integration-proxy

## Дата
2026-05-26

## Что сделано

### 1. Обновлена секция "Integrations Hub Admin API"

- Добавлен M16 note: монолит делегирует запросы к `lkfl-integration-proxy` через gRPC
- Указан facade pattern: Монолит → gRPC → proxy → ответ → HTTP response
- Обновлена строка backend: `integrationclient/` → gRPC → proxy
- Ссылка на ADR-035 добавлена
- HTTP API описание обновлено: `/admin/integrations/` → lkfl-server:8080 → gRPC → lkfl-integration-proxy:8090

### 2. Response Schema — Provider: добавлены proxy-поля

- `circuit_breaker_state` (enum: `closed`, `open`, `half_open`) — состояние circuit breaker провайдера
- `worker_pool_status` (object: `{ active: int, queued: int, max_concurrent: int }`) — статус worker pool провайдера
- Добавлено пояснение по полям после JSON блока

### 3. Error responses: добавлены gRPC error codes

| gRPC Code | HTTP Status | Описание |
|-----------|------------|----------|
| `UNAVAILABLE` | 503 | Proxy недоступен |
| `DEADLINE_EXCEEDED` | 504 | Timeout gRPC вызова |
| `INTERNAL` | 500 | Ошибка внутри proxy |
| `NOT_FOUND` | 404 | Провайдер не найден |

Пример JSON error response с `grpc_code` и `retry_after_seconds` добавлен.

### 4. Backend package mapping обновлён

- `External providers`: `internal/integrationclient/` → gRPC → `lkfl-integration-proxy` (было: `internal/integrations/`)
- Методы: `IntegrationClient.Activate()`, `IntegrationClient.SyncCatalog()` → gRPC → proxy (было: `ProviderGateway.Activate()`)
- Сводная таблица: External → `integrationclient/` → gRPC → proxy

### 5. История изменений

- Добавлен раздел v4 → v5 (M16 — Integration Proxy) с описанием всех изменений

## Изменённые файлы

| Файл | Изменения |
|------|----------|
| `doc/спецификация/api.md` | Секция Integrations Hub Admin API: note, response schema, error codes, backend mapping, история |
| `doc/задачи/M16-integration-proxy/T1606-api-spec/plan.yaml` | progress: 100%, все чекбоксы [x] |
| `doc/задачи/M16-integration-proxy/T1606-api-spec/report.md` | Этот файл |

## Консистентность с ADR-035

- ✅ Facade pattern: монолит → gRPC → proxy (ADR-035, Вариант 3)
- ✅ Circuit breaker states: closed, open, half_open (ADR-035, Circuit Breaker)
- ✅ Worker pool: active, queued, max_concurrent (ADR-035, Worker Pool per Provider)
- ✅ gRPC error codes маппинг (ADR-035, Fault Tolerance)
- ✅ Package: `integrationclient/` (ADR-035, Architecture)
- ✅ Proxy порт: 8090 gRPC (ADR-035, Ports)

## Замечания

Нет существующих endpoint'ов удалено — только добавлены notes и поля.
