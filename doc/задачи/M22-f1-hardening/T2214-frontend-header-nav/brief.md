# T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Текущий фронтенд — скелет на Mantine defaults (blue theme, sidebar, системный шрифт, эмодзи-иконки).
Прототип (`артефакты/Прототип ЛК физика(1).html`) — зрелый, детализированный дизайн с продуманной дизайн-системой.

**Цель:** привести фронтенд к визуальному соответствию прототипа — header-навигация, дизайн-система, карточки, страницы-заглушки с моками.

### Дизайн-система прототипа

```css
--green:        #00B33C;
--green-dark:   #009A33;
--green-light:  #F0FDF4;
--green-border: #BBF7D0;
--bg:           #F2F2F2;
--card:         #FFFFFF;
--text:         #1A1A1A;
--text-muted:   #6B7280;
--text-subtle:  #9CA3AF;
--border:       #EBEBEB;
--row:          #F9FAFB;
--radius-card:  14px;
--radius-btn:   6px;
--shadow-card:  0 1px 4px rgba(0,0,0,0.06);
```

Шрифт: **Inter** (400/500/600/700/800). Заголовки: `font-weight: 800`. Большие числа: `letter-spacing: -0.5px`.
Иконки: **Lucide** (SVG, все иконки — `lucide-react`).

### Что меняется

| Элемент | Сейчас | Станет |
|---------|--------|--------|
| Layout | `AppShell` (header + sidebar + main) | `AprilProductHeader` + `<main>` без sidebar |
| Навигация | `EmployeeNav` (sidebar, вертикальные кнопки) | Горизонтальные ссылки в `center` с underline active |
| Цвет primary | blue (Mantine default) | `#00B33C` (СДЭК-зелёный из прототипа) |
| Шрифт | системный стек | Inter (Google Fonts) |
| Заголовки | Mantine default | `font-weight: 800`, `letter-spacing: -0.5px` для big numbers |
| Иконки | эмодзи (❤️🛡️👥) | Lucide SVG (`lucide-react`) |
| Карточки льгот | Paper 160px с эмодзи | icon 44×44 bg-gray, название, провайдер, описание, footer |
| Фильтры каталога | FilterBar с selects | Filter pills (pill-кнопки с зелёным active) |
| Dashboard | заглушка «—» | моки по прототипу + stub-индикатор |
| Баллы / Документы / Поддержка | заглушки | заглушки по прототипу + stub-индикатор |
| Balance | нет | `BalancePill` в header `right` |
| Уведомления | нет | `BellIcon` в header `right` |
| Avatar | blue | зелёный круг 34px с инициалами |

## Файлы

### Новые
- `src/components/layout/HeaderNav.tsx` — навигация (5 ссылок с underline) для `center`
- `src/components/layout/HeaderRight.tsx` — balance-pill + bell + avatar для `right`
- `src/components/ui/StubBadge.tsx` — красный кружок с «?» + tooltip «Заглушка — данные появятся в F2»
- `src/components/ui/BrandTokens.css` — CSS переменные бренда из прототипа

### Изменяются
- `src/components/layout/Shell.tsx` — замена `AppShell` на `AprilProductHeader`
- `src/lib/theme.ts` — brand tokens, typography, primaryColor → custom green
- `src/lib/providers.tsx` → `AprilProviders` из `@ukituki-ps/april-ui`
- `src/components/catalog/EngagementCard.tsx` — Lucide иконки, дизайн по прототипу
- `src/components/catalog/FilterBar.tsx` — filter pills вместо selects
- `src/pages/Dashboard.tsx` — моки по прототипу + stub-индикатор
- `src/pages/Points.tsx` — заглушка по прототипу + stub-индикатор
- `src/pages/Documents.tsx` — заглушка по прототипу + stub-индикатор
- `src/pages/Support.tsx` — заглушка по прототипу + stub-индикатор
- `src/components/layout/EmployeeNav.tsx` — упрощение (только мобильный drawer)
- `src/components/layout/AdminNav.tsx` — упрощение (только мобильный drawer)
- `src/components/layout/UserMenu.tsx` — avatar → зелёный
- `src/main.tsx` — AprilProviders из DS, Inter preload
- `index.html` — Google Fonts Inter preload

### Удаляются
- Ничего полностью, `EmployeeNav`/`AdminNav` — только для мобильного drawer

## Что сделать

### 1. Brand tokens — CSS переменные прототипа

`src/components/ui/BrandTokens.css`:
```css
:root {
  --brand-green:        #00B33C;
  --brand-green-dark:   #009A33;
  --brand-green-light:  #F0FDF4;
  --brand-green-border: #BBF7D0;
  --brand-bg:           #F2F2F2;
  --brand-card:         #FFFFFF;
  --brand-text:         #1A1A1A;
  --brand-text-muted:   #6B7280;
  --brand-text-subtle:  #9CA3AF;
  --brand-border:       #EBEBEB;
  --brand-row:          #F9FAFB;
  --brand-radius-card:  14px;
  --brand-radius-btn:   6px;
  --brand-shadow-card:  0 1px 4px rgba(0,0,0,0.06);
}
```

Импортировать в `main.tsx`. White-label: в M22+ эти переменные будут переопределяться brand CSS из API tenant'а.

### 2. `src/lib/theme.ts` — брендированная тема

```ts
export function createAprilTheme(): MantineThemeOverride {
  return createTheme({
    // Custom green scale matching prototype #00B33C
    colors: {
      brand: ['#F0FDF4', '#DCFCE7', '#BBF7D0', '#86EFAC', '#4ADE80',
              '#22C55E', '#16A34A', '#00B33C', '#009A33', '#00651E'],
    },
    primaryColor: 'brand',
    fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
    headings: {
      fontFamily: 'Inter, sans-serif',
      fontWeight: '800',
    },
    defaultRadius: 'md',  // 14px
    other: {
      cardRadius: '14px',
      btnRadius: '6px',
      cardShadow: '0 1px 4px rgba(0,0,0,0.06)',
    },
  })
}
```

### 3. `src/lib/providers.tsx` — AprilProviders из DS

Заменить кастомный провайдер на `AprilProviders` из `@ukituki-ps/april-ui`.
В `main.tsx`: `AprilProviders` (DS) → `QueryClientProvider` → `App`.

### 4. `index.html` — Inter

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" rel="stylesheet">
```

### 5. `src/components/layout/Shell.tsx` — AprilProductHeader

```tsx
<div style={{ minHeight: '100vh', backgroundColor: 'var(--brand-bg)' }}>
  <AprilProductHeader
    left={<Logo onClick={() => navigate('/')} />}
    center={<HeaderNav isAdmin={isAdminRoute} />}
    right={<HeaderRight />}
    sticky
  />
  <main style={{ maxWidth: 1100, margin: '0 auto', padding: '28px 28px 56px' }}>
    <Outlet />
  </main>
</div>
```

### 6. `src/components/layout/HeaderNav.tsx` — навигация по прототипу

5 ссылок: Главная, Каталог льгот, Мои баллы, Документы, Поддержка.

Стили:
- `padding: 0 14px`, `font-size: 13px`, `font-weight: 500`
- inactive: `color: var(--brand-text-muted)`
- active: `color: var(--brand-text)`, `border-bottom: 2px solid var(--brand-green)`, `font-weight: 600`
- hover: `color: var(--brand-green)`

Для admin — `adminRoutes`.

### 7. `src/components/layout/HeaderRight.tsx`

```tsx
<Group gap={10}>
  <BalancePill />   {/* bg: var(--brand-green-light), border: var(--brand-green-border), radius: 20px, font-size: 12px, font-weight: 700, color: #166534 */}
  <BellIcon />      {/* 34x34, bg: var(--brand-row), border: 1px solid var(--brand-border) */}
  <UserMenu />      {/* avatar: 34px circle, bg: var(--brand-green), color: white, font-size: 11px, font-weight: 700 */}
</Group>
```

### 8. `src/components/ui/StubBadge.tsx` — индикатор заглушки

Красный кружок с «?» внутри. При наведении tooltip: «Заглушка — данные появятся в F2».

Используется на всех заглушечных страницах и карточках с моками.

```tsx
export function StubBadge() {
  return (
    <Tooltip label="Заглушка — данные появятся после подключения API">
      <div style={{
        width: 18, height: 18, borderRadius: '50%',
        background: '#FEE2E2', color: '#DC2626',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontSize: 11, fontWeight: 700, cursor: 'help',
      }}>?</div>
    </Tooltip>
  )
}
```

### 9. `src/components/catalog/EngagementCard.tsx` — Lucide + прототип

Заменить эмодзи на `lucide-react` иконки. Карточка по прототипу:

```
┌─────────────────────────┐
│  [icon 44×44 bg-gray]   │
│  Название (14px fw:700) │
│  Провайдер (11px muted) │
│  Описание (12px muted)  │
├─────────────────────────┤
│  Цена        [badge]    │
└─────────────────────────┘
```

Icon mapping:
```ts
const iconMap: Record<string, LucideIcon> = {
  'heart-pulse': HeartPulse,
  'shield-plus': ShieldPlus,
  users: Users,
  dumbbell: Dumbbell,
  bike: Bike,
  utensils: Utensils,
  'graduation-cap': GraduationCap,
  brain: Brain,
  languages: Languages,
  'shopping-bag': ShoppingBag,
  smile: Smile,
  coffee: Coffee,
}
```

### 10. `src/components/catalog/FilterBar.tsx` — filter pills

Заменить selects на pill-кнопки:
- `padding: 6px 14px`, `border-radius: 20px`, `font-size: 12px`, `font-weight: 600`
- inactive: `border: 1.5px solid var(--brand-border)`, `color: var(--brand-text-muted)`
- active: `background: var(--brand-green)`, `color: #fff`, `border-color: var(--brand-green)`

### 11. `src/pages/Dashboard.tsx` — моки по прототипу

Структура прототипа:
- **Heading:** «Привет, Алексей» + дата + статус пакета
- **3 stat-карточки:** баланс (зелёная highlight), активные льготы, до конца периода
- **Активные льготы:** список с иконками Lucide, бейджами статуса
- **Лента событий:** события с иконками (green/yellow/blue)
- **Быстрые действия:** grid 2×3, иконки Lucide

Все данные — моки. `StubBadge` на каждом блоке.

### 12. `src/pages/Points.tsx` — заглушка по прототипу

Структура:
- Зелёная карточка баланса (48px value)
- Прогресс-бары по категориям
- Транзакции с фильтрами (Все / Начисления / Списания)

Моки + `StubBadge`.

### 13. `src/pages/Documents.tsx` — заглушка по прототипу

Структура:
- Таблица: документ, тип, дата, статус, «Скачать»
- Бейджи: Заявление (blue), Согласие (gray), Polis (blue)

Моки + `StubBadge`.

### 14. `src/pages/Support.tsx` — заглушка по прототипу

Структура:
- FAQ аккордеон (левая колонка)
- Форма обращения (правая колонка)

Моки + `StubBadge`.

### 15. Мобильная навигация

`Burger` → `Drawer` (Mantine) с теми же ссылками.
На desktop (`> 768px`) burger скрыт.

### 16. `src/components/layout/UserMenu.tsx` — адаптация

- Avatar color → `var(--brand-green)`
- Размер: 34px circle с инициалами
- Dropdown: email + «Выйти»

## Зависимости

- `@ukituki-ps/april-ui` ≥ 0.1.13 (уже в package.json)
- `lucide-react` — проверить наличие (вероятно уже через april-ui)
- Нет зависимостей от других задач M22

## Критерии приёмки

### Header и layout
- [ ] `Shell.tsx` использует `AprilProductHeader` (не `AppShell`)
- [ ] Sidebar убран (навигация горизонтальная в header'е)
- [ ] 5 ссылок навигации с underline active-индикатором
- [ ] Balance-pill в правой зоне header'а
- [ ] Bell icon button в правой зоне header'а
- [ ] Avatar с инициалами (зелёный круг) + dropdown
- [ ] Sticky header (прилипает при скролле)
- [ ] Фон страницы `#F2F2F2`
- [ ] Content max-width 1100px, margin auto, padding 28px
- [ ] Мобильная навигация: Burger → Drawer

### Дизайн-система
- [ ] Brand CSS tokens (`--brand-*`) подключены
- [ ] Primary color → `#00B33C` (СДЭК-зелёный)
- [ ] Шрифт Inter подключён (Google Fonts)
- [ ] Заголовки: `font-weight: 800`
- [ ] Большие числа: `letter-spacing: -0.5px`
- [ ] `radius-card: 14px`, `radius-btn: 6px`

### Иконки
- [ ] Все эмодзи заменены на Lucide иконки
- [ ] Icon mapping в EngagementCard

### Каталог
- [ ] Карточки по дизайну прототипа (icon 44×44, footer с ценой + badge)
- [ ] Filter pills вместо selects (pill-кнопки с зелёным active)
- [ ] Grid 3 колонки (как в прототипе)

### Страницы-заглушки
- [ ] Dashboard: моки по прототипу (3 stat-карточки, льготы, события, быстрые действия)
- [ ] Points: моки по прототипу (баланс, прогресс-бары, транзакции)
- [ ] Documents: моки по прототипу (таблица с бейджами)
- [ ] Support: моки по прототипу (FAQ + форма)
- [ ] StubBadge на каждом заглушечном блоке

### Общее
- [ ] `npm run dev` → страница визуально соответствует прототипу
- [ ] E2E тесты не ломаются (или обновлены)
- [ ] Admin маршруты: собственная навигация в header'е

## Модалки (не в этой задаче)

Реализуются через универсальные компоненты дизайн-системы `@ukituki-ps/april-ui`.
ТЗ для DS-агента: `doc/задачи/M22-f1-hardening/T2214-frontend-header-nav/ds-components-tz.md`.

Компоненты, которые нужны в DS:
- `AprilWizard` — универсальный wizard (progress bar, steps, back/next, footer actions)
- `AprilWizardProgress` — горизонтальный progress bar с номерами шагов, label'ами, линиями, состояниями
- `AprilOptionCard` — selectable option card с border highlight
- `AprilTabbedContent` — табы + контент (для Benefit Detail)
- `AprilPolicyCard` — градиентная карточка полиса

После апгрейда DS — LKFL подключит компоненты и реализует 3 модалки прототипа:
- DMS Wizard (4 шага: опция → оплата → подтверждение → готово)
- Benefit Detail (табы: условия / полис / клиники на карте)
- Mat Capital Wizard (4 шага)
