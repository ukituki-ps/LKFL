// Package catalog — каталог энгейджментов (льготы/активности).
//
// model.go       — EngagementType, EngagementCategory, EngagementOffer, CatalogFilter
// repository.go  — Repository interface + pgx реализация
// service.go     — бизнес-логика (валидация, status transitions)
// handler.go     — HTTP handlers для публичного API каталога
package catalog

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"lkfl/internal/metrics"
	"lkfl/internal/tenant"
	shhttp "lkfl/shared/pkg/http"
)

// Handler — HTTP handlers для публичного API каталога.
type Handler struct {
	service *Service
	metrics *metrics.Metrics
}

// NewHandler создаёт Handler.
// Если m == nil, метрики не собираются.
func NewHandler(service *Service, m *metrics.Metrics) *Handler {
	return &Handler{service: service, metrics: m}
}

// EngagementTypeResponse — ответ для энгейджмента в публичном API.
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
	BadgeColor   string                      `json:"badge_color"`
	IconName     string                      `json:"icon_name,omitempty"`
	PriceDisplay string                      `json:"price_display,omitempty"`
}

// EngagementCategoryResponse — ответ для категории в публичном API.
type EngagementCategoryResponse struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon,omitempty"`
	SortOrder int       `json:"sort_order"`
}

// EngagementOfferResponse — ответ для оффера в публичном API.
type EngagementOfferResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CostCents   int64     `json:"cost_cents"`
	SortOrder   int       `json:"sort_order"`
}

// PaginationResponse — информация о пагинации в ответе.
type PaginationResponse struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListResponse — ответ для List endpoint'а.
type ListResponse struct {
	Data       []EngagementTypeResponse `json:"data"`
	Pagination PaginationResponse       `json:"pagination"`
}

// ToResponse преобразует EngagementType в EngagementTypeResponse.
func (et EngagementType) ToResponse() EngagementTypeResponse {
	r := EngagementTypeResponse{
		ID:           et.ID,
		Slug:         et.Slug,
		Name:         et.Name,
		Type:         et.Type,
		Status:       et.Status,
		CostCents:    et.CostCents,
		ImageURL:     et.ImageURL,
		Badge:        extractBadge(et),
		BadgeColor:   extractBadgeColor(et),
		IconName:     extractIconName(et),
		PriceDisplay: extractPriceDisplay(et),
	}
	if et.Description != nil {
		r.Description = *et.Description
	}
	if et.ProviderName != nil {
		r.ProviderName = *et.ProviderName
	}

	if et.Category != nil {
		catResp := EngagementCategoryResponse{
			ID:        et.Category.ID,
			Slug:      et.Category.Slug,
			Name:      et.Category.Name,
			SortOrder: et.Category.SortOrder,
		}
		if et.Category.Icon != nil {
			catResp.Icon = *et.Category.Icon
		}
		r.Category = &catResp
	}

	if len(et.Offers) > 0 {
		r.Offers = make([]EngagementOfferResponse, len(et.Offers))
		for i, o := range et.Offers {
			offerResp := EngagementOfferResponse{
				ID:        o.ID,
				Name:      o.Name,
				CostCents: o.CostCents,
				SortOrder: o.SortOrder,
			}
			if o.Description != nil {
				offerResp.Description = *o.Description
			}
			r.Offers[i] = offerResp
		}
	}

	return r
}

// extractBadge извлекает бейдж из metadata или вычисляет по статусу.
func extractBadge(et EngagementType) string {
	if v := et.Metadata.GetString("badge"); v != "" {
		return v
	}
	return computeBadge(et)
}

// extractBadgeColor извлекает цвет бейджа из metadata.
func extractBadgeColor(et EngagementType) string {
	return et.Metadata.GetString("badge_color")
}

// extractIconName извлекает имя Lucide-иконки из metadata.
func extractIconName(et EngagementType) string {
	return et.Metadata.GetString("icon_name")
}

// extractPriceDisplay извлекает строку цены из metadata.price.display.
func extractPriceDisplay(et EngagementType) string {
	price := et.Metadata.Get("price")
	if price == nil {
		return ""
	}
	if m, ok := price.(map[string]interface{}); ok {
		if d, ok := m["display"].(string); ok {
			return d
		}
	}
	return ""
}

// computeBadge вычисляет бейдж для энгейджмента.
// STUB: пока возвращает "Промо" для промо-статуса и "Доступна" для остальных.
// TODO: реализовать после M26 (Flow engine) — проверка user_engagements.
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

// List — список энгейджментов с фильтрами и пагинацией.
// GET /api/v1/engagements?type=benefit&status=active&category=fitness&search=йога&page=1&per_page=20
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Tenant isolation
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

	types, total, err := h.service.ListTypes(r.Context(), filter)
	if err != nil {
		if h.metrics != nil {
			h.metrics.CatalogQueryTotal.WithLabelValues("list", "error").Inc()
		}
		switch {
		case errors.Is(err, ErrInvalidType):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid type filter")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "list engagements: "+err.Error())
		}
		return
	}
	if h.metrics != nil {
		h.metrics.CatalogQueryTotal.WithLabelValues("list", "success").Inc()
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

// Get — детали энгейджмента по ID.
// GET /api/v1/engagements/:id
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid engagement ID")
		return
	}

	// Tenant isolation
	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	et, err := h.service.GetTypeByTenantID(r.Context(), tid, id)
	if err != nil {
		if h.metrics != nil {
			h.metrics.CatalogQueryTotal.WithLabelValues("get", "error").Inc()
		}
		switch {
		case errors.Is(err, ErrNotFound):
			shhttp.WriteJSONError(w, http.StatusNotFound, "engagement not found")
		case errors.Is(err, ErrTenantMismatch):
			shhttp.WriteJSONError(w, http.StatusNotFound, "engagement not found")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get engagement: "+err.Error())
		}
		return
	}
	if h.metrics != nil {
		h.metrics.CatalogQueryTotal.WithLabelValues("get", "success").Inc()
	}

	shhttp.WriteJSON(w, http.StatusOK, et.ToResponse())
}

// Categories — список категорий.
// GET /api/v1/engagements/categories
func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) {
	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	categories, err := h.service.GetCategories(r.Context(), tid)
	if err != nil {
		if h.metrics != nil {
			h.metrics.CatalogQueryTotal.WithLabelValues("categories", "error").Inc()
		}
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "get categories: "+err.Error())
		return
	}
	if h.metrics != nil {
		h.metrics.CatalogQueryTotal.WithLabelValues("categories", "success").Inc()
	}

	// Преобразуем в response
	res := make([]EngagementCategoryResponse, len(categories))
	for i, c := range categories {
		catResp := EngagementCategoryResponse{
			ID:        c.ID,
			Slug:      c.Slug,
			Name:      c.Name,
			SortOrder: c.SortOrder,
		}
		if c.Icon != nil {
			catResp.Icon = *c.Icon
		}
		res[i] = catResp
	}

	shhttp.WriteJSON(w, http.StatusOK, res)
}
