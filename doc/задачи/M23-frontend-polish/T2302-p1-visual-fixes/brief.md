# T2302 — P1 Visual Fixes: Dashboard + Header + Points polish

## Веха

M23 — Frontend Polish

## Тип

code

## Проблема

6 визуальных расхождений P1优先级 между прототипом и фронтендом.

### 1. Stat cards — flex Group вместо Grid

Прототип: `display: grid; grid-template-columns: repeat(3,1fr); gap: 12px`.
Фронтенд: `Group gap="md"` + `flex: 1, minWidth: 180`.

**Файл:** `Dashboard.tsx` строка 428-449.

### 2. Active benefits — нет ссылки «Весь каталог»

Прототип: `card-header` с `card-link "Весь каталог →"` справа.
Фронтенд: только заголовок + StubBadge (уже убран в T2301).

**Файл:** `Dashboard.tsx` строка 179-185.

### 3. Quick actions — нет белого квадрата иконки

Прототип: `quick-btn-icon 32×32`, `background: #fff`, `box-shadow: 0 1px 3px`.
Фронтенд: просто иконка без контейнера.

**Файл:** `Dashboard.tsx` строка 299-350.

### 4. Bell icon — нет border

Прототип: `nav-icon-btn`: `border: 1px solid var(--border)`.
Фронтенд: `ActionIcon` с `backgroundColor` без border.

**Файл:** `HeaderRight.tsx` строка 43-53.

### 5. Page headings — Title вместо h1 + p

Прототип: `<h1>` 24px/800 + `<p>` 13px dimmed.
Фронтенд: `Title order={2}` или `Text fw={600} size="lg"` + StubBadge (убран в T2301).

**Файлы:** Все страницы.

### 6. Benefit rows — размер иконки, badge, hover

Прототип: `benefit-icon 38×38`, `border-radius: 10px`, badge справа, `hover: background: green-light`.
Фронтенд: `32×32`, `border-radius: 8px`, нет badge, нет hover.

**Файл:** `Dashboard.tsx` строка 187-232.

## Что делать

### 1. Stat cards → Grid

```tsx
<div style={{
  display: 'grid',
  gridTemplateColumns: 'repeat(3, 1fr)',
  gap: 12,
}}>
  <StatCard ... />
  <StatCard ... />
  <StatCard ... />
</div>
```

Убрать `Group gap="md" wrap="wrap"`. Убрать `flex: 1, minWidth: 180` из StatCard.

### 2. Active benefits — ссылка «Весь каталог»

- Добавить `<Button variant="link" color="brand" onClick={() => navigate('/catalog')}>Весь каталог →</Button>`
- Стиль: `font-size: 12px, font-weight: 600, color: var(--brand-green)`

### 3. Quick actions — белый квадрат

- Обернуть иконку в `<div>`: `width: 32, height: 32, borderRadius: 8, background: '#fff', boxShadow: '0 1px 3px rgba(0,0,0,0.08)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--brand-green)'`

### 4. Bell icon — border

- `HeaderRight.tsx`: добавить `border: '1px solid var(--brand-border)'` к ActionIcon

### 5. Page headings

- Dashboard: `<h1>` заменить `Title order={2}` на `Title order={1}` (24px/800)
- Points/Documents/Support: аналогично
- Добавить `<p>` подзаголовок как в прототипе

### 6. Benefit rows

- Icon: `38×38`, `borderRadius: 10`
- Добавить Badge справа (active/awaiting)
- Hover: `onMouseEnter` → `background: var(--brand-green-light)`
- Meta строка: «АльфаСтрахование · до 31.12.2025»

## Требования

- Визуальное соответствие прототипу на уровне P1
- `tsc --noEmit` без ошибок
- `npm run build` без ошибок

## Критерии приёмки

- [ ] Stat cards — grid 3 колонки, gap 12px
- [ ] Active benefits — ссылка «Весь каталог»
- [ ] Quick actions — белый квадрат иконки 32×32 с тенью
- [ ] Bell icon — border 1px solid
- [ ] Page headings — h1 24px/800 + p подзаголовок
- [ ] Benefit rows — 38×38 иконки, badge, hover
- [ ] `tsc --noEmit` ✅
- [ ] `npm run build` ✅

## Зависимости

- **depends_on:** T2301 (StubBadge removal)
- **touches:** `Dashboard.tsx`, `HeaderRight.tsx`, `Points.tsx`, `Documents.tsx`, `Support.tsx`
