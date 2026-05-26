# T1103 — 1C → Billing payroll — отчёт

## Статус

✅ выполнено (документация)

## Что сделано

### `архитектура/модули.md` (основной артефакт #1):
- Billing секция: +`payroll/` модуль в таблицу модулей — новый API (`SubmitPayroll`, `ListPayrollStatements`, `GetPayrollStatus`)
- Provider Gateway: 1c/ удалён из таблиц (только benefit-providers)
- NATS JetStream: +`finance.payroll.submit` subject — producer = platform, consumer = billing
- Dependencies: "Billing → 1С напрямую (REST API через `internal/payroll/`, ADR-027) — без NATS"

### `архитектура/биллинг-движок.md` (основной артефакт #2):
- Добавлена секция "Payroll — 1C интеграция (M11 T1103, ADR-027)" перед "Управление правилами (API)"
- Описание API payroll/: SubmitPayroll, ListPayrollStatements, GetPayrollStatus
- NATS flow: Platform → `finance.payroll.submit` → Billing → 1C REST API
- Flow с payment-gateway: DMS upgrade → salary deduction vs card payment (clear separation)
- Зависимости: 1C REST API direct, NATS consumer, db/

### `архитектура/README.md` (cross-cutting):
- биллинг-движок.md описание: "+Payroll (1C) section"
- Ключевые решения: "Биллинг отдельно... M11 T1103: +payroll/ (1C бухгалтерия)"

### ADR:
- `adr/027-payroll-in-billing.md` — ХАДД: бухгалтерия = finance domain, closer to billing transactions

## Проблемы

- Billing зависит от 1C REST API напрямую. Нужен circuit breaker для 1C connections при реализации.
- `payment.payroll.submit` (payment-gateway payroll channel) остаётся в Payment Gateway — это пэйшлюз-канал удержания (отличается от 1C REST API).

## Следующие шаги

N/A — задача выполнена.
