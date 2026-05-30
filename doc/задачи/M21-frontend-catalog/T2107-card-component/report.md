# Отчёт

## Статус
выполнено

## Что сделано

Реализован компонент `EngagementCard` для отображения карточки льготы/активности в каталоге.

### Файлы

| Файл | Действие |
|------|----------|
| `src/components/catalog/EngagementCard.tsx` | Переписан (был stub) |
| `doc/задачи/M21-frontend-catalog/T2107-card-component/plan.yaml` | Обновлён (100%) |
| `doc/задачи/M21-frontend-catalog/T2107-card-component/report.md` | Заполнен |

### Реализованные функции

1. **EngagementCard** — карточка с:
   - Изображение с fallback-плейсхолдером (Paper с текстом «Нет изображения»)
   - Badge (Промо → yellow, Доступна → green) из API
   - Название категории (если есть)
   - Название льготы (lineClamp={1})
   - Описание (lineClamp={2})
   - Провайдер в футере
   - Стоимость (cost_cents → рубли через Intl.NumberFormat ru-RU)
   - Счётчик вариантов со склонением (при offers.length > 1)
   - Link к detail page: `/catalog/{slug}`

2. **EngagementGrid** — сетка карточек (CSS grid auto-fill, minmax 280px)

3. **formatCost** — форматирование центов в рубли

4. **getBadgeColor** — цвет бейджа по значению

5. **pluralizeOffers** — склонение «вариант/варианта/вариантов»

### Проверки

- `npm run build` — ✅ без ошибок
- `npm run lint` — ✅ без ошибок

### Замечания

- `Paper` не поддерживает проп `height` — высота задана через `style.height`
- Badge берётся из API (не вычисляется на клиенте)
- Цвета бейджа: Промо → yellow, Доступна → green (по задаче, отличается от brief.md где было «зелёный/серый»)
