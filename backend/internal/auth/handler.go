package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/redis/go-redis/v9"
	"lkfl/internal/metrics"
	sharedauth "lkfl/shared/pkg/auth"
	shhttp "lkfl/shared/pkg/http"
)

// Handler — HTTP handlers для аутентификации.
// Все URL используют host.docker.internal — единый адрес для браузера и бэкенда.
type Handler struct {
	verifier     *oidc.IDTokenVerifier
	redis        *redis.Client
	service      *Service
	issuer       string           // OIDC issuer (host.docker.internal) — для token verification
	publicIssuer string           // Public issuer — для browser-редиректов
	clientID     string
	clientSecret string
	tenantSlug   string           // Tenant slug из issuer URL (lkfl-sdek → sdek)
	metrics      *metrics.Metrics
}

// NewHandler создаёт Handler.
// Если m == nil, метрики не собираются.
// publicIssuer — URL видимый из браузера; если пуст, используется issuer.
func NewHandler(verifier *oidc.IDTokenVerifier, redis *redis.Client, service *Service, issuer, publicIssuer, clientID, clientSecret string, m *metrics.Metrics) *Handler {
	if publicIssuer == "" {
		publicIssuer = issuer
	}
	// Извлекаем tenant slug из issuer URL: .../realms/lkfl-sdek → sdek
	// Формат realm: lkfl-{tenant_slug}
	tenantSlug := ""
	parts := strings.Split(issuer, "/")
	for _, p := range parts {
		if strings.HasPrefix(p, "lkfl-") {
			tenantSlug = strings.TrimPrefix(p, "lkfl-")
			break
		}
	}
	return &Handler{
		verifier:     verifier,
		redis:        redis,
		service:      service,
		issuer:       issuer,
		publicIssuer: publicIssuer,
		clientID:     clientID,
		clientSecret: clientSecret,
		tenantSlug:   tenantSlug,
		metrics:      m,
	}
}

// generateState генерирует криптографически безопасный state-параметр (32 байта → 64 hex).
func generateState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// generatePKCEVerifier генерирует PKCE code_verifier (43-128 символов base64url).
func generatePKCEVerifier() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// pkceCodeChallenge вычисляет code_challenge из verifier (S256).
func pkceCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// LoginRedirect — редирект на Keycloak login page.
// GET /api/v1/auth/login
// Response: 302 → Keycloak authorize endpoint (с PKCE)
func (h *Handler) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	if h.metrics != nil {
		h.metrics.AuthLoginTotal.WithLabelValues("success").Inc()
	}
	// Генерируем state-параметр (CSRF protection)
	state := generateState()

	// Генерируем PKCE code_verifier (S256)
	verifier := generatePKCEVerifier()
	challenge := pkceCodeChallenge(verifier)

	// Сохраняем state + code_verifier в Redis (TTL 10 мин)
	stateKey := fmt.Sprintf("auth:state:%s", state)
	if err := h.redis.Set(r.Context(), stateKey, time.Now().Format(time.RFC3339), 10*time.Minute).Err(); err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "failed to generate login state")
		return
	}
	verifierKey := fmt.Sprintf("auth:pkce:%s", state)
	if err := h.redis.Set(r.Context(), verifierKey, verifier, 10*time.Minute).Err(); err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "failed to store PKCE verifier")
		return
	}

	// Собираем URL авторизации Keycloak
	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		redirect = "https://dev.april.ukituki.tech/callback"
	}

	// При retry=1 форсируем перелогин в Keycloak (prompt=login).
	// Используется когда callback получил 410 Gone — старая сессия истекла.
	prompt := ""
	if r.URL.Query().Get("retry") == "1" {
		prompt = "&prompt=login"
	}

	authorizeURL := fmt.Sprintf(
		"%s/protocol/openid-connect/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid+profile+email&state=%s&code_challenge=%s&code_challenge_method=S256%s",
		h.publicIssuer,
		h.clientID,
		url.QueryEscape(redirect),
		state,
		url.QueryEscape(challenge),
		prompt,
	)

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

// LoginCallback — Keycloak callback с authorization code.
// GET /api/v1/auth/callback?code=...&state=...
//
// Для браузерных запросов (Accept содержит text/html):
//   → 302 redirect на frontend /callback?code=...&state=...
//
// Для API запросов (Accept: application/json):
//   → обмен code на token, 200 с user data + session token
func (h *Handler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	// Браузерный запрос — редирект на фронтенд /callback с code и state
	if isBrowserRequest(r) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		frontendURL := fmt.Sprintf("/callback?code=%s&state=%s", url.QueryEscape(code), url.QueryEscape(state))
		http.Redirect(w, r, frontendURL, http.StatusFound)
		return
	}

	// API запрос от фронтенда — обмен code на token
	// 1. Проверяем state-параметр (CSRF protection)
	state := r.URL.Query().Get("state")
	stateKey := fmt.Sprintf("auth:state:%s", state)

	_, err := h.redis.Get(r.Context(), stateKey).Result()
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthCallbackTotal.WithLabelValues("failure").Inc()
		}
		// State истёк или не найден — возможно контейнер перезапускался.
		// Браузерный запрос (fallback) → 302 на /login с ошибкой.
		// API-запрос от фронтенда → 410 Gone (session expired).
		if isBrowserRequest(r) {
			http.Redirect(w, r, "/login?error=expired_session", http.StatusFound)
			return
		}
		shhttp.WriteJSONError(w, http.StatusGone, "session expired, please login again")
		return
	}
	// Удаляем state (одноразовое использование)
	h.redis.Del(r.Context(), stateKey)

	// 2. Получаем PKCE code_verifier из Redis
	verifierKey := fmt.Sprintf("auth:pkce:%s", state)
	verifierStr, err := h.redis.Get(r.Context(), verifierKey).Result()
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "PKCE verifier not found")
		return
	}
	h.redis.Del(r.Context(), verifierKey)

	// 3. Получаем authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		if h.metrics != nil {
			h.metrics.AuthCallbackTotal.WithLabelValues("error").Inc()
		}
		shhttp.WriteJSONError(w, http.StatusBadRequest, "no authorization code provided")
		return
	}

	// 4. Exchange authorization code for tokens via Keycloak token endpoint (с PKCE)
	tokenEndpoint := h.issuer + "/protocol/openid-connect/token"
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", h.clientID)
	form.Set("code", code)
	form.Set("redirect_uri", "https://dev.april.ukituki.tech/callback")
	form.Set("code_verifier", verifierStr)
	if h.clientSecret != "" {
		form.Set("client_secret", h.clientSecret)
	}

	resp, err := http.PostForm(tokenEndpoint, form)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadGateway, "token exchange failed: "+err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadGateway, "read token response failed")
		return
	}

	if resp.StatusCode != http.StatusOK {
		shhttp.WriteJSONError(w, resp.StatusCode, string(body))
		return
	}

	var tokenSet struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenSet); err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "parse token response: "+err.Error())
		return
	}

	// 5. Верифицируем ID token
	idToken, err := h.verifier.Verify(r.Context(), tokenSet.IDToken)
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthCallbackTotal.WithLabelValues("failure").Inc()
		}
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "invalid id_token: "+err.Error())
		return
	}

	// 6. Извлекаем claims и роли из ID token
	claims, roles, err := sharedauth.ExtractClaims(idToken)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "claim parsing error")
		return
	}

	// Если ролей нет в ID token, берём из access token (realm roles только в access token)
	if len(roles) == 0 {
		roles = sharedauth.ExtractRolesFromJWT(tokenSet.AccessToken)
	}

	

	// 7. Создаём/обновляем пользователя в БД
	user, err := h.service.CreateOrUpdateUser(r.Context(), claims, roles)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
		return
	}

	// 8. Устанавливаем сессию в Redis (TTL 24 часа)
	sessionKey := fmt.Sprintf("auth:session:%s", user.ID.String())
	h.redis.Set(r.Context(), sessionKey, tokenSet.IDToken, 24*time.Hour)

	// 9. Возвращаем данные пользователя + session token
	if h.metrics != nil {
		h.metrics.AuthCallbackTotal.WithLabelValues("success").Inc()
	}
	shhttp.WriteJSON(w, http.StatusOK, map[string]any{
		"user":  user.ToProfile(),
		"roles": roles,
		// Возвращаем ID token — middleware JWTMiddleware верифицирует
		// токены через oidc.IDTokenVerifier, который принимает только ID токены.
		// Access token не пройдёт верификацию → 401.
		"token": tokenSet.IDToken,
	})
}

// isBrowserRequest проверяет, является ли запрос от браузера (ожидает HTML).
func isBrowserRequest(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return true // браузер по умолчанию не шлёт Accept
	}
	for _, part := range strings.Split(accept, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "text/html") {
			return true
		}
	}
	return false
}

// Logout — инвалидация сессии.
// POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Удаляем сессию из Redis
	userID := sharedauth.UserIDFromContext(r.Context())
	if userID != "" {
		sessionKey := fmt.Sprintf("auth:session:%s", userID)
		h.redis.Del(r.Context(), sessionKey)
	}

	// Редирект на Keycloak logout
	redirect := r.URL.Query().Get("post_logout_redirect_uri")
	if redirect == "" {
		redirect = "http://localhost:5173/"
	}

	logoutURL := fmt.Sprintf(
		"%s/protocol/openid-connect/logout?id_token_hint=%s&post_logout_redirect_uri=%s",
		h.publicIssuer,
		url.QueryEscape(r.URL.Query().Get("id_token_hint")),
		url.QueryEscape(redirect),
	)

	http.Redirect(w, r, logoutURL, http.StatusFound)
}

// Me — текущий пользователь.
// GET /api/v1/auth/me
// Требует JWT middleware (токен в Authorization: Bearer <token>)
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	keycloakSub := sharedauth.UserIDFromContext(r.Context())
	if keycloakSub == "" {
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.service.GetUserByKeycloakSub(r.Context(), keycloakSub)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, user.ToProfile())
}
