# ADR-020 — Унификация NATS subjects registry + billing namespace

**Статус:** ❌ Superseded by ADR-024 (M12)

> ❌ **Superseded by ADR-024 (Modular Monolith, M12).** NATS subjects registry удалён. `nats-subjects.md` больше не существует. Межмодульная коммуникация → Go interfaces.

## Обновление M12

> **M12:** Эта ADR заменена [ADR-024](./024-modular-monolith.md). NATS subjects registry больше не используется — в mono-режиме межмодульная коммуникация через Go interfaces, не NATS. `nats-subjects.md` удалён.

## Контекст (исторический)

NATS subjects рассинхронизированы между документами:

1. `модули.md` — таблица platform↔billing (6 subjects: `billing.credit`, `billing.debit`, `billing.balance.*`, `billing.transactions.*`)
2. `модули.md` — таблица platform↔integrations (9 subjects: `integration.*`)
3. `пакеты-platform.md § billing/` — inline таблица billing events (другой формат названий)
4. `engagement.md` — billing events как `engagement.debit.*` и `billing.debit.*` (противоречие namespace)

### Проблема namespace:

`engagement.md` использует `engagement.debit.reserve` в Billing Events секции, но `модули.md` использует `billing.debit`. Это создаёт неоднозначность: какой namespace master?

## Решение (историческое)

1. Создать единый `nats-subjects.md` — master registry всех subjects
2. Установить `billing.*` как master namespace для всех billing-событий (замена `engagement.debit.*` → `billing.*`)
3. Все inline-tabl'ы → ссылки на `nats-subjects.md`
4. Split `billing.debit` на `billing.debit.reserve` + `billing.debit.confirm` для точности

> **M12:** `nats-subjects.md` удалён. NATS больше не используется в mono-режиме.

## Последствия (исторические)

- ✅ Один источник истины — `nats-subjects.md`
- ✅ Namespace consistency: `billing.*` для всех billing событий
- ✅ Cross-reference maintenance: один файл вместо 4
- ⚠️ Breaking change: `engagement.debit.*` → `billing.*` (требовало update NATS consumers)

## Статус

❌ Superseded (M12, T1204) — NATS subjects заменены Go interfaces. См. [ADR-024](./024-modular-monolith.md).