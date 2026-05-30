# T2106 — /catalog страница

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Главная страница каталога льгот/активностей. Отображает список энгейджментов с фильтрами, поиском и пагинацией.
Зависит от T2105 (Layout), T2109 (API layer), T2107 (EngagementCard).

### Backend endpoint (из `backend/internal/engagement/catalog/handler.go`)

```
GET /api/v1/engagements?type=benefit&status=active&category=fitness&search=йога&page=1&per_page=20
```

**Query params:**
- `type` — фильтр по типу: `benefit` | `activity`
- `status` — фильтр по статусу: `active` | `promo` | `draft` | `hidden` | `completed`
- `category` — фильтр по slug категории (например, `fitness`)
- `search` — текстовый поиск по name и description
- `page` — номер страницы (по умолчанию 1)
- `per_page` — элементов на странице (по умолчанию 20, максимум 100)

**Response** (из `handler.go`):

```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "yoga-studio",
      "name": "Йога в студии",
      "description": "Абонемент на йогу",
      "type": "benefit",
      "status": "active",
      "cost_cents": 150000,
      "provider_name": "FitLife",
      "image_url": "https://example.com/yoga.jpg",
      "category": {
        "id": "uuid",
        "slug": "fitness",
        "name": "Фитнес",
        "icon": "dumbbell",
        "sort_order": 1
      },
      "offers": [
        {
          "id": "uuid",
          "name": "Месечный",
          "description": "1 месяц",
          "cost_cents": 150000,
          "sort_order": 1
        }
      ],
      "badge": "Доступна"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

### Categories endpoint (из `handler.go`)

```
GET /api/v1/engagements/categories
```

**Response:**

```json
[
  {
    "id": "uuid",
    "slug": "fitness",
    "name": "Фитнес",
    "icon": "dumbbell",
    "sort_order": 1
  }
]
```

### Tenant isolation

Все запросы требуют JWT + tenant middleware (настроен в `backend/internal/app/server.go`).
Backend автоматически фильтрует по tenant из context.

## Что сделать

### `src/pages/Catalog.tsx` — страница каталога

```tsx
import { useCallback, useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useInfiniteQuery, useQuery } from '@tanstack/react-query'
import { getEngagements, getCategories } from '@/api/engagements'
import { EngagementCard } from '@/components/catalog/EngagementCard'
import { FilterBar } from '@/components/catalog/FilterBar'
import { SearchInput } from '@/components/catalog/SearchInput'

export function Catalog() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [search, setSearch] = useState('')

  // Фильтры из URL
  const type = searchParams.get('type') || ''
  const status = searchParams.get('status') || 'active'
  const category = searchParams.get('category') || ''
  const page = parseInt(searchParams.get('page') || '1')
  const perPage = parseInt(searchParams.get('per_page') || '20')

  // Загрузка категорий для фильтра
  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => getCategories(),
  })

  // Загрузка энгейджментов
  const { data, isLoading, isError } = useQuery({
    queryKey: ['engagements', type, status, category, search, page, perPage],
    queryFn: () => getEngagements({ type, status, category, search, page, per_page: perPage }),
  })

  // Debounced search
  const handleSearch = useCallback((value: string) => {
    setSearch(value)
    // Debounce через setTimeout или useDebounce (Mantine hooks)
  }, [])

  const handleFilterChange = useCallback((key: string, value: string) => {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (value) {
        next.set(key, value)
      } else {
        next.delete(key)
      }
      next.set('page', '1') // Reset page on filter change
      return next
    })
  }, [setSearchParams])

  if (isLoading) return <div>Загрузка каталога...</div>
  if (isError) return <div>Ошибка загрузки каталога</div>

  const engagements = data?.data || []
  const pagination = data?.pagination

  return (
    <div className="catalog-page">
      <h1>Каталог льгот</h1>

      <FilterBar
        categories={categories || []}
        type={type}
        status={status}
        category={category}
        onChange={handleFilterChange}
      />

      <SearchInput value={search} onChange={handleSearch} />

      <div className="catalog-grid">
        {engagements.length === 0 ? (
          <div>Нет доступных льгот</div>
        ) : (
          engagements.map(engagement => (
            <EngagementCard key={engagement.id} engagement={engagement} />
          ))
        )}
      </div>

      {pagination && (
        <Pagination
          page={pagination.page}
          perPage={pagination.per_page}
          total={pagination.total}
          totalPages={pagination.total_pages}
          onPageChange={(p) => setSearchParams(prev => { prev.set('page', String(p)); return prev })}
        />
      )}
    </div>
  )
}
```

### `src/components/catalog/FilterBar.tsx` — панель фильтров

Фильтры:
- **Категория** — dropdown с категориями из `/engagements/categories`
- **Тип** — toggle: `benefit` (льготы) / `activity` (активности)
- **Статус** — dropdown: `active` (по умолчанию), `promo`

### `src/components/catalog/SearchInput.tsx` — поиск

- Debounce 300ms (через `useDebouncedValue` из @mantine/hooks)
- Поиск по name и description (backend param `search`)

### `src/components/catalog/Pagination.tsx` — пагинация

Использует `PaginationResponse` из API:
- `page`, `per_page`, `total`, `total_pages`
- Навигация: prev/next + номера страниц

### States

| Состояние | UI |
|-----------|-----|
| Loading | Skeleton / spinner |
| Empty (no results) | «Нет доступных льгот» |
| Error | «Ошибка загрузки каталога» + кнопка «Повторить» |
| Has data | Сетка карточек + пагинация |

## Требования

- React Query для data fetching (useQuery)
- Фильтры в URL query params (shareable links)
- Debounce поиска 300ms
- Пагинация по PaginationResponse из backend
- Default filter: status=active, page=1, per_page=20
- Grid layout для карточек (responsive: 1/2/3/4 колонки)

## Критерии приёмки

- [ ] `src/pages/Catalog.tsx` — страница с фильтрами, поиском, пагинацией
- [ ] `src/components/catalog/FilterBar.tsx` — фильтры по категории, типу, статусу
- [ ] `src/components/catalog/SearchInput.tsx` — debounced поиск
- [ ] `src/components/catalog/Pagination.tsx` — навигация по страницам
- [ ] Загрузка категорий из `GET /api/v1/engagements/categories`
- [ ] Загрузка списка из `GET /api/v1/engagements` с параметрами
- [ ] Фильтры в URL query params
- [ ] Состояния: loading, empty, error, data
- [ ] PaginationResponse отображается корректно
- [ ] Grid layout адаптивный
