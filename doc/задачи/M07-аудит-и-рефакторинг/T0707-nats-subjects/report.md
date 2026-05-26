# T0707 — Унификация NATS subjects registry — отчёт

## Статус

✅ выполнено

## Что сделано

- Создан `архитектура/nats-subjects.md` — master registry всех subjects:
  - billing.* — 8 subjects (Platform ↔ Billing)
  - payment.* — 4 subjects (Platform ↔ Payment Gateway)
  - integration.* — 7 subjects (Platform ↔ Integrations)
  - Итого: 19 subjects
- Все billing-события используют namespace `billing.*` (не `engagement.debit.*`)
- `billing.debit` разбит на `billing.debit.reserve` + `billing.debit.confirm` + `billing.debit.reverse`
- Каждый subject: имя, direction, trigger, payload schema, consumer, response/ack
- `архитектура/модули.md` — inline tables → ссылка на nats-subjects.md
- `архитектура/биллинг-движок.md` — inline table → ссылка на registry
- `архитектура/engagement.md` — billing events → `billing.*` namespace + ссылка на registry
- `архитектура/интеграции.md` → ссылка на nats-subjects.md
- Создан ADR-020: обоснование унификации (3 источника → 1 registry, namespace consistency)
- `архитектура/README.md` — nats-subjects.md в таблице содержимого + ADR-020
- `задачи/README.md` — статус M07 обновлён

## Проблемы

- Breaking change: `engagement.debit.*` → `billing.*` (требует update NATS consumers при реализации кода)
