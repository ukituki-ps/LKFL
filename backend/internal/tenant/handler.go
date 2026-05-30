package tenant

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	shhttp "lkfl/shared/pkg/http"
)

// Handler — HTTP handlers для tenant CRUD (admin only).
type Handler struct {
	service *Service
}

// NewHandler создаёт новый Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create — POST /admin/tenants — создать tenant (admin only).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Slug   string `json:"slug"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	t := Tenant{
		Slug:   strings.ToLower(strings.TrimSpace(req.Slug)),
		Name:   req.Name,
		Status: req.Status,
	}

	tenant, err := h.service.CreateTenant(r.Context(), t)
	if err != nil {
		switch {
		case err == ErrInvalidSlug:
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid slug format: lowercase alphanumeric with hyphens")
		case err == ErrSlugExists:
			shhttp.WriteJSONError(w, http.StatusConflict, "tenant slug already exists")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "create tenant: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusCreated, tenant)
}

// List — GET /admin/tenants — список tenants (admin only).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter := TenantFilter{
		Status:  r.URL.Query().Get("status"),
		Page:    1,
		PerPage: 20,
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			filter.Page = v
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			filter.PerPage = v
		}
	}

	tenants, total, err := h.service.ListTenants(r.Context(), filter)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "list tenants: "+err.Error())
		return
	}

	// Pagination headers
	w.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
	w.Header().Set("X-Page", strconv.Itoa(filter.Page))
	w.Header().Set("X-Per-Page", strconv.Itoa(filter.PerPage))

	shhttp.WriteJSON(w, http.StatusOK, tenants)
}

// GetByID — GET /admin/tenants/:id — детали tenant (admin only).
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	tenant, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			shhttp.WriteJSONError(w, http.StatusNotFound, "tenant not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get tenant: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, tenant)
}

// Update — PUT /admin/tenants/:id — обновить tenant (admin only).
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	var req struct {
		Slug   string `json:"slug"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Сначала получаем текущие данные
	current, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			shhttp.WriteJSONError(w, http.StatusNotFound, "tenant not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get tenant: "+err.Error())
		}
		return
	}

	// Обновляем только переданные поля
	t := current
	if req.Slug != "" {
		t.Slug = strings.ToLower(strings.TrimSpace(req.Slug))
	}
	if req.Name != "" {
		t.Name = req.Name
	}
	if req.Status != "" {
		t.Status = req.Status
	}

	tenant, err := h.service.UpdateTenant(r.Context(), t)
	if err != nil {
		switch {
		case err == ErrInvalidSlug:
			shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid slug format")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "update tenant: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, tenant)
}

// Delete — DELETE /admin/tenants/:id — удалить tenant (admin only).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	if err := h.service.DeleteTenant(r.Context(), id); err != nil {
		if err == ErrNotFound {
			shhttp.WriteJSONError(w, http.StatusNotFound, "tenant not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "delete tenant: "+err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBrandConfig — GET /admin/tenants/:id/brand — brand config (admin only).
func (h *Handler) GetBrandConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	bc, err := h.service.GetBrandConfig(r.Context(), id)
	if err != nil {
		if err == ErrBrandNotFound {
			shhttp.WriteJSONError(w, http.StatusNotFound, "brand config not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get brand config: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, bc)
}

// UpsertBrandConfig — PUT /admin/tenants/:id/brand — upsert brand config (admin only).
func (h *Handler) UpsertBrandConfig(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid tenant ID")
		return
	}

	var req struct {
		PrimaryColor    string `json:"primary_color"`
		SecondaryColor  string `json:"secondary_color"`
		LogoURL         string `json:"logo_url"`
		FaviconURL      string `json:"favicon_url"`
		BrandName       string `json:"brand_name"`
		CSSVariables    JSONB  `json:"css_variables"`
		MetaTitle       string `json:"meta_title"`
		MetaDescription string `json:"meta_description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bc := BrandConfig{
		TenantID:       id,
		PrimaryColor:   req.PrimaryColor,
		SecondaryColor: req.SecondaryColor,
		CSSVariables:   req.CSSVariables,
	}

	if req.LogoURL != "" {
		bc.LogoURL = &req.LogoURL
	}
	if req.FaviconURL != "" {
		bc.FaviconURL = &req.FaviconURL
	}
	if req.BrandName != "" {
		bc.BrandName = &req.BrandName
	}
	if req.MetaTitle != "" {
		bc.MetaTitle = &req.MetaTitle
	}
	if req.MetaDescription != "" {
		bc.MetaDescription = &req.MetaDescription
	}

	result, err := h.service.UpsertBrandConfig(r.Context(), bc)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "upsert brand config: "+err.Error())
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, result)
}
