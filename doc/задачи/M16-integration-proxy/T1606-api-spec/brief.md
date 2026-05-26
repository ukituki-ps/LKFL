# T1606 — `спецификация/api.md` — admin endpoints → proxy

## Веха

M16-integration-proxy

## Контекст

12 admin endpoints Integrations Hub сейчас описаны как endpoint'ы монолита. Фактически монолит делегирует их proxy через gRPC.

## Что сделать

Обновить `спецификация/api.md`:
1. Секция "Integrations Hub Admin API" — добавить note: монолит делегирует запросы к proxy через gRPC
2. Response schema — добавить proxy-specific поля (circuit_breaker_state, worker_pool_status)
3. Error responses — добавить gRPC error codes (UNAVAILABLE — proxy down, DEADLINE_EXCEEDED — timeout)

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `спецификация/api.md` — Integrations Hub Admin API секция |

### Критерии приёмки

- [ ] Note: mono delegates to proxy via gRPC
- [ ] Response schema + proxy fields
- [ ] Error responses + gRPC codes
- [ ] Backend package mapping обновлён
