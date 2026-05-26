# T1205 — Переделка `спецификация/api.md`

## Контекст

`спецификация/api.md` описывает `/billing/v1/` endpoints и `/payments/v1/` endpoints как separate API'ы. При mono-литной архитектуре всё под `/api/v1/`.

## План

1. Убрать секцию "Billing API (/billing/v1/)" → endpoints перенести в `/api/v1/billing/`
2. Убрать секцию "Payment Gateway API (/payments/v1/)" → endpoints перенести в `/api/v1/payments/`
3. Health: 4 health endpoints → один `/api/healthz`
4. Error codes: убрать `SERVICE_UNAVAILABLE` для billing, integrations, payments (больше нет separate services)
5. Обновить frontend API paths — `/billing/v1/` → `/api/v1/billing/` (through Nginx)

## Ожидаемый результат

`спецификация/api.md` содержит единый набор endpoints под `/api/v1/`. Убраны упоминания отдельных сервисов.