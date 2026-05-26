# ТЗ для April Design System — компоненты ЛК физлица

> **M15 T1508:** АDR актуализирован. Было 11 «недостающих» компонентов → осталось 2 реальных gap'а.
> Остальные 9 компонентов покрываются существующими DS/Mantine компонентами или реализуются в LKFL.

**Дата:** 2026-05-25 (обновлено: 2026-05-26, M15 T1508)
**Автор:** Architect (LKFL)
**Прототип:** `артефакты/Прототип ЛК физика(1).html`
**Цель:** определить недостающие компоненты `@ukituki-ps/april-ui` для покрытия прототипа ЛК физлица

---

## 1. Gap-анализ: прототип vs текущая DS

### 1.1. Покрытие — что уже есть

| Элемент прототипа | Компонент DS | Статус |
|---|---|---|
| Карточки льгот (catalog grid) | `BenefitCard` | ✅ готов |
| Быстрые действия (dashboard) | `QuickActionsGrid` | ✅ готов |
| Прогресс-бары (points) | `GamifiedProgress` | ✅ готов |
| Бейджи достижений | `AchievementBadge` | ✅ готов |
| Пакеты льгот (DMS upgrade) | `SmartBundle` | ✅ готов |
| Фильтры + поиск каталога | `FacetedSearch` | ✅ готов |
| Пустые состояния | `EmptyStateIllustration` | ✅ готов |
| Модальные окна | `AprilModal` | ✅ готов |
| Статус-чипы | `StatusChip` | ✅ готов |
| Навигация sidebar (admin) | `ProductSidebarNavigation` | ✅ готов |
| Header toolbar (admin) | `ProductHeaderToolbar` | ✅ готов |
| Mobile bottom sheet | `AprilVaulBottomSheet` | ✅ готов |
| Mobile shell bar | `AprilMobileShellBar` | ✅ готов |
| Иконки | `AprilIcon` + `AprilIcon*` | ✅ готов |
| Card list column (admin CRUD) | `CardListColumn` | ✅ готов |
| Gradient segmented control | `AprilGradientSegmentedControl` | ✅ готов |
| JSON Schema Form (admin) | `AprilJsonSchemaForm` | ✅ готов |
| JSON Tree Editor (admin) | `AprilJsonTreeEditor` | ✅ готов |

### 1.2. Промежуточное покрытие — покрыто существующими компонентами

| Элемент прототипа | Было (ADR-029) | Стало (M15) | Решение |
|---|---|---|---|
| Табы в модалке льготы | `AprilTab` (нет) | Mantine `Tabs` + DS `AprilGradientSegmentedControl` | Mantine `Tabs` — достаточно |
| Стат-карточки dashboard | `StatCard` (нет) | Mantine `Paper` + токены | **LKFL-компонент** `components/StatCard.tsx` — простой, не нужен в DS |
| Навигация ЛК (top tabs) | `TopTabNavigation` (нет) | `ProductHeaderToolbar` + `AprilGradientSegmentedControl` | Использовать DS-компоненты |
| Баланс в header | `BalancePill` (нет) | Mantine `Badge` + токены | **LKFL-компонент** `components/BalancePill.tsx` — простой pill |

### 1.3. Реальные gap'ы — 2 компонента

| # | Компонент | Где используется | Решение |
|---|---|---|---|
| GAP-001 | `WizardContainer` (обёртка) | DMS upgrade, DMS relative, MatCapital — мультишаговые формы в модалках | **Обёртка в LKFL:** `modals/Wizard.tsx` над `Mantine Stepper` + `AprilModal`. Не добавлять в DS — специфичен для LKFL (JSON-driven renderer, ADR-019) |
| GAP-002 | `TransactionList` (компонент) | «Мои баллы» — история транзакций с иконками, фильтрами, суммами | **Компонент в LKFL:** `components/TransactionRow.tsx` + Mantine `Table`. Простой — не нужен в DS |

### 1.4. Компоненты, отменённые после пересмотра

| # | Компонент (было) | Статус | Причина отмены |
|---|---|---|---|
| DS-003 | `StatCard` | ❌ Отменён | Mantine `Paper` + токены. Прототип — простая карточка с числом. **YAGNI:** делать в LKFL как `components/StatCard.tsx` |
| DS-004 | `EventsFeed` | ❌ Отменён | `CardListColumn` покрывает список событий |
| DS-005 | `PolicyCard` | ❌ Отменён | Mantine `Card` + токены + градиент через CSS. Модалка ДМС — специфична для LKFL |
| DS-006 | `ClinicMapList` | ❌ Отменён | iframe карта + Mantine `Card` + список клиник. Не DS-компонент |
| DS-007 | `TopTabNavigation` | ❌ Отменён | `ProductHeaderToolbar` + `AprilGradientSegmentedControl` покрывают |
| DS-008 | `BalancePill` | ❌ Отменён | Mantine `Badge`/`Pill` + токены. Делать в LKFL как `components/BalancePill.tsx` |
| DS-009 | `DocumentRow` | ❌ Отменён | Mantine `Table` + токены. Делать в LKFL как `components/DocumentRow.tsx` |
| DS-010 | `SupportFAQ` | ❌ Отменён | Mantine `Accordion` — уже в прототипе (ADR-007) |
| DS-011 | `FilterPills` | ❌ Отменён | Mantine `Badge` + DS `FacetedSearch`. Делать в LKFL как `components/FilterPills.tsx` |

---

## 2. Компоненты LKFL vs DS — критерий решения

**Правило:** добавлять в DS (`@ukituki-ps/april-ui`) только если компонент используется в 2+ продуктах April экосистемы.

| Критерий | В DS | В LKFL |
|----------|------|--------|
| Используется в 2+ продуктах April | ✅ Да | ❌ Нет |
| Специфичен для льготной платформы | ❌ Нет | ✅ Да |
| Простой визуальный компонент (≤300 строк) | ❌ Нет | ✅ Да |
| Нужен в 2+ продуктах | ✅ Да | — |
| Универсальный паттерн (modal, card list, wizard) | ✅ Да | — |

**Компоненты LKFL (не в DS):**

| Компонент | Файл | Основание |
|-----------|------|-----------|
| `StatCard` | `components/StatCard.tsx` | Простой: Mantine `Paper` + токены + CSS (≤50 строк) |
| `BalancePill` | `components/BalancePill.tsx` | Простой: Mantine `Badge` + токены (≤30 строк) |
| `TransactionRow` | `components/TransactionRow.tsx` | Специфичен для льготной платформы (типы транзакций) |
| `DocumentRow` | `components/DocumentRow.tsx` | Специфичен (документы льготной платформы) |
| `FilterPills` | `components/FilterPills.tsx` | Простой: Mantine `Badge` + flex layout (≤40 строк) |
| `Wizard` | `modals/Wizard.tsx` | JSON-driven renderer — специфичен для LKFL (ADR-019) |
| `EventRow` | `components/EventRow.tsx` | Простой: icon + text + time (≤30 строк) |
| `QuickActionBtn` | `components/QuickActionBtn.tsx` | Простой: icon + text button (≤30 строк) |

---

## 2. Детальное ТЗ по каждому компоненту

> **⚠️ М15 T1508: этот раздел устарел.** Детальные спецификации DS-003 → DS-011 отменены — компоненты реализуются в LKFL или покрываются существующими DS/Mantine.
> Актуальные gap'ы: только GAP-001 (Wizard обёртка в LKFL) и GAP-002 (TransactionList компонент в LKFL).
> Спецификации DS-001 и DS-002 сохранены как историческая справка — реализации перенесены в LKFL.

---

### DS-001: `WizardContainer`

**Назначение:** универсальный контейнер для многошаговых форм в модальных окнах.

**Сценарии:**
- Апгрейд ДМС (4 шага: Опция → Оплата → Подтверждение → Готово)
- Добавление родственника к ДМС (4 шага)
- Материнский капитал (4 шага: Условия → Данные → Подтверждение → Готово)

**Layout:**
```
┌──────────────────────────────────────┐
│ WizardContainer                      │
│  ┌─ WizardProgress (sticky top)      │
│  │  ●─── ○─── ○─── ○               │
│  │  "Опция" "Оплата" "Подтв" "Готово"│
│  └───────────────────────────────────┘
│  ┌─ WizardStepContent (scrollable)   │
│  │  <контент текущего шага>          │
│  └───────────────────────────────────┘
│  ┌─ WizardFooter (sticky bottom)     │
│  │  [← НАЗАД]         [ДАЛЕЕ →]     │
│  └───────────────────────────────────┘
└──────────────────────────────────────┘
```

**Пропсы:**
```typescript
export interface WizardStepConfig {
  /** Уникальный ID шага. */
  id: string;
  /** Заголовок шага (в прогресс-баре). */
  label: string;
  /** Контент шага. */
  content: ReactNode;
  /** Валидация шага (optional). Возвращает boolean. */
  validate?: () => boolean;
}

export interface WizardContainerProps {
  /** Массив конфигураций шагов. */
  steps: WizardStepConfig[];
  /** Текущий шаг (контролируемый). */
  currentStep?: number;
  /** Начальный шаг (неконтролируемый). */
  initialStep?: number;
  /** Callback при смене шага. */
  onChange?: (step: number) => void;
  /** Текст кнопки "Далее". */
  nextLabel?: string;
  /** Текст кнопки "Назад". */
  backLabel?: string;
  /** Текст финальной кнопки (последний шаг). */
  finalLabel?: string;
  /** Кастомный футер (перекрывает стандартные кнопки). */
  footer?: ReactNode;
  /** Показать прогресс-бар. По умолчанию true. */
  showProgress?: boolean;
  /** Callback при достижении финального шага. */
  onFinalStep?: () => void;
}
```

**WizardProgress (подкомпонент):**
- Горизонтальная линия с кружками (номера шагов)
- Состояния: `done` (зелёный круг + белая цифра), `active` (обводка зелёная + цифра), `pending` (серый круг)
- Линии-соединители: зелёные между пройденными шагами, серые — впереди
- Мобильный fallback: при <480px — вертикальный прогресс или компактный "Шаг 2 из 4"

**WizardFooter (подкомпонент):**
- Back button: скрыт на шаге 1, виден на 2+
- Next button: меняет текст на последнем шаге → "ЗАКРЫТЬ" / "ОТПРАВИТЬ ЗАЯВКУ"
- Иконки: ← (back), → (next), ✓ (подтверждение на пред-финальном шаге)

**a11y:**
- `aria-label` на каждом шаге
- Tab order: кнопки footer в конце
- Escape → назад или закрытие (конфигурируемо)

**Импорт:** `import { WizardContainer } from '@april/ui'`

**Тесты:**
- Smoke: рендер N шагов, навигация forward/back
- Unit: валидация шага блокирует переход вперёд
- a11y: фокус на кнопках, aria-label

---

### DS-002: `TransactionList`

**Назначение:** список транзакций начислений/списаний баллов.

**Прототип:** раздел «Мои баллы», правая колонка. Каждая строка: иконка (цвет зелёный/серый) + описание + дата + сумма (цветовая: зелёная для начислений, серая для списаний).

**Layout строки:**
```
┌──────────────────────────────────────────────────────┐
│  [icon]  Описание                    ±1 250 б        │
│           Дата                          цвет: +/-    │
└──────────────────────────────────────────────────────┘
```

**Пропсы:**
```typescript
export type TransactionType = 'credit' | 'debit';

export interface TransactionItem {
  id: string;
  type: TransactionType;
  /** Название (например «Ежемесячное начисление»). */
  name: string;
  /** Сумма с знаком (например "+500 б", "−600 б"). */
  amount: string;
  /** Дата (форматированная строка). */
  date: string;
  /** Иконка Lucide (опционально, по умолчанию: plus-circle / минус). */
  icon?: ReactNode;
  /** Категория (для фильтрации). */
  category?: string;
}

export interface TransactionListProps {
  transactions: TransactionItem[];
  /** Фильтр: 'all' | 'credit' | 'debit'. */
  filter?: 'all' | 'credit' | 'debit';
  onFilterChange?: (filter: 'all' | 'credit' | 'debit') => void;
  /** Показать фильтры-табы сверху. */
  showFilters?: boolean;
  /** Callback при клике на строку. */
  onRowClick?: (item: TransactionItem) => void;
  /** Пустое состояние. */
  empty?: ReactNode;
}
```

**Визуальные требования:**
- Icon container: 36×36px, borderRadius 10px
  - credit: `#DCFCE7` bg, `#16A34A` icon color
  - debit: `var(--bg)` bg, `var(--text-subtle)` icon color
- Row hover: subtle background change
- Border-bottom между строками (последняя — без)
- Сумма: fontWeight 700, цвет по типу
- Фильтры: pill-кнопки (all / plus / minus), active — чёрный фон + белый текст

**a11y:**
- `role="list"` + `role="listitem"`
- `aria-label` на иконках
- Screen reader: читает тип, название, дату, сумму

**Импорт:** `import { TransactionList } from '@april/ui'`

---

### DS-003: `StatCard`

**Назначение:** компактная карточка с числовой метрикой для Dashboard.

**Прототип:** 3 карточки вверху Dashboard:
- «Баланс баллов» → 1 250 (зелёный фон, белый текст)
- «Активных льгот» → 4 (белый фон, тёмный текст)
- «До конца периода» → 47 дн (белый фон)

**Layout:**
```
┌──────────────────────────┐
│ [icon] Баланс баллов      │
│ 1 250                    │
│ +500 баллов в июне       │
└──────────────────────────┘
```

**Пропсы:**
```typescript
export type StatCardVariant = 'default' | 'accent';

export interface StatCardProps {
  /** Название метрики. */
  label: string;
  /** Числовое значение (строка для гибкости форматирования). */
  value: string;
  /** Иконка Lucide (опционально). */
  icon?: ReactNode;
  /** Подсказка / хинт под значением. */
  hint?: string;
  /** Вариант: 'default' (белый фон) или 'accent' (зелёный фон, белый текст). */
  variant?: StatCardVariant;
}
```

**Визуальные требования:**
- borderRadius: 14px, boxShadow: `0 1px 4px rgba(0,0,0,0.06)`
- Label: 11px, fontWeight 600, uppercase, letter-spacing 0.6px, иконка слева
- Value: 26px, fontWeight 800, letter-spacing -0.5px
- Hint: 11px, muted color
- Accent variant: `background: var(--green)`, all text white (label/hint — opacity 0.72)
- Hover: none (не кликабельно по умолчанию)

**Импорт:** `import { StatCard } from '@april/ui'`

---

### DS-004: `EventsFeed`

**Назначение:** лента событий/уведомлений на Dashboard.

**Прототип:** раздел «Лента событий» в Dashboard, 3 события: начисление баллов (зелёная иконка), ожидание подтверждения (жёлтая), новые льготы (синяя).

**Layout строки:**
```
┌──────────────────────────────────────────────┐
│  [icon]  Текст события с <b>акцентами</b>    │
│           Время                              │
└──────────────────────────────────────────────┘
```

**Пропсы:**
```typescript
export type EventIconVariant = 'success' | 'warning' | 'info';

export interface EventItem {
  id: string;
  /** Тип иконки (определяет цвет контейнера). */
  variant: EventIconVariant;
  /** Иконка Lucide. */
  icon: ReactNode;
  /** Текст события (ReactNode для поддержки bold/italic). */
  text: ReactNode;
  /** Время (например «Сегодня, 10:24»). */
  time: string;
}

export interface EventsFeedProps {
  events: EventItem[];
  /** Callback при клике на событие. */
  onEventClick?: (event: EventItem) => void;
  /** Пустое состояние. */
  empty?: ReactNode;
  /** Максимальное количество отображаемых событий. */
  maxItems?: number;
}
```

**Визуальные требования:**
- Icon container: 30×30px, borderRadius 8px
  - success: `#DCFCE7` bg, `#16A34A` icon
  - warning: `#FEF9C3` bg, `#CA8A04` icon
  - info: `#DBEAFE` bg, `#2563EB` icon
- Text: 12px, lineHeight 1.5, `<b>` — акцентированный текст
- Time: 10px, subtle color
- Border-bottom между строками

**Импорт:** `import { EventsFeed } from '@april/ui'`

---

### DS-005: `PolicyCard`

**Назначение:** визуализация полиса ДМС (или аналогичного документа) в модалке льготы.

**Прототип:** модалка «ДМС — Базовая» → таб «Мой полис». Градиентная зелёная карточка с номером полиса, застрахованным, сроком действия, программой.

**Layout:**
```
┌────────────────────────────────────────┐
│ POLICY ДМС                             │
│ 7740 8821 0034 5512                   │
│                                        │
│ Застрахованный     Действует до   Программа│
│ Алмазов А. А.      31.12.2025     Базовая │
└────────────────────────────────────────┘
[Скачать полис] [Поделиться]
```

**Пропсы:**
```typescript
export interface PolicyCardProps {
  /** Тип полиса (например «Полис ДМС»). */
  type: string;
  /** Номер полиса. */
  policyNumber: string;
  /** Данные полиса (ключ → значение). */
  fields: { label: string; value: string }[];
  /** Градиент фона (по умолчанию: teal gradient). */
  gradient?: 'teal' | 'green' | string;
  /** Показать кнопки действий. */
  showActions?: boolean;
  /** Callback "Скачать". */
  onDownload?: () => void;
  /** Callback "Поделиться". */
  onShare?: () => void;
  /** Описание под карточкой. */
  description?: ReactNode;
}
```

**Визуальные требования:**
- Фон: `linear-gradient(135deg, #00B33C 0%, #007A28 100%)` (brand-green overrideable)
- borderRadius: 14px, padding: 20px 22px
- Номер: 20px, fontWeight 800, letterSpacing 2px
- Fields: 3 колонки (flex), label — 10px uppercase opacity 0.7, value — 13px fontWeight 700
- Кнопки: `btn-primary` / `btn-outline` (или Mantine equivalent)

**Импорт:** `import { PolicyCard } from '@april/ui'`

---

### DS-006: `ClinicMapList`

**Назначение:** комбинация карты + список клиник рядом.

**Прототип:** модалка ДМС → таб «Клиники на карте». iframe OpenStreetMap сверху, список клиник снизу.

**Layout:**
```
┌────────────────────────────────────────┐
│ [карта — iframe, высота 220px]        │
├────────────────────────────────────────┤
│ 📍 Клиника АльфаМед — Центральная     │
│    ул. Тверская, 18к2 · Пн–Сб 8:00–21 │
├────────────────────────────────────────┤
│ 📍 МЦ «СМ-Клиника» Войковская         │
│    ул. Клары Цеткин, 33к28 · Пн–Вс ... │
└────────────────────────────────────────┘
```

**Пропсы:**
```typescript
export interface ClinicItem {
  id: string;
  name: string;
  address: string;
  schedule?: string;
  lat?: number;
  lng?: number;
}

export interface ClinicMapListProps {
  /** Клиники для отображения. */
  clinics: ClinicItem[];
  /** URL карты (iframe src). Если не задан — рендерит только список. */
  mapUrl?: string;
  /** Высота карты. По умолчанию 220px. */
  mapHeight?: string;
  /** Callback при клике на клинику. */
  onClinicClick?: (clinic: ClinicItem) => void;
  /** Показать список под картой. По умолчанию true. */
  showList?: boolean;
}
```

**Визуальные требования:**
- Карта: borderRadius 12px, overflow hidden, height 220px
- Список: gap 6px, padding снизу
- Item: flex, icon `map-pin`, bg `var(--row)`, borderRadius 8px, padding 9px 12px
- Name: 12px fontWeight 600, Address: 11px muted

**Замечание:** этот компонент может быть заменён на `react-map-gl` или `@vis.gl/react-google-maps` в будущем. Сейчас — iframe.

**Импорт:** `import { ClinicMapList } from '@april/ui'`

---

### DS-007: `TopTabNavigation`

**Назначение:** горизонтальная навигация в верхнем header (альтернатива sidebar для ЛК физлица).

**Прототип:** sticky header с горизонтальными табами (Главная, Каталог льгот, Мои баллы, Документы, Поддержка). Active tab — чёрный текст + зелёная подчёркивающая линия снизу (2px).

**Layout:**
```
┌─ logo ─────────────────────────────────────────────────────┐
│ Главная | Каталог льгот | Мои баллы | Документы | Поддержка │
│                                        [balance] [bell] [avatar] │
└────────────────────────────────────────────────────────────┘
```

**Пропсы:**
```typescript
export interface TopTabItem {
  /** Уникальный key (router path). */
  key: string;
  /** Заголовок таба. */
  label: string;
  /** Иконка Lucide (опционально). */
  icon?: ReactNode;
}

export interface TopTabNavigationProps {
  /** Табы. */
  items: TopTabItem[];
  /** Активный таб (контролируемый). */
  activeKey: string;
  /** Callback при смене таба. */
  onChange: (key: string) => void;
  /** Правая часть (balance pill, уведомления, аватар). */
  rightSection?: ReactNode;
  /** Логотип слева (если есть). */
  logo?: ReactNode;
  /** Высота панели. По умолчанию 58px. */
  height?: string | number;
  /** sticky positioning. По умолчанию true. */
  sticky?: boolean;
}
```

**Визуальные требования:**
- Sticky top: 0, z-index 100
- borderBottom: 1px solid `var(--border)`
- Tab: padding 0 14px, height 58px, fontSize 13px, fontWeight 500, muted color
- Active tab: `color: var(--text)`, `borderBottom: 2px solid var(--green)`, fontWeight 600
- Hover: `color: var(--green)`
- Flex layout: logo left, tabs center (flex: 1), rightSection margin-left: auto

**Мобильный fallback:**
- <768px: скрыть тексты табов → только иконки
- <480px: заменить на `AprilMobileShellBar` с табами в нижней панели

**Замечание:** если продукт выбирает sidebar (`ProductSidebarNavigation`), этот компонент не нужен. Решение зависит от финального UX-решения для ЛК.

**Импорт:** `import { TopTabNavigation } from '@april/ui'`

---

### DS-008: `BalancePill`

**Назначение:** компактный pill-индикатор баланса в header.

**Прототип:** `[🪙 1 250 баллов]` — зелёный фон, зелёная граница, тёмно-зелёный текст.

**Пропсы:**
```typescript
export interface BalancePillProps {
  /** Числовое значение. */
  value: number | string;
  /** Подпись единицы (например «баллов»). */
  unit?: string;
  /** Иконка Lucide. По умолчанию: coins. */
  icon?: ReactNode;
  /** Callback при клике. */
  onClick?: () => void;
}
```

**Визуальные требования:**
- bg: `#F0FDF4` (green-light), border: 1px solid `#BBF7D0` (green-border)
- borderRadius: 20px, padding: 5px 13px
- fontSize: 12px, fontWeight 700, color: `#166534`
- Icon: 13×13px, color: `#00B33C`
- Cursor: pointer (если onClick)

**Импорт:** `import { BalancePill } from '@april/ui'`

---

### DS-009: `DocumentRow`

**Назначение:** строка документа в таблице раздела «Документы».

**Прототип:** таблица с колонками: Документ, Тип, Дата, Статус, Действие (скачать). Каждая строка: название + мета, badge типа, дата, badge статуса, кнопка скачивания.

**Замечание:** можно использовать стандартную Mantine `Table`. `DocumentRow` — обёртка для визуального стиля строки.

**Пропсы:**
```typescript
export interface DocumentRowData {
  id: string;
  name: string;
  meta?: string;
  type: string;
  typeBadge?: 'blue' | 'gray' | 'green';
  date: string;
  status: string;
  statusBadge: 'green' | 'yellow' | 'blue' | 'gray';
  /** Доступна загрузка. */
  downloadable?: boolean;
}

export interface DocumentRowProps {
  document: DocumentRowData;
  onDownload?: (document: DocumentRowData) => void;
}
```

**Визуальные требования:**
- Название: 13px fontWeight 600, Мета: 11px subtle
- Badge: как в прототипе (blue/gray/green/yellow)
- Кнопка download: inline-flex, padding 6px 12px, bg `var(--row)`, border, borderRadius 6px
- Hover на кнопке: `var(--green-light)` bg, `var(--green)` text/border
- Row hover: bg `var(--row)`

**Импорт:** `import { DocumentRow } from '@april/ui'`

---

### DS-010: `SupportFAQ`

**Назначение:** FAQ аккордеон для раздела «Поддержка».

**Прототип:** 6 вопросов-ответов. Клик → раскрытие ответа + поворот иконки chevron. Один вопрос открыт одновременно.

**Замечание:** Mantine `Accordion` покрывает базовый функционал. Нужен `SupportFAQ` только если нужны кастомные иконки, hover-эффекты на вопрос (цвет → green) или специфичный layout.

**Пропсы:**
```typescript
export interface FAQItem {
  id: string;
  question: string;
  answer: ReactNode;
}

export interface SupportFAQProps {
  items: FAQItem[];
  /** Максимум открытых вопросов одновременно. По умолчанию 1. */
  multiple?: boolean;
  /** Иконка. По умолчанию: chevron-down. */
  icon?: ReactNode;
  /** Callback при раскрытии. */
  onChange?: (openedId: string | undefined) => void;
}
```

**Можно ли обойтись без нового компонента:** Да. Mantine `Accordion` с кастомным `chevron` через проп `chevron` полностью покрывает. Если команда DS решит, что кастомизация прототипа достаточно специфична — создать `SupportFAQ` как обёртку над `Accordion`.

**Приоритет:** 🟢 низкий. Реализовать в продукте на Mantine `Accordion`, если DS решит не делать обёртку.

**Импорт:** `import { SupportFAQ } from '@april/ui'` (опционально)

---

### DS-011: `FilterPills`

**Назначение:** inline фильтр-пили (категории) для каталога льгот.

**Прототип:** горизонтальный ряд кнопок-пилий: «Все», «ДМС», «Спорт», «Питание», «Развитие», «Мерч». Active — зелёный фон + белый текст.

**Пропсы:**
```typescript
export interface FilterPillItem {
  value: string;
  label: string;
  count?: number;
}

export interface FilterPillsProps {
  pills: FilterPillItem[];
  activeValue: string;
  onChange: (value: string) => void;
  /** Горизонтальный скролл при переполнении. По умолчанию true. */
  scrollable?: boolean;
}
```

**Визуальные требования:**
- Pill: padding 6px 14px, borderRadius 20px, fontSize 12px, fontWeight 600
- Inactive: border 1.5px solid `var(--border)`, bg `var(--card)`, color `var(--text-muted)`
- Active: bg `var(--green)`, color `#fff`, border-color `var(--green)`
- Hover (inactive): transition к зелёному
- Container: flex, gap 6px, wrap (или scrollable)

**Можно ли обойтись без нового компонента:** `FacetedSearch` уже имеет `ActiveFiltersPills`, но это badge с крестиком, а не интерактивные filter pills. Нужен отдельный компонент.

**Импорт:** `import { FilterPills } from '@april/ui'`

---

## 3. Приоритизация и порядок реализации

### Фаза 1 (M09-1) — критическая реализация

| Компонент | Причина |
|---|---|
| `StatCard` (DS-003) | Dashboard — первый экран пользователя |
| `FilterPills` (DS-011) | Каталог — ключевой сценарий поиска льгот |
| `TransactionList` (DS-002) | «Мои баллы» — обязательная страница |
| `BalancePill` (DS-008) | Header — видим на каждом экране |

### Фаза 2 (M09-2) — wizards + модалки льгот

| Компонент | Причина |
|---|---|
| `WizardContainer` (DS-001) | DMS upgrade/relative, MatCapital — ключевые бизнес-сценарии |
| `PolicyCard` (DS-005) | Модалка ДМС — полис пользователя |
| `ClinicMapList` (DS-006) | Модалка ДМС — клиники на карте |

### Фаза 3 (M09-3) — полировка

| Компонент | Причина |
|---|---|
| `EventsFeed` (DS-004) | Dashboard — вторичная информация |
| `TopTabNavigation` (DS-007) | Альтернативный header паттерн (если не sidebar) |
| `DocumentRow` (DS-009) | Раздел «Документы» |
| `SupportFAQ` (DS-010) | Раздел «Поддержка» — можно на Mantine Accordion |

---

## 4. Требования ко всем новым компонентам

### 4.1. Общие стандарты (согласно DESIGN_SYSTEM.md)

- **Density:** поддержка comfortable/compact через `useDensity()`
- **Theme:** light/dark через `useMantineColorScheme()`
- **Mobile:** корректное поведение на <768px
- **a11y:** видимый фокус, aria-label, tab order, WCAG 2.1 AA контраст
- **Icons:** только `lucide-react` через `AprilIcon`
- **Colors:** через Mantine theme (teal, gray, dark) + brand override
- **Экспорт:** тип + компонент в `packages/ui/src/index.ts`
- **Витрина:** секция в `UIKit` showcase
- **Тесты:** smoke-тест минимален, unit-тест для критической логики

### 4.2. Структура файлов каждого компонента

```
packages/ui/src/components/
├── StatCard.tsx
├── StatCardSection.tsx         ← витрина в UIKit
├── TransactionList.tsx
├── TransactionListSection.tsx
├── WizardContainer.tsx
├── WizardContainerSection.tsx
├── WizardProgress.tsx          ← подкомпонент
├── WizardFooter.tsx            ← подкомпонент
...
```

### 4.3. Реэкспорт

Каждый новый компонент — добавить в `packages/ui/src/index.ts`:
```typescript
export { StatCard } from './components/StatCard';
export type { StatCardProps, StatCardVariant } from './components/StatCard';
```

---

## 5. Связь с прототипом — карта использования

| Страница прототипа | Компоненты DS (готовые) | Компоненты DS (новые) |
|---|---|---|
| Dashboard | `QuickActionsGrid`, `StatusChip` | `StatCard`, `EventsFeed` |
| Каталог льгот | `FacetedSearch`, `BenefitCard`, `StatusChip` | `FilterPills` |
| Мои баллы | `GamifiedProgress` | `TransactionList`, `StatCard` |
| Документы | `StatusChip` | `DocumentRow` |
| Поддержка | — | `SupportFAQ` |
| Header | `ProductHeaderToolbar` | `TopTabNavigation`, `BalancePill` |
| Модалка ДМС | `AprilModal`, `StatusChip` | `PolicyCard`, `ClinicMapList` |
| DMS Wizard | `AprilModal` | `WizardContainer` |
| MatCapital Wizard | `AprilModal` | `WizardContainer` |

---

## 6. Примечания для команды DS

### 6.1. Brand override

Все новые компоненты используют **Mantine theme colors** (`teal.6`, `gray.0`, `gray.6` и т.д.), а не хардкод hex. Бренд СДЭК переопределяет цвета через CSS variables (`--april-accent-*`), и компоненты автоматически подхватят изменения.

Не использовать `#00B33C` в коде компонентов — это цвет бренда клиента, не DS.

### 6.2. WizardContainer — самый сложный компонент

`WizardContainer` — единственный компонент с внутренней state-логикой (шаги, валидация, навигация). Рекомендуется:
- Сделать `WizardProgress` и `WizardFooter` отдельными экспортируемыми подкомпонентами (для кастомизации)
- Поддерживать как контролируемый (`currentStep`), так и неконтролируемый (`initialStep`) режимы
- Не привязываться к конкретной бизнес-логике (DMS/MatCapital) — только шаблон

### 6.3. TopTabNavigation vs ProductSidebarNavigation

`TopTabNavigation` создаётся **только если** продукт ЛК выбирает горизонтальную навигацию. Если решение — sidebar (как в DS по умолчанию), этот компонент можно не делать.

Рекомендация: использовать sidebar + `ProductHeaderToolbar` для ЛК. Горизонтальные табы — только если заказчик жёдко настаивает.

### 6.4. SupportFAQ

Можно реализовать в продукте на Mantine `Accordion` без нового компонента DS. Создать `SupportFAQ` в DS только если команда решит стандартизировать паттерн FAQ.
