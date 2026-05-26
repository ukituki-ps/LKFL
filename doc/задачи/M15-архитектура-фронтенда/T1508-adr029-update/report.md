# Отчёт T1508 — Обновление ADR-029 (DS gap analysis: 11 → 2)

## Статус

✅ Завершена

## Что сделано

1. **Пересмотрены все 11 компонентов:** DS-001 → DS-011
2. **Таблицы §1.1, §1.2, §1.3 обновлены** в `архитектура/adr/029-ds-components-gap-tz.md`
3. **Добавлен §1.4** — «Компоненты, отменённые после пересмотра» (9 компонентов)
4. **Добавлен §2** — «Компоненты LKFL vs DS» — критерий решения (8 LKFL-компонентов)
5. **Добавлен хедер** — пометка M15 T1508 с датой обновления

## Итог: 11 → 2

| Статус | Кол-во | Компоненты |
|--------|--------|------------|
| ✅ Покрыто DS/Mantine | 9 | StatCard, EventsFeed, PolicyCard, ClinicMapList, TopTabNavigation, BalancePill, DocumentRow, SupportFAQ, FilterPills |
| 🔴 Реальные gap'ы | 2 | WizardContainer (LKFL обёртка), TransactionList (LKFL компонент) |
