# T1104 — NATS consumer update — отчёт

## Статус

✅ выполнено (документация)

## Что сделано

### `архитектура/nats-subjects.md` (основной артефакт):
- Namespace Rules обновлены: 4 → 5 namespaces
  - `provider.*` — Platform ↔ Provider Gateway (был `integration.*`)
  - `finance.payroll.*` — Platform ↔ Billing (был `integration.payroll.*`)
  - `llm.*` — Billing ↔ Platform (был implicit)
- Integrations секция полностью переписана → "Provider Gateway (`provider.*`)" + "Finance Payroll (`finance.payroll.*`)"
- Provider Gateway subjects: 4 subject'а (was 7, минус hr и payroll)
  - consumer renamed: integrations → provider-gateway
- Payroll: 1 subject (`finance.payroll.submit`) — consumer = billing
- Сводная таблица: 5 namespaces, 19 subjects total, 5 сервисов (4 Go + 1 in-process)
- "Что НЕ в NATS" обновлен: hr-sync-daily → Asynq (was: NATS)

### `архитектура/модули.md` (cross-cutting):
- NATS JetStream секция: старая таблица `integration.*` → новая `provider.*`
- Новая подтаблица M11 T1103 Payroll: `finance.payroll.submit`
- Dependencies: HR-sync и 1C убраны из NATS → Provider Gateway
- Rule обновлена: "Platform и Billing не обращаются к Provider Gateway для HR и 1C"

### ADR:
- `adr/028-nats-rename.md` — ХАДД: namespace rename rationale, migration window, stream archive strategy

## Проблемы

- NATS JetStream streams `integration.*` нужно archive после cut-over
- Новые streams `provider.*` и `finance.payroll.*` создать до deploy
- Все consumers переключить на новые subject'ы (migration window)

## Следующие шаги

N/A — задача выполнена. M11 полностью завершена.
