# T2003 — Public API: Каталог

## Веха

M20-catalog

## Тип

code

## Контекст

Public API endpoints для каталога. Доступен всем аутентифицированным пользователям.

## Что сделать

### `internal/engagement/catalog/handler.go`

```go
package catalog

// List — список энгейджментов с фильтрами
// GET /api/v1/engagements?type=benefit&status=active&category=fitness&search=йога&page=1&per_page=20
// Response:
// {
//   "data": [EngagementTypeResponse],
//   "pagination": { "page": 1, "per_page": 20, "total": 150, "total_pages": 8 }
// }
func (h *Handler) List(w http.ResponseWriter, r *http.Request)

// Get — детали энгейджмента
// GET /api/v1/engagements/:id
// Response: EngagementTypeResponse с category, offers
func (h *Handler) Get(w http.ResponseWriter, r *http.Request)

// Categories — список категорий
// GET /api/v1/engagements/categories
// Response: [EngagementCategoryResponse]
func (h *Handler) Categories(w http.ResponseWriter, r *http.Request)
```

### Response format

```go
type EngagementTypeResponse struct {
    ID           uuid.UUID                    `json:"id"`
    Slug         string                       `json:"slug"`
    Name         string                       `json:"name"`
    Description  string                       `json:"description,omitempty"`
    Type         string                       `json:"type"`
    Status       string                       `json:"status"`
    CostCents    *int64                       `json:"cost_cents,omitempty"`
    ProviderName string                       `json:"provider_name,omitempty"`
    ImageURL     *string                      `json:"image_url,omitempty"`
    Category     *EngagementCategoryResponse  `json:"category,omitempty"`
    Offers       []EngagementOfferResponse    `json:"offers,omitempty"`
    Badge        string                       `json:"badge"` // Активна, Доступна, Новинка, Промо
}

type PaginationResponse struct {
    Page      int `json:"page"`
    PerPage   int `json:"per_page"`
    Total     int64 `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

### Badge logic

```go
func computeBadge(et EngagementType, userEngagements []UserEngagement) string {
    // Если у пользователя уже активна эта льгота → "Активна"
    // Если status = promo → "Промо"
    // Если status = active и есть новые → "Новинка"
    // Иначе → "Доступна"
}
```

## Требования

- GET /api/v1/engagements — список с фильтрами + pagination
- GET /api/v1/engagements/:id — детали с category + offers
- GET /api/v1/engagements/categories — список категорий
- JWT middleware + tenant middleware
- Badge computation (Активна/Доступна/Новинка/Промо)
- Pagination response format
- Error handling: 404 not found, 400 bad request

## Критерии приёмки

- [ ] `handler.go` — List, Get, Categories
- [ ] GET /api/v1/engagements — фильтры + pagination
- [ ] GET /api/v1/engagements/:id — детали
- [ ] GET /api/v1/engagements/categories — категории
- [ ] Badge computation
- [ ] Pagination response
- [ ] Error handling
- [ ] Unit tests
