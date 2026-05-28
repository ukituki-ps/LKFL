package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRBACMiddleware_Authorized проверяет, что пользователь с подходящей ролью
// получает доступ к защищённому ресурсу.
func TestRBACMiddleware_Authorized(t *testing.T) {
	requiredRoles := []string{"admin", "manager"}

	// Handler, который всегда возвращает 200
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RBACMiddleware(requiredRoles)
	handler := mw(next)

	tests := []struct {
		name       string
		userRoles  []string
		wantStatus int
	}{
		{
			name:       "exact role match — admin",
			userRoles:  []string{"admin"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "exact role match — manager",
			userRoles:  []string{"manager"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "multiple roles including match",
			userRoles:  []string{"employee", "admin", "viewer"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "all required roles present",
			userRoles:  []string{"admin", "manager"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			ctx := withRoles(req.Context(), tt.userRoles)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("RBACMiddleware() status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

// TestRBACMiddleware_Unauthorized проверяет, что пользователь без подходящей
// роли получает 403 Forbidden.
func TestRBACMiddleware_Unauthorized(t *testing.T) {
	requiredRoles := []string{"admin", "manager"}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for unauthorized user")
	})

	mw := RBACMiddleware(requiredRoles)
	handler := mw(next)

	tests := []struct {
		name      string
		userRoles []string
	}{
		{
			name:      "no matching role",
			userRoles: []string{"employee", "viewer"},
		},
		{
			name:      "empty roles",
			userRoles: []string{},
		},
		{
			name:      "nil roles",
			userRoles: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			ctx := withRoles(req.Context(), tt.userRoles)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("RBACMiddleware() status = %d, want %d", rr.Code, http.StatusForbidden)
			}

			// Проверяем, что ответ — JSON с полем "error"
			if !strings.Contains(rr.Body.String(), `"error"`) {
				t.Errorf("RBACMiddleware() response should contain JSON error, got: %s", rr.Body.String())
			}
		})
	}
}

// TestRolesFromContext проверяет работу хелпера RolesFromContext.
func TestRolesFromContext(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantRoles []string
	}{
		{
			name:      "roles present",
			ctx:       withRoles(context.Background(), []string{"admin", "employee"}),
			wantRoles: []string{"admin", "employee"},
		},
		{
			name:      "no roles in context",
			ctx:       context.Background(),
			wantRoles: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RolesFromContext(tt.ctx)

			if len(got) != len(tt.wantRoles) {
				t.Errorf("RolesFromContext() = %v, want %v", got, tt.wantRoles)
				return
			}

			for i := range got {
				if got[i] != tt.wantRoles[i] {
					t.Errorf("RolesFromContext()[%d] = %q, want %q", i, got[i], tt.wantRoles[i])
				}
			}
		})
	}
}

// TestUserIDFromContext проверяет работу хелпера UserIDFromContext.
func TestUserIDFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "claims present",
			ctx: context.WithValue(context.Background(), ClaimsKey, Claims{
				Subject: "user-123",
				Email:   "test@example.com",
			}),
			want: "user-123",
		},
		{
			name: "no claims in context",
			ctx:  context.Background(),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserIDFromContext(tt.ctx)
			if got != tt.want {
				t.Errorf("UserIDFromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestExtractKeycloakRoles проверяет извлечение ролей из Keycloak raw claims.
func TestExtractKeycloakRoles(t *testing.T) {
	tests := []struct {
		name string
		raw  map[string]interface{}
		want []string
	}{
		{
			name: "single client with roles",
			raw: map[string]interface{}{
				"resource_access": map[string]interface{}{
					"lkfl-spa": map[string]interface{}{
						"roles": []interface{}{"admin", "employee"},
					},
				},
			},
			want: []string{"admin", "employee"},
		},
		{
			name: "no resource_access",
			raw: map[string]interface{}{
				"sub": "user-123",
			},
			want: nil,
		},
		{
			name: "empty resource_access",
			raw: map[string]interface{}{
				"resource_access": map[string]interface{}{},
			},
			want: nil,
		},
		{
			name: "multiple clients — first with roles wins",
			raw: map[string]interface{}{
				"resource_access": map[string]interface{}{
					"lkfl-service": map[string]interface{}{
						"roles": []interface{}{"service-role"},
					},
					"lkfl-spa": map[string]interface{}{
						"roles": []interface{}{"admin"},
					},
				},
			},
			want: nil, // Will return one of them (map iteration is non-deterministic, but not nil)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractKeycloakRoles(tt.raw)

			// Для теста "multiple clients" просто проверяем, что не nil
			if tt.name == "multiple clients — first with roles wins" {
				if got == nil {
					t.Error("expected roles from one of the clients, got nil")
				}
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("extractKeycloakRoles() = %v, want %v", got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("extractKeycloakRoles()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestWriteJSONError проверяет, что writeJSONError возвращает корректный JSON.
func TestWriteJSONError(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSONError(rr, http.StatusUnauthorized, "unauthorized")

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	if !strings.Contains(rr.Body.String(), `"error":"unauthorized"`) {
		t.Errorf("body should contain error field, got: %s", rr.Body.String())
	}
}
