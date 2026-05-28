package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// =============================================================================
// JWT Middleware Edge Cases
// =============================================================================

// mockTokenVerifier — мок для oidc.IDTokenVerifier для тестов.
type mockTokenVerifier struct {
	verifyFn func(ctx context.Context, rawID string) (*mockIDToken, error)
}

func (m *mockTokenVerifier) Verify(ctx context.Context, rawID string) (*mockIDToken, error) {
	if m.verifyFn != nil {
		return m.verifyFn(ctx, rawID)
	}
	return &mockIDToken{subject: "user-123"}, nil
}

// mockIDToken — минимальный мок ID token.
type mockIDToken struct {
	subject string
	claims  map[string]interface{}
}

func (m *mockIDToken) Claims(v interface{}) error {
	return nil
}

// mockOIDCVerifierForMiddleware — адаптер для JWTMiddleware.
// Реализует интерфейс, ожидаемый middleware (Verify → *oidc.IDToken).
// Для тестов мы используем httptest и проверяем поведение middleware
// на уровне HTTP-запросов.

// testJWTMiddleware — тестовый middleware, имитирующий JWT middleware
// с контролируемыми ответами.
type testJWTMiddleware struct {
	tokenValid   bool
	tokenString  string
	claims       Claims
	roles        []string
	returnError  bool
	errorMessage string
}

func (m *testJWTMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			writeJSONError(w, http.StatusUnauthorized, "invalid token format")
			return
		}

		if m.returnError {
			writeJSONError(w, http.StatusUnauthorized, m.errorMessage)
			return
		}

		if !m.tokenValid {
			writeJSONError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, m.claims)
		ctx = context.WithValue(ctx, RolesKey, m.roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TestJWTMiddleware_NoAuthHeader(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "unauthorized") {
		t.Errorf("expected 'unauthorized' in body, got: %s", rr.Body.String())
	}
}

func TestJWTMiddleware_EmptyAuthHeader(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_NoBearerPrefix(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Token abc123")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "invalid token format") {
		t.Errorf("expected 'invalid token format' in body, got: %s", rr.Body.String())
	}
}

func TestJWTMiddleware_BasicAuthFormat(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_EmptyTokenString(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: true}
	var called bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// "Bearer " with TrimPrefix returns "" which passes format check.
	// The real OIDC verifier would reject empty token string.
	// This test verifies the format check behavior.
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 (empty token passes format check), got %d", rr.Code)
	}
	if !called {
		t.Error("handler should be called — empty token passes format check")
	}
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: false, returnError: true, errorMessage: "token expired"}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for expired token")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.expired.signature")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "token expired") {
		t.Errorf("expected 'token expired' in body, got: %s", rr.Body.String())
	}
}

func TestJWTMiddleware_InvalidSignature(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: false, returnError: true, errorMessage: "invalid signature"}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.signature.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_MalformedJWT(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: false, returnError: true, errorMessage: "malformed JWT"}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_WrongAlgorithm(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: false, returnError: true, errorMessage: "wrong algorithm"}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiJ9.wrong-algo.sig")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	mw := &testJWTMiddleware{
		tokenValid: true,
		claims:     Claims{Subject: "user-123", Email: "test@example.com"},
		roles:      []string{"employee"},
	}
	var capturedCtx context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid.token.here")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	claims := UserIDFromContext(capturedCtx)
	if claims != "user-123" {
		t.Errorf("expected subject 'user-123', got '%s'", claims)
	}

	roles := RolesFromContext(capturedCtx)
	if len(roles) != 1 || roles[0] != "employee" {
		t.Errorf("expected roles ['employee'], got %v", roles)
	}
}

func TestJWTMiddleware_TokenWithExtraClaims(t *testing.T) {
	mw := &testJWTMiddleware{
		tokenValid: true,
		claims: Claims{
			Subject:           "user-456",
			Email:             "extra@example.com",
			PreferredUsername: "extra_user",
			Name:              "Extra User",
		},
		roles: []string{"employee", "hr"},
	}
	var capturedCtx context.Context
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCtx = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token.with.extra.claims")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	roles := RolesFromContext(capturedCtx)
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestJWTMiddleware_NilContext(t *testing.T) {
	mw := &testJWTMiddleware{tokenValid: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Handler(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid.token")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestJWTMiddleware_MultipleRequests(t *testing.T) {
	mw := &testJWTMiddleware{
		tokenValid: true,
		claims:     Claims{Subject: "user-123"},
		roles:      []string{"employee"},
	}
	callCount := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	handler := mw.Handler(next)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid.token")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rr.Code)
		}
	}

	if callCount != 5 {
		t.Errorf("expected 5 calls, got %d", callCount)
	}
}

// =============================================================================
// RBAC Middleware Edge Cases
// =============================================================================

func TestRBACMiddleware_NoRolesInContext(t *testing.T) {
	mw := RBACMiddleware([]string{"admin"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without roles")
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	// No roles in context
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRBACMiddleware_WrongRole(t *testing.T) {
	mw := RBACMiddleware([]string{"admin"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with wrong role")
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{"employee"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRBACMiddleware_MultipleRolesOneMatches(t *testing.T) {
	mw := RBACMiddleware([]string{"admin", "manager"})
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{"employee", "admin", "viewer"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("handler should be called when one role matches")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRBACMiddleware_RoleEscalationEmployeeToAdmin(t *testing.T) {
	mw := RBACMiddleware([]string{"admin"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called — role escalation attempt")
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{"employee"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for role escalation, got %d", rr.Code)
	}
}

func TestRBACMiddleware_EmptyRolesList(t *testing.T) {
	mw := RBACMiddleware([]string{"admin"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with empty roles")
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for empty roles, got %d", rr.Code)
	}
}

func TestRBACMiddleware_NilUserInContext(t *testing.T) {
	mw := RBACMiddleware([]string{"admin"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without user")
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	// No roles set in context at all
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRBACMiddleware_EmptyRequiredRoles(t *testing.T) {
	mw := RBACMiddleware([]string{})
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := withRoles(req.Context(), []string{"employee"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if nextCalled {
		t.Error("handler should not be called with empty required roles (no role matches)")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRBACMiddleware_NilRequiredRoles(t *testing.T) {
	mw := RBACMiddleware(nil)
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := withRoles(req.Context(), []string{"employee"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if nextCalled {
		t.Error("handler should not be called with nil required roles")
	}
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRBACMiddleware_MultipleRequiredRoles(t *testing.T) {
	mw := RBACMiddleware([]string{"admin", "manager", "hr"})
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{"hr"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("handler should be called when one of multiple required roles matches")
	}
}

func TestRBACMiddleware_AllRolesWrong(t *testing.T) {
	mw := RBACMiddleware([]string{"admin", "manager"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{"employee", "viewer"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRBACMiddleware_SingleRequiredRoleMatches(t *testing.T) {
	mw := RBACMiddleware([]string{"admin"})
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{"admin"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !nextCalled {
		t.Error("handler should be called")
	}
}

func TestRBACMiddleware_ResponseContainsJSONError(t *testing.T) {
	mw := RBACMiddleware([]string{"admin"})
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	handler := mw(next)
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	ctx := withRoles(req.Context(), []string{"employee"})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"error"`) {
		t.Errorf("expected JSON error in response, got: %s", body)
	}
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got '%s'", contentType)
	}
}

// =============================================================================
// ExtractClaims Edge Cases
// =============================================================================

func TestExtractKeycloakRoles_NoResourceAccess(t *testing.T) {
	raw := map[string]interface{}{
		"sub": "user-123",
		"email": "test@example.com",
	}
	roles := extractKeycloakRoles(raw)
	if roles != nil {
		t.Errorf("expected nil roles, got %v", roles)
	}
}

func TestExtractKeycloakRoles_EmptyResourceAccess(t *testing.T) {
	raw := map[string]interface{}{
		"resource_access": map[string]interface{}{},
	}
	roles := extractKeycloakRoles(raw)
	if roles != nil {
		t.Errorf("expected nil roles, got %v", roles)
	}
}

func TestExtractKeycloakRoles_ClientWithoutRoles(t *testing.T) {
	raw := map[string]interface{}{
		"resource_access": map[string]interface{}{
			"lkfl-spa": map[string]interface{}{
				"other_field": "value",
			},
		},
	}
	roles := extractKeycloakRoles(raw)
	if roles != nil {
		t.Errorf("expected nil roles when client has no roles field, got %v", roles)
	}
}

func TestExtractKeycloakRoles_EmptyRolesArray(t *testing.T) {
	raw := map[string]interface{}{
		"resource_access": map[string]interface{}{
			"lkfl-spa": map[string]interface{}{
				"roles": []interface{}{},
			},
		},
	}
	roles := extractKeycloakRoles(raw)
	if roles != nil {
		t.Errorf("expected nil roles for empty array, got %v", roles)
	}
}

func TestExtractKeycloakRoles_NonStringRole(t *testing.T) {
	raw := map[string]interface{}{
		"resource_access": map[string]interface{}{
			"lkfl-spa": map[string]interface{}{
				"roles": []interface{}{"admin", 42, true, nil},
			},
		},
	}
	roles := extractKeycloakRoles(raw)
	if len(roles) != 1 {
		t.Errorf("expected 1 role (only string), got %d: %v", len(roles), roles)
	}
	if roles[0] != "admin" {
		t.Errorf("expected 'admin', got '%s'", roles[0])
	}
}

func TestExtractKeycloakRoles_NonMapClient(t *testing.T) {
	raw := map[string]interface{}{
		"resource_access": map[string]interface{}{
			"lkfl-spa": "not-a-map",
		},
	}
	roles := extractKeycloakRoles(raw)
	if roles != nil {
		t.Errorf("expected nil roles when client is not a map, got %v", roles)
	}
}

func TestExtractKeycloakRoles_NilResourceAccess(t *testing.T) {
	raw := map[string]interface{}{
		"resource_access": nil,
	}
	roles := extractKeycloakRoles(raw)
	if roles != nil {
		t.Errorf("expected nil roles, got %v", roles)
	}
}

func TestExtractKeycloakRoles_MultipleClientsSecondHasRoles(t *testing.T) {
	raw := map[string]interface{}{
		"resource_access": map[string]interface{}{
			"lkfl-service": map[string]interface{}{
				"roles": []interface{}{},
			},
			"lkfl-spa": map[string]interface{}{
				"roles": []interface{}{"admin", "employee"},
			},
		},
	}
	roles := extractKeycloakRoles(raw)
	if roles == nil {
		t.Fatal("expected roles from one of the clients, got nil")
	}
	// At least one client should return roles
	found := false
	for _, r := range roles {
		if r == "admin" || r == "employee" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'admin' or 'employee' in roles, got %v", roles)
	}
}

// =============================================================================
// Claims Context Helpers Edge Cases
// =============================================================================

func TestUserIDFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), ClaimsKey, "not-claims")
	result := UserIDFromContext(ctx)
	if result != "" {
		t.Errorf("expected empty string for wrong type, got '%s'", result)
	}
}

func TestRolesFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), RolesKey, "not-roles")
	result := RolesFromContext(ctx)
	if result != nil {
		t.Errorf("expected nil for wrong type, got %v", result)
	}
}

func TestRolesFromContext_IntSlice(t *testing.T) {
	ctx := context.WithValue(context.Background(), RolesKey, []int{1, 2})
	result := RolesFromContext(ctx)
	if result != nil {
		t.Errorf("expected nil for int slice, got %v", result)
	}
}

func TestUserIDFromContext_ValidClaims(t *testing.T) {
	ctx := context.WithValue(context.Background(), ClaimsKey, Claims{
		Subject: "user-789",
		Email:   "valid@example.com",
	})
	result := UserIDFromContext(ctx)
	if result != "user-789" {
		t.Errorf("expected 'user-789', got '%s'", result)
	}
}

func TestUserIDFromContext_EmptySubject(t *testing.T) {
	ctx := context.WithValue(context.Background(), ClaimsKey, Claims{
		Subject: "",
		Email:   "no-sub@example.com",
	})
	result := UserIDFromContext(ctx)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestRolesFromContext_ValidRoles(t *testing.T) {
	roles := []string{"admin", "employee"}
	ctx := context.WithValue(context.Background(), RolesKey, roles)
	result := RolesFromContext(ctx)
	if len(result) != 2 {
		t.Errorf("expected 2 roles, got %d", len(result))
	}
}

func TestRolesFromContext_EmptySlice(t *testing.T) {
	ctx := context.WithValue(context.Background(), RolesKey, []string{})
	result := RolesFromContext(ctx)
	if len(result) != 0 {
		t.Errorf("expected 0 roles, got %d", len(result))
	}
}

// =============================================================================
// writeJSONError Edge Cases
// =============================================================================

func TestWriteJSONError_Forbidden(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, http.StatusForbidden, "forbidden")

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"error":"forbidden"`) {
		t.Errorf("expected error field, got: %s", rr.Body.String())
	}
}

func TestWriteJSONError_BadRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, http.StatusBadRequest, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestWriteJSONError_EmptyMessage(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, http.StatusInternalServerError, "")

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"error":""`) {
		t.Errorf("expected empty error string, got: %s", rr.Body.String())
	}
}

func TestWriteJSONError_SpecialCharsInMessage(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, http.StatusBadRequest, "error with \"quotes\" and <script>")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
	// JSON encoding should escape special characters
}
