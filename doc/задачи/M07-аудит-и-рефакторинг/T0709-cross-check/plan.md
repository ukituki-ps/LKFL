# T0709 — Общая проверка: перекрёстные ссылки и консистентность

## План

- [x] 1. Перечислить all .md files в doc/архитектура/, doc/контекст/, doc/спецификация/
- [x] 2. Сверить каждую table of content → file exists
- [x] 3. Проверить all ADR numbers: ADR-001 → ADR-020, no gaps
- [x] 4. Grep all `\]\(` refs → verify each target file exists
- [x] 5. Fix stale links (old paths → new paths after M07)
- [x] 6. Check service count: all refs say 4 Go services
- [x] 7. Check namespace: billing.* uniform everywhere
- [x] 8. Check настраиваемость.md matrix vs architectural decisions
- [x] 9. Закрыть M07 in `задачи/README.md`

## Зависимости

- ALL T0701-T0708 — зависит от всех предыдущих задач
