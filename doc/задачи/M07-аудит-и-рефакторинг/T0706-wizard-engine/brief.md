# T0706 — WizardEngine для фронтенда

## Веха

M07-аудит-и-рефакторинг

## Контекст

Фронтенд содержит 2 wizard-компонента:
- `DmsWizard.tsx` — 4 шага для ДМС (Опция → Оплата → Подтверждение → Готово)
- `MatCapitalWizard.tsx` — 4 шага для маткапитала (Условия → Данные ребёнка → Подтверждение → Готово)

**Проблема:**
Оба wizard-а жёстко привязаны к workflow СДЭК. Новый tenant с другими wizard-ами требует нового TSX-компонента.

Противоречит white-label философии: `контекст/философия.md` — «Новый tenant без изменения кода».

**Решение:**

Создать конфигурируемую `WizardEngine`:
```json
{
  "id": "wizard-dms",
  "name": "Подключение ДМС",
  "steps": [
    {
      "order": 1,
      "type": "option_selection",
      "title": "Выберите программу",
      "component": "OptionCardGroup",
      "props": { "optionsSource": "/api/v1/engagement-offers?type=dms" }
    },
    {
      "order": 2,
      "type": "payment",
      "title": "Способ оплаты",
      "component": "PaymentMethodSelector",
      "props": { "methods": ["payroll_deduction", "card"] }
    },
    {
      "order": 3,
      "type": "confirmation",
      "title": "Подтверждение",
      "component": "DocumentPreview",
      "props": { "templateId": "dms-application" }
    },
    {
      "order": 4,
      "type": "completion",
      "title": "Готово",
      "component": "SuccessMessage",
      "props": { "redirectUrl": "/" }
    }
  ]
}
```

Тогда DmsWizard и MatCapitalWizard исчезают как компоненты. На их месте `Wizard.tsx` рендерит любой wizard из JSON-config.

### Файлы-мишени

| Действие | Файл |
|---|---|
| Новая структура frontend | `архитектура/модули.md` | Wizard.tsx + WizardEngine (store) |
| Убрать DmsWizard | `архитектура/модули.md` — modals table | только Wizard.tsx, BenefitDetail.tsx |
| API spec | `спецификация/api.md` — endpoint `/api/v1/wizards` (get wizard config) |
| Спецификация артефакты | `спецификация/артефакты.md` — generic wizard вместо 2 конкретных |
| Настраиваемость | `контекст/настраиваемость.md` — "Новый wizard" → configure без кода (добавить в матрицу wizard как configurable) |
| Journeys сотрудник | `спецификация/journeys/сотрудник.md` — wizard steps generic, не СДЭК-specific |
| Создать ADR | `архитектура/adr/ADR-019-wizard-engine.md` |
| Обновить README архитектуры | `архитектура/README.md` — ADR-019 |
| Обновить README вехи | `задачи/README.md` — статус M07 |

### Критерии приёмки

- [x] DmsWizard.tsx и MatCapitalWizard.tsx убраны из таблицы модальных окон
- [x] Wizard.tsx добавлен как универсальный компонент
- [x] WizardEngine store добавлен в `src/store/` (пятый store: wizards.ts)
- [x] JSON-schema wizard config документирован
- [x] `спецификация/api.md` — `/api/v1/wizards` endpoint
- [x] Создан ADR-019
- [x] Файлы-мишени все перечислены выше
