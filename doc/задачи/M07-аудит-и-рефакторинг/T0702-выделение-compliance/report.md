# T0702 — Выделение compliance (cascade + audit) из consent/ — отчёт

## Статус

✅ выполнено

## Что сделано

- `архитектура/пакеты-platform.md` — добавлен `compliance/` как 9-й пакет
- `compliance/` задокументирован: cascade.go (CascadeRevoke), audit.go (AuditTrail), retention.go (EnforceRetention)
- `consent/` уменьшен: удалён CascadeRevoke, оставлены только lifecycle (Grant, Revoke, List, CheckGranted)
- DI граф обновлён: compliance зависит от user/, engagement/, notification/, db/
- `архитектура/модули.md` — Platform: 8 → 9 пакетов
- `архитектура/безопасность.md` — audit trail section обновлена (ссылка на compliance/)
- Создан ADR-015: обоснование выноса compliance из consent (3 домена каскадного удаления)
- `архитектура/README.md` — ADR-015 добавлен в таблицу
- `задачи/README.md` — статус M07 обновлён
- Asynq worker `consent-revoke` → переназначен на compliance/ (CascadeRevoke)

## Проблемы

- Нет
