# T0707 — Унификация NATS subjects registry (включая billing namespace)

## Веха

M07-аудит-и-рефакторинг

## Контекст

NATS subjects описаны в трёх местах с рассинхронизациями:

1. **`архитектура/модули.md`** — Billing ↔ Platform (6 subjects) + Platform ↔ Integrations (9 subjects)
2. **`архитектура/биллинг-движок.md`** — Billing ↔ Platform (6 subjects, другой формат таблицы)
3. **`архитектура/engagement.md`** — billing events внутри engagement/ (5 subjects с namespace `engagement.debit.*`)

**Пример рассинхронизации:**
- модули.md: `billing.debit` — Platform → Billing (один subject)
- engagement.md: `engagement.debit.confirm`, `engagement.debit.reserve` — 3 отдельных event'а в другом namespace
- billing-движок.md: `billing.debit` — тот же event, но без reserve/confirm distinction

**Проблема namespace:**
- `модули.md` и `биллинг-движок.md` используют `billing.*` namespace
- `engagement.md` использует `engagement.debit.*` namespace
- Разработчик читает один файл → не видит полную картину → wrong subject name в коде

**Решение namespace:**
1. Все billing-события используют namespace `billing.*` (не `engagement.debit.*`)
2. `billing.debit` разбит на `billing.debit.reserve` (заморозка) + `billing.debit.confirm` (подтверждение)
3. Создан единый billing-events section в billing-движок.md, ссылающийся на nats-subjects.md

**Решение registry:**
Создать единый registry NATS subjects как отдельный документ:
`архитектура/nats-subjects.md`

Формат:
```
Subject: billing.credit
Direction: platform → billing
Trigger: engagement → completed (activity)
Payload: { userId, amount, source }
Consumer: billing/transaction/
Response: billing.credit.confirmed (sync)
DLQ: billing.credit.failed
```

### Файлы-мишени

| Действие | Файл |
|--|--|--|
| Новый файл | `архитектура/nats-subjects.md` — полный registry всех subjects |
| Обновить | `архитектура/модули.md` — replace inline tables → link to nats-subjects.md |
| Обновить | `архитектура/биллинг-движок.md` — replace inline table → link + billing-events section |
| Обновить | `архитектура/engagement.md` — billing events → `billing.*` namespace + link |
| Обновить | `архитектура/интеграции.md` → link to nats-subjects |
| Создать ADR | `архитектура/adr/ADR-020-nats-subjects-registry.md` |
| Обновить README архитектуры | `архитектура/README.md` — nats-subjects.md в таблице содержимого + ADR-020 |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] Создан `архитектура/nats-subjects.md` с полным списком всех subjects
- [x] Каждый subject: имя, direction, trigger, payload schema, consumer, response/ack, DLQ
- [x] Все 3 источника (модули.md, billing-движок.md, engagement.md) обновлены → ссылка на registry
- [x] Нет рассинхронизаций: каждый subject появляется 1 раз в registry, везде ссылка
- [x] Все billing-события используют namespace `billing.*` (не `engagement.debit.*`)
- [x] `billing.debit` разбит на `billing.debit.reserve` + `billing.debit.confirm`
- [x] billing-движок.md содержит billing-events section ссылающийся на registry
- [x] Создан ADR-020
- [x] Файлы-мишени все перечислены выше
