# Архитектура фронтенда — Mobile + Forms

> Отдельный документ для мобильных и форм. Ссылается из `фронтенд.md` §§I-J.

---

## I. Mobile-архитектура

### I.1. `AprilMobileShellBar` — единая нижняя панель

**Назначение:** замена desktop header/sidebar на мобильных (<768px).

**MUST:**
- Один активный контекст — экран ИЛИ модалка, не оба одновременно
- `position: fixed` к viewport (root) или `absolute` (внутренний слой)
- z-index: `APRIL_MOBILE_SHELL_BAR_Z_INDEX` (над sheet оверлеями)
- Слоты: `leading` (Назад), `center` (табы/действия), trailing (поиск/меню)

**MUST NOT:**
- Не использовать одновременно с `AprilModal` (desktop модалка не отображается на mobile)
- Не менять z-index без согласования с DS

**Desktop → Mobile mapping:**

| Desktop | Mobile |
|---------|--------|
| Header (sticky top) | `AprilMobileShellBar` (fixed bottom) |
| Sidebar navigation | Bottom tabs |
| `AprilModal` (центрированный) | `AprilVaulBottomSheet` (снизу) |
| Dropdown menus | Bottom sheet или fullscreen page |

### I.2. Модальности

| Контекст | Desktop | Mobile |
|----------|---------|--------|
| Модалка | `AprilModal` (центрированный, scrollable body, headerActions) | `AprilVaulBottomSheet` (свайп закрытие, действия в shell bar) |
| Fallback mobile | — | `AprilMobileBottomSheet` (Mantine Drawer, без Vaul жестов) |
| Wizard | `AprilModal` + Mantine `Stepper` | `AprilVaulBottomSheet` + вертикальный прогресс |

**Правило переключения:**
```tsx
import { useMantineTheme } from '@mantine/core'

const isMobile = useMantineTheme().other?.breakpoints?.mobile // <768px

<ModalWrapper isOpen={opened} onClose={close}>
  {isMobile ? (
    <AprilVaulBottomSheet open={opened} onDismiss={close}>
      {children}
    </AprilVaulBottomSheet>
  ) : (
    <AprilModal opened={opened} onClose={close}>
      {children}
    </AprilModal>
  )}
</ModalWrapper>
```

### I.3. Breakpoints

| Breakpoint | Ширина | Поведение |
|------------|--------|-----------|
| `xl` (≥1280px) | Full desktop | Sidebar 250px + контент (admin) / Header + контент (employee) |
| `md` (768–1279px) | Tablet | Collapsible sidebar (admin) / Header + контент (employee) |
| `<768px` | Mobile | `AprilMobileShellBar`, полноэкранные страницы, bottom sheet модалки |

**Mantine integration:**
```ts
// theme.ts
createAprilTheme({
  other: {
    breakpoints: {
      mobile: '768px',
      tablet: '1280px',
    },
  },
})
```

### I.4. Touch-ориентиры

| Требование | Значение | Где |
|------------|----------|-----|
| Мин. зона касания | 44×44px | Все кнопки, ссылки, табы |
| Safe area bottom | `env(safe-area-inset-bottom)` | `AprilMobileShellBar`, bottom sheets |
| Content padding bottom | `aprilMobileShellBarContentPaddingBottom()` | Все страницы (чтобы контент не прятался за shell bar) |
| Font-size у инпутов | ≥16px | iOS Safari zoom prevention |

**CSS:**
```css
/* Safe area */
:root {
  --safe-area-bottom: env(safe-area-inset-bottom, 0px);
}

/* Content padding — чтобы контент не прятался за shell bar */
.page-content-mobile {
  padding-bottom: calc(56px + var(--safe-area-bottom));
}

/* iOS Safari — запрет zoom на инпуты */
input, textarea, select {
  font-size: 16px;
}
```

### I.5. Жесты

| Жест | Поведение |
|------|-----------|
| Android back button | Согласован с «Назад» в `AprilMobileShellBar`. Закрывает модалку → предыдущий экран → exit |
| iOS swipe-back | Включён. Работает с react-router `back()` |
| Pull-to-refresh | Только осознанно (каталог, нотификации). Не на всех страницах |
| Swipe-down (bottom sheet) | Закрытие `AprilVaulBottomSheet` |
| Swipe-back (bottom sheet) | Закрытие через shell bar «Назад» |

**Android back — реализация:**
```tsx
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'

function useAndroidBack(onClose?: () => void) {
  const navigate = useNavigate()

  useEffect(() => {
    const handler = () => {
      if (onClose) {
        onClose() // закрыть модалку/sheet
      } else {
        navigate(-1) // предыдущий экран
      }
    }
    window.addEventListener('popstate', handler)
    return () => window.removeEventListener('popstate', handler)
  }, [navigate, onClose])
}
```

---

## II. Forms-архитектура

### II.1. Бизнес-формы (регистрация, wizard шаги, обращения)

**Стек:** Zod + react-hook-form (peer deps DS).

**Паттерн:**
```tsx
// Пример: форма обращения в поддержку
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

const supportFormSchema = z.object({
  topic: z.enum(['ball', 'benefit', 'document', 'tech', 'other']),
  message: z.string().min(10).max(5000),
})

type SupportForm = z.infer<typeof supportFormSchema>

export function SupportForm() {
  const { register, handleSubmit, formState: { errors } } = useForm<SupportForm>({
    resolver: zodResolver(supportFormSchema),
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <Select {...register('topic')} error={errors.topic?.message} />
      <Textarea {...register('message')} error={errors.message?.message} />
      <Button type="submit">Отправить</Button>
    </form>
  )
}
```

**Валидация:**
- Клиентская: Zod schema (sync)
- Серверная: API validation (async, показываем ошибки Mantine `error` + `description`)

**Error display:**
```tsx
<Textarea
  {...register('message')}
  error={errors.message?.message}
  description={serverError?.message}
/>
```

### II.2. Admin-формы (CRUD карточек, провайдеров, правил)

**Два подхода:**

| Тип формы | Инструмент | Когда |
|-----------|-----------|-------|
| Простые (≤5 полей) | Zod + react-hook-form | Когда JSON Schema overkill |
| Сложные (динамические) | `AprilJsonSchemaForm` (RJSF + Ajv8) | Когда форма определяется JSON Schema |

**`AprilJsonSchemaForm` — пример:**
```tsx
// Admin форма редактирования провайдера
import { AprilJsonSchemaForm } from '@april/ui'

const providerSchema = {
  type: 'object',
  properties: {
    name: { type: 'string', title: 'Название' },
    apiUrl: { type: 'string', format: 'uri', title: 'API URL' },
    authType: { type: 'string', enum: ['bearer', 'api-key'] },
    retryPolicy: { type: 'object', properties: { maxRetries: { type: 'integer' } } },
  },
  required: ['name', 'apiUrl', 'authType'],
}

function ProviderEditForm() {
  return (
    <AprilJsonSchemaForm
      schema={providerSchema}
      formData={provider}
      onSubmit={handleSave}
    />
  )
}
```

**Решение:** использовать `AprilJsonSchemaForm` для admin-форм со сложной структурой. Простые формы — Zod + react-hook-form.

### II.3. Wizard-формы (DMS upgrade, MatCapital)

**Архитектура:**
- Каждая шаг — отдельная форма с Zod schema
- Состояние wizard — Zustand `useWizardsStore`
- Валидация на каждом шаге → блокировка forward
- Review-шаг → summary всех данных

**Zustand store:**
```ts
interface WizardState {
  activeWizard: string | null // wizard ID
  steps: Record<string, WizardStepData> // данные по шагам
  currentStep: number
  setStepData: (wizardId: string, stepId: string, data: WizardStepData) => void
  nextStep: () => boolean // false если валидация не прошла
  prevStep: () => void
  reset: (wizardId: string) => void
}

export const useWizardsStore = create<WizardState>((set, get) => ({
  activeWizard: null,
  steps: {},
  currentStep: 0,
  setStepData: (wizardId, stepId, data) => {
    set(state => ({
      steps: { ...state.steps, [`${wizardId}:${stepId}`]: data },
    }))
  },
  nextStep: () => {
    const { currentStep, activeWizard } = get()
    // Validate current step
    if (!validateStep(activeWizard, currentStep)) return false
    set(state => ({ currentStep: state.currentStep + 1 }))
    return true
  },
  // ...
}))
```

**JSON конфигурация wizard (ADR-019):**
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
  "onComplete": { "type": "api", "method": "POST", "url": "/api/v1/user-engagements" }
}
```

**Mobile:** wizard рендерится в `AprilVaulBottomSheet`. Прогресс — вертикальный («Шаг 2 из 4»).

### II.4. Survey-формы (M14)

**Backend-модель:** ADR-025 (Survey Engine, 633 строки). Frontend — только UI-слой.

**Компоненты вопросов:**

| Тип вопроса | Компонент | Валидация |
|-------------|-----------|-----------|
| Select (dropdown) | `SurveyQuestionSelect` | required |
| Radio (одиночный выбор) | `SurveyQuestionRadio` | required |
| Checkbox (множественный выбор) | `SurveyQuestionCheckbox` | min/max count |
| Text (свободный ответ) | `SurveyQuestionText` | required, min/max length |
| Rating (1-5 звёзд) | `SurveyQuestionRating` | required |
| Date | `SurveyQuestionDate` | required, min/max date |

**Бранчинг:**
```tsx
// Resolver определяет следующий вопрос на основе ответов
const nextQuestion = surveyResolver.getNextQuestion(currentQuestionId, answers)
// Backend: surveyResolver.GetNextQuestion() → вопрос с branching logic
```

**Optimistic submit:**
```tsx
// React Query mutation с optimistic update
const { mutate } = useMutation({
  mutationFn: submitSurveyAnswer,
  onMutate: async (answer) => {
    await queryClient.cancelQueries({ queryKey: ['survey', surveyId] })
    const previous = queryClient.getQueryData(['survey', surveyId])
    // Optimistic: показываем следующий вопрос сразу
    queryClient.setQueryData(['survey', surveyId], (old) => ({
      ...old,
      currentQuestion: nextQuestion,
      answers: [...old.answers, answer],
    }))
    return { previous }
  },
  onError: (err, answer, context) => {
    // Rollback
    queryClient.setQueryData(['survey', surveyId], context.previous)
  },
  onSettled: () => {
    queryClient.invalidateQueries({ queryKey: ['survey', surveyId] })
  },
})
```

**Mobile:** survey рендерится как полноэкранная страница (не модалка). Прогресс-бар сверху. `AprilMobileShellBar` — «Назад» + «Далее».

---

## III. Сводная таблица

| Сценарий | Desktop | Mobile | Формы |
|----------|---------|--------|-------|
| Страницы | Header + контент | `AprilMobileShellBar` + полноэкранные | — |
| Модалки | `AprilModal` | `AprilVaulBottomSheet` | — |
| Бизнес-формы | — | — | Zod + react-hook-form |
| Admin CRUD | — | — | `AprilJsonSchemaForm` / Zod |
| Wizard | `AprilModal` + Stepper | `AprilVaulBottomSheet` + вертикальный прогресс | Zod per step |
| Survey | Полноэкранная страница | Полноэкранная страница | Dynamic components + optimistic |
