# ТЗ для Design System Agent — @ukituki-ps/april-ui

> Прототип-референс: `артефакты/Прототип ЛК физика(1).html`

---

## Цель

Добавить в `@ukituki-ps/april-ui` универсальные компоненты, извлечённые из прототипа LKFL.
Компоненты — **white-label**, используют CSS tokens, не привязаны к бизнес-логике LKFL.

Принцип: **если компонент может использоваться в любом продукте на April DS — он в DS. Если он специфичен для LKFL — он в продукте.**

---

## Компоненты для добавления

### 1. AprilFilterPills

**Что это:** Горизонтальный ряд pill-кнопок-фильтров.

```tsx
type AprilFilterPillsProps = {
  items: { value: string; label: string }[];
  active: string;
  onChange: (value: string) => void;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа):**

| Токен | Значение |
|-------|----------|
| `padding` | 6px 14px |
| `border-radius` | 20px |
| `font-size` | 12px |
| `font-weight` | 600 |
| inactive | `border: 1.5px solid var(--border)`, `color: var(--text-muted)`, `background: transparent` |
| active | `background: primary`, `color: #fff`, `border-color: primary` |
| hover | `border-color: primary`, `color: primary` |

**Где используется в прототипе:** Каталог льгот (строка 230-237), транзакции (строка 276-282).

---

### 2. AprilWizard

**Что это:** Универсальный multi-step wizard с progress bar, body и footer.

```tsx
type AprilWizardStep = {
  id: string;
  label: string;
  content: ReactNode;
  footer?: ReactNode;
};

type AprilWizardProps = {
  steps: AprilWizardStep[];
  current: string;
  onChange: (step: string) => void;
  /** Показывать progress bar. Default true. */
  withProgress?: boolean;
  /** Контент wizard'а. */
  children?: ReactNode;
  /** Кастомный footer (кнопки назад/далее). Default — пустой. */
  footer?: ReactNode;
  'data-testid'?: string;
};
```

**Архитектура:**

```
┌──────────────────────────────┐
│ Wizard Progress (опционально) │  ← AprilWizardProgress
├──────────────────────────────┤
│ Step content (children)      │  ← slot, контент продукта
├──────────────────────────────┤
│ Footer (back / next / done)  │  ← slot, кнопки продукта
└──────────────────────────────┘
```

Wizard **не управляет state** (какой шаг активен) — это делает продукт. Wizard только рендерит.

---

### 3. AprilWizardProgress

**Что это:** Горизонтальный progress bar с шагами (круги + линии + label'ы).

```tsx
type AprilWizardProgressStep = {
  id: string;
  label: string;
  status: 'pending' | 'active' | 'done';
};

type AprilWizardProgressProps = {
  steps: AprilWizardProgressStep[];
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 362-379):**

| Элемент | Стили |
|---------|-------|
| Container | `padding: 14px 24px`, `background: var(--row)` |
| Step circle | `width: 24px`, `height: 24px`, `border-radius: 50%`, `font-size: 10px`, `font-weight: 700` |
| pending | `background: var(--border)`, `border: 2px solid var(--border)`, `color: var(--text-subtle)` |
| active | `border: 2px solid primary`, `color: primary`, `background: #fff`, `font-weight: 800` |
| done | `background: primary`, `border-color: primary`, `color: #fff` |
| Label | `font-size: 11px`, `font-weight: 600` — inactive: `--text-subtle`, active/done: `--text` |
| Line | `height: 2px`, `background: var(--border)`, `margin: 0 8px`, `min-width: 16px` — done: `background: primary` |

---

### 4. AprilStatCard

**Что это:** Карточка статистики с label, значением и hint.

```tsx
type AprilStatCardProps = {
  /** Label (например "Баланс баллов"). */
  label: ReactNode;
  /** Основное значение (например "1 250"). */
  value: ReactNode;
  /** Дополнительный текст под значением (например "+500 баллов в июне"). */
  hint?: ReactNode;
  /** Зелёная карточка (фон primary, белый текст). */
  variant?: 'default' | 'highlight';
  /** Иконка Lucide слева от label. */
  icon?: AprilLucideIcon;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 176-184):**

| Токен | Значение |
|-------|----------|
| `background` (default) | `var(--card)` |
| `background` (highlight) | `primary` |
| `border-radius` | `var(--radius-card)` (14px) |
| `padding` | 14px 18px |
| Label | `font-size: 11px`, `font-weight: 600`, `color: --text-subtle`, `text-transform: uppercase`, `letter-spacing: 0.6px` |
| Value | `font-size: 26px`, `font-weight: 800`, `letter-spacing: -0.5px` |
| Hint | `font-size: 11px`, `color: --text-subtle` |
| highlight label | `color: rgba(255,255,255,0.72)` |
| highlight value | `color: #fff` |
| highlight hint | `color: rgba(255,255,255,0.65)` |

---

### 5. AprilCard

**Что это:** Карточка с header (title + action link) и scrollable body.

```tsx
type AprilCardProps = {
  /** Заголовок в header. */
  title: ReactNode;
  /** Ссылка/кнопка справа в header (например "Весь каталог →"). */
  action?: ReactNode;
  /** Иконка Lucide слева от title. */
  icon?: AprilLucideIcon;
  children: ReactNode;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 55-78):**

| Токен | Значение |
|-------|----------|
| `background` | `var(--card)` |
| `border-radius` | `var(--radius-card)` (14px) |
| `box-shadow` | `var(--shadow-card)` (`0 1px 4px rgba(0,0,0,0.06)`) |
| Header | `padding: 16px 20px`, `border-bottom: 1px solid var(--row)` |
| Title | `font-size: 14px`, `font-weight: 700`, `display: flex`, `gap: 8px` |
| Title icon | `width: 15px`, `height: 15px`, `color: var(--text-muted)` |
| Action link | `font-size: 12px`, `font-weight: 600`, `color: primary` |

---

### 6. AprilBenefitRow

**Что это:** Строка в списке льгот (иконка + название + мета + badge).

```tsx
type AprilBenefitRowProps = {
  /** Иконка Lucide. */
  icon: AprilLucideIcon;
  /** Название льготы. */
  name: string;
  /** Мета-информация (провайдер, дата). */
  meta?: string;
  /** Бейдж справа. */
  badge?: ReactNode;
  /** Click handler. */
  onClick?: () => void;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 187-193):**

| Токен | Значение |
|-------|----------|
| `padding` | 13px 20px |
| `border-bottom` | `1px solid var(--row)` |
| `hover` | `background: primary-light` |
| Icon | `width: 38px`, `height: 38px`, `border-radius: 10px`, `background: var(--bg)`, `color: primary` |
| Icon inner | `width: 18px`, `height: 18px` |
| Name | `font-size: 13px`, `font-weight: 600` |
| Meta | `font-size: 11px`, `color: var(--text-subtle)` |

---

### 7. AprilEventRow

**Что это:** Строка в ленте событий (иконка + текст + время).

```tsx
type AprilEventRowProps = {
  /** Иконка Lucide. */
  icon: AprilLucideIcon;
  /** Цветовая схема иконки. */
  variant?: 'green' | 'yellow' | 'blue';
  /** Текст события. */
  text: ReactNode;
  /** Время. */
  time?: string;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 194-203):**

| Токен | Значение |
|-------|----------|
| `padding` | 11px 20px |
| `border-bottom` | `1px solid var(--row)` |
| Icon | `width: 30px`, `height: 30px`, `border-radius: 8px` |
| Icon inner | `width: 14px`, `height: 14px` |
| green | `background: #DCFCE7`, `color: #16A34A` |
| yellow | `background: #FEF9C3`, `color: #CA8A04` |
| blue | `background: #DBEAFE`, `color: #2563EB` |
| Text | `font-size: 12px`, `color: #374151`, `line-height: 1.5` |
| Time | `font-size: 10px`, `color: var(--text-subtle)` |

---

### 8. AprilQuickButton

**Что это:** Кнопка быстрого действия (иконка + текст, в grid).

```tsx
type AprilQuickButtonProps = {
  /** Иконка Lucide. */
  icon: AprilLucideIcon;
  /** Текст кнопки. */
  text: string;
  /** Click handler. */
  onClick?: () => void;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 204-218):**

| Токен | Значение |
|-------|----------|
| `background` | `var(--row)` |
| `border-radius` | 10px |
| `padding` | 12px |
| `hover` | `background: primary-light` |
| Icon container | `width: 32px`, `height: 32px`, `border-radius: 8px`, `background: #fff`, `color: primary`, `box-shadow: 0 1px 3px rgba(0,0,0,0.08)` |
| Icon inner | `width: 16px`, `height: 16px` |
| Text | `font-size: 11px`, `font-weight: 600`, `color: #374151`, `line-height: 1.35` |

---

### 9. AprilBalanceCard

**Что это:** Зелёная карточка баланса с прогресс-барами по категориям.

```tsx
type AprilBalanceCategory = {
  label: string;
  value: string;
  percentage: number;  // 0-100
};

type AprilBalanceCardProps = {
  /** Label (например "Доступный баланс"). */
  label: string;
  /** Основное значение (например "1 250"). */
  value: string;
  /** Подзаголовок (например "Следующее начисление: +500 в июне"). */
  subtitle?: string;
  /** Категории с прогресс-барами. */
  categories?: AprilBalanceCategory[];
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 265-274):**

| Токен | Значение |
|-------|----------|
| `background` | `primary` |
| `border-radius` | 14px |
| `padding` | 28px 24px |
| `color` | `#fff` |
| Label | `font-size: 12px`, `font-weight: 600`, `opacity: 0.75`, `text-transform: uppercase`, `letter-spacing: 0.5px` |
| Value | `font-size: 48px`, `font-weight: 800`, `letter-spacing: -2px` |
| Subtitle | `font-size: 12px`, `opacity: 0.7` |
| Category label | `font-size: 12px` — name: `opacity: 0.85`, value: `font-weight: 700` |
| Progress bar | `height: 6px`, `background: rgba(255,255,255,0.25)`, `border-radius: 3px` |
| Progress fill | `height: 100%`, `background: #fff`, `border-radius: 3px` |

---

### 10. AprilTransactionRow

**Что это:** Строка транзакции с иконкой, описанием и суммой.

```tsx
type AprilTransactionRowProps = {
  /** Иконка Lucide. */
  icon: AprilLucideIcon;
  /** Тип транзакции. */
  type: 'plus' | 'minus';
  /** Название транзакции. */
  name: string;
  /** Дата. */
  date?: string;
  /** Сумма. */
  amount: string;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 283-294):**

| Токен | Значение |
|-------|----------|
| `padding` | 13px 18px |
| `border-bottom` | `1px solid var(--row)` |
| Icon | `width: 36px`, `height: 36px`, `border-radius: 10px` |
| Icon inner | `width: 16px`, `height: 16px` |
| plus icon | `background: #DCFCE7`, `color: #16A34A` |
| minus icon | `background: var(--bg)`, `color: var(--text-subtle)` |
| Name | `font-size: 13px`, `font-weight: 600` |
| Date | `font-size: 11px`, `color: var(--text-subtle)` |
| Amount plus | `font-size: 14px`, `font-weight: 700`, `color: #16A34A` |
| Amount minus | `font-size: 14px`, `font-weight: 700`, `color: var(--text-muted)` |

---

### 11. AprilFaqItem

**Что это:** Аккордеон FAQ (вопрос + раскрывающийся ответ).

```tsx
type AprilFaqItemProps = {
  /** Вопрос. */
  question: string;
  /** Ответ (ReactNode — можно разметка). */
  answer: ReactNode;
  /** Управляемое состояние. */
  opened?: boolean;
  onOpenedChange?: (opened: boolean) => void;
  /** Начальное состояние. */
  defaultOpened?: boolean;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 320-332):**

| Токен | Значение |
|-------|----------|
| `border-bottom` | `1px solid var(--row)` |
| Question | `padding: 16px 20px`, `font-size: 13px`, `font-weight: 600` |
| Question hover | `color: primary` |
| Question active | `color: primary` |
| Chevron | `width: 16px`, `height: 16px`, `transition: transform 0.2s` — opened: `rotate(180deg)`, `color: primary` |
| Answer | `padding: 0 20px 16px`, `font-size: 13px`, `color: var(--text-muted)`, `line-height: 1.6` |

---

### 12. AprilOptionCard

**Что это:** Выбриваемая карточка опции (для wizard step 1).

```tsx
type AprilOptionCardProps = {
  /** Название опции. */
  name: string;
  /** Описание. */
  description?: string;
  /** Цена (правая часть). */
  price?: ReactNode;
  /** Выбрана ли опция. */
  selected?: boolean;
  onChange?: (selected: boolean) => void;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 387-398):**

| Токен | Значение |
|-------|----------|
| `border` | `2px solid var(--border)` |
| `border-radius` | 10px |
| `padding` | 14px 16px |
| hover | `border-color: primary`, `background: primary-light` |
| selected | `border-color: primary`, `background: primary-light` |
| Name | `font-size: 13px`, `font-weight: 700` |
| Description | `font-size: 12px`, `color: var(--text-muted)` |
| Price | `font-size: 14px`, `font-weight: 800`, `color: primary` |

---

### 13. AprilPayOptionCard

**Что это:** Карточка варианта оплаты (иконка + название + описание).

```tsx
type AprilPayOptionCardProps = {
  /** Иконка Lucide. */
  icon: AprilLucideIcon;
  /** Название. */
  name: string;
  /** Описание. */
  description: string;
  /** Выбран. */
  selected?: boolean;
  onChange?: (selected: boolean) => void;
  className?: string;
  'data-testid'?: string;
};
```

**Спецификация (из прототипа, строка 400-416):**

| Токен | Значение |
|-------|----------|
| `border` | `2px solid var(--border)` |
| `border-radius` | 10px |
| `padding` | 16px |
| hover | `border-color: primary` |
| selected | `border-color: primary`, `background: primary-light` |
| Icon | `width: 38px`, `height: 38px`, `border-radius: 10px`, `background: var(--bg)`, `color: primary` |
| Icon inner | `width: 18px`, `height: 18px` |
| Name | `font-size: 13px`, `font-weight: 700` |
| Description | `font-size: 12px`, `color: var(--text-muted)`, `line-height: 1.4` |

---

### 14. AprilFormInput, AprilFormTextarea, AprilFormSelect

**Что это:** Формы по дизайну прототипа.

```tsx
type AprilFormInputProps = {
  placeholder?: string;
  value?: string;
  onChange?: (value: string) => void;
  error?: string;
  label?: ReactNode;
  className?: string;
};
```

**Спецификация (из прототипа, строка 113-123):**

| Токен | Значение |
|-------|----------|
| `background` | `var(--bg)` |
| `border` | `1.5px solid var(--border)` |
| `border-radius` | 8px |
| `padding` | 10px 14px |
| `font-size` | 13px |
| focus | `border-color: primary` |
| error | `border-color: --danger` |
| Label | `font-size: 12px`, `font-weight: 600`, `color: var(--text-muted)` |

---

### 15. AprilConfirmCheckbox

**Что это:** Чекбокс подтверждения с label.

```tsx
type AprilConfirmCheckboxProps = {
  checked?: boolean;
  onChange?: (checked: boolean) => void;
  label: ReactNode;
  className?: string;
};
```

**Спецификация (из прототипа, строка 426-428):**

| Токен | Значение |
|-------|----------|
| Checkbox | `width: 16px`, `height: 16px`, `accent-color: primary` |
| Label | `font-size: 12px`, `color: var(--text-muted)`, `line-height: 1.5` |

---

### 16. AprilSuccessScreen

**Что это:** Экран успешного завершения wizard.

```tsx
type AprilSuccessScreenProps = {
  /** Заголовок. */
  title: string;
  /** Описание. */
  description?: ReactNode;
  /** Кастомная иконка. Default — Check (зелёный круг). */
  icon?: AprilLucideIcon;
  className?: string;
};
```

**Спецификация (из прототипа, строка 431-439):**

| Токен | Значение |
|-------|----------|
| Icon container | `width: 64px`, `height: 64px`, `background: #DCFCE7`, `border-radius: 50%`, `color: primary` |
| Icon inner | `width: 32px`, `height: 32px` |
| Title | `font-size: 18px`, `font-weight: 800` |
| Description | `font-size: 13px`, `color: var(--text-muted)`, `line-height: 1.5` |
| Layout | `text-align: center`, `padding: 20px 0` |

---

### 17. AprilConfirmDoc

**Что это:** Блок превью документа для confirmation step.

```tsx
type AprilConfirmDocProps = {
  /** Заголовок документа. */
  title: string;
  /** Содержимое документа. */
  content: ReactNode;
  className?: string;
};
```

**Спецификация (из прототипа, строка 419-424):**

| Токен | Значение |
|-------|----------|
| `background` | `var(--row)` |
| `border` | `1px solid var(--border)` |
| `border-radius` | 10px |
| `padding` | 16px |
| `font-size` | 12px |
| `color` | `var(--text-muted)` |
| `line-height` | 1.6 |
| Title | `color: var(--text)`, `font-size: 13px`, `font-weight: 700` |

---

## Не добавлять в DS (продуктовые компоненты)

Следующие элементы специфичны для LKFL и реализуются в продукте, не в DS:

| Компонент | Почему не в DS |
|-----------|----------------|
| BenefitCard (каталог) | Привязан к домену льгот (provider, price, badge) |
| PolicyCard (полис) | Специфичен для ДМС (градиент, номер полиса, мета) |
| ClinicMap | Привязан к гео-данным, iframe карты |
| ClinicList | Привязан к домену клиник |
| MatAmount (мат. капитал) | Бизнес-логика LKFL |
| MatInfoBanner | Контент-зависимый |
| DMSConditions (список условий) | Контент-зависимый |

Эти компоненты LKFL строит **на основе** компонентов DS (AprilCard, AprilBenefitRow, AprilBadge и т.д.).

---

## CSS tokens (если не существуют)

Убедиться, что в `@ukituki-ps/april-tokens` есть следующие CSS переменные:

```css
:root {
  /* Если нет — добавить */
  --april-row:          #F9FAFB;
  --april-border:       #EBEBEB;
  --april-text-subtle:  #9CA3AF;
  --april-radius-card:  14px;
  --april-radius-btn:   6px;
  --april-shadow-card:  0 1px 4px rgba(0,0,0,0.06);
}
```

**Важно:** tokens используют **neutral** названия (`--april-row`, не `--sdek-bg-gray`). Бренд-цвет (`#00B33C`) приходит через Mantine `primaryColor` (который продукт переопределяет).

---

## Lucide иконки

Все иконки — `lucide-react` (уже зависимость april-ui через `AprilIcon`).

Иконки, используемые в прототипе:

```
bell, brain, calendar, check, check-circle, chevron-right, circle,
coins, coffee, dumbbell, gift, graduation-cap, heart-pulse,
info, languages, lock, map-pin, plus, plus-circle, search,
shield-check, shield-plus, shopping-bag, smartphone, sparkle,
user-plus, users, arrow-up-circle, baby, clock, smart-speaker,
x, x-circle, file-text, download, send
```

Если какой-то иконки нет в текущем export — добавить.

---

## KPI

- [ ] Все 17 компонентов экспортированы из `@ukituki-ps/april-ui`
- [ ] TypeScript types (`index.d.ts`) — полные пропсы, без `any`
- [ ] Все компоненты используют `AprilIcon` (не сырой Lucide)
- [ ] Все размеры/цвета — через CSS tokens (не hardcoded)
- [ ] Primary color — через Mantine `primaryColor` (не `#00B33C`)
- [ ] Компоненты работают в `comfortable` и `compact` density
- [ ] `data-testid` проп на каждом компоненте
- [ ] semver bump: minor (добавление новых публичных компонентов)
