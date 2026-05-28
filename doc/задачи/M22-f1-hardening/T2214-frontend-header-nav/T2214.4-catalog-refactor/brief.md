# T2214.4 — Каталог: EngagementCard + FilterBar + Lucide

## Родительская задача

T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Кратко

Рефакторинг карточки льготы (EngagementCard) и фильтров (FilterBar) по дизайну прототипа. Замена эмодзи на Lucide иконки.

---

## Зависимости

- **T2214.1** (Brand tokens) — нужен `--brand-*` для цветов
- **T2214.3** (Pages stubs) — устанавливает `lucide-react`

## Что сделать

### 1. `src/components/catalog/EngagementCard.tsx` — Lucide + прототип

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

> **Иконочная стратегия:** использовать `lucide-react` напрямую (не `AprilIcon` обёртку DS). DS v0.1.13 экспортирует ~25 иконок, прототипу нужно ~40. См. T2214.3.

### 2. `src/components/catalog/FilterBar.tsx` — filter pills

Заменить `Select`-dropdown на pill-кнопки:
- `padding: 6px 14px`, `border-radius: 20px`, `font-size: 12px`, `font-weight: 600`
- inactive: `border: 1.5px solid var(--brand-border)`, `color: var(--brand-text-muted)`
- active: `background: var(--brand-green)`, `color: #fff`, `border-color: var(--brand-green)`

> **Компонент DS:** `AprilFilterPills` доступен начиная с v0.1.15. Текущая версия `0.1.13` — не имеет. Реализовать кастомно. В M22+ можно заменить на DS-компонент.

### 3. `src/pages/Catalog.tsx` — Grid 3 колонки

Обновить grid: `grid-template-columns: repeat(3, 1fr)` (как в прототипе).

### 4. Обновление тестов `EngagementCard.test.tsx`

Рефакторинг карточки меняет структуру DOM. Обновить тесты:
- Проверить Lucide иконки (по `data-testid` или наличие SVG)
- Проверить новый layout (icon 44×44, footer с ценой)
- Убрать проверки на «Нет изображения» (заменён на Lucide иконку)

## Файлы

### Изменяются
- `src/components/catalog/EngagementCard.tsx` — Lucide иконки, дизайн по прототипу
- `src/components/catalog/FilterBar.tsx` — filter pills вместо Select
- `src/components/catalog/EngagementCard.test.tsx` — обновление тестов
- `src/pages/Catalog.tsx` — grid 3 колонки

## Критерии приёмки

- [ ] Карточки по дизайну прототипа (icon 44×44, footer с ценой + badge)
- [ ] Filter pills вместо selects (pill-кнопки с зелёным active)
- [ ] Grid 3 колонки (как в прототипе)
- [ ] Все эмодзи заменены на Lucide иконки
- [ ] Icon mapping в EngagementCard
- [ ] Тесты EngagementCard.test.tsx обновлены и проходят
- [ ] `npm run dev` → каталог визуально соответствует прототипу
