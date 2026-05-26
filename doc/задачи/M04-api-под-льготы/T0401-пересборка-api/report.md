# T0401 — Отчёт о выполнении

## Статус

выполнено

## Что сделать

Пересобрать API-контракты под новую архитектуру льгот (5 сущностей).
Детали → [brief.md](./brief.md)

## Результат

### Переписано

1. **`спецификация/api.md`** — полный реврайт (~400 строк):
   - Catalog/Benefits → 4 endpoint-а (продукт + тарифы)
   - Benefit Plans → 2 endpoint-а (`/benefit-plans/:id`, check-eligibility)
   - User Benefits → 7 endpoint-ов (activate, deactivate, upgrade, steps)
   - Collections → 3 endpoint-а (plans, не benefits)
   - Activities → 5 endpoint-ов (universal `/submit` вместо `/survey/submit` + `/attend`)
   - DMS (11 endpoints) — **убраны** (теперь generic через user-benefits + activation-flow)
   - Matkapital (3 endpoints) — **убраны** (теперь generic через user-benefits + activation-flow)
   - Admin: Benefits → 7 endpoint-ов (CRUD продукта)
   - Admin: Benefit Plans → 6 endpoint-ов (CRUD тарифов)
   - Admin: Activation Flows → 5 endpoint-ов (CRUD потоков)
   - Admin: Collections → 5 endpoint-ов
   - Итого: **128 endpoints** (было 116), 28 модулей (было 24)

2. **`спецификация/артефакты.md`**:
   - S10a «ДМС Wizard» → S10a «Activation Wizard» (generic, рендерит steps из activation_flow)
   - S10c «Маткапитал Wizard» → объединён в S10a (убран как отдельный артефакт)
   - S10b «Детали льготы» → табы из `plan_metadata`, не захардкожены
   - Артефактов: 26 (было 27)

### Обновлено

3. **`задачи/README.md`** — M03 и M04 в таблице вех
4. **`задачи/M04-api-под-льготы/overview.md`** — статус вехи

## Критерии приёмки

- [x] `api.md` — все endpoints обновлены под 5 сущностей льгот
- [x] `api.md` — DMS/matkapital убраны как отдельные модули (generic через activation flow)
- [x] `api.md` — добавлены endpoints для benefit-plans, activation-flows, user-benefits
- [x] `артефакты.md` — S10a generic wizard, S10b обновлён
- [x] Сводная таблица API обновлена (128 endpoints, 28 модулей)
- [x] Endpoints согласованы с journeys (J02–J08, J13a)

## Проблемы

Нет.

## Следующие шаги

Нет — задача завершена.
