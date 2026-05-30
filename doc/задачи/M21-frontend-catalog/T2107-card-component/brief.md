# T2107 — Карточка льготы (EngagementCard)

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Компонент карточки энгейджмента для отображения в каталоге.
Зависит от T2102 (AprilProviders — MantineProvider, тема).

### Тип данных (из `backend/internal/engagement/catalog/handler.go`)

```go
type EngagementTypeResponse struct {
    ID           uuid.UUID                   `json:"id"`
    Slug         string                      `json:"slug"`
    Name         string                      `json:"name"`
    Description  string                      `json:"description,omitempty"`
    Type         string                      `json:"type"` // "benefit" | "activity"
    Status       string                      `json:"status"` // "draft"|"active"|"promo"|"hidden"|"completed"
    CostCents    *int64                      `json:"cost_cents,omitempty"`
    ProviderName string                      `json:"provider_name,omitempty"`
    ImageURL     *string                     `json:"image_url,omitempty"`
    Category     *EngagementCategoryResponse `json:"category,omitempty"`
    Offers       []EngagementOfferResponse   `json:"offers,omitempty"`
    Badge        string                      `json:"badge"` // "Промо" | "Доступна"
}
```

### Badge logic (из `handler.go` — computeBadge)

```go
func computeBadge(et EngagementType) string {
    switch et.Status {
    case StatusPromo:
        return "Промо"
    case StatusActive:
        return "Доступна"
    default:
        return "Доступна"
    }
}
```

- `status == "promo"` → бейдж «Промо»
- `status == "active"` и остальные → бейдж «Доступна»
- TODO M26: «Активна» (проверка user_engagements), «Новинка» (created_at < 7 дней)

### Cost formatting

`cost_cents` — стоимость в центах (копейках). Для отображения: `cost_cents / 100` рублей.
Пример: `150000` → `1 500 ₽`.

## Что сделать

### TypeScript тип (соответствует Go struct)

```ts
// src/api/types.ts
export interface EngagementCategoryResponse {
  id: string
  slug: string
  name: string
  icon?: string
  sort_order: number
}

export interface EngagementOfferResponse {
  id: string
  name: string
  description?: string
  cost_cents: number
  sort_order: number
}

export interface EngagementTypeResponse {
  id: string
  slug: string
  name: string
  description?: string
  type: 'benefit' | 'activity'
  status: 'draft' | 'active' | 'promo' | 'hidden' | 'completed'
  cost_cents?: number
  provider_name?: string
  image_url?: string
  category?: EngagementCategoryResponse
  offers?: EngagementOfferResponse[]
  badge: string // "Промо" | "Доступна"
}
```

### `src/components/catalog/EngagementCard.tsx`

```tsx
import { Card, Image, Badge, Text, Group, Box } from '@mantine/core'
import { Link } from 'react-router-dom'
import type { EngagementTypeResponse } from '@/api/types'

interface EngagementCardProps {
  engagement: EngagementTypeResponse
}

// Форматирование стоимости из центов в рубли
function formatCost(cents: number): string {
  const rubles = Math.floor(cents / 100)
  return new Intl.NumberFormat('ru-RU').format(rubles) + ' ₽'
}

// Цвет бейджа
function getBadgeColor(badge: string): string {
  switch (badge) {
    case 'Промо':
      return 'yellow'
    case 'Доступна':
      return 'green'
    default:
      return 'gray'
  }
}

export function EngagementCard({ engagement }: EngagementCardProps) {
  return (
    <Card withBorder padding="lg" radius="md">
      <Link to={`/catalog/${engagement.id}`}>
        {/* Image */}
        {engagement.image_url && (
          <Image
            src={engagement.image_url}
            alt={engagement.name}
            height={160}
            radius="md"
            mb="md"
          />
        )}

        {/* Badge */}
        <Group justify="space-between" mb="xs">
          <Badge color={getBadgeColor(engagement.badge)} size="sm">
            {engagement.badge}
          </Badge>

          {/* Category */}
          {engagement.category && (
            <Text size="xs" c="dimmed">
              {engagement.category.icon && <span>{engagement.category.icon} </span>}
              {engagement.category.name}
            </Text>
          )}
        </Group>

        {/* Name */}
        <Text fw={600} size="lg" mb="xs">
          {engagement.name}
        </Text>

        {/* Description */}
        {engagement.description && (
          <Text size="sm" c="dimmed" lineClamp={2} mb="md">
            {engagement.description}
          </Text>
        )}

        {/* Footer */}
        <Group justify="space-between">
          {/* Provider */}
          {engagement.provider_name && (
            <Text size="xs" c="dimmed">
              {engagement.provider_name}
            </Text>
          )}

          {/* Cost */}
          {engagement.cost_cents !== undefined && (
            <Text fw={600} size="sm">
              {formatCost(engagement.cost_cents)}
            </Text>
          )}
        </Group>

        {/* Offers count */}
        {engagement.offers && engagement.offers.length > 1 && (
          <Text size="xs" c="dimmed" mt="xs">
            {engagement.offers.length} варианта
          </Text>
        )}
      </Link>
    </Card>
  )
}
```

### Grid layout (в Catalog.tsx)

```tsx
<div className="catalog-grid" style={{
  display: 'grid',
  gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
  gap: '16px',
}}>
  {engagements.map(e => <EngagementCard key={e.id} engagement={e} />)}
</div>
```

## Требования

- Тип EngagementTypeResponse точно соответствует Go struct из handler.go
- Badge берётся из API (не вычисляется на клиенте)
- Цвет бейджа: Промо → yellow, Доступна → green
- Cost formatting: cost_cents / 100 с форматированием ru-RU
- Image с fallback (если image_url не задан)
- Link к detail page: `/catalog/{id}`
- Category icon + name (если category !== null)
- Provider name отображается в футере
- Offers count если > 1

## Критерии приёмки

- [ ] `src/components/catalog/EngagementCard.tsx` — компонент карточки
- [ ] Отображение: name, description, cost, provider_name, image_url
- [ ] Badge из API (Промо / Доступна) с цветовой кодировкой
- [ ] Category icon + name
- [ ] formatCost: cost_cents → рубли (Intl.NumberFormat ru-RU)
- [ ] Link `/catalog/{id}` для detail
- [ ] Fallback при отсутствии image_url
- [ ] Offers count при > 1 варианте
- [ ] Тип EngagementTypeResponse соответствует Go struct handler.go
