# T2004 — Admin API: Каталог

## Веха

M20-catalog

## Тип

code

## Контекст

Admin API для управления каталогом. RBAC: catalog_manager, admin.

**После M19 — routing:**
- Routes монтируются в `server.go` admin group `/admin/` (T1904)
- Middleware chain: `sharedauth.JWTMiddleware(verifier)` + `sharedauth.RBACMiddleware(["catalog_manager", "admin"])` + `tenant.AdminTenantMiddleware()`
- Response: `shhttp.WriteJSON()`, `shhttp.WriteJSONError()` из `shared/pkg/http`
- Pattern: `user/handler.go` admin методы (T1905) — hasRole check, tenant isolation

**⚠️ Delete protection — STUB:**
- `DELETE /admin/engagements/types/:id` — проверка 0 user_engagements требует таблицу `user_engagements` (F2, M26)
- Пока: DELETE без проверки (soft delete через status=hidden)
- TODO: добавить проверку после M26

## Что сделать

### Admin handlers

```go
// Categories CRUD
// POST /admin/engagements/categories
func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request)
// PUT /admin/engagements/categories/:id
func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request)
// DELETE /admin/engagements/categories/:id
func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request)

// Types CRUD
// POST /admin/engagements/types
func (h *AdminHandler) CreateType(w http.ResponseWriter, r *http.Request)
// GET /admin/engagements/types?page=1&per_page=20
func (h *AdminHandler) ListTypes(w http.ResponseWriter, r *http.Request)
// GET /admin/engagements/types/:id
func (h *AdminHandler) GetType(w http.ResponseWriter, r *http.Request)
// PUT /admin/engagements/types/:id
func (h *AdminHandler) UpdateType(w http.ResponseWriter, r *http.Request)
// DELETE /admin/engagements/types/:id — только при 0 активациях
func (h *AdminHandler) DeleteType(w http.ResponseWriter, r *http.Request)
// PATCH /admin/engagements/types/:id/status — смена статуса
func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request)

// Offers CRUD
// POST /admin/engagements/types/:typeId/offers
func (h *AdminHandler) CreateOffer(w http.ResponseWriter, r *http.Request)
// PUT /admin/engagements/types/:typeId/offers/:id
func (h *AdminHandler) UpdateOffer(w http.ResponseWriter, r *http.Request)
// DELETE /admin/engagements/types/:typeId/offers/:id
func (h *AdminHandler) DeleteOffer(w http.ResponseWriter, r *http.Request)
```

### Status transitions

```
draft → active → promo → active → hidden → active → completed
```

- Draft → Active: публикация
- Active → Promo: продвижение (баннер + приоритет + push)
- Promo → Active: конец периода продвижения
- Active → Hidden: скрытие (не удаление)
- Hidden → Active: повторная публикация
- Active/Hidden → Completed: завершение (когда партнёр закрыт)

### Validation

- Slug uniqueness в рамках tenant
- Category exists перед creation type
- Type exists перед creation offer
- Delete type — только при 0 user_engagements (проверка в F2)

## Требования

- RBAC: catalog_manager, admin
- Status transition validation
- Slug uniqueness
- Delete protection (0 активаций)
- Request validation (validator)
- Error responses

## Критерии приёмки

- [ ] Categories CRUD
- [ ] Types CRUD
- [ ] Offers CRUD
- [ ] Status transitions
- [ ] Slug uniqueness validation
- [ ] Delete protection
- [ ] Request validation
- [ ] Unit tests
