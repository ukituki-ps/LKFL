# T0702 — Выделение compliance (cascade + audit) из consent/

## План

- [x] 1. Прочитать текущую структуру `consent/` из `архитектура/пакеты-platform.md`
- [x] 2. Определить boundary: CascadeRevoke → compliance, audit trail → compliance, retention → compliance
- [x] 3. Обновить `пакеты-platform.md` — добавить `compliance/` как 9-й пакет
- [x] 4. Обновить DI граф — compliance зависит от user, engagement, notification, db/
- [x] 5. Обновить `архитектура/модули.md` — Platform: 8 → 9 пакетов
- [x] 6. Обновить `архитектура/безопасность.md` — audit trail → compliance/ reference
- [x] 7. Создать ADR-015: почему compliance выделен из consent
- [x] 8. Обновить `архитектура/README.md` — ADR-015 в таблице
- [x] 9. Обновить `задачи/README.md` — статус M07

## Зависимости

- Нет (compliance зависит от существующей структуры engagement, а не от eligibility extraction)
