# Отчёт T1505 — ADR-033: Testing Strategy

## Статус

✅ Завершена

## Что сделано

- Создан `архитектура/adr/033-frontend-testing.md` (полный ADR в формате ХАДД)
- Решение: **Vitest + RTL** (unit) + **Playwright** (E2E)
- Граница: DS-компоненты тестируются в DS repo, LKFL — только свои компоненты
