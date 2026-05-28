package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"lkfl/internal/tenant"
	"lkfl/shared/pkg/auth"
	shhttp "lkfl/shared/pkg/http"
)

// Handler — HTTP handlers для пользователей.
type Handler struct {
	service *Service
}

// NewHandler создаёт Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Me — профиль текущего пользователя (employee).
// GET /api/v1/users/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	keycloakSub := auth.UserIDFromContext(r.Context())
	if keycloakSub == "" {
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetByKeycloakSub(r.Context(), keycloakSub)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "user not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get profile: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, user.ToProfile())
}

// UpdateMe — обновление профиля (employee).
// PUT /api/v1/users/me
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	keycloakSub := auth.UserIDFromContext(r.Context())
	if keycloakSub == "" {
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Получаем текущего пользователя
	current, err := h.service.GetByKeycloakSub(r.Context(), keycloakSub)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "user not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get user: "+err.Error())
		}
		return
	}

	var req struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), current.ID, req.Email, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserDeactivated):
			shhttp.WriteJSONError(w, http.StatusForbidden, "cannot update deactivated profile")
		case errors.Is(err, ErrDuplicateEmail):
			shhttp.WriteJSONError(w, http.StatusConflict, "email already taken")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "update profile: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, user.ToProfile())
}

// AdminList — список пользователей (admin/hr).
// GET /admin/users?page=1&per_page=20&status=active&search=email
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	// Проверка прав: admin или hr
	roles := auth.RolesFromContext(r.Context())
	if !hasRole(roles, RoleAdmin) && !hasRole(roles, RoleHR) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "admin or hr role required")
		return
	}

	// Tenant isolation — tenant ID берётся из context (установлен middleware)
	tid := tenant.TenantIDFromContext(r.Context())
	if tid == uuid.Nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "tenant context missing")
		return
	}

	filter := UserFilter{
		Status:  r.URL.Query().Get("status"),
		Search:  r.URL.Query().Get("search"),
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

	// Устанавливаем tenant ID в context для tenant isolation в repository
	ctx := tenant.TenantContext(r.Context(), tid)

	users, total, err := h.service.List(ctx, filter)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "list users: "+err.Error())
		return
	}

	// Преобразуем в профили
	profiles := make([]UserProfile, len(users))
	for i, u := range users {
		profiles[i] = u.ToProfile()
	}

	// Pagination headers
	w.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
	w.Header().Set("X-Page", strconv.Itoa(filter.Page))
	w.Header().Set("X-Per-Page", strconv.Itoa(filter.PerPage))

	shhttp.WriteJSON(w, http.StatusOK, profiles)
}

// AdminGet — детали пользователя (admin/hr).
// GET /admin/users/:id
func (h *Handler) AdminGet(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasRole(roles, RoleAdmin) && !hasRole(roles, RoleHR) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "admin or hr role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			shhttp.WriteJSONError(w, http.StatusNotFound, "user not found")
		} else {
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "get user: "+err.Error())
		}
		return
	}

	// Проверяем, что пользователь belongs to текущего tenant
	tid := tenant.TenantIDFromContext(r.Context())
	if tid != uuid.Nil && user.TenantID != tid {
		shhttp.WriteJSONError(w, http.StatusForbidden, "user belongs to another tenant")
		return
	}

	// Возвращаем полный объект (без keycloak_sub из-за json:"-")
	shhttp.WriteJSON(w, http.StatusOK, user)
}

// AdminUpdate — обновление пользователя (admin/hr).
// PUT /admin/users/:id
func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasRole(roles, RoleAdmin) && !hasRole(roles, RoleHR) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "admin or hr role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), id, req.Email, req.FirstName, req.LastName, req.Phone)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shhttp.WriteJSONError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, ErrUserDeactivated):
			shhttp.WriteJSONError(w, http.StatusForbidden, "cannot update deactivated user")
		case errors.Is(err, ErrDuplicateEmail):
			shhttp.WriteJSONError(w, http.StatusConflict, "email already taken")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "update user: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, user)
}

// AdminDeactivate — деактивация пользователя (admin/hr).
// POST /admin/users/:id/deactivate
func (h *Handler) AdminDeactivate(w http.ResponseWriter, r *http.Request) {
	// Проверка прав
	roles := auth.RolesFromContext(r.Context())
	if !hasRole(roles, RoleAdmin) && !hasRole(roles, RoleHR) {
		shhttp.WriteJSONError(w, http.StatusForbidden, "admin or hr role required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.service.Deactivate(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			shhttp.WriteJSONError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, ErrAlreadyDeactivated):
			shhttp.WriteJSONError(w, http.StatusConflict, "user is already deactivated")
		case errors.Is(err, ErrInvalidStatus):
			shhttp.WriteJSONError(w, http.StatusBadRequest, "user cannot be deactivated from current status")
		default:
			shhttp.WriteJSONError(w, http.StatusInternalServerError, "deactivate user: "+err.Error())
		}
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, user)
}

// hasRole проверяет наличие роли в списке.
func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
