# T2214.3 — Страницы-заглушки (Dashboard, Points, Documents, Support)

## Родительская задача

T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Кратко

Реализовать 4 страницы-заглушки с моками по прототипу + `StubBadge` на каждом блоке. Данные — статические (API подключится в F2).

---

## Зависимости

- **T2214.1** (Brand tokens) — нужен `--brand-*` для цветов
- **T2214.2** (Shell + Header) — нужен layout с `AprilProductHeader`

## Что сделать

### 1. `src/components/ui/StubBadge.tsx` — индикатор заглушки

```tsx
export function StubBadge() {
  // Показывается ТОЛЬКО в dev-режиме
  if (import.meta.env.PROD) return null
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

**Важно:** компонент рендерится только когда `import.meta.env.DEV` === `true`. В production — `null`.

### 2. `src/pages/Dashboard.tsx` — моки по прототипу

**Убрать `useQuery`** (заменить на статические моки). Структура:

- **Heading:** «Привет, Алексей» + дата + статус пакета
- **3 stat-карточки:** баланс (зелёная highlight), активные льготы, до конца периода
- **Активные льготы:** список с иконками Lucide, бейджами статуса
- **Лента событий:** события с иконками (green/yellow/blue)
- **Быстрые действия:** grid 2×3, иконки Lucide

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

Использовать `lucide-react` напрямую (не через `AprilIcon` обёртку DS). Все иконки импортировать из `lucide-react`:

```tsx
import {
  HeartPulse, ShieldPlus, Users, Dumbbell, Bike, Utensils,
  GraduationCap, Brain, Languages, ShoppingBag, Smile, Coffee,
  Bell, Coins, Calendar, Clock, MapPin, Download, Send,
  ChevronRight, ChevronDown, FileText, Search,
} from 'lucide-react'
```

> **Почему не AprilIcon:** DS v0.1.13 экспортирует лишь ~25 иконок. Прототипу нужно ~40. Прямые импорты Lucide — проще, без обёрток.

## Зависимости (npm)

- `lucide-react` — добавить в `package.json` dependencies
- Нет других новых зависимостей

## Файлы

### Новые
- `src/components/ui/StubBadge.tsx`

### Изменяются
- `src/pages/Dashboard.tsx` — полный реврайт (моки по прототипу)
- `src/pages/Points.tsx` — заглушка по прототипу
- `src/pages/Documents.tsx` — заглушка по прототипу
- `src/pages/Support.tsx` — заглушка по прототипу
- `frontend/package.json` — +`lucide-react`

## Критерии приёмки

- [ ] Dashboard: моки по прототипу (3 stat-карточки, льготы, события, быстрые действия)
- [ ] Points: моки по прототипу (баланс, прогресс-бары, транзакции)
- [ ] Documents: моки по прототипу (таблица с бейджами)
- [ ] Support: моки по прототипу (FAQ + форма)
- [ ] StubBadge на каждом заглушечном блоке (виден только в dev)
- [ ] StubBadge не виден при `import.meta.env.PROD`
- [ ] Все эмодзи заменены на Lucide иконки
- [ ] `npm run dev` → страницы визуально соответствуют прототипу
