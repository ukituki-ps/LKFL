# T2002 — Отчёт

## Статус

✅ Выполнено

## Дата

2026-05-26

## Что сделано

### Созданы файлы

1. **`backend/internal/engagement/catalog/model.go`** — модели:
   - `EngagementType` — тип энгейджмента (льгота/активность)
   - `EngagementCategory` — категория энгейджментов
   - `EngagementOffer` — оффер (тариф/план)
   - `CatalogFilter` — фильтр для списка
   - Константы статусов: `StatusDraft`, `StatusActive`, `StatusPromo`, `StatusHidden`, `StatusCompleted`
   - Константы типов: `TypeBenefit`, `TypeActivity`

2. **`backend/internal/engagement/catalog/repository.go`** — Repository interface + pgx реализация:
   - `Repository` interface с 12 методами (Public, Admin Categories, Admin Types, Offers)
   - `pgRepository` — concrete type с `*pgxpool.Pool`
   - `NewRepository` — конструктор
   - Compile-time check: `var _ Repository = (*pgRepository)(nil)`
   - `ListTypes` — LEFT JOIN с категориями, N-фильтры (type, status, category, search ILIKE), пагинация, promo-first ordering
   - `GetTypeByID` — simple SELECT
   - `GetCategories` — SELECT ORDER BY sort_order
   - Admin CRUD для категорий и типов
   - CRUD для офферов
   - Soft delete через `status=hidden` для типов
   - Ошибки: `ErrNotFound`, `ErrCategoryNotFound`, `ErrOfferNotFound`, `ErrDuplicateSlug`
   - Empty slice pattern повсеместно

3. **`backend/internal/engagement/catalog/service.go`** — бизнес-логика:
   - `Service` struct с `Repository` (interface)
   - `NewService` конструктор
   - `ListTypes` — валидация фильтров и пагинации
   - `GetTypeByID` — делегирование + загрузка офферов
   - `GetCategories`, `GetOffersByType` — делегирование
   - `AdminCreateCategory` — проверка slug uniqueness
   - `AdminUpdateCategory` — проверка slug uniqueness
   - `AdminCreateType` — валидация type/status, проверка category exists, slug uniqueness
   - `AdminUpdateType` — валидация + проверка существования
   - `AdminDeleteType` — soft delete (TODO: проверка 0 активаций в F2)
   - `AdminCreateOffer` — проверка type exists
   - `AdminUpdateOffer`, `AdminDeleteOffer`
   - Ошибки: `ErrInvalidFilter`, `ErrInvalidStatus`, `ErrInvalidType`

4. **`backend/internal/engagement/catalog/service_test.go`** — unit-тесты:
   - `mockRepository` — полная мока всех методов Repository
   - 11 тестовых функций, 22 sub-теста
   - Покрытие: ListTypes, GetTypeByID, GetCategories, GetOffersByType, Admin CRUD, валидация, константы

## Критерии приёмки

| Критерий | Статус |
|----------|--------|
| model.go — structs | ✅ |
| repository.go — interface + pgx impl | ✅ |
| service.go — business logic | ✅ |
| ListTypes с N-фильтрами | ✅ |
| GetTypeByID с offers + category | ✅ |
| GetCategories | ✅ |
| Admin CRUD | ✅ |
| Pagination | ✅ |
| Promo first ordering | ✅ |
| Unit tests | ✅ (11 функций, 22 sub-теста) |

## Проверка

- `go build ./...` — чистая компиляция ✅
- `go test ./...` — все тесты зелёные ✅
- `tenant.JSONB` переиспользован из `lkfl/internal/tenant/model.go` ✅
- Паттерны user/repository.go и user/service.go соблюдены ✅
- Empty slice pattern повсеместно ✅
- Error wrapping: `fmt.Errorf("operation: %w", err)` ✅
- «Три нуля» — нет привязки к бренду/провайдеру ✅

## Замечания

- AdminDeleteType реализован как soft delete (status=hidden). Проверка на 0 активаций отложена на F2.
- Проверка slug uniqueness для типов использует ListTypes+Search (не оптимизально). В будущем добавить `GetBySlug` в Repository.
