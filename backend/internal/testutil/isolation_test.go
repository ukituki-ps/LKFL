//go:build integration

package testutil_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"lkfl/internal/testutil"
)

func setupIsolationTest(t *testing.T) (*testutil.TestServer, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ts, cleanup, err := testutil.SetupAllWithServer(ctx)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return ts, cleanup
}

// TestMultiTenantIsolation_DataSeparation — данные tenant A ≠ данные tenant B
func TestMultiTenantIsolation_DataSeparation(t *testing.T) {
	ts, cleanup := setupIsolationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant A with data
	tenantA, _ := ts.CreateTenant(ctx, "iso-a", "Isolation A")
	catA, _ := ts.CreateCategory(ctx, tenantA, "cat-a", "Category A")
	ts.CreateEngagementType(ctx, tenantA, catA, "only-a", "Only A", "benefit", "active")

	// Create tenant B with data
	tenantB, _ := ts.CreateTenant(ctx, "iso-b", "Isolation B")
	catB, _ := ts.CreateCategory(ctx, tenantB, "cat-b", "Category B")
	ts.CreateEngagementType(ctx, tenantB, catB, "only-b", "Only B", "benefit", "active")

	// Verify isolation via DB
	var countA, countB int64
	ts.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM lkfl_platform.engagement_types WHERE tenant_id = $1", tenantA,
	).Scan(&countA)
	ts.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM lkfl_platform.engagement_types WHERE tenant_id = $1", tenantB,
	).Scan(&countB)

	if countA != 1 {
		t.Errorf("tenant A: expected 1 engagement, got %d", countA)
	}
	if countB != 1 {
		t.Errorf("tenant B: expected 1 engagement, got %d", countB)
	}

	// Cross-check: A should NOT see B's data
	ts.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM lkfl_platform.engagement_types WHERE tenant_id = $1 AND slug = $2",
		tenantA, "only-b",
	).Scan(&countA)
	if countA != 0 {
		t.Errorf("tenant A should NOT see 'only-b', got count=%d", countA)
	}

	ts.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM lkfl_platform.engagement_types WHERE tenant_id = $1 AND slug = $2",
		tenantB, "only-a",
	).Scan(&countB)
	if countB != 0 {
		t.Errorf("tenant B should NOT see 'only-a', got count=%d", countB)
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestMultiTenantIsolation_UserIsolation — пользователи tenant A ≠ пользователи tenant B
func TestMultiTenantIsolation_UserIsolation(t *testing.T) {
	ts, cleanup := setupIsolationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenants
	tenantA, _ := ts.CreateTenant(ctx, "user-iso-a", "User Isolation A")
	tenantB, _ := ts.CreateTenant(ctx, "user-iso-b", "User Isolation B")

	// Create users
	userA, _ := ts.CreateUser(ctx, tenantA, "user@a.local", "User", "A", "kc-user-a")
	userB, _ := ts.CreateUser(ctx, tenantB, "user@b.local", "User", "B", "kc-user-b")

	// Verify user A belongs to tenant A
	var tenantID uuid.UUID
	ts.DB.QueryRow(ctx,
		"SELECT tenant_id FROM lkfl_platform.users WHERE id = $1", userA,
	).Scan(&tenantID)
	if tenantID != tenantA {
		t.Errorf("user A: expected tenant %s, got %s", tenantA, tenantID)
	}

	// Verify user B belongs to tenant B
	ts.DB.QueryRow(ctx,
		"SELECT tenant_id FROM lkfl_platform.users WHERE id = $1", userB,
	).Scan(&tenantID)
	if tenantID != tenantB {
		t.Errorf("user B: expected tenant %s, got %s", tenantB, tenantID)
	}

	// Cross-check: tenant A should NOT have user B
	var count int64
	ts.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM lkfl_platform.users WHERE tenant_id = $1 AND id = $2",
		tenantA, userB,
	).Scan(&count)
	if count != 0 {
		t.Error("tenant A should NOT have user B")
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestMultiTenantIsolation_AdminSeesAll — admin видит все tenants
func TestMultiTenantIsolation_AdminSeesAll(t *testing.T) {
	ts, cleanup := setupIsolationTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create multiple tenants
	_, _ = ts.CreateTenant(ctx, "admin-all-a", "Admin All A")
	_, _ = ts.CreateTenant(ctx, "admin-all-b", "Admin All B")
	_, _ = ts.CreateTenant(ctx, "admin-all-c", "Admin All C")

	// Admin should see all tenants
	resp, err := ts.GetWithToken("/admin/tenants", adminToken)
	if err != nil {
		t.Fatalf("list tenants: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
		return
	}

	var tenants []struct {
		Slug string `json:"slug"`
	}
	json.NewDecoder(resp.Body).Decode(&tenants)
	if len(tenants) < 3 {
		t.Errorf("admin should see at least 3 tenants, got %d", len(tenants))
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestMultiTenantIsolation_BrandConfigIsolation — brand config изолирован по tenant
func TestMultiTenantIsolation_BrandConfigIsolation(t *testing.T) {
	ts, cleanup := setupIsolationTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create tenants
	tenantA, _ := ts.CreateTenant(ctx, "brand-iso-a", "Brand Isolation A")
	tenantB, _ := ts.CreateTenant(ctx, "brand-iso-b", "Brand Isolation B")

	// Create brand configs
	ts.CreateBrandConfig(ctx, tenantA, "#FF0000", "#00FF00")
	ts.CreateBrandConfig(ctx, tenantB, "#0000FF", "#FFFF00")

	// Verify via API
	resp, err := ts.GetWithToken("/admin/tenants/"+tenantA.String()+"/brand", adminToken)
	if err != nil {
		t.Fatalf("get brand A: %v", err)
	}
	var brandA struct {
		PrimaryColor string `json:"primary_color"`
	}
	json.NewDecoder(resp.Body).Decode(&brandA)
	resp.Body.Close()

	if brandA.PrimaryColor != "#FF0000" {
		t.Errorf("tenant A brand: expected '#FF0000', got '%s'", brandA.PrimaryColor)
	}

	resp, err = ts.GetWithToken("/admin/tenants/"+tenantB.String()+"/brand", adminToken)
	if err != nil {
		t.Fatalf("get brand B: %v", err)
	}
	var brandB struct {
		PrimaryColor string `json:"primary_color"`
	}
	json.NewDecoder(resp.Body).Decode(&brandB)
	resp.Body.Close()

	if brandB.PrimaryColor != "#0000FF" {
		t.Errorf("tenant B brand: expected '#0000FF', got '%s'", brandB.PrimaryColor)
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}

// TestMultiTenantIsolation_CascadeDelete — удаление tenant удаляет все данные (CASCADE)
func TestMultiTenantIsolation_CascadeDelete(t *testing.T) {
	ts, cleanup := setupIsolationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant with data
	tenantID, _ := ts.CreateTenant(ctx, "cascade-del", "Cascade Delete")
	catID, _ := ts.CreateCategory(ctx, tenantID, "cascade-cat", "Cascade Cat")
	typeID, _ := ts.CreateEngagementType(ctx, tenantID, catID, "cascade-type", "Cascade Type", "benefit", "active")
	userID, _ := ts.CreateUser(ctx, tenantID, "cascade@test.local", "Cascade", "User", "kc-cascade")

	// Delete tenant
	_, err := ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants WHERE id = $1", tenantID)
	if err != nil {
		t.Fatalf("delete tenant: %v", err)
	}

	// Verify CASCADE: all related data deleted
	var catCount, typeCount, userCount int64
	ts.DB.QueryRow(ctx, "SELECT COUNT(*) FROM lkfl_platform.engagement_categories WHERE id = $1", catID).Scan(&catCount)
	ts.DB.QueryRow(ctx, "SELECT COUNT(*) FROM lkfl_platform.engagement_types WHERE id = $1", typeID).Scan(&typeCount)
	ts.DB.QueryRow(ctx, "SELECT COUNT(*) FROM lkfl_platform.users WHERE id = $1", userID).Scan(&userCount)

	if catCount != 0 {
		t.Errorf("category should be CASCADE deleted, got count=%d", catCount)
	}
	if typeCount != 0 {
		t.Errorf("engagement type should be CASCADE deleted, got count=%d", typeCount)
	}
	if userCount != 0 {
		t.Errorf("user should be CASCADE deleted, got count=%d", userCount)
	}
}

// TestMultiTenantIsolation_TenantHeader — X-Tenant-ID header для tenant resolution
func TestMultiTenantIsolation_TenantHeader(t *testing.T) {
	ts, cleanup := setupIsolationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant
	tenantID, _ := ts.CreateTenant(ctx, "header-test", "Header Test")

	// Create user
	userID, _ := ts.CreateUser(ctx, tenantID, "header@test.local", "Header", "User", "kc-header")
	ts.AddUserRole(ctx, userID, "admin")

	// Request with X-Tenant-ID header
	req, _ := http.NewRequest("GET", ts.URL+"/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+testutil.TestTokenAdmin("kc-header"))
	req.Header.Set("X-Tenant-ID", "header-test")

	resp, err := ts.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 200 with users from the specified tenant
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 with X-Tenant-ID header, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}

	// Cleanup
	ts.DB.Exec(ctx, "DELETE FROM lkfl_platform.tenants")
}
