# T0701 — Разбиение engagement/ на catalog + flow + collections + eligibility — отчёт

## Статус

✅ выполнено

## Что сделано

- `архитектура/пакеты-platform.md` — добавлен `eligibility/` как 8-й пакет (позже 9-й с compliance)
- `eligibility/` задокументирован: engine.go (Check, EvaluateRule, EvaluateGroup) + types.go
- `engagement/` уменьшен: удалён eligibility.go, оставлены catalog + flow + collections + billing_events
- DI граф обновлён: flow зависит от eligibility через DI (не через import подпакета)
- `архитектура/модули.md` — Platform: 7 → 8 → 9 пакетов (с учётом M07)
- `архитектура/engagement.md` — eligibility section обновлена (ссылка на eligibility/ пакет)
- `спецификация/api.md` — eligibility endpoints вынесены в отдельный раздел
- Создан ADR-014: обоснование выноса eligibility в отдельный пакет (3 вызывающих домена)
- `архитектура/README.md` — ADR-014 добавлен в таблицу
- `задачи/README.md` — статус M07 обновлён

## Проблемы

- Была stale ASCII-диаграмма в `пакеты-platform.md` (стр.557), где eligibility показан как подпакет engagement/ — исправлено в ходе аудита
