# T2305 — Catalog Visual Polish — Report

## Статус

✅ Выполнено

## Выполненная работа (2026-06-02)

### 1. Filter pills — FilterBar.tsx

- Убрал зависимость от `AprilFilterPills` — реализовал кастомный компонент `FilterPill`
- `border-radius: 20px`
- `border: 1.5px solid` — цвет зависит от active/inactive состояния
- Active: `background: var(--brand-green, #00B33C)`, `color: #fff`, `border-color: var(--brand-green, #00B33C)`
- Inactive: `background: var(--brand-card)`, `color: var(--brand-text-muted)`, `border-color: var(--brand-border)`
- `font-size: 12px`, `font-weight: 600`, `padding: 6px 14px`
- Добавлена плавная анимация `transition: all 0.15s`

### 2. Search box — SearchInput.tsx

- Добавлена иконка `AprilIconSearch` слева через `leftSection`
- `border: 1.5px solid var(--brand-border)` через `styles={{ input: {...} }}`
- `border-radius: 8px` через `styles`
- `padding: 8px 14px` через `styles`
- `font-size: 13px` через `style`

### 3. Grid — EngagementGrid (в EngagementCard.tsx)

- `grid-template-columns: repeat(3, 1fr)` — было уже
- `gap: 14px` — исправлено с 16px на 14px

### 4. Card hover — EngagementCard.tsx

- Был уже реализован: `transition: transform 0.15s, box-shadow 0.15s`
- Hover: `transform: translateY(-2px)`, `box-shadow: 0 4px 16px rgba(0,0,0,0.1)`

### 5. Card layout — EngagementCard.tsx

- Icon container: `width: 44`, `height: 44`, `borderRadius: 12` (was: 44px height + 12px padding → теперь точно 44×44)
- Name: `fontSize: 14px`, `fontWeight: 700` (было: size="md")
- Provider: `fontSize: 11px`, `color: var(--brand-text-subtle)` (было: size="xs", c="dimmed")
- Desc: `fontSize: 12px`, `color: var(--brand-text-muted)`, `lineHeight: 1.5` (было: size="sm", c="dimmed")
- Footer: `border-top: 1px solid var(--brand-row)` — было уже

### 6. Badge цвета — EngagementCard.tsx

- Убрал `MantineBadge` — заменил на кастомный компонент `<Badge>` (span)
- Кастомная палитра `badgeColors`:
  - `green`: `#DCFCE7` bg, `#166534` text
  - `yellow`: `#FEF9C3` bg, `#854D0E` text
  - `gray`: `#F3F4F6` bg, `#4B5563` text
  - `blue`: `#DBEAFE` bg, `#1D4ED8` text
- Поддержка `size="xs"` и `size="sm"` с разными padding/font-size

## Изменённые файлы

| Файл | Тип изменения |
|------|--------------|
| `frontend/src/components/catalog/FilterBar.tsx` | Переписан: кастомные filter pills |
| `frontend/src/components/catalog/SearchInput.tsx` | Добавлена иконка, точные стили |
| `frontend/src/components/catalog/EngagementCard.tsx` | Badge цвета, размеры шрифтов, grid gap |

## Валидация

- ✅ `tsc --noEmit` — без ошибок
- ✅ `npm run build` — без ошибок (3.46s, 3250 модулей)

## Замечания

- Тесты `EngagementCard.test.tsx` не требуют изменений — используют текстовые проверки, не зависящие от стилей
- `Catalog.tsx` не менялся — импорты совместимы
