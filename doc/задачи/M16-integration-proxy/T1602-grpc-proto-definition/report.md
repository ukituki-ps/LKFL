# T1602 — gRPC Proto Definition — Отчёт

## Веха

M16-integration-proxy

## Дата выполнения

2026-05-26

## Что сделано

Добавлена полная секция **«gRPC Contract»** в `doc/архитектура/интеграции.md` (после секции "ProviderAdapter interface", перед "Generic REST adapter").

### Содержание секции

1. **Package и импорты** — `integration.v1`, `go_package` path, импорты `timestamp` и `empty`.

2. **Enums:**
   - `ProviderStatus` — 4 значения: ACTIVE, INACTIVE, ERROR, IN_PROGRESS
   - `CircuitBreakerState` — 3 значения: CLOSED, OPEN, HALF_OPEN
   - `JobStatus` — 4 значения: QUEUED, PROCESSING, COMPLETED, FAILED

3. **Service IntegrationService** — 10 RPC методов:
   - Синхронные: `GetProviderStatus`, `GetCatalog`, `HealthCheck`
   - Асинхронные: `Activate`, `Deactivate`
   - Admin: `ListProviders`, `GetProvider`, `UpdateProvider`, `TriggerSync`, `GetSyncLogs`

4. **Message types** (полные определения с полями, типами, комментариями на русском):
   - `ActivateRequest` / `ActivateResponse`
   - `DeactivateRequest` / `DeactivateResponse`
   - `ProviderStatusRequest` / `ProviderStatusResponse`
   - `CatalogRequest` / `CatalogResponse` / `CatalogItem`
   - `HealthCheckRequest` / `HealthCheckResponse` / `ProviderHealthStatus`
   - `ListProvidersRequest` / `ListProvidersResponse` / `ProviderInfo`
   - `GetProviderRequest` / `GetProviderResponse` / `ProviderConfigResponse` / `ProviderStats`
   - `UpdateProviderRequest` / `UpdateProviderResponse`
   - `TriggerSyncRequest` / `TriggerSyncResponse`
   - `GetSyncLogsRequest` / `GetSyncLogsResponse` / `SyncLogEntry`

5. **Таблица RPC методов** — режим (sync/async), timeout, описание.

6. **Flow активации** — пошаговая диаграмма асинхронной активации.

## Консистентность с ADR-035

- Все 10 RPC методов из ADR-035 включены
- Разделение sync/async операций соответствует таблице ADR-035
- Circuit breaker параметры (states) согласованы
- Flow активации соответствует flow из ADR-035
- Message types расширяют черновой proto из ADR-035 (добавлены все недостающие типы)

## Изменённые файлы

| Файл | Действие |
|------|----------|
| `doc/архитектура/интеграции.md` | Добавлена секция «gRPC Contract» (~250 строк) |
| `doc/задачи/M16-integration-proxy/T1602-grpc-proto-definition/plan.yaml` | Обновлён progress: 100%, все чекбоксы [x] |
| `doc/задачи/M16-integration-proxy/T1602-grpc-proto-definition/report.md` | Создан (этот файл) |

## Замечания

Нет.
