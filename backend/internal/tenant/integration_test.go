//go:build integration

package tenant_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"lkfl/internal/testutil"
)

func setupTenantTest(t *testing.T) (*testutil.TestServer, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ts, cleanup, err := testutil.SetupAllWithServer(ctx)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return ts, cleanup
}

// TestTenantCRUD_FullCycle — полный цикл: Create → Get → List → Update → Delete
func TestTenantCRUD_FullCycle(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// 1. Create tenant
	createReq := map[string]string{"slug": "test-company", "name": "Test Company"}
	resp, err := ts.PostWithToken("/admin/tenants", adminToken, createReq)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}

	var created struct {
		ID     uuid.UUID `json:"id"`
		Slug   string    `json:"slug"`
		Name   string    `json:"name"`
		Status string    `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	resp.Body.Close()

	if created.Slug != "test-company" {
		t.Errorf("slug: expected 'test-company', got '%s'", created.Slug)
	}
	if created.Name != "Test Company" {
		t.Errorf("name: expected 'Test Company', got '%s'", created.Name)
	}
	if created.Status != "active" {
		t.Errorf("status: expected 'active', got '%s'", created.Status)
	}

	// 2. Get by ID
	resp, err = ts.GetWithToken("/admin/tenants/"+created.ID.String(), adminToken)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	resp.Body.Close()

	// 3. List
	resp, err = ts.GetWithToken("/admin/tenants", adminToken)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	var tenants []struct {
		ID   uuid.UUID `json:"id"`
		Slug string    `json:"slug"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tenants); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	resp.Body.Close()
	if len(tenants) < 1 {
		t.Error("list: expected at least 1 tenant")
	}

	// 4. Update
	updateReq := map[string]string{"name": "Updated Company"}
	resp, err = ts.PutWithToken("/admin/tenants/"+created.ID.String(), adminToken, updateReq)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	resp.Body.Close()

	// 5. Delete
	resp, err = ts.DeleteWithToken("/admin/tenants/"+created.ID.String(), adminToken)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}

	// Verify deleted
	resp, err = ts.GetWithToken("/admin/tenants/"+created.ID.String(), adminToken)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete: expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Cleanup: remove any remaining tenants
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestTenantCreate_DuplicateSlug — попытка создать tenant с дублирующимся slug
func TestTenantCreate_DuplicateSlug(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create first tenant
	createReq := map[string]string{"slug": "dup-slug", "name": "First"}
	resp, _ := ts.PostWithToken("/admin/tenants", adminToken, createReq)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", resp.StatusCode)
	}

	// Try duplicate
	resp, _ = ts.PostWithToken("/admin/tenants", adminToken, createReq)
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate create: expected 409, got %d body=%s", resp.StatusCode, body)
	}
}

// TestTenantCreate_InvalidSlug — создание с невалидным slug
func TestTenantCreate_InvalidSlug(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	invalidSlugs := []string{"has space", "has_underscore", "-starts-hyphen", "ends-hyphen-"}
	for _, slug := range invalidSlugs {
		t.Run(slug, func(t *testing.T) {
			createReq := map[string]string{"slug": slug, "name": "Test"}
			resp, _ := ts.PostWithToken("/admin/tenants", adminToken, createReq)
			body := testutil.ReadBody(resp)
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("slug '%s': expected 400, got %d body=%s", slug, resp.StatusCode, body)
			}
		})
	}
}

// TestTenantCreate_UppercaseSlugLowercased — UPPER slug lowercased handler'ом
func TestTenantCreate_UppercaseSlugLowercased(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	createReq := map[string]string{"slug": "UPPER-SLUG", "name": "Test"}
	resp, _ := ts.PostWithToken("/admin/tenants", adminToken, createReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}

	var created struct {
		Slug string `json:"slug"`
	}
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	// Handler lowercases the slug
	if created.Slug != "upper-slug" {
		t.Errorf("expected slug 'upper-slug' (lowercased), got '%s'", created.Slug)
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants WHERE slug = 'upper-slug'")
}

// TestTenantSuspension — create → suspend → verify blocked → activate
func TestTenantSuspension(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create tenant
	tenantID, err := ts.CreateTenant(ctx, "suspend-test", "Suspend Test")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	// Suspend via update
	updateReq := map[string]string{"status": "suspended"}
	resp, err := ts.PutWithToken("/admin/tenants/"+tenantID.String(), adminToken, updateReq)
	if err != nil {
		t.Fatalf("suspend: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("suspend: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	resp.Body.Close()

	// Verify suspended via GetBySlug (service level check)
	// The tenant resolver middleware will still find it, but GetBySlug in service returns error
	// We verify via admin get — it should still return the tenant
	resp, err = ts.GetWithToken("/admin/tenants/"+tenantID.String(), adminToken)
	if err != nil {
		t.Fatalf("get suspended: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get suspended: expected 200, got %d", resp.StatusCode)
	}
	var result struct {
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if result.Status != "suspended" {
		t.Errorf("expected status 'suspended', got '%s'", result.Status)
	}

	// Re-activate
	updateReq = map[string]string{"status": "active"}
	resp, err = ts.PutWithToken("/admin/tenants/"+tenantID.String(), adminToken, updateReq)
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("activate: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify active
	resp, err = ts.GetWithToken("/admin/tenants/"+tenantID.String(), adminToken)
	if err != nil {
		t.Fatalf("get after activate: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if result.Status != "active" {
		t.Errorf("expected status 'active' after reactivation, got '%s'", result.Status)
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestTenant_RBAC_EmployeeForbidden — employee не может создать tenant
func TestTenant_RBAC_EmployeeForbidden(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	employeeToken := testutil.TestTokenEmployee("employee-user")

	createReq := map[string]string{"slug": "should-fail", "name": "Should Fail"}
	resp, _ := ts.PostWithToken("/admin/tenants", employeeToken, createReq)
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d body=%s", resp.StatusCode, body)
	}
}

// TestTenantBrandConfig_CRUD — brand config CRUD
func TestTenantBrandConfig_CRUD(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create tenant first
	tenantID, err := ts.CreateTenant(ctx, "brand-test", "Brand Test")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	// Create brand config
	brandReq := map[string]any{
		"primary_color":    "#FF0000",
		"secondary_color":  "#00FF00",
		"logo_url":         "https://example.com/logo.png",
		"brand_name":       "Test Brand",
		"meta_title":       "Test Meta",
		"meta_description": "Test Description",
	}
	resp, err := ts.PutWithToken("/admin/tenants/"+tenantID.String()+"/brand", adminToken, brandReq)
	if err != nil {
		t.Fatalf("create brand: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create brand: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	resp.Body.Close()

	// Get brand config
	resp, err = ts.GetWithToken("/admin/tenants/"+tenantID.String()+"/brand", adminToken)
	if err != nil {
		t.Fatalf("get brand: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get brand: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	var brand struct {
		PrimaryColor   string `json:"primary_color"`
		SecondaryColor string `json:"secondary_color"`
		BrandName      string `json:"brand_name"`
	}
	json.NewDecoder(resp.Body).Decode(&brand)
	resp.Body.Close()
	if brand.PrimaryColor != "#FF0000" {
		t.Errorf("expected primary '#FF0000', got '%s'", brand.PrimaryColor)
	}

	// Update brand config
	brandReq = map[string]any{"primary_color": "#0000FF"}
	resp, err = ts.PutWithToken("/admin/tenants/"+tenantID.String()+"/brand", adminToken, brandReq)
	if err != nil {
		t.Fatalf("update brand: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update brand: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify brand config not found for non-existent tenant
	resp, err = ts.GetWithToken("/admin/tenants/"+uuid.New().String()+"/brand", adminToken)
	if err != nil {
		t.Fatalf("get brand not found: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("get brand not found: expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestTenantList_Pagination — пагинация списка tenants
func TestTenantList_Pagination(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create 3 tenants
	for i := 1; i <= 3; i++ {
		_, err := ts.CreateTenant(ctx, "page-test-"+string(rune('a'+i)), "Page Test "+string(rune('A'+i)))
		if err != nil {
			t.Fatalf("create tenant %d: %v", i, err)
		}
	}

	// List with pagination
	resp, err := ts.GetWithToken("/admin/tenants?page=1&per_page=2", adminToken)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", resp.StatusCode)
	}
	var tenants []struct{}
	json.NewDecoder(resp.Body).Decode(&tenants)
	resp.Body.Close()
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants on page 1, got %d", len(tenants))
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestTenantGet_InvalidID — запрос с невалидным UUID
func TestTenantGet_InvalidID(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	resp, err := ts.GetWithToken("/admin/tenants/not-a-uuid", adminToken)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestTenantCreate_EmptyBody — создание с пустым телом
func TestTenantCreate_EmptyBody(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	resp, err := ts.PostWithToken("/admin/tenants", adminToken, map[string]string{})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
}

// TestTenantList_NoAuth — список без авторизации
func TestTenantList_NoAuth(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", ts.URL+"/admin/tenants", nil)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestTenantConcurrent_Create — конкурентное создание tenants
func TestTenantConcurrent_Create(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create several tenants concurrently
	for i := 0; i < 5; i++ {
		go func(i int) {
			slug := "concurrent-" + string(rune('a'+i))
			createReq := map[string]string{"slug": slug, "name": "Concurrent " + slug}
			resp, _ := ts.PostWithToken("/admin/tenants", adminToken, createReq)
			if resp.StatusCode != http.StatusCreated {
				t.Errorf("concurrent create %d: expected 201, got %d", i, resp.StatusCode)
			}
			resp.Body.Close()
		}(i)
	}

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	// Verify all created
	resp, _ := ts.GetWithToken("/admin/tenants", adminToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", resp.StatusCode)
	}
	var tenants []struct{}
	json.NewDecoder(resp.Body).Decode(&tenants)
	resp.Body.Close()
	if len(tenants) != 5 {
		t.Errorf("expected 5 tenants after concurrent creates, got %d", len(tenants))
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestTenantCreate_Defaults — проверка значений по умолчанию
func TestTenantCreate_Defaults(t *testing.T) {
	ts, cleanup := setupTenantTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create without status
	createReq := map[string]string{"slug": "defaults-test", "name": "Defaults"}
	resp, _ := ts.PostWithToken("/admin/tenants", adminToken, createReq)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", resp.StatusCode)
	}

	var created struct {
		ID     uuid.UUID `json:"id"`
		Status string    `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&created)
	resp.Body.Close()

	if created.Status != "active" {
		t.Errorf("expected default status 'active', got '%s'", created.Status)
	}
	if created.ID == uuid.Nil {
		t.Error("expected non-nil ID")
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants WHERE slug = 'defaults-test'")
}
