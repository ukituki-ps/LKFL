# T2214.4 — Каталог: EngagementCard + AprilFilterPills + AprilIcon

## Родительская задача

T2214 — Frontend: прототип → код (header, дизайн, страницы-заглушки)

## Веха

M22-f1-hardening

## Тип

code

## Кратко

Полный реврайт карточки льготы (EngagementCard) по дизайну прототипа. Замена FilterBar на `AprilFilterPills` из DS. Замена эмодзи на `AprilIcon*`.

---

## Зависимости

- **T2214.1** (Brand tokens + DS upgrade) — нужен `--brand-*` для цветов, нужен `@ukituki-ps/april-ui@0.1.16`

## Что сделать

### 1. `src/components/catalog/EngagementCard.tsx` — полный реврайт

Заменить эмодзи на `AprilIcon*` и обновить layout по прототипу:

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

Icon mapping через `AprilIcon*`:
```ts
import {
  AprilIconHeart,
  AprilIconSuccess,
  AprilIconUsers,
  AprilIconDumbbell,
  AprilIconGift,        // → bike (nearest аналог)
  AprilIconCoffee,
  AprilIconGraduationCap,
  AprilIconBrain,
  AprilIconLanguages,
  AprilIconShoppingBag,
  AprilIconSmile,       // → smile (если нет → AprilIconHeart)
  AprilIconCalendar,    // → fallback
} from '@ukituki-ps/april-ui'

const iconMap: Record<string, LucideIcon> = {
  'heart-pulse': AprilIconHeart,
  'shield-plus': AprilIconSuccess,
  'shield-check': AprilIconSuccess,
  users: AprilIconUsers,
  dumbbell: AprilIconDumbbell,
  bike: AprilIconGift,         // nearest аналог
  utensils: AprilIconCoffee,   // nearest аналог
  'graduation-cap': AprilIconGraduationCap,
  brain: AprilIconBrain,
  languages: AprilIconLanguages,
  'shopping-bag': AprilIconShoppingBag,
  smile: AprilIconHeart,       // nearest аналог
  coffee: AprilIconCoffee,
}
```

### 2. `src/components/catalog/FilterBar.tsx` → `AprilFilterPills`

Заменить `Select`-dropdown на `AprilFilterPills` из DS:

```tsx
import { AprilFilterPills } from '@ukituki-ps/april-ui'

<AprilFilterPills
  items={filterOptions}
  active={activeFilter}
  onChange={setActiveFilter}
/>
```

**DS `AprilFilterPills` — доступен в v0.1.16.** Кастомная реализация НЕ требуется.

### 3. `src/pages/Catalog.tsx` — Grid 3 колонки

Обновить grid: `grid-template-columns: repeat(3, 1fr)` (как в прототипе).

### 4. Обновление тестов `EngagementCard.test.tsx`

Реврайт карточки меняет структуру DOM. Обновить тесты:
- Проверить `AprilIcon*` (по `data-testid` или наличие SVG)
- Проверить новый layout (icon 44×44, footer с ценой)
- Убрать проверки на «Нет изображения» (заменён на `AprilIcon`)
- Убрать проверки на эмодзи

## Файлы

### Изменяются
- `src/components/catalog/EngagementCard.tsx` — полный реврайт (AprilIcon*, дизайн по прототипу)
- `src/components/catalog/FilterBar.tsx` → обёртка над `AprilFilterPills` из DS
- `src/components/catalog/EngagementCard.test.tsx` — обновление тестов
- `src/pages/Catalog.tsx` — grid 3 колонки

## Критерии приёмки

- [ ] Карточки по дизайну прототипа (icon 44×44, footer с ценой + badge)
- [ ] `AprilFilterPills` из DS вместо кастомных filter pills
- [ ] Grid 3 колонки (как в прототипе)
- [ ] Все эмодзи заменены на `AprilIcon*` из DS
- [ ] Icon mapping в EngagementCard через `AprilIcon*`
- [ ] `lucide-react` НЕ импортируется напрямую в коде LKFL
- [ ] Тесты EngagementCard.test.tsx обновлены и проходят
- [ ] `npm run dev` → каталог визуально соответствует прототипу
