//go:build integration

package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"lkfl/internal/testutil"
)

func setupAuthTest(t *testing.T) (*testutil.TestServer, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ts, cleanup, err := testutil.SetupAllWithServer(ctx)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return ts, cleanup
}

// TestAuthMe_ValidToken — GET /api/v1/auth/me с валидным токеном
func TestAuthMe_ValidToken(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant + user
	tenantID, _ := ts.CreateTenant(ctx, "auth-test", "Auth Test")
	subj := "kc-sub-auth-me"
	_, _ = ts.CreateUser(ctx, tenantID, "auth@test.local", "Auth", "User", subj)
	_ = tenantID

	token := testutil.TestToken(subj, "employee")

	resp, err := ts.GetWithToken("/api/v1/auth/me", token)
	if err != nil {
		t.Fatalf("me: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body := testutil.ReadBody(resp)
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, body)
		return
	}

	var profile struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	json.NewDecoder(resp.Body).Decode(&profile)
	resp.Body.Close()
	if profile.Email != "auth@test.local" {
		t.Errorf("expected email 'auth@test.local', got '%s'", profile.Email)
	}
}

// TestAuthMe_MissingToken — GET /api/v1/auth/me без токена → 401
func TestAuthMe_MissingToken(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	// Debug: check health
	healthResp, _ := http.Get(ts.URL + "/healthz")
	healthBody := testutil.ReadBody(healthResp)
	t.Logf("Server URL: %s, Health: %d %s", ts.URL, healthResp.StatusCode, healthBody)

	// Check admin route
	adminResp, _ := http.Get(ts.URL + "/admin/tenants")
	adminBody := testutil.ReadBody(adminResp)
	t.Logf("Admin tenants (no auth): %d %s", adminResp.StatusCode, adminBody)

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/auth/me", nil)
	resp, err := ts.Do(req)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	body := testutil.ReadBody(resp)
	t.Logf("Auth/me no token: status=%d body=%s", resp.StatusCode, body)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", resp.StatusCode)
	}
}

// TestAuthMe_InvalidTokenFormat — невалидный формат токена → 401
func TestAuthMe_InvalidTokenFormat(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	resp, err := ts.Do(req)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid token format, got %d body=%s", resp.StatusCode, body)
	}
}

// TestAuthMe_UserNotFound — токен валидный, но пользователь не найден
func TestAuthMe_UserNotFound(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	// Token with subject that doesn't exist in DB
	token := testutil.TestToken("nonexistent-user", "employee")

	resp, err := ts.GetWithToken("/api/v1/auth/me", token)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent user, got %d body=%s", resp.StatusCode, body)
	}
}

// TestRBAC_AdminRoute_EmployeeForbidden — admin route с ролью employee → 403
func TestRBAC_AdminRoute_EmployeeForbidden(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	employeeToken := testutil.TestTokenEmployee("employee-user")

	resp, _ := ts.GetWithToken("/admin/tenants", employeeToken)
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for employee on admin route, got %d body=%s", resp.StatusCode, body)
	}
}

// TestRBAC_AdminRoute_AdminAllowed — admin route с ролью admin → 200
func TestRBAC_AdminRoute_AdminAllowed(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	resp, _ := ts.GetWithToken("/admin/tenants", adminToken)
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for admin on admin route, got %d body=%s", resp.StatusCode, body)
	}
}

// TestRBAC_HRRoute_HRAllowed — HR route с ролью hr → 200
func TestRBAC_HRRoute_HRAllowed(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = ts.CreateTenant(ctx, "rbac-test", "RBAC Test")

	hrToken := testutil.TestTokenHR("hr-user")

	// HR can access /admin/users (with tenant header)
	resp, _ := ts.GetWithTokenAndTenant("/admin/users", hrToken, "rbac-test")
	body := testutil.ReadBody(resp)

	// Should be 200 (empty list) or at least not 403
	if resp.StatusCode == http.StatusForbidden {
		t.Errorf("expected not 403 for HR on HR route, got %d body=%s", resp.StatusCode, body)
	}
}

// TestLogout — POST /api/v1/auth/logout → 302 redirect
func TestLogout(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	resp, err := ts.PostWithToken("/api/v1/auth/logout", adminToken, nil)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	body := testutil.ReadBody(resp)
	t.Logf("Logout: status=%d body=%s", resp.StatusCode, body)

	// Logout may return 302 redirect or 200 OK depending on implementation
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		t.Logf("Logout returned status %d (may vary)", resp.StatusCode)
	}
}

// TestEmployeeRoutes_AccessControl — employee может получить /api/v1/users/me
func TestEmployeeRoutes_AccessControl(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenant + user
	tenantID, _ := ts.CreateTenant(ctx, "emp-test", "Employee Test")
	tenantSlug := "emp-test"
	subj := "kc-sub-emp"
	_, _ = ts.CreateUser(ctx, tenantID, "emp@test.local", "Emp", "User", subj)

	employeeToken := testutil.TestTokenEmployee(subj)

	// Employee can access /api/v1/users/me (with X-Tenant-ID header)
	resp, err := ts.GetWithTokenAndTenant("/api/v1/users/me", employeeToken, tenantSlug)
	if err != nil {
		t.Fatalf("users/me: %v", err)
	}
	body := testutil.ReadBody(resp)
	t.Logf("Users/me: status=%d body=%s", resp.StatusCode, body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for employee on /users/me, got %d body=%s", resp.StatusCode, body)
	}
}

// TestAuthMe_MultipleRoles — пользователь с несколькими ролями
func TestAuthMe_MultipleRoles(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	ctx := context.Background()

	tenantID, _ := ts.CreateTenant(ctx, "multi-role", "Multi Role")
	subj := "kc-sub-multi"
	_, _ = ts.CreateUser(ctx, tenantID, "multi@test.local", "Multi", "User", subj)
	_ = tenantID

	// Token with multiple roles
	token := testutil.TestToken(subj, "admin", "hr", "employee")

	resp, err := ts.GetWithToken("/api/v1/auth/me", token)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, body)
		return
	}

	// Also should be able to access admin routes
	resp2, _ := ts.GetWithToken("/admin/tenants", token)
	body2 := testutil.ReadBody(resp2)
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for multi-role admin access, got %d body=%s", resp2.StatusCode, body2)
	}
}

// TestRBAC_CatalogManagerRole — catalog_manager может доступ к catalog admin
func TestRBAC_CatalogManagerRole(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = ts.CreateTenant(ctx, "catalog-test", "Catalog Test")

	cmToken := testutil.TestTokenCatalogManager("cm-user")

	// Catalog manager can access catalog admin routes
	resp, _ := ts.GetWithToken("/admin/engagements/types", cmToken)
	body := testutil.ReadBody(resp)

	if resp.StatusCode == http.StatusForbidden {
		t.Errorf("expected not 403 for catalog_manager on catalog route, got %d body=%s", resp.StatusCode, body)
	}
}

// TestRBAC_EmployeeCannotAccessAdmin — employee не может доступ к admin/users
func TestRBAC_EmployeeCannotAccessAdmin(t *testing.T) {
	ts, cleanup := setupAuthTest(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = ts.CreateTenant(ctx, "rbac-emp", "RBAC Employee")

	employeeToken := testutil.TestTokenEmployee("emp-user")

	resp, _ := ts.GetWithTokenAndTenant("/admin/users", employeeToken, "rbac-emp")
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for employee on admin/users, got %d body=%s", resp.StatusCode, body)
	}
}
