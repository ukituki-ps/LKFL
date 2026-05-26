# T0706 — WizardEngine для фронтенда — отчёт

## Статус

✅ выполнено

## Что сделано

- `архитектура/модули.md` — modals table обновлена:
  - Убраны DmsWizard.tsx и MatCapitalWizard.tsx из таблицы модальных окон
  - Добавлен Wizard.tsx как универсальный JSON-driven renderer
  - wizard-configs/ добавлен (dms-upgrade.json, matkapital.json, ...)
- WizardEngine store добавлен: `src/store/wizards.ts` (пятый store)
- JSON schema wizard config задокументирована (steps[], validation, onComplete action)
- `спецификация/api.md` — добавлен раздел «Wizard Config»:
  - `/wizards` — GET, список wizard конфигов
  - `/wizards/:id` — GET, детали конфига
  - `/wizards/:id/validate` — POST, server-side validation
  - Итого 3 endpoints
- `спецификация/артефакты.md` — generic wizard вместо 2 конкретных
- `контекст/настраиваемость.md` — wizard как configurable (не requires-code)
- Создан ADR-019: обоснование generic wizard (общие элементы 2 wizard-ов → 1 renderer)
- `архитектура/README.md` — ADR-019 добавлен в таблицу
- `задачи/README.md` — статус M07 обновлён

## Проблемы

- Нет
