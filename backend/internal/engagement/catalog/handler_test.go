package catalog

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"lkfl/internal/tenant"
)

// strPtr — вспомогательная функция для тестов: создаёт *string из string.
func strPtr(s string) *string { return &s }

// makeTestRequest создаёт тестовый запрос с tenant context.
func makeTestRequest(method, path string, tid uuid.UUID, queryParams url.Values) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	if queryParams != nil {
		req.URL.RawQuery = queryParams.Encode()
	}
	ctx := tenant.TenantContext(req.Context(), tid)
	return req.WithContext(ctx)
}

// makeTestRequestWithParam создаёт запрос с tenant context + chi URL param.
func makeTestRequestWithParam(method, path string, tid uuid.UUID, paramKey, paramValue string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	ctx := tenant.TenantContext(req.Context(), tid)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(paramKey, paramValue)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

	return req.WithContext(ctx)
}

// --- List tests ---

func TestHandler_List_EmptyCatalog(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeTestRequest("GET", "/api/v1/engagements", tid, nil)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Data))
	}
	if resp.Pagination.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.TotalPages != 1 {
		t.Errorf("expected total_pages 1 (min), got %d", resp.Pagination.TotalPages)
	}
}

func TestHandler_List_WithItems(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	et := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "gym",
		Name:     "Абонемент в спортзал",
	}
	repo.types[et.ID] = et

	w := httptest.NewRecorder()
	req := makeTestRequest("GET", "/api/v1/engagements", tid, nil)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Data))
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Pagination.Total)
	}
	if resp.Pagination.Page != 1 {
		t.Errorf("expected page 1, got %d", resp.Pagination.Page)
	}
}

func TestHandler_List_FilterByType(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	benefit := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "benefit-1",
		Name:     "Benefit 1",
	}
	activity := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeActivity,
		Status:   StatusActive,
		Slug:     "activity-1",
		Name:     "Activity 1",
	}
	repo.types[benefit.ID] = benefit
	repo.types[activity.ID] = activity

	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("type", TypeBenefit)
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("expected total 1 (only benefit), got %d", resp.Pagination.Total)
	}
}

func TestHandler_List_InvalidType(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("type", "invalid")
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var errResp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if _, ok := errResp["error"]; !ok {
		t.Error("expected error field in response")
	}
}

func TestHandler_List_NoTenant(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/engagements", nil)
	// No tenant context

	h.List(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Pagination(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("page", "2")
	q.Set("per_page", "10")
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Pagination.Page != 2 {
		t.Errorf("expected page 2, got %d", resp.Pagination.Page)
	}
	if resp.Pagination.PerPage != 10 {
		t.Errorf("expected per_page 10, got %d", resp.Pagination.PerPage)
	}
}

func TestHandler_List_PerPageCapped(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("per_page", "200")
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Pagination.PerPage != 100 {
		t.Errorf("expected per_page capped to 100, got %d", resp.Pagination.PerPage)
	}
}

func TestHandler_List_CategoryFilter(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	cat := EngagementCategory{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "fitness",
		Name:     "Фитнес",
	}
	et := EngagementType{
		ID:         uuid.New(),
		TenantID:   tid,
		CategoryID: cat.ID,
		Type:       TypeBenefit,
		Status:     StatusActive,
		Slug:       "gym",
		Name:       "Спортзал",
		Category:   &cat,
	}
	repo.types[et.ID] = et

	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("category", "fitness")
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Pagination.Total)
	}
}

func TestHandler_List_SearchFilter(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	et1 := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "yoga",
		Name:     "Йога-студия",
	}
	et2 := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "gym",
		Name:     "Спортзал",
	}
	repo.types[et1.ID] = et1
	repo.types[et2.ID] = et2

	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("search", "йога")
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// Mock repo doesn't filter by search, but handler passes it through
	// The search filtering happens in the real repository
}

func TestHandler_List_StatusFilter(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	active := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "active-1",
		Name:     "Active",
	}
	promo := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusPromo,
		Slug:     "promo-1",
		Name:     "Promo",
	}
	repo.types[active.ID] = active
	repo.types[promo.ID] = promo

	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("status", StatusPromo)
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Pagination.Total != 1 {
		t.Errorf("expected total 1 (only promo), got %d", resp.Pagination.Total)
	}
}

// --- Get tests ---

func TestHandler_Get_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "gym",
		Name:     "Абонемент в спортзал",
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	req := makeTestRequestWithParam("GET", "/api/v1/engagements/"+typeID.String(), tid, "id", typeID.String())

	h.Get(w, req)

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
	if resp.Name != "Абонемент в спортзал" {
		t.Errorf("expected name 'Абонемент в спортзал', got %s", resp.Name)
	}
	if resp.Badge != "Доступна" {
		t.Errorf("expected badge 'Доступна', got %s", resp.Badge)
	}
}

func TestHandler_Get_WithCategory(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	cat := &EngagementCategory{
		ID:        uuid.New(),
		TenantID:  tid,
		Slug:      "fitness",
		Name:      "Фитнес",
		Icon:      strPtr("dumbbell"),
		SortOrder: 1,
	}
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "gym",
		Name:     "Спортзал",
		Category: cat,
	}
	repo.types[typeID] = et

	w := httptest.NewRecorder()
	req := makeTestRequestWithParam("GET", "/api/v1/engagements/"+typeID.String(), tid, "id", typeID.String())

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementTypeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Category == nil {
		t.Fatal("expected category to be present")
	}
	if resp.Category.Slug != "fitness" {
		t.Errorf("expected category slug 'fitness', got %s", resp.Category.Slug)
	}
}

func TestHandler_Get_WithOffers(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "gym",
		Name:     "Спортзал",
	}
	repo.types[typeID] = et

	offer := EngagementOffer{
		ID:               uuid.New(),
		EngagementTypeID: typeID,
		Name:             "Basic",
		CostCents:        1000,
		SortOrder:        1,
	}
	repo.offers[offer.ID] = offer

	w := httptest.NewRecorder()
	req := makeTestRequestWithParam("GET", "/api/v1/engagements/"+typeID.String(), tid, "id", typeID.String())

	h.Get(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp EngagementTypeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Offers) != 1 {
		t.Errorf("expected 1 offer, got %d", len(resp.Offers))
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	notFoundID := uuid.New()

	w := httptest.NewRecorder()
	req := makeTestRequestWithParam("GET", "/api/v1/engagements/"+notFoundID.String(), tid, "id", notFoundID.String())

	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeTestRequestWithParam("GET", "/api/v1/engagements/not-a-uuid", tid, "id", "not-a-uuid")

	h.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_TenantMismatch(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	// Type belongs to a different tenant
	otherTenant := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: otherTenant,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "gym",
		Name:     "Абонемент в спортзал",
	}
	repo.types[typeID] = et

	// Request from a different tenant
	w := httptest.NewRecorder()
	req := makeTestRequestWithParam("GET", "/api/v1/engagements/"+typeID.String(), uuid.New(), "id", typeID.String())

	h.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (tenant mismatch), got %d", w.Code)
	}
}

func TestHandler_Get_NoTenant(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	typeID := uuid.New()
	w := httptest.NewRecorder()
	// No tenant context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", typeID.String())
	req := httptest.NewRequest("GET", "/api/v1/engagements/"+typeID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.Get(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (no tenant), got %d", w.Code)
	}
}

// --- Categories tests ---

func TestHandler_Categories_Empty(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	w := httptest.NewRecorder()
	req := makeTestRequest("GET", "/api/v1/engagements/categories", tid, nil)

	h.Categories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp []EngagementCategoryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected 0 categories, got %d", len(resp))
	}
	if resp == nil {
		t.Error("expected non-nil empty array")
	}
}

func TestHandler_Categories_WithItems(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	cat1 := EngagementCategory{
		ID:        uuid.New(),
		TenantID:  tid,
		Slug:      "fitness",
		Name:      "Фитнес",
		Icon:      strPtr("dumbbell"),
		SortOrder: 1,
	}
	cat2 := EngagementCategory{
		ID:        uuid.New(),
		TenantID:  tid,
		Slug:      "food",
		Name:      "Питание",
		Icon:      strPtr("utensils"),
		SortOrder: 2,
	}
	repo.categories[cat1.ID] = cat1
	repo.categories[cat2.ID] = cat2

	w := httptest.NewRecorder()
	req := makeTestRequest("GET", "/api/v1/engagements/categories", tid, nil)

	h.Categories(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp []EngagementCategoryResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 categories, got %d", len(resp))
	}
}

func TestHandler_Categories_NoTenant(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/engagements/categories", nil)
	// No tenant context

	h.Categories(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// --- Badge tests ---

func TestComputeBadge_Promo(t *testing.T) {
	et := EngagementType{Status: StatusPromo}
	badge := computeBadge(et)
	if badge != "Промо" {
		t.Errorf("expected 'Промо', got %s", badge)
	}
}

func TestComputeBadge_Active(t *testing.T) {
	et := EngagementType{Status: StatusActive}
	badge := computeBadge(et)
	if badge != "Доступна" {
		t.Errorf("expected 'Доступна', got %s", badge)
	}
}

func TestComputeBadge_Draft(t *testing.T) {
	et := EngagementType{Status: StatusDraft}
	badge := computeBadge(et)
	if badge != "Доступна" {
		t.Errorf("expected 'Доступна' for draft (default), got %s", badge)
	}
}

func TestComputeBadge_Hidden(t *testing.T) {
	et := EngagementType{Status: StatusHidden}
	badge := computeBadge(et)
	if badge != "Доступна" {
		t.Errorf("expected 'Доступна' for hidden (default), got %s", badge)
	}
}

// --- ToResponse tests ---

func TestToResponse_WithCategory(t *testing.T) {
	cat := &EngagementCategory{
		ID:        uuid.New(),
		Slug:      "fitness",
		Name:      "Фитнес",
		Icon:      strPtr("dumbbell"),
		SortOrder: 1,
	}
	cost := int64(1500)
	et := EngagementType{
		ID:           uuid.New(),
		Slug:         "gym",
		Name:         "Спортзал",
		Description:  strPtr("Абонемент"),
		Type:         TypeBenefit,
		Status:       StatusActive,
		CostCents:    &cost,
		ProviderName: strPtr("FitLife"),
		Category:     cat,
	}

	resp := et.ToResponse()
	if resp.Category == nil {
		t.Fatal("expected category to be set")
	}
	if resp.Category.Slug != "fitness" {
		t.Errorf("expected category slug 'fitness', got %s", resp.Category.Slug)
	}
	if resp.Badge != "Доступна" {
		t.Errorf("expected badge 'Доступна', got %s", resp.Badge)
	}
	if resp.CostCents == nil || *resp.CostCents != 1500 {
		t.Errorf("expected cost_cents 1500, got %v", resp.CostCents)
	}
}

func TestToResponse_WithOffers(t *testing.T) {
	et := EngagementType{
		ID:     uuid.New(),
		Slug:   "gym",
		Name:   "Спортзал",
		Type:   TypeBenefit,
		Status: StatusActive,
		Offers: []EngagementOffer{
			{ID: uuid.New(), Name: "Basic", CostCents: 1000, SortOrder: 1},
			{ID: uuid.New(), Name: "Premium", Description: strPtr("Full access"), CostCents: 2000, SortOrder: 2},
		},
	}

	resp := et.ToResponse()
	if len(resp.Offers) != 2 {
		t.Errorf("expected 2 offers, got %d", len(resp.Offers))
	}
	if resp.Offers[0].Name != "Basic" {
		t.Errorf("expected first offer 'Basic', got %s", resp.Offers[0].Name)
	}
	if resp.Offers[1].Description != "Full access" {
		t.Errorf("expected second offer description 'Full access', got %s", resp.Offers[1].Description)
	}
}

func TestToResponse_WithoutCategory(t *testing.T) {
	et := EngagementType{
		ID:     uuid.New(),
		Slug:   "gym",
		Name:   "Спортзал",
		Type:   TypeBenefit,
		Status: StatusActive,
	}

	resp := et.ToResponse()
	if resp.Category != nil {
		t.Error("expected nil category")
	}
	if len(resp.Offers) != 0 {
		t.Error("expected empty offers")
	}
}

func TestToResponse_PromoBadge(t *testing.T) {
	et := EngagementType{
		ID:     uuid.New(),
		Slug:   "promo-gym",
		Name:   "Промо Спортзал",
		Type:   TypeBenefit,
		Status: StatusPromo,
	}

	resp := et.ToResponse()
	if resp.Badge != "Промо" {
		t.Errorf("expected badge 'Промо', got %s", resp.Badge)
	}
}

// --- Integration-style tests ---

func TestHandler_List_TotalPagesCalculation(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	// Add 5 items
	for i := 0; i < 5; i++ {
		et := EngagementType{
			ID:       uuid.New(),
			TenantID: tid,
			Type:     TypeBenefit,
			Status:   StatusActive,
			Slug:     "item-" + string(rune('a'+i)),
			Name:     "Item " + string(rune('A'+i)),
		}
		repo.types[et.ID] = et
	}

	// per_page=2 → total_pages = ceil(5/2) = 3
	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("per_page", "2")
	req := makeTestRequest("GET", "/api/v1/engagements", tid, q)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Pagination.TotalPages != 3 {
		t.Errorf("expected total_pages 3 (ceil(5/2)), got %d", resp.Pagination.TotalPages)
	}
	if resp.Pagination.Total != 5 {
		t.Errorf("expected total 5, got %d", resp.Pagination.Total)
	}
}

func TestHandler_List_ResponseFields(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc, nil)

	tid := uuid.New()
	cat := EngagementCategory{
		ID:        uuid.New(),
		TenantID:  tid,
		Slug:      "fitness",
		Name:      "Фитнес",
		Icon:      strPtr("dumbbell"),
		SortOrder: 1,
	}
	cost := int64(1500)
	imgURL := "https://example.com/gym.jpg"
	et := EngagementType{
		ID:           uuid.New(),
		TenantID:     tid,
		CategoryID:   cat.ID,
		Slug:         "gym-premium",
		Name:         "Премиум Спортзал",
		Description:  strPtr("Полный доступ"),
		Type:         TypeBenefit,
		Status:       StatusPromo,
		CostCents:    &cost,
		ProviderName: strPtr("FitLife"),
		ImageURL:     &imgURL,
		Category:     &cat,
	}
	repo.types[et.ID] = et

	w := httptest.NewRecorder()
	req := makeTestRequest("GET", "/api/v1/engagements", tid, nil)

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Data))
	}

	item := resp.Data[0]
	if item.ID != et.ID {
		t.Errorf("ID mismatch")
	}
	if item.Slug != "gym-premium" {
		t.Errorf("expected slug 'gym-premium', got %s", item.Slug)
	}
	if item.Name != "Премиум Спортзал" {
		t.Errorf("expected name 'Премиум Спортзал', got %s", item.Name)
	}
	if item.Description != "Полный доступ" {
		t.Errorf("expected description 'Полный доступ', got %s", item.Description)
	}
	if item.Type != TypeBenefit {
		t.Errorf("expected type 'benefit', got %s", item.Type)
	}
	if item.Status != StatusPromo {
		t.Errorf("expected status 'promo', got %s", item.Status)
	}
	if item.CostCents == nil || *item.CostCents != 1500 {
		t.Errorf("expected cost_cents 1500, got %v", item.CostCents)
	}
	if item.ProviderName != "FitLife" {
		t.Errorf("expected provider 'FitLife', got %s", item.ProviderName)
	}
	if item.ImageURL == nil || *item.ImageURL != imgURL {
		t.Errorf("expected image_url %s, got %v", imgURL, item.ImageURL)
	}
	if item.Badge != "Промо" {
		t.Errorf("expected badge 'Промо', got %s", item.Badge)
	}
	if item.Category == nil || item.Category.Slug != "fitness" {
		t.Error("expected category fitness")
	}
}
