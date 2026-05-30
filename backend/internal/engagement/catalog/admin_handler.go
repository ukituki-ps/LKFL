// Package catalog — каталог энгейджментов (льготы/активности).
//
// admin_handler.go — HTTP handlers для admin API каталога.
// RBAC: catalog_manager, admin.
package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"lkfl/internal/tenant"
	"lkfl/internal/user"
	"lkfl/shared/pkg/auth"
	shhttp "lkfl/shared/pkg/http"
)

// AdminHandler — HTTP handlers для admin API каталога.
type AdminHandler struct {
	service *Service
	cache   *Cache
}

// NewAdminHandler создаёт AdminHandler.
// Если cache == nil, инвалидация кэша отключена.
func NewAdminHandler(service *Service, cache *Cache) *AdminHandler {
	return &AdminHandler{service: service, cache: cache}
}

// invalidateCache инвалидирует кэш каталога для tenant'а.
// nil-safe: если cache == nil, ничего не делает.
func (h *AdminHandler) invalidateCache(ctx context.Context, tenantID string) {
	if h.cache != nil {
		_ = h.cache.Invalidate(ctx, tenantID)
	}
}

// hasCatalogRole проверяет наличие роли catalog_manager или admin.
func hasCatalogRole(roles []string) bool {
	for _, r := range roles {
		if r == user.RoleCatalogManager || r == user.RoleAdmin {
			return true
		}
	}
	return false
}

// validTransitions — допустимые переходы статусов.
var validTransitions = map[string][]string{
	StatusDraft:     {StatusActive},
	StatusActive:    {StatusPromo, StatusHidden, StatusCompleted},
	StatusPromo:     {StatusActive},
	StatusHidden:    {StatusActive},
	StatusCompleted: {}, // terminal
}

// --- Request body types ---

// CreateCategoryRequest — запрос для создания категории.
type CreateCategoryRequest struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// UpdateCategoryRequest — запрос для обновления категории.
type UpdateCategoryRequest struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

// CreateTypeRequest — запрос для создания типа энгейджмента.
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

// UpdateTypeRequest — запрос для обновления типа энгейджмента.
type UpdateTypeRequest struct {
	CategoryID   uuid.UUID `json:"category_id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Type         string    `json:"type"`
	CostCents    *int64    `json:"cost_cents"`
	ProviderName string    `json:"provider_name"`
	ImageURL     *string   `json:"image_url"`
}

// UpdateStatusRequest — запрос для смены статуса.
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// CreateOfferRequest — запрос для создания оффера.
type CreateOfferRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CostCents   int64  `json:"cost_cents"`
	SortOrder   int    `json:"sort_order"`
}

// UpdateOfferRequest — запрос для обновления оффера.
type UpdateOfferRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CostCents   int64  `json:"cost_cents"`
	SortOrder   int    `json:"sort_order"`
}

// --- Categories CRUD ---

// CreateCategory создаёт новую категорию.
// POST /admin/engagements/categories
func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat := EngagementCategory{
		TenantID:  tid,
		Slug:      req.Slug,
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}
	if req.Icon != "" {
		cat.Icon = &req.Icon
	}

	result, err := h.service.AdminCreateCategory(r.Context(), cat)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateSlug):
			shhttp.WriteJSONError(w, http.StatusConflict, "category slug already exists")
		case errors.Is(err, ErrInvalidFilter):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "slug and name are required")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "create category: "+err.Error())
		}
		return
	}

	h.invalidateCache(r.Context(), tid.String())
	shhttp.WriteJSON(w, http.StatusCreated, result)
}

// UpdateCategory обновляет категорию.
// PUT /admin/engagements/categories/:id
func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cat := EngagementCategory{
		ID:        id,
		TenantID:  tid,
		Slug:      req.Slug,
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}
	if req.Icon != "" {
		cat.Icon = &req.Icon
	}

	result, err := h.service.AdminUpdateCategory(r.Context(), cat)
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			shhttp.WriteJSONError(w, http.StatusNotFound, "category not found")
		case errors.Is(err, ErrDuplicateSlug):
			shhttp.WriteJSONError(w, http.StatusConflict, "category slug already exists")
		case errors.Is(err, ErrInvalidFilter):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "slug and name are required")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "update category: "+err.Error())
		}
		return
	}

	h.invalidateCache(r.Context(), tid.String())
	shhttp.WriteJSON(w, http.StatusOK, result)
}

// DeleteCategory удаляет категорию.
// DELETE /admin/engagements/categories/:id
func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid category ID")
		return
	}

	err = h.service.AdminDeleteCategory(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "category not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "delete category: "+err.Error())
		}
		return
	}

	// Invalidate cache — нужно получить tenantID из context
	tidDel := tenant.TenantIDFromContext(r.Context())
	if tidDel != uuid.Nil {
		h.invalidateCache(r.Context(), tidDel.String())
	}
	shhttp.WriteJSON(w, http.StatusNoContent, nil)
}

// --- Types CRUD ---

// CreateType создаёт новый тип энгейджмента.
// POST /admin/engagements/types
func (h *AdminHandler) CreateType(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	var req CreateTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	et := EngagementType{
		TenantID:   tid,
		CategoryID: req.CategoryID,
		Slug:       req.Slug,
		Name:       req.Name,
		Type:       req.Type,
		Status:     req.Status,
		CostCents:  req.CostCents,
		ImageURL:   req.ImageURL,
	}
	if req.Description != "" {
		et.Description = &req.Description
	}
	if req.ProviderName != "" {
		et.ProviderName = &req.ProviderName
	}

	result, err := h.service.AdminCreateType(r.Context(), et)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidType):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement type")
		case errors.Is(err, ErrInvalidStatus):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid status")
		case errors.Is(err, ErrDuplicateSlug):
			shhttp.WriteJSONError(w, http.StatusConflict, "slug already exists")
		case errors.Is(err, ErrCategoryNotFound):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "category not found")
		case errors.Is(err, ErrInvalidFilter):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "slug and name are required")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "create engagement type: "+err.Error())
		}
		return
	}

	h.invalidateCache(r.Context(), tid.String())
	shhttp.WriteJSON(w, http.StatusCreated, result.ToResponse())
}

// ListTypes возвращает список всех типов энгейджментов (все статусы).
// GET /admin/engagements/types?page=1&per_page=20&status=active&type=benefit&category=fitness&search=йога
func (h *AdminHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	q := r.URL.Query()

	// Пагинация
	page := 1
	if p := q.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	perPage := 20
	if pp := q.Get("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			perPage = v
		}
	}
	if perPage > 100 {
		perPage = 100
	}

	filter := CatalogFilter{
		TenantID: tid,
		Type:     q.Get("type"),
		Status:   q.Get("status"),
		Category: q.Get("category"),
		Search:   q.Get("search"),
		Page:     page,
		PerPage:  perPage,
	}

	types, total, err := h.service.AdminListTypes(r.Context(), filter)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidType):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid type filter")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "list engagements: "+err.Error())
		}
		return
	}

	// Преобразуем в response
	data := make([]EngagementTypeResponse, len(types))
	for i, t := range types {
		data[i] = t.ToResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	shhttp.WriteJSON(w, http.StatusOK, ListResponse{
		Data: data,
		Pagination: PaginationResponse{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetType возвращает детали типа энгейджмента.
// GET /admin/engagements/types/:id
func (h *AdminHandler) GetType(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement ID")
		return
	}

	et, err := h.service.GetTypeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "engagement not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get engagement: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, et.ToResponse())
}

// UpdateType обновляет тип энгейджмента.
// PUT /admin/engagements/types/:id
func (h *AdminHandler) UpdateType(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement ID")
		return
	}

	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	var req UpdateTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Получаем текущий тип
	current, err := h.service.GetTypeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "engagement not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get engagement: "+err.Error())
		}
		return
	}

	// Merge полей из запроса (partial update — не nullify пустые строки)
	if req.Slug != "" {
		current.Slug = req.Slug
	}
	if req.Name != "" {
		current.Name = req.Name
	}
	// Description: не перезаписываем пустой строкой (partial merge)
	// Если нужно очистить description, клиент должен явно передать ""
	// через отдельный endpoint или флаг. Пока оставляем как есть.
	if req.Description != "" {
		current.Description = &req.Description
	}
	if req.Type != "" {
		current.Type = req.Type
	}
	if req.CategoryID != uuid.Nil {
		current.CategoryID = req.CategoryID
	}
	if req.CostCents != nil {
		current.CostCents = req.CostCents
	}
	if req.ProviderName != "" {
		current.ProviderName = &req.ProviderName
	}
	if req.ImageURL != nil {
		current.ImageURL = req.ImageURL
	}
	current.TenantID = tid

	updated, err := h.service.AdminUpdateType(r.Context(), current)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shhttp.WriteJSONError(w, http.StatusNotFound, "engagement not found")
		case errors.Is(err, ErrInvalidType):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement type")
		case errors.Is(err, ErrInvalidStatus):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid status")
		case errors.Is(err, ErrDuplicateSlug):
			shhttp.WriteJSONError(w, http.StatusConflict, "slug already exists")
		case errors.Is(err, ErrCategoryNotFound):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "category not found")
		case errors.Is(err, ErrInvalidFilter):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "slug and name are required")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "update engagement: "+err.Error())
		}
		return
	}

	h.invalidateCache(r.Context(), tid.String())
	shhttp.WriteJSON(w, http.StatusOK, updated.ToResponse())
}

// DeleteType удаляет тип энгейджмента (soft delete → status=hidden).
// DELETE /admin/engagements/types/:id
// TODO: добавить проверку 0 активаций после M26 (Flow engine).
func (h *AdminHandler) DeleteType(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement ID")
		return
	}

	// TODO: проверить что нет активных активаций (F2, M26)

	err = h.service.AdminDeleteType(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "engagement not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "delete engagement: "+err.Error())
		}
		return
	}

	// Invalidate cache
	tidDel := tenant.TenantIDFromContext(r.Context())
	if tidDel != uuid.Nil {
		h.invalidateCache(r.Context(), tidDel.String())
	}
	shhttp.WriteJSON(w, http.StatusNoContent, nil)
}

// UpdateStatus меняет статус типа энгейджмента с валидацией переходов.
// PATCH /admin/engagements/types/:id/status
func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement ID")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Получаем текущий тип
	current, err := h.service.GetTypeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "engagement not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get engagement: "+err.Error())
		}
		return
	}

	// Валидация перехода статуса
	allowed, ok := validTransitions[current.Status]
	if !ok {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "unknown current status: "+current.Status)
		return
	}
	transitionAllowed := false
	for _, s := range allowed {
		if s == req.Status {
			transitionAllowed = true
			break
		}
	}
	if !transitionAllowed {
		shhttp.WriteJSONError(w, http.StatusBadRequest,
			"invalid status transition: "+current.Status+" → "+req.Status)
		return
	}

	// Обновляем статус
	current.Status = req.Status
	updated, err := h.service.AdminUpdateType(r.Context(), current)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "update status: "+err.Error())
		return
	}

	// Invalidate cache
	tidStatus := tenant.TenantIDFromContext(r.Context())
	if tidStatus != uuid.Nil {
		h.invalidateCache(r.Context(), tidStatus.String())
	}
	shhttp.WriteJSON(w, http.StatusOK, updated.ToResponse())
}

// --- Offers CRUD ---

// CreateOffer создаёт новый оффер.
// POST /admin/engagements/types/:typeId/offers
func (h *AdminHandler) CreateOffer(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	typeIDStr := chi.URLParam(r, "typeId")
	typeID, err := uuid.Parse(typeIDStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement type ID")
		return
	}

	var req CreateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	offer := EngagementOffer{
		TenantID:         tid,
		EngagementTypeID: typeID,
		Name:             req.Name,
		CostCents:        req.CostCents,
		SortOrder:        req.SortOrder,
	}
	if req.Description != "" {
		offer.Description = &req.Description
	}

	result, err := h.service.AdminCreateOffer(r.Context(), offer)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "engagement type not found")
		case errors.Is(err, ErrInvalidFilter):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "offer name is required")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "create offer: "+err.Error())
		}
		return
	}

	h.invalidateCache(r.Context(), tid.String())
	shhttp.WriteJSON(w, http.StatusCreated, result)
}

// UpdateOffer обновляет оффер.
// PUT /admin/engagements/types/:typeId/offers/:id
func (h *AdminHandler) UpdateOffer(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid offer ID")
		return
	}

	var req UpdateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	offer := EngagementOffer{
		ID:        id,
		Name:      req.Name,
		CostCents: req.CostCents,
		SortOrder: req.SortOrder,
	}
	if req.Description != "" {
		offer.Description = &req.Description
	}

	result, err := h.service.AdminUpdateOffer(r.Context(), offer)
	if err != nil {
		switch {
		case errors.Is(err, ErrOfferNotFound):
			shhttp.WriteJSONError(w, http.StatusNotFound, "offer not found")
		case errors.Is(err, ErrInvalidFilter):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "offer name is required")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "update offer: "+err.Error())
		}
		return
	}

	// Invalidate cache
	tidOffer := tenant.TenantIDFromContext(r.Context())
	if tidOffer != uuid.Nil {
		h.invalidateCache(r.Context(), tidOffer.String())
	}
	shhttp.WriteJSON(w, http.StatusOK, result)
}

// DeleteOffer удаляет оффер.
// DELETE /admin/engagements/types/:typeId/offers/:id
func (h *AdminHandler) DeleteOffer(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasCatalogRole(roles) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "catalog_manager or admin role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid offer ID")
		return
	}

	err = h.service.AdminDeleteOffer(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrOfferNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "offer not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "delete offer: "+err.Error())
		}
		return
	}

	// Invalidate cache
	tidDel := tenant.TenantIDFromContext(r.Context())
	if tidDel != uuid.Nil {
		h.invalidateCache(r.Context(), tidDel.String())
	}
	shhttp.WriteJSON(w, http.StatusNoContent, nil)
}
