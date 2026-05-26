# Отчёт T1503 — ADR-031: Data Fetching Strategy

## Статус

✅ Завершена

## Что сделано

- Создан `архитектура/adr/031-api-data-fetching.md` (полный ADR в формате ХАДД)
- Решение: **React Query** (`@tanstack/react-query`)
- Обоснование: 118 endpoints → caching критичен; DevTools; optimistic updates; polling
