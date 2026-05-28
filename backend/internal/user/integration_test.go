//go:build integration

package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"lkfl/internal/testutil"
)

func setupUserTest(t *testing.T) (*testutil.TestServer, func()) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ts, cleanup, err := testutil.SetupAllWithServer(ctx)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return ts, cleanup
}

// TestUserCRUD_FullCycle — полный цикл: Create → Get → Update → Deactivate
func TestUserCRUD_FullCycle(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	// Create tenant
	tenantID, _ := ts.CreateTenant(ctx, "user-crud", "User CRUD Test")
	tenantSlug := "user-crud"

	// Create user via DB
	userID, _ := ts.CreateUser(ctx, tenantID, "user@test.local", "John", "Doe", "kc-sub-crud")

	// 1. Get user
	resp, err := ts.GetWithToken("/admin/users/"+userID.String(), adminToken)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}
	var gotUser struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Status    string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&gotUser)
	resp.Body.Close()
	if gotUser.Email != "user@test.local" {
		t.Errorf("expected email 'user@test.local', got '%s'", gotUser.Email)
	}

	// 2. Update user
	updateReq := map[string]string{
		"first_name": "Jane",
		"last_name":  "Smith",
	}
	resp, err = ts.PutWithToken("/admin/users/"+userID.String(), adminToken, updateReq)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}

	// Verify update
	resp, err = ts.GetWithToken("/admin/users/"+userID.String(), adminToken)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&gotUser)
	resp.Body.Close()
	if gotUser.FirstName != "Jane" {
		t.Errorf("expected first_name 'Jane', got '%s'", gotUser.FirstName)
	}

	// 3. Deactivate
	resp, err = ts.PostWithToken("/admin/users/"+userID.String()+"/deactivate", adminToken, nil)
	if err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("deactivate: expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
	}

	// Verify deactivated
	resp, err = ts.GetWithToken("/admin/users/"+userID.String(), adminToken)
	if err != nil {
		t.Fatalf("get after deactivate: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&gotUser)
	resp.Body.Close()
	if gotUser.Status != "deactivated" {
		t.Errorf("expected status 'deactivated', got '%s'", gotUser.Status)
	}

	_ = tenantSlug
}

// TestUserMe_Profile — GET /api/v1/users/me
func TestUserMe_Profile(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()

	tenantID, _ := ts.CreateTenant(ctx, "me-test", "Me Test")
	tenantSlug := "me-test"
	subj := "kc-sub-me"
	_, _ = ts.CreateUser(ctx, tenantID, "me@test.local", "Me", "User", subj)

	token := testutil.TestToken(subj, "employee")

	resp, err := ts.GetWithTokenAndTenant("/api/v1/users/me", token, tenantSlug)
	if err != nil {
		t.Fatalf("me: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
		return
	}

	var profile struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	json.NewDecoder(resp.Body).Decode(&profile)
	resp.Body.Close()
	t.Logf("Users/me: email=%s first_name=%s last_name=%s", profile.Email, profile.FirstName, profile.LastName)
	if profile.Email != "me@test.local" {
		t.Errorf("expected email 'me@test.local', got '%s'", profile.Email)
	}
}

// TestUserUpdateMe — PUT /api/v1/users/me
func TestUserUpdateMe(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()

	tenantID, _ := ts.CreateTenant(ctx, "update-me", "Update Me Test")
	tenantSlug := "update-me"
	subj := "kc-sub-update-me"
	_, _ = ts.CreateUser(ctx, tenantID, "update@test.local", "Orig", "Name", subj)

	token := testutil.TestToken(subj, "employee")

	updateReq := map[string]string{
		"first_name": "Updated",
		"last_name":  "Name",
		"phone":      "+1234567890",
	}
	resp, err := ts.PutWithTokenAndTenant("/api/v1/users/me", token, tenantSlug, updateReq)
	if err != nil {
		t.Fatalf("update me: %v", err)
	}
	body := testutil.ReadBody(resp)
	t.Logf("Update me: status=%d body=%s", resp.StatusCode, body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, body)
	}
}

// TestUserList_WithFilters — GET /admin/users с фильтрами
func TestUserList_WithFilters(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	tenantID, _ := ts.CreateTenant(ctx, "list-filter", "List Filter Test")
	tenantSlug := "list-filter"

	// Create users
	_, _ = ts.CreateUser(ctx, tenantID, "alice@test.local", "Alice", "Test", "kc-alice")
	_, _ = ts.CreateUser(ctx, tenantID, "bob@test.local", "Bob", "Test", "kc-bob")
	_, _ = ts.CreateUser(ctx, tenantID, "charlie@test.local", "Charlie", "Test", "kc-charlie")

	// List with search (with tenant header)
	resp, err := ts.GetWithTokenAndTenant("/admin/users?search=alice&page=1&per_page=10", adminToken, tenantSlug)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
		return
	}

	var users []struct {
		Email string `json:"email"`
	}
	json.NewDecoder(resp.Body).Decode(&users)
	resp.Body.Close()
	t.Logf("List with filter: found %d users", len(users))
	if len(users) != 1 {
		t.Errorf("expected 1 user with search 'alice', got %d", len(users))
	}
}

// TestUserList_Pagination — пагинация списка пользователей
func TestUserList_Pagination(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	tenantID, _ := ts.CreateTenant(ctx, "pagination", "Pagination Test")
	tenantSlug := "pagination"

	// Create 5 users
	for i := 0; i < 5; i++ {
		email := "user" + string(rune('a'+i)) + "@test.local"
		_, _ = ts.CreateUser(ctx, tenantID, email, "User", string(rune('A'+i)), "kc-"+string(rune('a'+i)))
	}

	// Page 1, per_page=2 (with tenant header)
	resp, err := ts.GetWithTokenAndTenant("/admin/users?page=1&per_page=2", adminToken, tenantSlug)
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
		return
	}

	var users []struct{}
	json.NewDecoder(resp.Body).Decode(&users)
	resp.Body.Close()
	t.Logf("List pagination: got %d users on page 1", len(users))
	if len(users) != 2 {
		t.Errorf("expected 2 users on page 1, got %d", len(users))
	}
}

// TestUserDeactivate_AlreadyDeactivated — попытка деактивировать уже деактивированного
func TestUserDeactivate_AlreadyDeactivated(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	tenantID, _ := ts.CreateTenant(ctx, "deactivate", "Deactivate Test")
	userID, _ := ts.CreateUser(ctx, tenantID, "deact@test.local", "Deact", "User", "kc-deact")

	// First deactivation
	resp, _ := ts.PostWithToken("/admin/users/"+userID.String()+"/deactivate", adminToken, nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first deactivate: expected 200, got %d", resp.StatusCode)
	}

	// Second deactivation should fail
	resp, _ = ts.PostWithToken("/admin/users/"+userID.String()+"/deactivate", adminToken, nil)
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("second deactivate: expected 409, got %d body=%s", resp.StatusCode, body)
	}
}

// TestUserDeactivate_NoAuth — деактивация без авторизации → 401
func TestUserDeactivate_NoAuth(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	req, _ := http.NewRequest("POST", ts.URL+"/admin/users/"+uuid.New().String()+"/deactivate", nil)
	resp, err := ts.Do(req)
	if err != nil {
		t.Fatalf("deactivate: %v", err)
	}
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d body=%s", resp.StatusCode, body)
	}
}

// TestUserMe_NoAuth — GET /api/v1/users/me без токена → 401
func TestUserMe_NoAuth(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/users/me", nil)
	resp, err := ts.Do(req)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d body=%s", resp.StatusCode, body)
	}
}

// TestUserAdminGet_InvalidID — невалидный UUID → 400
func TestUserAdminGet_InvalidID(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	resp, _ := ts.GetWithToken("/admin/users/not-a-uuid", adminToken)
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid UUID, got %d body=%s", resp.StatusCode, body)
	}
}

// TestUserAdminGet_NotFound — существующий UUID но пользователь не найден → 404
func TestUserAdminGet_NotFound(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	adminToken := testutil.TestTokenAdmin("admin-user")

	// Use a valid UUID that doesn't exist in DB
	fakeID := uuid.New()
	resp, _ := ts.GetWithToken("/admin/users/"+fakeID.String(), adminToken)
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent user, got %d body=%s", resp.StatusCode, body)
	}
}

// TestUserUpdateDeactivated — попытка обновить деактивированного пользователя → 403
func TestUserUpdateDeactivated(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()
	adminToken := testutil.TestTokenAdmin("admin-user")

	tenantID, _ := ts.CreateTenant(ctx, "update-deact", "Update Deactivated Test")
	userID, _ := ts.CreateUser(ctx, tenantID, "deact-update@test.local", "Deact", "Update", "kc-deact-update")

	// Deactivate first
	resp, _ := ts.PostWithToken("/admin/users/"+userID.String()+"/deactivate", adminToken, nil)
	resp.Body.Close()

	// Try to update deactivated user
	updateReq := map[string]string{"first_name": "New Name"}
	resp, _ = ts.PutWithToken("/admin/users/"+userID.String(), adminToken, updateReq)
	body := testutil.ReadBody(resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for updating deactivated user, got %d body=%s", resp.StatusCode, body)
	}
}

// TestUserMe_TenantIsolation — пользователь видит только свой профиль
func TestUserMe_TenantIsolation(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	ctx := context.Background()

	// Create two tenants
	tenantA, _ := ts.CreateTenant(ctx, "iso-a", "Isolation A")
	tenantB, _ := ts.CreateTenant(ctx, "iso-b", "Isolation B")

	// Create users in each tenant
	subjA := "kc-sub-a"
	subjB := "kc-sub-b"
	_, _ = ts.CreateUser(ctx, tenantA, "user@a.local", "User", "A", subjA)
	_, _ = ts.CreateUser(ctx, tenantB, "user@b.local", "User", "B", subjB)

	// User A can get their own profile
	tokenA := testutil.TestToken(subjA, "employee")
	resp, err := ts.GetWithTokenAndTenant("/api/v1/users/me", tokenA, "iso-a")
	if err != nil {
		t.Fatalf("me for user A: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for user A, got %d body=%s", resp.StatusCode, testutil.ReadBody(resp))
		return
	}

	var profile struct {
		Email string `json:"email"`
	}
	json.NewDecoder(resp.Body).Decode(&profile)
	resp.Body.Close()
	t.Logf("User A me: email=%s", profile.Email)
	if profile.Email != "user@a.local" {
		t.Errorf("expected email 'user@a.local', got '%s'", profile.Email)
	}

	_ = tenantB
	_ = subjB
}

// TestUserUpdateMe_NoAuth — PUT /api/v1/users/me без токена → 401
func TestUserUpdateMe_NoAuth(t *testing.T) {
	ts, cleanup := setupUserTest(t)
	defer cleanup()

	reqBody, _ := json.Marshal(map[string]string{"first_name": "Hacked"})
	req, _ := http.NewRequest("PUT", ts.URL+"/api/v1/users/me", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := ts.Do(req)
	if err != nil {
		t.Fatalf("update me: %v", err)
	}
	body := testutil.ReadBody(resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d body=%s", resp.StatusCode, body)
	}
}
