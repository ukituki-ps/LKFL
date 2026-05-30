# T2003 — Отчёт

## Статус

✅ Выполнено

## Созданные файлы

| Файл | Описание |
|------|----------|
| `backend/internal/engagement/catalog/handler.go` | HTTP handlers для публичного API каталога |
| `backend/internal/engagement/catalog/handler_test.go` | Unit-тесты для handler'ов |

## Реализовано

### Handler struct
- `Handler` с привязкой к `*Service`
- `NewHandler(service *Service) *Handler` — конструктор

### Endpoint'ы

1. **`GET /api/v1/engagements`** (List)
   - Фильтры: `type`, `status`, `category`, `search`
   - Пагинация: `page` (default 1), `per_page` (default 20, max 100)
   - Response: `{data: [], pagination: {}}` с `total_pages = ceil(total/per_page)`
   - Error: 400 invalid type, 400 no tenant, 500 internal

2. **`GET /api/v1/engagements/:id`** (Get)
   - UUID из URL через `chi.URLParam`
   - Response: `EngagementTypeResponse` с category и offers
   - Error: 400 invalid UUID, 404 not found, 500 internal

3. **`GET /api/v1/engagements/categories`** (Categories)
   - Response: массив `EngagementCategoryResponse`
   - Error: 400 no tenant, 500 internal

### Response types
- `EngagementTypeResponse` — полный ответ с category, offers, badge
- `EngagementCategoryResponse` — категория
- `EngagementOfferResponse` — оффер
- `PaginationResponse` — пагинация
- `ListResponse` — обёртка List с data + pagination

### Badge computation (STUB)
- `computeBadge(et EngagementType) string` — возвращает "Промо" для promo, "Доступна" для остальных
- TODO: после M26 (Flow engine) — проверка user_engagements для бейджа "Активна"

### ToResponse
- `EngagementType.ToResponse()` — конвертация model → response с category и offers

### Error handling
- 400: invalid filter, missing tenant, invalid UUID
- 404: engagement not found
- 500: internal service/repository errors
- Используется `shhttp.WriteJSONError()` для всех ошибок

## Тесты

28 unit-тестов (все зелёные):
- List: empty catalog, with items, filter by type, invalid type, no tenant, pagination, per_page cap, category filter, search filter, status filter, total_pages calculation, response fields
- Get: success, with category, with offers, not found, invalid ID
- Categories: empty, with items, no tenant
- Badge: promo, active, draft, hidden
- ToResponse: with category, with offers, without category, promo badge

## Проверка критериев приёмки

- [x] `handler.go` — List, Get, Categories
- [x] GET /api/v1/engagements — фильтры + pagination
- [x] GET /api/v1/engagements/:id — детали
- [x] GET /api/v1/engagements/categories — категории
- [x] Badge computation
- [x] Pagination response
- [x] Error handling
- [x] Unit tests

## Компиляция

```
go build ./...          # ✅ чистая
go test ./...           # ✅ 47 тестов (28 handler + 19 service)
```

## Замечания

- Badge computation — STUB (только "Промо"/"Доступна"). Полная логика "Активна"/"Новинка" после M26.
- Tenant isolation через `tenant.TenantIDFromContext(r.Context())` — middleware уже устанавливает tenant ID.
- Паттерн handler полностью следует `user/handler.go`.
