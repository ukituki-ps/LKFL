# Отчёт

## Статус
выполнено

## Что сделано

Реализована страница каталога `/catalog` с фильтрами, поиском и пагинацией.

### Созданные/изменённые файлы

1. **`src/components/catalog/SearchInput.tsx`** — поле поиска с debounce 300ms через `useDebouncedValue` из `@mantine/hooks`. Синхронизирует внутреннее состояние с внешним `value` prop.

2. **`src/components/catalog/FilterBar.tsx`** — панель фильтров с тремя Select-компонентами:
   - Тип: Льготы (benefit) / Активности (activity)
   - Статус: Активные (active) / Промо (promo)
   - Категория: динамический список из API

3. **`src/components/catalog/Pagination.tsx`** — навигация по страницам с:
   - Отображением диапазона элементов (1–20 из 45)
   - Кнопками «Назад» / «Далее»
   - Номерами страниц с эллипсисом для больших наборов
   - Автоскрытием при totalPages <= 1

4. **`src/pages/Catalog.tsx`** — страница каталога:
   - Фильтры синхронизированы с URL query params (shareable links)
   - React Query `useQuery` для загрузки категорий и энгейджментов
   - Три состояния: loading (Loader), error (retry button), data
   - Empty state при отсутствии результатов
   - Debounced поиск (300ms) через SearchInput

### Технические детали

- Все фильтры в URL query params — ссылки можно шарить
- Default: status=active, page=1, per_page=20
- Сброс page=1 при изменении фильтров
- TypeScript: строгий режим, noUnusedLocals, noUnusedParameters
- Mantine 7: w prop вместо minWidth для Select

### Проверки

- `npm run build` — ✅ чистый
- `npm run lint` — ✅ чистый

## Время
~30 минут

## Замечания
- Mantine Select не поддерживает `minWidth` напрямую, использован `w` prop
- `Badge` импортирован в FilterBar из спецификации, но не используется (удалён для чистоты)
