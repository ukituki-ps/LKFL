package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// =============================================================================
// Mock Redis Client для unit-тестов
// =============================================================================

// mockRedisClient — мок Redis клиента.
type mockRedisClient struct {
	mu   sync.RWMutex
	data map[string]string
}

func newMockRedis() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string]string),
	}
}

// setErr — возвращает ошибку Set (nil по умолчанию).
var mockSetErr error = nil

func (m *mockRedisClient) Set(_ context.Context, key string, value interface{}, _ time.Duration) *redis.StatusCmd {
	if mockSetErr != nil {
		return redis.NewStatusResult("", mockSetErr)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	switch v := value.(type) {
	case string:
		m.data[key] = v
	case []byte:
		m.data[key] = string(v)
	default:
		m.data[key] = ""
	}
	return redis.NewStatusResult("OK", nil)
}

func (m *mockRedisClient) Get(_ context.Context, key string) *redis.StringCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[key]
	if !exists {
		return redis.NewStringResult("", redis.Nil)
	}
	cmd := redis.NewStringResult(value, nil)
	return cmd
}

func (m *mockRedisClient) Del(_ context.Context, keys ...string) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			delete(m.data, key)
			count++
		}
	}
	return redis.NewIntResult(int64(count), nil)
}

func (m *mockRedisClient) Scan(_ context.Context, _ uint64, match string, _ int64) *redis.ScanCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var results []string
	prefix := match[:len(match)-1] // remove trailing *
	for key := range m.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			results = append(results, key)
		}
	}
	return redis.NewScanCmdResult(results, 0, nil)
}

// =============================================================================
// SessionStore Tests
// =============================================================================

func TestSessionStore_Create(t *testing.T) {
	m := newMockRedis()
	store := NewSessionStore(m, 24*time.Hour)

	data := SessionData{
		UserID:    "kc-sub-123",
		TenantID:  "sdek",
		CreatedAt: time.Now(),
	}

	token, err := store.Create(context.Background(), data)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}

	if len(token) != 64 {
		t.Errorf("expected token length 64 (32 bytes hex), got %d", len(token))
	}

	// Check Redis key exists
	key := sessionKey(token)
	m.mu.RLock()
	_, exists := m.data[key]
	m.mu.RUnlock()
	if !exists {
		t.Error("expected session data in Redis")
	}
}

func TestSessionStore_Get(t *testing.T) {
	m := newMockRedis()
	store := NewSessionStore(m, 24*time.Hour)

	data := SessionData{
		UserID:    "kc-sub-123",
		TenantID:  "sdek",
		CreatedAt: time.Now(),
	}

	token, err := store.Create(context.Background(), data)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	result, err := store.Get(context.Background(), token)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}

	if result.UserID != "kc-sub-123" {
		t.Errorf("expected UserID 'kc-sub-123', got '%s'", result.UserID)
	}
	if result.TenantID != "sdek" {
		t.Errorf("expected TenantID 'sdek', got '%s'", result.TenantID)
	}
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	m := newMockRedis()
	store := NewSessionStore(m, 24*time.Hour)

	_, err := store.Get(context.Background(), "nonexistent-token")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSessionStore_Delete(t *testing.T) {
	m := newMockRedis()
	store := NewSessionStore(m, 24*time.Hour)

	data := SessionData{UserID: "kc-sub-123", TenantID: "sdek", CreatedAt: time.Now()}
	token, err := store.Create(context.Background(), data)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	err = store.Delete(context.Background(), token)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err = store.Get(context.Background(), token)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestSessionStore_DeleteByUserID(t *testing.T) {
	m := newMockRedis()
	store := NewSessionStore(m, 24*time.Hour)

	token1, _ := store.Create(context.Background(), SessionData{
		UserID: "kc-sub-123", TenantID: "sdek", CreatedAt: time.Now(),
	})
	token2, _ := store.Create(context.Background(), SessionData{
		UserID: "kc-sub-123", TenantID: "sdek", CreatedAt: time.Now(),
	})
	token3, _ := store.Create(context.Background(), SessionData{
		UserID: "kc-sub-456", TenantID: "sdek", CreatedAt: time.Now(),
	})

	err := store.DeleteByUserID(context.Background(), "kc-sub-123")
	if err != nil {
		t.Fatalf("DeleteByUserID error: %v", err)
	}

	_, err = store.Get(context.Background(), token1)
	if err == nil {
		t.Error("expected session 1 to be deleted")
	}
	_, err = store.Get(context.Background(), token2)
	if err == nil {
		t.Error("expected session 2 to be deleted")
	}

	_, err = store.Get(context.Background(), token3)
	if err != nil {
		t.Error("expected session 3 to still exist")
	}
}

func TestSessionStore_Uniqueness(t *testing.T) {
	m := newMockRedis()
	store := NewSessionStore(m, 24*time.Hour)

	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := store.Create(context.Background(), SessionData{
			UserID: "kc-sub-test", TenantID: "sdek", CreatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("Create error at iteration %d: %v", i, err)
		}
		if tokens[token] {
			t.Fatalf("duplicate token generated: %s", token)
		}
		tokens[token] = true
	}
}

// =============================================================================
// TokenStore Tests
// =============================================================================

func TestTokenStore_SaveAndGet(t *testing.T) {
	m := newMockRedis()
	store := NewTokenStore(m)

	err := store.SaveTokens(context.Background(), "kc-sub-123", "access-token-value", "refresh-token-value")
	if err != nil {
		t.Fatalf("SaveTokens error: %v", err)
	}

	accessToken, err := store.GetAccessToken(context.Background(), "kc-sub-123")
	if err != nil {
		t.Fatalf("GetAccessToken error: %v", err)
	}
	if accessToken != "access-token-value" {
		t.Errorf("expected 'access-token-value', got '%s'", accessToken)
	}

	refreshToken, err := store.GetRefreshToken(context.Background(), "kc-sub-123")
	if err != nil {
		t.Fatalf("GetRefreshToken error: %v", err)
	}
	if refreshToken != "refresh-token-value" {
		t.Errorf("expected 'refresh-token-value', got '%s'", refreshToken)
	}
}

func TestTokenStore_Delete(t *testing.T) {
	m := newMockRedis()
	store := NewTokenStore(m)

	_ = store.SaveTokens(context.Background(), "kc-sub-123", "access", "refresh")

	err := store.Delete(context.Background(), "kc-sub-123")
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	_, err = store.GetAccessToken(context.Background(), "kc-sub-123")
	if err == nil {
		t.Error("expected error after delete")
	}

	_, err = store.GetRefreshToken(context.Background(), "kc-sub-123")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestTokenStore_NotFound(t *testing.T) {
	m := newMockRedis()
	store := NewTokenStore(m)

	_, err := store.GetAccessToken(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}

	_, err = store.GetRefreshToken(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

func TestTokenStore_UpdateAccessToken(t *testing.T) {
	m := newMockRedis()
	store := NewTokenStore(m)

	_ = store.SaveTokens(context.Background(), "kc-sub-123", "old-access", "old-refresh")

	err := store.UpdateAccessToken(context.Background(), "kc-sub-123", "new-access")
	if err != nil {
		t.Fatalf("UpdateAccessToken error: %v", err)
	}

	accessToken, _ := store.GetAccessToken(context.Background(), "kc-sub-123")
	if accessToken != "new-access" {
		t.Errorf("expected 'new-access', got '%s'", accessToken)
	}

	refreshToken, _ := store.GetRefreshToken(context.Background(), "kc-sub-123")
	if refreshToken != "old-refresh" {
		t.Errorf("expected 'old-refresh', got '%s'", refreshToken)
	}
}

// =============================================================================
// TokenRefresher Tests
// =============================================================================

func mockTokenServer(refreshTokenValid bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if refreshTokenValid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"access_token": "new-access-token-abc123",
				"refresh_token": "new-refresh-token-xyz789",
				"expires_in": 900,
				"token_type": "Bearer"
			}`))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Refresh token expired"}`))
		}
	}))
}

func TestTokenRefresher_Success(t *testing.T) {
	ts := mockTokenServer(true)
	defer ts.Close()

	m := newMockRedis()
	tokenStore := NewTokenStore(m)
	refresher := NewTokenRefresher(ts.URL, "lkfl-spa", "", tokenStore)

	_ = tokenStore.SaveTokens(context.Background(), "kc-sub-test", "old-access", "old-refresh")

	accessToken, refreshToken, err := refresher.Refresh(context.Background(), "kc-sub-test")
	if err != nil {
		t.Fatalf("Refresh error: %v", err)
	}

	if accessToken != "new-access-token-abc123" {
		t.Errorf("expected new access token, got '%s'", accessToken)
	}
	if refreshToken != "new-refresh-token-xyz789" {
		t.Errorf("expected new refresh token, got '%s'", refreshToken)
	}
}

func TestTokenRefresher_ExpiredRefreshToken(t *testing.T) {
	ts := mockTokenServer(false)
	defer ts.Close()

	m := newMockRedis()
	tokenStore := NewTokenStore(m)
	refresher := NewTokenRefresher(ts.URL, "lkfl-spa", "", tokenStore)

	_ = tokenStore.SaveTokens(context.Background(), "kc-sub-test", "old-access", "expired-refresh")

	_, _, err := refresher.Refresh(context.Background(), "kc-sub-test")
	if err == nil {
		t.Error("expected error for expired refresh token")
	}
}

func TestTokenRefresher_NetworkError(t *testing.T) {
	m := newMockRedis()
	tokenStore := NewTokenStore(m)
	refresher := NewTokenRefresher("http://localhost:59999/realms/test", "lkfl-spa", "", tokenStore)

	_ = tokenStore.SaveTokens(context.Background(), "kc-sub-test", "access", "refresh")

	_, _, err := refresher.Refresh(context.Background(), "kc-sub-test")
	if err == nil {
		t.Error("expected network error")
	}
}

func TestTokenRefresher_MissingRefreshToken(t *testing.T) {
	ts := mockTokenServer(true)
	defer ts.Close()

	m := newMockRedis()
	tokenStore := NewTokenStore(m)
	refresher := NewTokenRefresher(ts.URL, "lkfl-spa", "", tokenStore)

	_, _, err := refresher.Refresh(context.Background(), "kc-sub-nonesuch")
	if err == nil {
		t.Error("expected error when no refresh token exists")
	}
}

// =============================================================================
// JWT decode helpers tests
// =============================================================================

func TestDecodeAccessToken_ValidPayload(t *testing.T) {
	payload := `{"sub":"kc-sub-123","email":"test@example.com","exp":9999999999,"realm_access":{"roles":["admin","employee"]}}`
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	testToken := "eyJhbGciOiJSUzI1NiJ9." + encodedPayload + ".fake-signature"

	rawClaims, roles, err := decodeAccessToken(testToken)
	if err != nil {
		t.Fatalf("decodeAccessToken error: %v", err)
	}

	if sub, ok := rawClaims["sub"].(string); !ok || sub != "kc-sub-123" {
		t.Errorf("expected sub 'kc-sub-123', got '%v'", rawClaims["sub"])
	}

	if len(roles) == 0 {
		t.Error("expected roles from access token")
	}
}

func TestDecodeAccessToken_InvalidFormat(t *testing.T) {
	_, _, err := decodeAccessToken("not-a-jwt")
	if err == nil {
		t.Error("expected error for invalid token format")
	}
}

func TestIsTokenExpired_NotExpired(t *testing.T) {
	claims := map[string]interface{}{
		"exp": float64(9999999999),
	}
	if isTokenExpired(claims) {
		t.Error("expected token to not be expired")
	}
}

func TestIsTokenExpired_Expired(t *testing.T) {
	claims := map[string]interface{}{
		"exp": float64(0),
	}
	if !isTokenExpired(claims) {
		t.Error("expected token to be expired")
	}
}

func TestIsTokenExpired_NoExpClaim(t *testing.T) {
	claims := map[string]interface{}{
		"sub": "user-123",
	}
	if isTokenExpired(claims) {
		t.Error("expected token without exp to not be considered expired")
	}
}

// =============================================================================
// ExtractSessionCookie Tests
// =============================================================================

func TestExtractSessionCookie_Present(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token-123"})

	result := ExtractSessionCookie(req)
	if result != "session-token-123" {
		t.Errorf("expected 'session-token-123', got '%s'", result)
	}
}

func TestExtractSessionCookie_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	result := ExtractSessionCookie(req)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestExtractSessionCookie_EmptyValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ""})

	result := ExtractSessionCookie(req)
	if result != "" {
		t.Errorf("expected empty string for empty cookie, got '%s'", result)
	}
}

// =============================================================================
// BuildClaims Tests
// =============================================================================

func TestBuildClaims(t *testing.T) {
	rawClaims := map[string]interface{}{
		"sub":                "kc-sub-123",
		"email":              "test@example.com",
		"preferred_username": "testuser",
		"name":               "Test User",
		"given_name":         "Test",
		"family_name":        "User",
		"iss":                "https://keycloak.example.com/realms/lkfl-sdek",
	}

	sessionData := SessionData{
		UserID:   "kc-sub-123",
		TenantID: "sdek",
	}

	claims := buildClaims(rawClaims, sessionData)

	if claims.Subject != "kc-sub-123" {
		t.Errorf("expected Subject 'kc-sub-123', got '%s'", claims.Subject)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.TenantID != "sdek" {
		t.Errorf("expected TenantID 'sdek', got '%s'", claims.TenantID)
	}
	if claims.PreferredUsername != "testuser" {
		t.Errorf("expected PreferredUsername 'testuser', got '%s'", claims.PreferredUsername)
	}
}

// =============================================================================
// SessionData JSON roundtrip
// =============================================================================

func TestSessionData_JSONRoundtrip(t *testing.T) {
	data := SessionData{
		UserID:    "kc-sub-test",
		TenantID:  "acme",
		CreatedAt: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
	}

	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded SessionData
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.UserID != data.UserID {
		t.Errorf("UserID mismatch: %q vs %q", decoded.UserID, data.UserID)
	}
	if decoded.TenantID != data.TenantID {
		t.Errorf("TenantID mismatch: %q vs %q", decoded.TenantID, data.TenantID)
	}
}


