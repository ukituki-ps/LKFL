# T2305 — Catalog Visual Polish: filter pills, search, card hover, grid

## Веха

M23 — Frontend Polish

## Тип

code

## Проблема

Страница каталога (`/catalog`) визуально не соответствует прототипу в нескольких деталях.

### 1. Filter pills

Прототип: `filter-pill` — `padding: 6px 14px`, `border-radius: 20px`, `border: 1.5px solid`, `font-size: 12px`, `font-weight: 600`. Active: green bg + white text + green border.
Фронтенд: `FilterBar` — нужно проверить и привести к прототипу.

### 2. Search box

Прототип: `background: var(--card)`, `border: 1.5px solid var(--border)`, `border-radius: 8px`, `padding: 8px 14px`, иконка search + input.
Фронтенд: `SearchInput` — нужно проверить и привести к прототипу.

### 3. Grid columns

Прототип: `grid-template-columns: repeat(3, 1fr)`, gap 14px.
Фронтенд: `EngagementGrid` — нужно проверить.

### 4. Benefit card hover

Прототип: `transform: translateY(-2px)`, `box-shadow: 0 4px 16px rgba(0,0,0,0.1)`.
Фронтенд: нужно проверить и добавить если нет.

### 5. Benefit card layout

Прототип:
- Icon: `44×44`, `borderRadius: 12`, `background: var(--bg)`, `color: var(--green)`
- Name: `font-size: 14px`, `font-weight: 700`
- Provider: `font-size: 11px`, `color: var(--text-subtle)`
- Desc: `font-size: 12px`, `color: var(--text-muted)`, `line-height: 1.5`
- Footer: `border-top: 1px solid var(--row)`, цена + badge

### 6. Badge цвета

Прототип:
- `badge-green`: `#DCFCE7` bg, `#166534` text
- `badge-yellow`: `#FEF9C3` bg, `#854D0E` text
- `badge-gray`: `#F3F4F6` bg, `#4B5563` text
- `badge-blue`: `#DBEAFE` bg, `#1D4ED8` text

## Что делать

### 1. Filter pills — привести к прототипу

Проверить `FilterBar.tsx`:
- `border-radius: 20px`
- `border: 1.5px solid var(--brand-border)`
- Active: `background: var(--brand-green)`, `color: #fff`, `border-color: var(--brand-green)`
- Inactive: `background: var(--brand-card)`, `color: var(--brand-text-muted)`
- `font-size: 12px`, `font-weight: 600`
- `padding: 6px 14px`

### 2. Search box — привести к прототипу

Проверить `SearchInput.tsx`:
- `border: 1.5px solid var(--brand-border)`
- `border-radius: 8px`
- `padding: 8px 14px`
- Иконка search слева
- `font-size: 13px`

### 3. Grid — 3 колонки

Проверить `EngagementGrid.tsx`:
- `grid-template-columns: repeat(3, 1fr)`
- `gap: 14px`

### 4. Card hover

Проверить/добавить в `EngagementCard.tsx`:
- `transition: transform 0.15s, box-shadow 0.15s`
- Hover: `transform: translateY(-2px)`, `box-shadow: 0 4px 16px rgba(0,0,0,0.1)`

### 5. Card layout

Проверить/исправить:
- Icon container: `44×44`, `borderRadius: 12`
- Name: `fontSize: 14px`, `fontWeight: 700`
- Provider: `fontSize: 11px`, `color: var(--brand-text-subtle)`
- Desc: `fontSize: 12px`, `color: var(--brand-text-muted)`, `lineHeight: 1.5`
- Footer: `border-top: 1px solid var(--brand-row)`

### 6. Badge цвета

Создать/использовать кастомные badge цвета:

```tsx
const badgeColors = {
  green: { bg: '#DCFCE7', color: '#166534' },
  yellow: { bg: '#FEF9C3', color: '#854D0E' },
  gray: { bg: '#F3F4F6', color: '#4B5563' },
  blue: { bg: '#DBEAFE', color: '#1D4ED8' },
}
```

## Требования

- Каталог визуально соответствует прототипу
- `tsc --noEmit` без ошибок
- `npm run build` без ошибок

## Критерии приёмки

- [ ] Filter pills: border-radius 20px, active green, inactive gray
- [ ] Search box: 1.5px border, 8px radius, icon слева
- [ ] Grid: 3 колонки, gap 14px
- [ ] Card hover: translateY(-2px), shadow
- [ ] Card layout: icon 44×44, name 14px/700, provider 11px, desc 12px
- [ ] Card footer: border-top
- [ ] Badge цвета как в прототипе (green/yellow/gray/blue)
- [ ] `tsc --noEmit` ✅
- [ ] `npm run build` ✅

## Зависимости

- **depends_on:** T2301 (StubBadge removal)
- **touches:** `FilterBar.tsx`, `SearchInput.tsx`, `EngagementGrid.tsx`, `EngagementCard.tsx`
