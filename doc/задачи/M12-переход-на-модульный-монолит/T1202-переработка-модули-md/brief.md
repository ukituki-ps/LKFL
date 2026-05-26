# T1202 — Переделка `архитектура/модули.md`

## Контекст

`модули.md` описывает 4 Go-сервиса (platform, billing, integrations, payment-gateway) + LLM Proxy + React SPA. После M12 всё это → один бинарник `lkfl-server` с 13 internal пакетами.

## План

Полностью переписать файл:
1. Tree structure: `backend/` вместо `platform/` + `billing/` + `integrations/` + `payment-gateway/`
2. Убрать "Сервис 2. Billing", "Сервис 3. Integrations", "Сервис 4. Payment Gateway", "Сервис 6. LLM"
3. Добавить billing/, integrations/, payments/ как internal пакеты Platform
4. NATS таблицы → DI через interfaces (контракт Go-типов)
5. Nginx routes: `/billing/v1/` → `/api/v1/billing/`, `/payments/v1/` → `/api/v1/payments/`
6. Ports: убрать :8081, :8082, :8084, :8085 → остался :8080 + :8083 (asynq dashboard)
7. Release policy: один бинарник
8. Зависимости между сервисами → зависимости между пакетами

## Ожидаемый результат

`архитектура/модули.md` переписана для modular monolith. Все упоминания отдельных сервисов убраны или заменены.