# T2214.3 — Страницы-заглушки (Dashboard, Points, Documents, Support)

## Родительская задача

T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Кратко

Реализовать 4 страницы-заглушки с моками по прототипу + `StubBadge` на каждом блоке. Данные — статические (API подключится в F2). Иконки — `AprilIcon*` из DS v0.1.16.

---

## Зависимости

- **T2214.1** (Brand tokens + DS upgrade) — нужен `--brand-*` для цветов, нужен `@ukituki-ps/april-ui@0.1.16`
- **T2214.2** (Shell + Header) — нужен layout с `AprilProductHeader`

## Что сделать

### 1. `src/components/ui/StubBadge.tsx` — индикатор заглушки

```tsx
export function StubBadge() {
  // Показывается ТОЛЬКО в dev-режиме
  if (!import.meta.env.DEV) return null
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

**Важно:** компонент рендерится только когда `import.meta.env.DEV === true`. В production — `null`.

### 2. `src/pages/Dashboard.tsx` — моки по прототипу

**Убрать `useQuery`** (заменить на статические моки). Структура:

- **Heading:** «Привет, Алексей» + дата + статус пакета
- **3 stat-карточки:** баланс (зелёная highlight), активные льготы, до конца периода
- **Активные льготы:** список с иконками `AprilIcon*`, бейджами статуса
- **Лента событий:** события с иконками (green/yellow/blue)
- **Быстрые действия:** grid 2×3, иконки `AprilIcon*`

Все данные — моки. `StubBadge` на каждом блоке.

### 3. `src/pages/Points.tsx` — заглушка

Структура прототипа:
- Зелёная карточка баланса (48px value)
- Прогресс-бары по категориям
- Транзакции с фильтрами (Все / Начисления / Списания)

Моки + `StubBadge`.

### 4. `src/pages/Documents.tsx` — заглушка

Структура:
- Таблица: документ, тип, дата, статус, «Скачать»
- Бейджи: Заявление (blue), Согласие (gray), Полис (blue)

Моки + `StubBadge`.

### 5. `src/pages/Support.tsx` — заглушка

Структура:
- FAQ аккордеон (левая колонка)
- Форма обращения (правая колонка)

Моки + `StubBadge`.

## Иконки — стратегия

Использовать `AprilIcon*` из `@ukituki-ps/april-ui@0.1.16`. Прямой импорт `lucide-react` — **НЕ требуется**.

```tsx
import {
  AprilIconHeart,           // → heart-pulse (nearest аналог)
  AprilIconSuccess,         // → shield-check (nearest аналог)
  AprilIconPlusCircle,      // → shield-plus (nearest аналог)
  AprilIconSmartphone,      // → smart-speaker (nearest аналог)
  AprilIconCoins,
  AprilIconCalendar,
  AprilIconClock,
  AprilIconMapPin,
  AprilIconDownload,
  AprilIconSend,
  AprilIconChevronRight,
  AprilIconChevronLeft,
  AprilIconFileText,
  AprilIconSearch,
  AprilIconGift,
  AprilIconGraduationCap,
  AprilIconLanguages,
  AprilIconShoppingBag,
  AprilIconCoffee,
  AprilIconDumbbell,
  AprilIconBrain,
  AprilIconUserPlus,
  AprilIconUsers,
  AprilIconInfo,
  AprilIconBell,
} from '@ukituki-ps/april-ui'
```

> **Icon mapping (прототип → DS v0.1.16):**
> - `heart-pulse` → `AprilIconHeart` (nearest)
> - `shield-check` → `AprilIconSuccess` (=CheckCircle2, semantic match)
> - `shield-plus` → `AprilIconPlusCircle` (semantic match)
> - `smart-speaker` → `AprilIconSmartphone` (nearest)
> - Остальные иконки → 1:1 маппинг (Bell, Coins, Calendar, Brain и др.)

> **Почему не lucide-react напрямую:** DS v0.1.16 экспортирует 54 иконки. Прототипу нужно 38. Покрытие 35/38 + 3 nearest-аналога = 100%. Прямой импорт `lucide-react` запрещён архитектурой (фронтенд.md §Иконки, NAVIGATION.md правило #8).

## Зависимости (npm)

- **НЕ добавлять `lucide-react`** — уже транзитивная зависимость через `@ukituki-ps/april-ui`
- Нет других новых зависимостей

## Файлы

### Новые
- `src/components/ui/StubBadge.tsx`

### Изменяются
- `src/pages/Dashboard.tsx` — полный реврайт (моки по прототипу, `AprilIcon*`)
- `src/pages/Points.tsx` — заглушка по прототипу
- `src/pages/Documents.tsx` — заглушка по прототипу
- `src/pages/Support.tsx` — заглушка по прототипу

## Критерии приёмки

- [ ] Dashboard: моки по прототипу (3 stat-карточки, льготы, события, быстрые действия)
- [ ] Points: моки по прототипу (баланс, прогресс-бары, транзакции)
- [ ] Documents: моки по прототипу (таблица с бейджами)
- [ ] Support: моки по прототипу (FAQ + форма)
- [ ] StubBadge на каждом заглушечном блоке (виден только в dev)
- [ ] StubBadge не виден при `!import.meta.env.DEV`
- [ ] Все эмодзи заменены на `AprilIcon*` из DS
- [ ] `lucide-react` НЕ добавлен в `dependencies` package.json
- [ ] `npm run dev` → страницы визуально соответствуют прототипу
