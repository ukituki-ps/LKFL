# ADR-019 — WizardEngine для фронтенда (JSON-driven generic wizard)

## Контекст

Фронтенд содержит два wizard-компонента, жёстко привязанных к СДЭК:

1. `DmsWizard` (4 шага: Опция → Оплата → ПодConfirmation → Готово)
2. `MatCapitalWizard` (4 шага: Условия → Данные ребёнка → ПодConfirmation → Готовo)

Общие элементы:
- Мультишаговая последовательность
- Навигация вперёд/назад с валидацией
- Суммарный review-шаг перед подачей
- Закрытие по успеху/отмене

Различия:
- Данные формы (JSON schema)
- Конкретные шаги (API endpoints)
- UI-компоненты на каждом шаге

## Решение

Создать generic `WizardEngine` + `Wizard.tsx`:

### WizardEngine (store):
```typescript
interface WizardState {
  currentStep: number;
  formData: Record<string, any>;
  validation: Record<string, string>;
  isSubmitting: boolean;
}

interface WizardConfig {
  id: string;
  name: string;
  steps: WizardStep[];
  onComplete: WizardAction;
}
```

### Wizard.tsx (generic step renderer):
```tsx
<Wizard config={wizardConfig} onComplete={handler} />
```

Каждый конкретный wizard — это только JSON-конфиг (не Go-код, не React-компонент):

```json
{
  "id": "dms-upgrade",
  "name": "Подключение ДМС",
  "steps": [
    { "id": "option", "component": "OptionSelector", "validate": "required" },
    { "id": "payment-method", "component": "PaymentMethod", "validate": "card_or_payroll" },
    { "id": "confirmation", "component": "ReviewSummary", "validate": "all_fields" },
    { "id": "done", "component": "SuccessScreen" }
  ],
  "onComplete": { "type": "api", "method": "POST", "url": "/user-engagements" }
}
```

### Что меняется:
- `DmsWizard.tsx` → JSON config + Wizard renderer
- `MatCapitalWizard.tsx` → JSON config + Wizard renderer
- `Wizard.tsx` — generic шаблон (переиспользуется для любого multi-step workflow)

## Последствия

- ✅ 0 нового кода для новых wizard'ов — только JSON config
- ✅ Конфигурируемость: HR/catalog_manager может менять шаги через admin UI (в будущем)
- ✅ White-label: разные tenant'ы — разные wizard'ы, один renderer
- ⚠️ Усложнение: если wizard имеет custom логику — нужен custom component (fallback)

## Статус

✅ Accepted (M07, T0706)
