# T2110 — OpenAPI Codegen

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Генерация TypeScript типов из OpenAPI спецификации backend M20.
Зависит от T2101 (Vite bootstrap — package.json, scripts).

### Go response types (из `backend/internal/engagement/catalog/handler.go`)

Типы должны быть сгенерированы так, чтобы точно соответствовать Go struct:

```go
type EngagementTypeResponse struct {
    ID           uuid.UUID                   `json:"id"`
    Slug         string                      `json:"slug"`
    Name         string                      `json:"name"`
    Description  string                      `json:"description,omitempty"`
    Type         string                      `json:"type"`
    Status       string                      `json:"status"`
    CostCents    *int64                      `json:"cost_cents,omitempty"`
    ProviderName string                      `json:"provider_name,omitempty"`
    ImageURL     *string                     `json:"image_url,omitempty"`
    Category     *EngagementCategoryResponse `json:"category,omitempty"`
    Offers       []EngagementOfferResponse   `json:"offers,omitempty"`
    Badge        string                      `json:"badge"`
}

type EngagementCategoryResponse struct {
    ID        uuid.UUID `json:"id"`
    Slug      string    `json:"slug"`
    Name      string    `json:"name"`
    Icon      string    `json:"icon,omitempty"`
    SortOrder int       `json:"sort_order"`
}

type EngagementOfferResponse struct {
    ID          uuid.UUID `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    CostCents   int64     `json:"cost_cents"`
    SortOrder   int       `json:"sort_order"`
}

type PaginationResponse struct {
    Page       int  `json:"page"`
    PerPage    int  `json:"per_page"`
    Total      int64 `json:"total"`
    TotalPages int  `json:"total_pages"`
}

type ListResponse struct {
    Data       []EngagementTypeResponse `json:"data"`
    Pagination PaginationResponse       `json:"pagination"`
}
```

### Request types (из `backend/internal/engagement/catalog/admin_handler.go`)

```go
type CreateCategoryRequest struct {
    Slug      string `json:"slug"`
    Name      string `json:"name"`
    Icon      string `json:"icon"`
    SortOrder int    `json:"sort_order"`
}

type CreateTypeRequest struct {
    CategoryID   uuid.UUID `json:"category_id"`
    Slug         string    `json:"slug"`
    Name         string    `json:"name"`
    Description  string    `json:"description"`
    Type         string    `json:"type"`
    Status       string    `json:"status"`
    CostCents    *int64    `json:"cost_cents"`
    ProviderName string    `json:"provider_name"`
    ImageURL     *string   `json:"image_url"`
}

type UpdateStatusRequest struct {
    Status string `json:"status"`
}

type CreateOfferRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    CostCents   int64  `json:"cost_cents"`
    SortOrder   int    `json:"sort_order"`
}
```

## Что сделать

### Установить openapi-typescript

```bash
cd frontend
npm install -D openapi-typescript
```

### Добавить скрипт в package.json

```json
{
  "scripts": {
    "generate-types": "openapi-typescript ./openapi/spec.yaml -o ./src/api/types.generated.ts"
  }
}
```

### OpenAPI spec

Создать `frontend/openapi/spec.yaml` с описанием всех M20 endpoints.
TODO: после реализации OpenAPI spec generator в backend (M23) заменить на автоматическую генерацию.

### Структура сгенерированных типов

Файл `src/api/types.generated.ts` должен содержать:

```ts
// Сгенерированные типы — не редактировать вручную
// Источник: openapi/spec.yaml
// Команда: npm run generate-types

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
  type: string
  status: string
  cost_cents?: number
  provider_name?: string
  image_url?: string
  category?: EngagementCategoryResponse
  offers?: EngagementOfferResponse[]
  badge: string
}

export interface PaginationResponse {
  page: number
  per_page: number
  total: number
  total_pages: number
}

export interface ListResponse {
  data: EngagementTypeResponse[]
  pagination: PaginationResponse
}
```

### .gitignore

```
# Не игнорируем сгенерированные типы
# src/api/types.generated.ts — коммитится в репозиторий
```

## Требования

- openapi-typescript для генерации
- Скрипт `npm run generate-types` в package.json
- Типы в `src/api/types.generated.ts`
- Типы соответствуют Go struct из handler.go и admin_handler.go
- Сгенерированные файлы коммитятся в репозиторий
- OpenAPI spec в `frontend/openapi/spec.yaml`

## Критерии приёмки

- [ ] `openapi-typescript` установлен как devDependency
- [ ] Скрипт `npm run generate-types` в package.json
- [ ] `frontend/openapi/spec.yaml` — OpenAPI spec для M20 endpoints
- [ ] `src/api/types.generated.ts` — сгенерированные типы
- [ ] Типы точно соответствуют Go response/request types
- [ ] uuid.UUID → string (TypeScript не имеет native UUID)
- [ ] *int64 → number | undefined (optional)
- [ ] int64 → number
- [ ] Сгенерированные файлы в git
