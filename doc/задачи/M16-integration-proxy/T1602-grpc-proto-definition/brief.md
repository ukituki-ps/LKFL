# T1602 — gRPC Proto Definition

## Веха

M16-integration-proxy

## Контекст

ADR-035 определил архитектуру Integration Proxy. Нужна детальная спецификация gRPC contract между монолитом и proxy.

## Что сделать

Описать полный `.proto` файл для `integration.v1.IntegrationService`:
- 8 RPC методов (Activate, Deactivate, GetProviderStatus, GetCatalog, HealthCheck, ListProviders, GetProvider, UpdateProvider, TriggerSync, GetSyncLogs)
- Все message types с полями, типами, комментариями
- Enum для статусов, состояний circuit breaker

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Создать | `архитектура/интеграции.md` — секция "gRPC Contract" с полным proto |

### Критерии приёмки

- [ ] Proto definition описан в документации
- [ ] Все message types определены
- [ ] Enum для статусов (active, inactive, error, in_progress)
- [ ] Enum для circuit breaker states (closed, open, half-open)
- [ ] Комментарии к каждому полю
