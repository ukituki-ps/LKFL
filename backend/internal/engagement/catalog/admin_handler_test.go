package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"lkfl/internal/tenant"
	"lkfl/internal/user"
	"lkfl/shared/pkg/auth"
)

// makeAdminRequest создаёт тестовый admin запрос с tenant context и ролями.
func makeAdminRequest(method, path string, tid uuid.UUID, roles []string, body interface{}) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	ctx := tenant.TenantContext(req.Context(), tid)
	ctx = context.WithValue(ctx, auth.RolesKey, roles)
	return req.WithContext(ctx)
}

// makeAdminRequestWithParam создаёт admin запрос с chi URL param.
func makeAdminRequestWithParam(method, path string, tid uuid.UUID, roles []string, paramKey, paramValue string, body interface{}) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	ctx := tenant.TenantContext(req.Context(), tid)
	ctx = context.WithValue(ctx, auth.RolesKey, roles)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(paramKey, paramValue)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return req.WithContext(ctx)
}

// makeAdminRequestWithTypeParam создаёт admin запрос с typeId param (для офферов).
func makeAdminRequestWithTypeParam(method, path string, tid uuid.UUID, roles []string, typeID, offerID string, body interface{}) *http.Request {
	var reader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader([]byte{})
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	ctx := tenant.TenantContext(req.Context(), tid)
	ctx = context.WithValue(ctx, auth.RolesKey, roles)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("typeId", typeID)
	rctx.URLParams.Add("id", offerID)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return req.WithContext(ctx)
}

var adminRoles = []string{user.RoleCatalogManager}

// --- hasCatalogRole tests ---

func TestHasCatalogRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  bool
	}{
		{"catalog_manager", []string{user.RoleCatalogManager}, true},
		{"admin", []string{user.RoleAdmin}, true},
		{"both roles", []string{user.RoleCatalogManager, user.RoleAdmin}, true},
		{"employee only", []string{user.RoleEmployee}, false},
		{"hr only", []string{user.RoleHR}, false},
		{"empty", []string{}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasCatalogRole(tt.roles)
			if got != tt.want {
				t.Errorf("hasCatalogRole(%v) = %v, want %v", tt.roles, got, tt.want)
			}
		})
	}
}

// --- validTransitions tests ---

func TestValidTransitions(t *testing.T) {
	tests := []struct {
		from    string
		to      string
		allowed bool
	}{
		{StatusDraft, StatusActive, true},
		{StatusDraft, StatusPromo, false},
		{StatusActive, StatusPromo, true},
		{StatusActive, StatusHidden, true},
		{StatusActive, StatusCompleted, true},
		{StatusActive, StatusDraft, false},
		{StatusPromo, StatusActive, true},
		{StatusPromo, StatusHidden, false},
		{StatusHidden, StatusActive, true},
		{StatusHidden, StatusPromo, false},
		{StatusCompleted, StatusActive, false}, // terminal
		{StatusCompleted, StatusHidden, false}, // terminal
	}

	for _, tt := range tests {
		t.Run(tt.from+"->"+tt.to, func(t *testing.T) {
			allowed, ok := validTransitions[tt.from]
			if !ok {
				t.Fatalf("unknown status: %s", tt.from)
			}
			found := false
			for _, s := range allowed {
				if s == tt.to {
					found = true
					break
				}
			}
			if found != tt.allowed {
				t.Errorf("transition %s -> %s: allowed=%v, want=%v", tt.from, tt.to, found, tt.allowed)
			}
		})
	}
}

// --- RBAC tests ---

func TestAdminHandler_CreateCategory_NoRole(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/categories", tid, []string{user.RoleEmployee}, nil)

	h.CreateCategory(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestAdminHandler_CreateType_NoRole(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/types", tid, []string{user.RoleHR}, nil)

	h.CreateType(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestAdminHandler_ListTypes_NoRole(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("GET", "/admin/engagements/types", tid, []string{user.RoleEmployee}, nil)

	h.ListTypes(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

// --- Category tests ---

func TestAdminHandler_CreateCategory_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/categories", tid, adminRoles, CreateCategoryRequest{
		Slug:      "health",
		Name:      "Здоровье",
		Icon:      "heart",
		SortOrder: 1,
	})

	h.CreateCategory(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var cat EngagementCategory
	if err := json.NewDecoder(w.Body).Decode(&cat); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if cat.Slug != "health" {
		t.Errorf("expected slug 'health', got %s", cat.Slug)
	}
	if cat.TenantID != tid {
		t.Errorf("expected tenant ID %s, got %s", tid, cat.TenantID)
	}
}

func TestAdminHandler_CreateCategory_DuplicateSlug(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	// Add existing category
	existing := EngagementCategory{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "health",
		Name:     "Здоровье",
	}
	repo.categories[existing.ID] = existing

	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/categories", tid, adminRoles, CreateCategoryRequest{
		Slug: "health",
		Name: "Дубликат",
	})

	h.CreateCategory(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestAdminHandler_CreateCategory_NoTenant(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/engagements/categories", bytes.NewReader([]byte("{}")))
	ctx := context.WithValue(req.Context(), auth.RolesKey, adminRoles)
	req = req.WithContext(ctx)
	// No tenant context

	h.CreateCategory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateCategory_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	catID := uuid.New()
	cat := EngagementCategory{
		ID:       catID,
		TenantID: tid,
		Slug:     "health",
		Name:     "Здоровье",
	}
	repo.categories[catID] = cat

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("PUT", "/admin/engagements/categories/"+catID.String(), tid, adminRoles, "id", catID.String(), UpdateCategoryRequest{
		Slug:      "health",
		Name:      "Здоровье обновлённое",
		Icon:      "heart-pulse",
		SortOrder: 2,
	})

	h.UpdateCategory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementCategory
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "Здоровье обновлённое" {
		t.Errorf("expected updated name, got %s", resp.Name)
	}
}

func TestAdminHandler_UpdateCategory_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	notFoundID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("PUT", "/admin/engagements/categories/"+notFoundID.String(), tid, adminRoles, "id", notFoundID.String(), UpdateCategoryRequest{
		Slug: "health",
		Name: "Здоровье",
	})

	h.UpdateCategory(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAdminHandler_DeleteCategory_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	catID := uuid.New()
	cat := EngagementCategory{
		ID:       catID,
		TenantID: tid,
		Slug:     "health",
		Name:     "Здоровье",
	}
	repo.categories[catID] = cat

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("DELETE", "/admin/engagements/categories/"+catID.String(), tid, adminRoles, "id", catID.String(), nil)

	h.DeleteCategory(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestAdminHandler_DeleteCategory_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	notFoundID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("DELETE", "/admin/engagements/categories/"+notFoundID.String(), tid, adminRoles, "id", notFoundID.String(), nil)

	h.DeleteCategory(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Type tests ---

func TestAdminHandler_CreateType_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/types", tid, adminRoles, CreateTypeRequest{
		Slug:         "gym-membership",
		Name:         "Абонемент в спортзал",
		Description:  "Полный доступ",
		Type:         TypeBenefit,
		ProviderName: "FitLife",
	})

	h.CreateType(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp EngagementTypeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Slug != "gym-membership" {
		t.Errorf("expected slug 'gym-membership', got %s", resp.Slug)
	}
	if resp.Type != TypeBenefit {
		t.Errorf("expected type 'benefit', got %s", resp.Type)
	}
	// Default status should be draft
	if resp.Status != StatusDraft {
		t.Errorf("expected default status 'draft', got %s", resp.Status)
	}
}

func TestAdminHandler_CreateType_WithCategory(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	catID := uuid.New()
	cat := EngagementCategory{
		ID:       catID,
		TenantID: tid,
		Slug:     "fitness",
		Name:     "Фитнес",
	}
	repo.categories[catID] = cat

	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/types", tid, adminRoles, CreateTypeRequest{
		CategoryID: catID,
		Slug:       "gym",
		Name:       "Спортзал",
	})

	h.CreateType(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestAdminHandler_CreateType_CategoryNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/types", tid, adminRoles, CreateTypeRequest{
		CategoryID: uuid.New(), // Non-existent category
		Slug:       "gym",
		Name:       "Спортзал",
	})

	h.CreateType(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_CreateType_DuplicateSlug(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	// Add existing type with same slug
	existing := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Type:     TypeBenefit,
		Status:   StatusActive,
	}
	repo.types[existing.ID] = existing

	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/types", tid, adminRoles, CreateTypeRequest{
		Slug: "gym",
		Name: "Другой спортзал",
	})

	h.CreateType(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestAdminHandler_CreateType_InvalidType(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/types", tid, adminRoles, CreateTypeRequest{
		Slug: "test",
		Name: "Test",
		Type: "invalid",
	})

	h.CreateType(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_CreateType_InvalidStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/types", tid, adminRoles, CreateTypeRequest{
		Slug:   "test",
		Name:   "Test",
		Status: "invalid",
	})

	h.CreateType(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_ListTypes_AllStatuses(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	// Add types with different statuses
	draft := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "draft-type",
		Name:     "Draft",
		Type:     TypeBenefit,
		Status:   StatusDraft,
	}
	active := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "active-type",
		Name:     "Active",
		Type:     TypeBenefit,
		Status:   StatusActive,
	}
	hidden := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "hidden-type",
		Name:     "Hidden",
		Type:     TypeBenefit,
		Status:   StatusHidden,
	}
	repo.types[draft.ID] = draft
	repo.types[active.ID] = active
	repo.types[hidden.ID] = hidden

	w := httptest.NewRecorder()
	req := makeAdminRequest("GET", "/admin/engagements/types", tid, adminRoles, nil)

	h.ListTypes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// Admin sees all statuses (draft, active, hidden)
	if resp.Pagination.Total != 3 {
		t.Errorf("expected total 3 (all statuses), got %d", resp.Pagination.Total)
	}
}

func TestAdminHandler_ListTypes_FilterByStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	active := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "active-type",
		Name:     "Active",
		Type:     TypeBenefit,
		Status:   StatusActive,
	}
	draft := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "draft-type",
		Name:     "Draft",
		Type:     TypeBenefit,
		Status:   StatusDraft,
	}
	repo.types[active.ID] = active
	repo.types[draft.ID] = draft

	w := httptest.NewRecorder()
	req := makeAdminRequest("GET", "/admin/engagements/types?status=active", tid, adminRoles, nil)

	h.ListTypes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("expected total 1 (only active), got %d", resp.Pagination.Total)
	}
}

func TestAdminHandler_GetType_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Type:     TypeBenefit,
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("GET", "/admin/engagements/types/"+typeID.String(), tid, adminRoles, "id", typeID.String(), nil)

	h.GetType(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementTypeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ID != typeID {
		t.Errorf("expected ID %s, got %s", typeID, resp.ID)
	}
}

func TestAdminHandler_GetType_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	notFoundID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("GET", "/admin/engagements/types/"+notFoundID.String(), tid, adminRoles, "id", notFoundID.String(), nil)

	h.GetType(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAdminHandler_GetType_InvalidID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("GET", "/admin/engagements/types/not-a-uuid", tid, adminRoles, "id", "not-a-uuid", nil)

	h.GetType(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateType_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Type:     TypeBenefit,
		Status:   StatusDraft,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("PUT", "/admin/engagements/types/"+typeID.String(), tid, adminRoles, "id", typeID.String(), UpdateTypeRequest{
		Name: "Премиум Спортзал",
	})

	h.UpdateType(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementTypeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "Премиум Спортзал" {
		t.Errorf("expected updated name, got %s", resp.Name)
	}
}

func TestAdminHandler_UpdateType_DuplicateSlug(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID1 := uuid.New()
	typeID2 := uuid.New()

	et1 := EngagementType{
		ID:       typeID1,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	et2 := EngagementType{
		ID:       typeID2,
		TenantID: tid,
		Slug:     "fitness",
		Name:     "Фитнес",
		Status:   StatusActive,
	}
	repo.types[typeID1] = et1
	repo.types[typeID2] = et2

	// Попытка изменить slug et2 на "gym" (занято et1)
	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("PUT", "/admin/engagements/types/"+typeID2.String(), tid, adminRoles, "id", typeID2.String(), UpdateTypeRequest{
		Slug: "gym",
	})

	h.UpdateType(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409 (duplicate slug), got %d", w.Code)
	}
}

func TestAdminHandler_UpdateType_PreserveProviderName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:           typeID,
		TenantID:     tid,
		Slug:         "gym",
		Name:         "Спортзал",
		Type:         TypeBenefit,
		Status:       StatusDraft,
		ProviderName: strPtr("FitLife"),
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	// Update только name — provider_name не передан
	req := makeAdminRequestWithParam("PUT", "/admin/engagements/types/"+typeID.String(), tid, adminRoles, "id", typeID.String(), UpdateTypeRequest{
		Name: "Премиум Спортзал",
	})

	h.UpdateType(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementTypeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// ProviderName должен сохраниться
	if resp.ProviderName != "FitLife" {
		t.Errorf("expected ProviderName 'FitLife' preserved, got '%s'", resp.ProviderName)
	}
}

func TestAdminHandler_DeleteType_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("DELETE", "/admin/engagements/types/"+typeID.String(), tid, adminRoles, "id", typeID.String(), nil)

	h.DeleteType(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestAdminHandler_DeleteType_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	notFoundID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("DELETE", "/admin/engagements/types/"+notFoundID.String(), tid, adminRoles, "id", notFoundID.String(), nil)

	h.DeleteType(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// --- Status transition tests ---

func TestAdminHandler_UpdateStatus_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusDraft,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("PATCH", "/admin/engagements/types/"+typeID.String()+"/status", tid, adminRoles, "id", typeID.String(), UpdateStatusRequest{
		Status: StatusActive,
	})

	h.UpdateStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementTypeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Status != StatusActive {
		t.Errorf("expected status 'active', got %s", resp.Status)
	}
}

func TestAdminHandler_UpdateStatus_InvalidTransition(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusDraft,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	// Draft → Promo is not allowed
	req := makeAdminRequestWithParam("PATCH", "/admin/engagements/types/"+typeID.String()+"/status", tid, adminRoles, "id", typeID.String(), UpdateStatusRequest{
		Status: StatusPromo,
	})

	h.UpdateStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateStatus_TerminalStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusCompleted,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	// Completed → anything is not allowed
	req := makeAdminRequestWithParam("PATCH", "/admin/engagements/types/"+typeID.String()+"/status", tid, adminRoles, "id", typeID.String(), UpdateStatusRequest{
		Status: StatusActive,
	})

	h.UpdateStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateStatus_FullLifecycle(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusDraft,
	}
	repo.types[typeID] = et

	transitions := []struct {
		to       string
		expected int
	}{
		{StatusActive, http.StatusOK},    // draft → active
		{StatusPromo, http.StatusOK},     // active → promo
		{StatusActive, http.StatusOK},    // promo → active
		{StatusHidden, http.StatusOK},    // active → hidden
		{StatusActive, http.StatusOK},    // hidden → active
		{StatusCompleted, http.StatusOK}, // active → completed
	}

	for i, tr := range transitions {
		w := httptest.NewRecorder()
		req := makeAdminRequestWithParam("PATCH", "/admin/engagements/types/"+typeID.String()+"/status", tid, adminRoles, "id", typeID.String(), UpdateStatusRequest{
			Status: tr.to,
		})

		h.UpdateStatus(w, req)

		if w.Code != tr.expected {
			t.Errorf("transition %d (%s → %s): expected %d, got %d", i, et.Status, tr.to, tr.expected, w.Code)
		} else {
			et.Status = tr.to
		}
	}
}

// --- Offer tests ---

func TestAdminHandler_CreateOffer_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	req := makeAdminRequestWithTypeParam("POST", "/admin/engagements/types/"+typeID.String()+"/offers", tid, adminRoles, typeID.String(), "", CreateOfferRequest{
		Name:        "Basic Plan",
		Description: "Базовый доступ",
		CostCents:   1000,
		SortOrder:   1,
	})

	h.CreateOffer(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var offer EngagementOffer
	if err := json.NewDecoder(w.Body).Decode(&offer); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if offer.Name != "Basic Plan" {
		t.Errorf("expected name 'Basic Plan', got %s", offer.Name)
	}
	if offer.CostCents != 1000 {
		t.Errorf("expected cost_cents 1000, got %d", offer.CostCents)
	}
}

func TestAdminHandler_CreateOffer_TypeNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	notFoundTypeID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithTypeParam("POST", "/admin/engagements/types/"+notFoundTypeID.String()+"/offers", tid, adminRoles, notFoundTypeID.String(), "", CreateOfferRequest{
		Name:      "Basic Plan",
		CostCents: 1000,
	})

	h.CreateOffer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateOffer_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	offerID := uuid.New()

	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	offer := EngagementOffer{
		ID:               offerID,
		TenantID:         tid,
		EngagementTypeID: typeID,
		Name:             "Basic Plan",
		CostCents:        1000,
	}
	repo.offers[offerID] = offer

	w := httptest.NewRecorder()
	req := makeAdminRequestWithTypeParam("PUT", "/admin/engagements/types/"+typeID.String()+"/offers/"+offerID.String(), tid, adminRoles, typeID.String(), offerID.String(), UpdateOfferRequest{
		Name:        "Premium Plan",
		Description: "Полный доступ",
		CostCents:   2000,
		SortOrder:   2,
	})

	h.UpdateOffer(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementOffer
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Name != "Premium Plan" {
		t.Errorf("expected name 'Premium Plan', got %s", resp.Name)
	}
}

func TestAdminHandler_UpdateOffer_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	notFoundOfferID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithTypeParam("PUT", "/admin/engagements/types/"+typeID.String()+"/offers/"+notFoundOfferID.String(), tid, adminRoles, typeID.String(), notFoundOfferID.String(), UpdateOfferRequest{
		Name:      "Updated",
		CostCents: 1000,
	})

	h.UpdateOffer(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAdminHandler_DeleteOffer_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	offerID := uuid.New()

	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	offer := EngagementOffer{
		ID:               offerID,
		TenantID:         tid,
		EngagementTypeID: typeID,
		Name:             "Basic Plan",
	}
	repo.offers[offerID] = offer

	w := httptest.NewRecorder()
	req := makeAdminRequestWithTypeParam("DELETE", "/admin/engagements/types/"+typeID.String()+"/offers/"+offerID.String(), tid, adminRoles, typeID.String(), offerID.String(), nil)

	h.DeleteOffer(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestAdminHandler_DeleteOffer_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	notFoundOfferID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithTypeParam("DELETE", "/admin/engagements/types/"+typeID.String()+"/offers/"+notFoundOfferID.String(), tid, adminRoles, typeID.String(), notFoundOfferID.String(), nil)

	h.DeleteOffer(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAdminHandler_DeleteOffer_InvalidID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()

	w := httptest.NewRecorder()
	req := makeAdminRequestWithTypeParam("DELETE", "/admin/engagements/types/"+typeID.String()+"/offers/not-a-uuid", tid, adminRoles, typeID.String(), "not-a-uuid", nil)

	h.DeleteOffer(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- NoTenant tests ---

func TestAdminHandler_CreateType_NoTenant(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/engagements/types", bytes.NewReader([]byte("{}")))
	ctx := context.WithValue(req.Context(), auth.RolesKey, adminRoles)
	req = req.WithContext(ctx)
	// No tenant context

	h.CreateType(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_UpdateType_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	// Type doesn't exist in repo

	w := httptest.NewRecorder()
	req := makeAdminRequestWithParam("PUT", "/admin/engagements/types/"+typeID.String(), tid, adminRoles, "id", typeID.String(), UpdateTypeRequest{
		Name: "Updated",
	})

	h.UpdateType(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (type not found), got %d", w.Code)
	}
}

// --- Invalid request body tests ---

func TestAdminHandler_CreateCategory_InvalidBody(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/engagements/categories", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := tenant.TenantContext(req.Context(), tid)
	ctx = context.WithValue(ctx, auth.RolesKey, adminRoles)
	req = req.WithContext(ctx)

	h.CreateCategory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminHandler_CreateType_InvalidBody(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/engagements/types", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := tenant.TenantContext(req.Context(), tid)
	ctx = context.WithValue(ctx, auth.RolesKey, adminRoles)
	req = req.WithContext(ctx)

	h.CreateType(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Admin with RoleAdmin tests ---

func TestAdminHandler_AdminRoleAllowed(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	tid := uuid.New()
	adminRoles := []string{user.RoleAdmin}

	w := httptest.NewRecorder()
	req := makeAdminRequest("POST", "/admin/engagements/categories", tid, adminRoles, CreateCategoryRequest{
		Slug: "health",
		Name: "Здоровье",
	})

	h.CreateCategory(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 with admin role, got %d", w.Code)
	}
}
