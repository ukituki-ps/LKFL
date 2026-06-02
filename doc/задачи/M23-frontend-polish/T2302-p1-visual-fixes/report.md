# T2302 — Отчёт: P1 Visual Fixes

## Статус

✅ **Выполнено** — 2026-06-02, перереализация после потери кода.

## Что сделано

### 1. Stat cards → Grid

- `Dashboard.tsx`: `Group gap="md" wrap="wrap"` → `<div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>`
- Убраны `flex: 1, minWidth: 180` из StatCard и Skeleton-заглушек
- Error-сообщение получает `gridColumn: '1 / -1'` для полного охвата сетки

### 2. Active benefits — ссылка «Весь каталог»

- `Dashboard.tsx`, `ActiveBenefitsList`: заголовок → `Group justify="space-between"` + `Button variant="link"` → `navigate('/catalog')`
- Стиль: `fontSize: 12, fontWeight: 600, color: var(--brand-green)`
- Добавлен `import { useNavigate } from 'react-router-dom'`

### 3. Quick actions — белый квадрат иконки

- `Dashboard.tsx`, `QuickActionsGrid`: иконка обернута в `<div>`:
  - `width: 32, height: 32, borderRadius: 8`
  - `background: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.08)'`
  - `display: 'flex', alignItems: 'center', justifyContent: 'center'`
  - `color: 'var(--brand-green)'`

### 4. Bell icon — border

- `HeaderRight.tsx`: ActionIcon получил `border: '1px solid var(--brand-border)'`

### 5. Page headings — h1 + p

| Страница | Было | Стало | Подзаголовок |
|----------|------|-------|-------------|
| Dashboard | `Title order={2}` | `Title order={1}` (24px/800) | `{today} · Добро пожаловать в личный кабинет льгот` |
| Points | `Text fw={600} size="lg"` | `Title order={1}` (24px/800) | `История начислений и списаний` |
| Documents | `Group + AprilIconFileText + Text fw={600}` | `Title order={1}` (24px/800) | `Заявления, согласия и сформированные документы` |
| Support | `Group + AprilIconHelp + Text fw={600}` | `Title order={1}` (24px/800) | `Частые вопросы и обратная связь` |

- Убраны неиспользуемые импорты: `Group`, `AprilIconFileText`, `AprilIconHelp`
- Добавлен импорт `Title` в Points, Documents, Support

### 6. Benefit rows

- `Dashboard.tsx`, `ActiveBenefitsList`:
  - Icon: `32×32, borderRadius: 8` → `38×38, borderRadius: 10`, `size={18}` вместо `size={16}`
  - Hover: `onMouseEnter/onMouseLeave` → `background: var(--brand-green-light, #F0FDF4)`
  - Meta строка: `{b.provider} · до 31.12.2025` (ранее только `{b.provider}`)
  - Badge — уже был (сохранён)
  - Обёртка: `<Group>` → `<div>` с hover + `<Group>` внутри

## Валидация

- ✅ `tsc --noEmit` — без ошибок
- ✅ `npm run build` — без ошибок (vite v6.4.2, 3.46s)

## Файлы

| Файл | Изменения |
|------|-----------|
| `frontend/src/pages/Dashboard.tsx` | +1 import, StatCard flex→grid, ActiveBenefitsList header+rows+hover, QuickActions icon wrap, greeting h1+p |
| `frontend/src/pages/Points.tsx` | +1 import (Title), heading → h1+p |
| `frontend/src/pages/Documents.tsx` | +1 import (Title), -2 imports (Group, AprilIconFileText), heading → h1+p |
| `frontend/src/pages/Support.tsx` | +1 import (Title), -1 import (AprilIconHelp), heading → h1+p |
| `frontend/src/components/layout/HeaderRight.tsx` | Bell ActionIcon + border |

## Замечания

- Код полностью перереализован с нуля (предыдущая реализация была потеряна после `git reset --hard`)
- Все 6 пунктов P1 выполнены согласно спецификации из `brief.md`
