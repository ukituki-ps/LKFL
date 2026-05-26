# T1103 — 1C → Billing payroll

## Веха

M11-split-integrations

## Контекст

`integrations/1c/` документируется как:
- `SubmitPayroll()` — передача заявлений на удержание из ЗП в 1C

Сейчас 1C находится в Integrations Hub, который также содержит provider adapters.

NATS flow (до M11):
```
platform publishes: integration.payroll.submit  → integrations subscribes
```

**Проблема:**
1C — бухгалтерия (удержание из ЗП). Это финансовая операция, которая:
- Привязана к платежам (DMS upgrade → удержание из ЗП)
- Нуждается в audit trail вместе с transaction history
- Должна быть close к billing/payment domain

Если Integrations упал → провайдеры недоступны → 1C тоже недоступен. 1C ≠ vendor activation. Это разные SLA.

**Решение — перенести 1C в Billing:**
```
billing/
  internal/
    account/
    transaction/
    period/
    rule_engine/
    payroll/          ← НОВЫЙ: 1C integration (было integrations/1c/)
    db/
```

Payroll engine: `SubmitPayroll(stmt *PayrollStatement)`, `ListPayrollStatements(tenantId string) []PayrollStatement`, `GetPayrollStatus(id string) (*PayrollStatus, error)`.

Billing становится consumer payroll subject. T1104 переименует `integration.payroll.submit` → `finance.payroll.submit`.

### Файлы-мишени

| Действие | Файл |
|---|--|
| payroll/ → billing | `архитектура/модули.md` — Billing +1 модуль: payroll/ |
| Убрать из Integrations | `архитектура/модули.md` — 1c/ удалён |
| NATS consumer | `архитектура/nats-subjects.md` — payroll subject consumer → billing (T1104 переименует namespace) |
| Обновить биллинг-движок | `архитектура/биллинг-движок.md` — payroll section (1C, удержание из ЗП) |
| Создать ADR | `архитектура/adr/027-payroll-in-billing.md` |
| Обновить README архитектуры | `архитектура/README.md` — ADR-027 |

### Критерии приёмки

- [ ] `архитектура/модули.md` — Billing имеет payroll/ (SubmitPayroll(stmt), ListPayrollStatements(), GetPayrollStatus(id))
- [ ] integrations/1c/ удалён из таблицы модулей
- [ ] `архитектура/nats-subjects.md` — integration.payroll.submit consumer = billing (было integrations)
- [ ] `архитектура/nats-subjects.md` — payload description обновлён: "Передача заявления на удержание ЗП → 1C via billing/payroll/"
- [ ] `архитектура/биллинг-движок.md` — payroll section добавлен (1C, uдержание из ЗП, flow с payment-gateway)
- [ ] Создан ADR-027: обоснование (1C = бухгалтерия = finance domain, closer to billing, separate SLA from providers)
- [ ] `архитектура/README.md` — ADR-027 в таблицу
