package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// --- HostResolver Tests ---

func TestHostResolver_Subdomain(t *testing.T) {
	tenant := Tenant{
		ID:     uuid.New(),
		Slug:   "sdek",
		Name:   "СДЭК",
		Status: "active",
	}

	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			if slug == "sdek" {
				return tenant, nil
			}
			return Tenant{}, ErrNotFound
		},
	}

	svc := newTestService(repo)
	resolver := NewHostResolver(svc, nil, nil) // без Redis

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "sdek.example.com"

	result, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "sdek" {
		t.Errorf("expected slug 'sdek', got '%s'", result.Slug)
	}
}

func TestHostResolver_FallbackHeader(t *testing.T) {
	tenant := Tenant{
		ID:     uuid.New(),
		Slug:   "mytenant",
		Name:   "My Tenant",
		Status: "active",
	}

	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			if slug == "mytenant" {
				return tenant, nil
			}
			return Tenant{}, ErrNotFound
		},
	}

	svc := newTestService(repo)
	resolver := NewHostResolver(svc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "www.example.com"
	req.Header.Set("X-Tenant-ID", "mytenant")

	result, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "mytenant" {
		t.Errorf("expected slug 'mytenant', got '%s'", result.Slug)
	}
}

func TestHostResolver_LocalhostFallbackHeader(t *testing.T) {
	tenant := Tenant{
		ID:     uuid.New(),
		Slug:   "dev",
		Name:   "Dev",
		Status: "active",
	}

	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			if slug == "dev" {
				return tenant, nil
			}
			return Tenant{}, ErrNotFound
		},
	}

	svc := newTestService(repo)
	resolver := NewHostResolver(svc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "localhost:8080"
	req.Header.Set("X-Tenant-ID", "dev")

	result, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "dev" {
		t.Errorf("expected slug 'dev', got '%s'", result.Slug)
	}
}

func TestHostResolver_NoHeaderOnLocalhost(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)
	resolver := NewHostResolver(svc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "localhost:8080"
	// Без X-Tenant-ID header

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error when no tenant header on localhost")
	}
	if !strings.Contains(err.Error(), "X-Tenant-ID") {
		t.Errorf("expected X-Tenant-ID error, got: %v", err)
	}
}

func TestHostResolver_TenantNotFound(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)
	resolver := NewHostResolver(svc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "nonexistent.example.com"

	_, err := resolver.Resolve(req)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

// --- PathResolver Tests ---

func TestPathResolver_ValidPath(t *testing.T) {
	tenant := Tenant{
		ID:     uuid.New(),
		Slug:   "sdek",
		Name:   "СДЭК",
		Status: "active",
	}

	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			if slug == "sdek" {
				return tenant, nil
			}
			return Tenant{}, ErrNotFound
		},
	}

	svc := newTestService(repo)
	resolver := NewPathResolver(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/t/sdek/api/v1/benefits", nil)

	result, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "sdek" {
		t.Errorf("expected slug 'sdek', got '%s'", result.Slug)
	}
}

func TestPathResolver_NoPrefix(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)
	resolver := NewPathResolver(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/benefits", nil)

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error when path doesn't start with /t/")
	}
}

func TestPathResolver_EmptySlug(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)
	resolver := NewPathResolver(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/t/", nil)

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error when slug is empty")
	}
}

// --- Middleware Tests ---

func TestMiddleware_Success(t *testing.T) {
	tenant := Tenant{
		ID:     uuid.New(),
		Slug:   "sdek",
		Name:   "СДЭК",
		Status: "active",
	}

	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return tenant, nil
		},
	}

	svc := newTestService(repo)
	resolver := NewHostResolver(svc, nil, nil)

	var capturedID uuid.UUID
	var capturedSlug string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = TenantIDFromContext(r.Context())
		capturedSlug = TenantSlugFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mw := Middleware(resolver)
	wrapped := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "sdek.example.com"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if capturedID != tenant.ID {
		t.Errorf("expected tenant ID %s, got %s", tenant.ID, capturedID)
	}
	if capturedSlug != "sdek" {
		t.Errorf("expected slug 'sdek', got '%s'", capturedSlug)
	}
}

func TestMiddleware_NotFound(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)
	resolver := NewHostResolver(svc, nil, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	mw := Middleware(resolver)
	wrapped := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "nonexistent.example.com"
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// --- Context Helper Tests ---

func TestTenantIDFromContext(t *testing.T) {
	tenantID := uuid.New()
	ctx := TenantContext(context.Background(), tenantID)

	result := TenantIDFromContext(ctx)
	if result != tenantID {
		t.Errorf("expected %s, got %s", tenantID, result)
	}
}

func TestTenantIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	result := TenantIDFromContext(ctx)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil, got %s", result)
	}
}

func TestTenantSlugFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), TenantSlugKey, "sdek")
	result := TenantSlugFromContext(ctx)
	if result != "sdek" {
		t.Errorf("expected 'sdek', got '%s'", result)
	}
}

func TestTenantSlugFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	result := TenantSlugFromContext(ctx)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

// --- SkipPaths Tests ---

func TestSkipPaths_SkipsHealthz(t *testing.T) {
	skipped := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skipped = true
		w.WriteHeader(http.StatusOK)
	})

	mw := SkipPaths([]string{"/healthz", "/metrics", "/admin/"})
	wrapped := mw(handler)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if !skipped {
		t.Error("handler should have been called for /healthz")
	}
}

// --- Redis Cache Tests ---

func TestHostResolver_WithRedisCache(t *testing.T) {
	// Создаём mock Redis через Redis stub
	// Для этого теста используем реальную Redis если доступна, иначе пропускаем
	// В CI/CD без Redis — тест пропускается

	tenant := Tenant{
		ID:     uuid.New(),
		Slug:   "cached",
		Name:   "Cached",
		Status: "active",
	}

	// Создаём Redis client с пустым URL — будет ошибка при подключении
	// Поэтому для unit теста просто проверяем логику без реального Redis
	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return tenant, nil
		},
	}

	svc := newTestService(repo)

	// Redis client в disconnected mode
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // несуществующий порт
	})

	resolver := NewHostResolver(svc, rdb, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Host = "cached.example.com"

	result, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "cached" {
		t.Errorf("expected slug 'cached', got '%s'", result.Slug)
	}

	// Redis кэш не работает без сервера, но resolver должен
	// успешно вернуть tenant из fallback (DB через repo)
}

func TestHostResolver_CacheSerialization(t *testing.T) {
	tenant := Tenant{
		ID:       uuid.New(),
		Slug:     "test",
		Name:     "Test",
		Status:   "active",
		Settings: JSONB{"key": "value"},
	}

	// Проверяем что tenant можно сериализовать/десериализовать для Redis
	data, err := json.Marshal(tenant)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Tenant
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.ID != tenant.ID {
		t.Errorf("ID mismatch after roundtrip")
	}
	if decoded.Slug != tenant.Slug {
		t.Errorf("Slug mismatch after roundtrip")
	}
}
