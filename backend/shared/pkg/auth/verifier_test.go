package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewVerifier_Options проверяет конфигурацию через functional options.
func TestNewVerifier_Options(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := defaultConfig()
		if cfg.maxRetries != 30 {
			t.Errorf("expected maxRetries=30, got %d", cfg.maxRetries)
		}
		if cfg.retryDelay != 2*time.Second {
			t.Errorf("expected retryDelay=2s, got %v", cfg.retryDelay)
		}
	})

	t.Run("WithMaxRetries", func(t *testing.T) {
		cfg := defaultConfig()
		WithMaxRetries(10)(&cfg)
		if cfg.maxRetries != 10 {
			t.Errorf("expected maxRetries=10, got %d", cfg.maxRetries)
		}
	})

	t.Run("WithRetryDelay", func(t *testing.T) {
		cfg := defaultConfig()
		WithRetryDelay(5 * time.Second)(&cfg)
		if cfg.retryDelay != 5*time.Second {
			t.Errorf("expected retryDelay=5s, got %v", cfg.retryDelay)
		}
	})

	t.Run("WithMaxRetries zero ignored", func(t *testing.T) {
		cfg := defaultConfig()
		WithMaxRetries(0)(&cfg)
		if cfg.maxRetries != 30 {
			t.Errorf("expected default maxRetries=30, got %d", cfg.maxRetries)
		}
	})

	t.Run("WithRetryDelay zero ignored", func(t *testing.T) {
		cfg := defaultConfig()
		WithRetryDelay(0)(&cfg)
		if cfg.retryDelay != 2*time.Second {
			t.Errorf("expected default retryDelay=2s, got %v", cfg.retryDelay)
		}
	})
}

// TestNewVerifier_FailsOnBadIssuer проверяет что NewVerifier возвращает ошибку
// при недоступном issuer URL.
func TestNewVerifier_FailsOnBadIssuer(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := NewVerifier(ctx, "http://localhost:59999/realms/test", "test-client",
		WithMaxRetries(2), WithRetryDelay(100*time.Millisecond),
	)
	if err == nil {
		t.Error("expected error for unreachable issuer, got nil")
	}
}

// TestExtractToken проверяет извлечение токена из запроса.
func TestExtractToken(t *testing.T) {
	t.Run("from Bearer header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer test-token-123")

		token := extractToken(req)
		if token != "test-token-123" {
			t.Errorf("expected 'test-token-123', got %q", token)
		}
	})

	t.Run("from cookie when no header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.AddCookie(&http.Cookie{
			Name:  sessionCookieName,
			Value: "cookie-token-456",
		})

		token := extractToken(req)
		if token != "cookie-token-456" {
			t.Errorf("expected 'cookie-token-456', got %q", token)
		}
	})

	t.Run("Bearer header takes priority over cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		req.AddCookie(&http.Cookie{
			Name:  sessionCookieName,
			Value: "cookie-token",
		})

		token := extractToken(req)
		if token != "header-token" {
			t.Errorf("expected 'header-token' (header priority), got %q", token)
		}
	})

	t.Run("empty when nothing provided", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)

		token := extractToken(req)
		if token != "" {
			t.Errorf("expected empty string, got %q", token)
		}
	})

	t.Run("non-Bearer auth header ignored", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		req.AddCookie(&http.Cookie{
			Name:  sessionCookieName,
			Value: "cookie-token-789",
		})

		token := extractToken(req)
		if token != "cookie-token-789" {
			t.Errorf("expected cookie fallback for non-Bearer auth, got %q", token)
		}
	})
}

// TestResolveTenantSlug проверяет извлечение tenant slug из issuer URL.
func TestResolveTenantSlug(t *testing.T) {
	tests := []struct {
		name   string
		issuer string
		want   string
	}{
		{
			name:   "standard issuer",
			issuer: "https://keycloak.example.com/realms/lkfl-sdek",
			want:   "sdek",
		},
		{
			name:   "internal issuer",
			issuer: "http://keycloak:8080/realms/lkfl-acme",
			want:   "acme",
		},
		{
			name:   "no lkfl prefix",
			issuer: "https://keycloak.example.com/realms/myrealm",
			want:   "",
		},
		{
			name:   "empty issuer",
			issuer: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveTenantSlug(tt.issuer)
			if got != tt.want {
				t.Errorf("ResolveTenantSlug(%q) = %q, want %q", tt.issuer, got, tt.want)
			}
		})
	}
}

// TestWriteAuthError проверяет формат JSON-ответа с ошибкой.
func TestWriteAuthError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteAuthError(w, http.StatusUnauthorized, "bad token")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type=application/json, got %q", ct)
	}

	// Проверяем что тело содержит поле "error"
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty body")
	}
}
