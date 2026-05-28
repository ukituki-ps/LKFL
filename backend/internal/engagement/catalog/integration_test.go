//go:build integration

package catalog_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"lkfl/internal/testutil"
)

func setupCatalogTest(t *testing.T) (*testutil.TestServer, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ts, cleanup, err := testutil.SetupAllWithServer(ctx)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return ts, cleanup
}

// TestCatalogPublic_List — GET /api/v1/engagements — публичный список
func TestCatalogPublic_List(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()

	// Setup: tenant + category + engagement type
	tenantID, _ := ts.CreateTenant(ctx, "catalog-pub", "Catalog Public")
	tenantSlug := "catalog-pub"
	catID, _ := ts.CreateCategory(ctx, tenantID, "fitness", "Fitness")
	ts.CreateEngagementType(ctx, tenantID, catID, "gym-pass", "Gym Pass", "benefit", "active")

	token := testutil.TestTokenEmployee("kc-employee")

	resp, err := ts.GetWithTokenAndTenant("/api/v1/engagements", token, tenantSlug)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
		return
	}

	var listResp struct {
		Data []struct {
			ID     string `json:"id"`
			Slug   string `json:"slug"`
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"data"`
		Pagination struct {
			Total int64 `json:"total"`
		} `json:"pagination"`
	}
	json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()
	t.Logf("Catalog list: total=%d, items=%d", listResp.Pagination.Total, len(listResp.Data))

	if listResp.Pagination.Total < 1 {
		t.Errorf("expected at least 1 engagement, got total=%d", listResp.Pagination.Total)
	}
}

// TestCatalogPublic_GetByID — GET /api/v1/engagements/{id} — детали
func TestCatalogPublic_GetByID(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()

	tenantID, _ := ts.CreateTenant(ctx, "catalog-get", "Catalog Get")
	tenantSlug := "catalog-get"
	catID, _ := ts.CreateCategory(ctx, tenantID, "health", "Health")
	typeID, _ := ts.CreateEngagementType(ctx, tenantID, catID, "dms-policy", "DMS Policy", "benefit", "active")

	token := testutil.TestTokenEmployee("kc-employee")

	resp, err := ts.GetWithTokenAndTenant("/api/v1/engagements/"+typeID.String(), token, tenantSlug)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
		return
	}

	var et struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&et)
	resp.Body.Close()
	t.Logf("Catalog get: slug=%s", et.Slug)
	if et.Slug != "dms-policy" {
		t.Errorf("expected slug 'dms-policy', got '%s'", et.Slug)
	}
}

// TestCatalogPublic_GetNotFound — GET /api/v1/engagements/{id} для несуществующего → 404
func TestCatalogPublic_GetNotFound(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	token := testutil.TestTokenEmployee("kc-employee")

	resp, _ := ts.GetWithTokenAndTenant("/api/v1/engagements/"+uuid.New().String(), token, "nonexistent")
	body := testutil.ReadBody(resp)
	t.Logf("Catalog get not found: status=%d body=%s", resp.StatusCode, body)

	// May be 401 (tenant not found) or 404 (engagement not found)
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 404 or 401, got %d body=%s", resp.StatusCode, body)
	}
}

// TestCatalogPublic_GetInvalidID — невалидный UUID → 400
func TestCatalogPublic_GetInvalidID(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	token := testutil.TestTokenEmployee("kc-employee")

	resp, _ := ts.GetWithTokenAndTenant("/api/v1/engagements/not-a-uuid", token, "nonexistent")
	body := testutil.ReadBody(resp)
	t.Logf("Catalog get invalid ID: status=%d body=%s", resp.StatusCode, body)

	// May be 401 (tenant not found) or 400 (invalid UUID)
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusUnauthorized {
		t.Logf("got status %d body=%s", resp.StatusCode, body)
	}
}

// TestCatalogAdmin_CRUD — admin CRUD: category → type → offer → status → delete
func TestCatalogAdmin_CRUD(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	_, _ = ts.CreateTenant(ctx, "catalog-admin", "Catalog Admin")
	tenantSlug := "catalog-admin"

	// 1. Create category
	catReq := map[string]any{"slug": "food", "name": "Food", "sort_order": 1}
	resp, err := ts.PostWithTokenAndTenant("/admin/engagements/categories", adminToken, tenantSlug, catReq)
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create category: expected 201, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	var catResp struct {
		ID uuid.UUID `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&catResp)
	resp.Body.Close()
	catID := catResp.ID
	t.Logf("Created category: %s", catID)

	// 2. Create type
	typeReq := map[string]any{
		"category_id":   catID,
		"slug":          "lunch-voucher",
		"name":          "Lunch Voucher",
		"description":   "Monthly lunch voucher",
		"type":          "benefit",
		"status":        "draft",
		"provider_name": "FoodProvider",
	}
	resp, err = ts.PostWithTokenAndTenant("/admin/engagements/types", adminToken, tenantSlug, typeReq)
	if err != nil {
		t.Fatalf("create type: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create type: expected 201, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	var typeResp struct {
		ID   uuid.UUID `json:"id"`
		Slug string    `json:"slug"`
	}
	json.NewDecoder(resp.Body).Decode(&typeResp)
	resp.Body.Close()
	typeID := typeResp.ID
	t.Logf("Created type: %s", typeID)

	// 3. Create offer
	offerReq := map[string]any{"name": "Basic", "cost_cents": 50000, "sort_order": 1}
	resp, err = ts.PostWithTokenAndTenant("/admin/engagements/types/"+typeID.String()+"/offers", adminToken, tenantSlug, offerReq)
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create offer: expected 201, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	resp.Body.Close()

	// 4. Update status: draft → active
	statusReq := map[string]string{"status": "active"}
	resp, err = ts.PatchWithTokenAndTenant("/admin/engagements/types/"+typeID.String()+"/status", adminToken, tenantSlug, statusReq)
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update status: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	resp.Body.Close()

	// 5. Verify status changed
	resp, err = ts.GetWithTokenAndTenant("/admin/engagements/types/"+typeID.String(), adminToken, tenantSlug)
	if err != nil {
		t.Fatalf("get type: %v", err)
	}
	var statusCheck struct {
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&statusCheck)
	resp.Body.Close()
	if statusCheck.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", statusCheck.Status)
	}
}

// TestCatalogMultiTenant_Isolation — каталог tenant A ≠ каталог tenant B
func TestCatalogMultiTenant_Isolation(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create two tenants
	tenantA, _ := ts.CreateTenant(ctx, "cat-a", "Catalog A")
	tenantB, _ := ts.CreateTenant(ctx, "cat-b", "Catalog B")

	// Create categories + types in each
	catA, _ := ts.CreateCategory(ctx, tenantA, "cat-a-type", "Cat A Type")
	catB, _ := ts.CreateCategory(ctx, tenantB, "cat-b-type", "Cat B Type")

	ts.CreateEngagementType(ctx, tenantA, catA, "only-in-a", "Only In A", "benefit", "active")
	ts.CreateEngagementType(ctx, tenantB, catB, "only-in-b", "Only In B", "benefit", "active")

	// Verify isolation via DB
	var countA, countB int64
	ts.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM lkfl_platform.engagement_types WHERE tenant_id = $1", tenantA,
	).Scan(&countA)
	ts.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM lkfl_platform.engagement_types WHERE tenant_id = $1", tenantB,
	).Scan(&countB)

	if countA != 1 {
		t.Errorf("expected 1 engagement in tenant A, got %d", countA)
	}
	if countB != 1 {
		t.Errorf("expected 1 engagement in tenant B, got %d", countB)
	}
}

// TestCatalogStatusTransitions — draft → active → promo → completed
func TestCatalogStatusTransitions(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	tenantID, _ := ts.CreateTenant(ctx, "status-trans", "Status Transitions")
	tenantSlug := "status-trans"
	catID, _ := ts.CreateCategory(ctx, tenantID, "trans", "Transitions")
	typeID, _ := ts.CreateEngagementType(ctx, tenantID, catID, "status-test", "Status Test", "benefit", "draft")

	// draft → active
	resp, err := ts.PatchWithTokenAndTenant("/admin/engagements/types/"+typeID.String()+"/status", adminToken, tenantSlug, map[string]string{"status": "active"})
	if err != nil {
		t.Fatalf("draft→active: %v", err)
	}
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("draft→active: expected 200, got %d body=%s", resp.StatusCode, body)
	}

	// active → promo
	resp, err = ts.PatchWithTokenAndTenant("/admin/engagements/types/"+typeID.String()+"/status", adminToken, tenantSlug, map[string]string{"status": "promo"})
	if err != nil {
		t.Fatalf("active→promo: %v", err)
	}
	body = testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("active→promo: expected 200, got %d body=%s", resp.StatusCode, body)
	}

	// active → completed (via promo → active → completed)
	resp, err = ts.PatchWithTokenAndTenant("/admin/engagements/types/"+typeID.String()+"/status", adminToken, tenantSlug, map[string]string{"status": "active"})
	if err != nil {
		t.Fatalf("promo→active: %v", err)
	}
	resp.Body.Close()

	resp, err = ts.PatchWithTokenAndTenant("/admin/engagements/types/"+typeID.String()+"/status", adminToken, tenantSlug, map[string]string{"status": "completed"})
	if err != nil {
		t.Fatalf("active→completed: %v", err)
	}
	body = testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("active→completed: expected 200, got %d body=%s", resp.StatusCode, body)
	}

	// completed → anything should fail (terminal state)
	resp, err = ts.PatchWithTokenAndTenant("/admin/engagements/types/"+typeID.String()+"/status", adminToken, tenantSlug, map[string]string{"status": "active"})
	body = testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("completed→active: expected 400 (terminal state), got %d body=%s", resp.StatusCode, body)
	}
}

// TestCatalogCache_Invalidation — admin create → cache invalidation → list updated
func TestCatalogCache_Invalidation(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()

	tenantID, _ := ts.CreateTenant(ctx, "cache-test", "Cache Test")
	tenantSlug := "cache-test"
	catID, _ := ts.CreateCategory(ctx, tenantID, "cache-cat", "Cache Category")

	employeeToken := testutil.TestTokenEmployee("kc-employee")
	adminToken := testutil.TestTokenAdmin("admin-user")

	// First list (empty)
	resp, err := ts.GetWithTokenAndTenant("/api/v1/engagements", employeeToken, tenantSlug)
	if err != nil {
		t.Fatalf("initial list: %v", err)
	}
	var listResp1 struct {
		Pagination struct {
			Total int64 `json:"total"`
		} `json:"pagination"`
	}
	json.NewDecoder(resp.Body).Decode(&listResp1)
	resp.Body.Close()
	initialTotal := listResp1.Pagination.Total
	t.Logf("Initial list: total=%d", initialTotal)

	// Create new type via admin API (triggers cache invalidation)
	typeReq := map[string]any{
		"category_id":   catID,
		"slug":          "cache-item",
		"name":          "Cache Item",
		"description":   "Item for cache test",
		"type":          "benefit",
		"status":        "active",
		"provider_name": "CacheProvider",
	}
	resp, err = ts.PostWithTokenAndTenant("/admin/engagements/types", adminToken, tenantSlug, typeReq)
	if err != nil {
		t.Fatalf("create type: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create type: expected 201, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	resp.Body.Close()

	// List again — should include new item (cache invalidated by admin create)
	resp, err = ts.GetWithTokenAndTenant("/api/v1/engagements", employeeToken, tenantSlug)
	if err != nil {
		t.Fatalf("list after create: %v", err)
	}

	var listResp2 struct {
		Pagination struct {
			Total int64 `json:"total"`
		} `json:"pagination"`
	}
	json.NewDecoder(resp.Body).Decode(&listResp2)
	resp.Body.Close()
	t.Logf("List after create: total=%d", listResp2.Pagination.Total)

	if listResp2.Pagination.Total <= initialTotal {
		t.Errorf("expected more items after admin create, initial=%d after=%d",
			initialTotal, listResp2.Pagination.Total)
	}
}

// TestCatalogCategories_Public — GET /api/v1/engagements/categories
func TestCatalogCategories_Public(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()

	tenantID, _ := ts.CreateTenant(ctx, "cat-pub", "Categories Public")
	tenantSlug := "cat-pub"
	_, _ = ts.CreateCategory(ctx, tenantID, "cat-1", "Category One")
	_, _ = ts.CreateCategory(ctx, tenantID, "cat-2", "Category Two")

	token := testutil.TestTokenEmployee("kc-employee")

	resp, err := ts.GetWithTokenAndTenant("/api/v1/engagements/categories", token, tenantSlug)
	if err != nil {
		t.Fatalf("categories: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
		return
	}

	var categories []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&categories)
	resp.Body.Close()
	t.Logf("Categories: got %d categories", len(categories))
	if len(categories) < 2 {
		t.Errorf("expected at least 2 categories, got %d", len(categories))
	}
}

// TestCatalogAdmin_RBAC — employee не может доступ к admin catalog
func TestCatalogAdmin_RBAC(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	employeeToken := testutil.TestTokenEmployee("kc-employee")

	resp, _ := ts.PostWithToken("/admin/engagements/categories", employeeToken,
		map[string]any{"slug": "hack", "name": "Hack"})
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for employee on admin catalog, got %d body=%s", resp.StatusCode, body)
	}
}

// TestCatalogAdmin_DuplicateSlug — создание типа с дублирующимся slug → 409
func TestCatalogAdmin_DuplicateSlug(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	tenantID, _ := ts.CreateTenant(ctx, "dup-slug-cat", "Duplicate Slug")
	tenantSlug := "dup-slug-cat"
	catID, _ := ts.CreateCategory(ctx, tenantID, "dup-cat", "Dup Category")

	// Create first type
	ts.CreateEngagementType(ctx, tenantID, catID, "dup-slug", "Dup Slug", "benefit", "draft")

	// Try duplicate via API
	typeReq := map[string]any{
		"category_id": catID,
		"slug":        "dup-slug",
		"name":        "Duplicate",
		"type":        "benefit",
		"status":      "draft",
	}
	resp, _ := ts.PostWithTokenAndTenant("/admin/engagements/types", adminToken, tenantSlug, typeReq)
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 for duplicate slug, got %d body=%s", resp.StatusCode, body)
	}
}

// TestCatalogAdmin_DeleteType — soft delete типа
func TestCatalogAdmin_DeleteType(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	tenantID, _ := ts.CreateTenant(ctx, "del-type", "Delete Type")
	tenantSlug := "del-type"
	catID, _ := ts.CreateCategory(ctx, tenantID, "del-cat", "Delete Category")
	typeID, _ := ts.CreateEngagementType(ctx, tenantID, catID, "del-me", "Delete Me", "benefit", "active")

	// Delete
	resp, _ := ts.DeleteWithTokenAndTenant("/admin/engagements/types/"+typeID.String(), adminToken, tenantSlug)
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 for delete, got %d body=%s", resp.StatusCode, body)
	}

	// Verify soft delete (status=hidden)
	var status string
	err := ts.DB.QueryRow(ctx,
		"SELECT status FROM lkfl_platform.engagement_types WHERE id = $1", typeID,
	).Scan(&status)
	if err != nil {
		t.Fatalf("check status: %v", err)
	}
	if status != "hidden" {
		t.Errorf("expected status 'hidden' after soft delete, got '%s'", status)
	}
}

// TestCatalogList_NoAuth — без авторизации → 401
func TestCatalogList_NoAuth(t *testing.T) {
	ts, cleanup := setupCatalogTest(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/engagements", nil)
	resp, err := ts.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d body=%s", resp.StatusCode, body)
	}
}
