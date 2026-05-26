# ADR-027: 1C Payroll → Billing payroll/ (перенос из Integrations)

**Статус:** ❌ Superseded by ADR-024 (Modular Monolith)
**Дата:** 2026-05-25
**Контекст:** M11 T1103 — split integrations по доменам

## Ситуация

`integrations/1c/` — бухгалтерия (удержание из ЗП) — находится в Integrations Hub = wrong domain:
```
Platform → NATS `integration.payroll.submit` → Integrations → 1C REST API
```

**Проблемы:**
- 1C — бухгалтерия (удержание из ЗП). Это финансовая операция, которая:
  - Привязана к платежам (DMS upgrade → удержание из ЗП)
  - Нуждается в audit trail вместе с transaction history
  - Должна быть close к billing/payment domain
- Если Provider Gateway падает → 1C тоже падает. 1C ≠ vendor activation. Разные SLA.
- Смешивание бухгалтерии с benefit-провайдерами нарушает SRP.

## Решение

Перенести 1C в Billing `internal/billing/payroll/` (подпакет):
```
Platform → NATS `finance.payroll.submit` → Billing payroll.SubmitPayroll(stmt) → 1C REST API
```

**Изменения:**
- `архитектура/модули.md` — Billing +payroll/ (SubmitPayroll, ListPayrollStatements, GetPayrollStatus)
- `архитектура/модули.md` — 1c/ удалён из Provider Gateway
- `архитектура/nats-subjects.md` — `finance.payroll.submit` (consumer = billing)
- `архитектура/пакеты-platform.md § billing/` — +Payroll section (1C integration, flow с payment-gateway)
- NATS namespace rename: `integration.payroll.submit` → `finance.payroll.submit` (выполняется T1104)

**API payroll/:**
- `SubmitPayroll(stmt *PayrollStatement)` — передать заявление на удержание
- `ListPayrollStatements(tenantId string) []PayrollStatement` — список заявлений
- `GetPayrollStatus(id string) (*PayrollStatus, error)` — статус заявления

**Зависимости:**
- 1C REST API — прямой вызов (не через Provider Gateway)
- NATS consumer — subject `finance.payroll.submit` (M11 T1104)
- `db/` — PostgreSQL (payroll_statements table)

**Почему Billing, а не Payment Gateway:**
- 1C ≠ PCI DSS (не обрабатывает карточные данные)
- 1C = бухгалтерия = finance domain → ближе к billing (transaction audit trail)
- Payment Gateway — PCI DSS isolation (только card payments)

**Последствия:**
- ✅ Бухгалтерия belongs к finance domain (domain correctness)
- ✅ 1C-audit trail рядом с billing transactions
- ✅ 1C SLA ≠ provider SLA (decoupling)
- ⚠️ Billing теперь зависит от 1C REST API напрямую
- ⚠️ Нужно добавить circuit breaker для 1C connections