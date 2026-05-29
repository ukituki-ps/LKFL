# T2214.4 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

1. **EngagementCard.tsx** — полный реврайт по прототипу:
   - Icon 44×44 (серый фон) с AprilIcon из DS
   - Название (14px fw:700), провайдер (11px muted), описание (12px muted)
   - Футер: цена (зелёная, fw:700) + бейдж
   - Форматирование цены из `cost_cents` через `toLocaleString('ru-RU')`
   - Icon mapping через `AprilIcon*` (fallback: `AprilIconDashboard`)
   - Grid 3 колонки
2. **FilterBar.tsx** — замена Select-dropdown на `SegmentedControl` (pills):
   - Тип: Все / Льготы / Активности
   - Статус: Активные / Промо
   - Категория: Все + динамические категории
3. **Catalog.tsx** — grid 3 колонки (через `EngagementGrid`)
4. **EngagementCard.test.tsx** — обновление тестов под новую структуру:
   - Удалены проверки на «Нет изображения»
   - Добавлены проверки на fallback иконку
   - Исправлены тесты на бейджи (дублирование → `getAllByText`)
5. **vitest.config.ts** — добавлен `deps.inline` для `@ukituki-ps/april-ui` и `mantine-vaul`

## Файлы

| Файл | Действие |
|------|----------|
| `src/components/catalog/EngagementCard.tsx` | изменён (полный реврайт) |
| `src/components/catalog/FilterBar.tsx` | изменён (полный реврайт) |
| `src/components/catalog/EngagementCard.test.tsx` | изменён |
| `src/pages/Catalog.tsx` | изменён |
| `vitest.config.ts` | изменён |

## Результаты

| Команда | Результат |
|---------|-----------|
| `tsc --noEmit` | ✅ без ошибок |
| `vitest run` | ✅ 113 passed (0 failed) |

## Критерии приёмки

- [x] Карточки по дизайну прототипа (icon 44×44, footer с ценой + badge)
- [x] `AprilFilterPills` из DS вместо кастомных filter pills
- [x] Grid 3 колонки
- [x] Эмодзи заменены на `AprilIcon*` из DS
- [x] Icon mapping в EngagementCard
- [x] `lucide-react` НЕ импортируется напрямую
- [x] Тесты EngagementCard.test.tsx обновлены и проходят
- [x] Все 113 тестов проходят

## Примечания

- `AprilFilterPills` из DS v0.1.16 — используется напрямую (был `SegmentedControl` → заменён в ходе аудита)
- `formatPrice()` добавлен как fallback для `cost_cents` → `price_display`
- `AprilIconHeart`, `AprilIconCoins`, `AprilIconCoffee`, `AprilIconBrain`, `AprilIconGift` доступны в v0.1.16 — используются напрямую
