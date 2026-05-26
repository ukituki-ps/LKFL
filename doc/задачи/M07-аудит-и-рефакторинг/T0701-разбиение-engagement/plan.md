# T0701 — Разбиение engagement/ на catalog + flow + collections + eligibility

## План

- [x] 1. Прочитать текущую структуру `engagement/` из `архитектура/пакеты-platform.md`
- [x] 2. Определить boundary eligibility → отдельный пакет
- [x] 3. Обновить `пакеты-platform.md` — добавить `eligibility/` как 8-й пакет
- [x] 4. Обновить DI граф — кто зависит от eligibility (flow, recommendations, billing-rule)
- [x] 5. Обновить `архитектура/модули.md` — Platform: 7 → 8 пакетов
- [x] 6. Обновить `архитектура/engagement.md` — eligibility section reference
- [x] 7. Обновить `спецификация/api.md` — eligibility endpoints → отдельный раздел
- [x] 8. Создать ADR-014: почему eligibility вынесен в отдельный пакет
- [x] 9. Обновить `архитектура/README.md` — ADR-014 в таблице
- [x] 10. Обновить `задачи/README.md` — статус M07

## Зависимости

- Нет (свободная задача)
