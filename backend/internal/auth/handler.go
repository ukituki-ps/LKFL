package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/redis/go-redis/v9"
	"lkfl/internal/metrics"
	sharedauth "lkfl/shared/pkg/auth"
	shhttp "lkfl/shared/pkg/http"
)

const protoHTTPS = "https"

// sessionCookieName — имя httpOnly cookie для хранения сессионного токена.
const sessionCookieName = "lkfl_session"

// Handler — HTTP handlers для аутентификации.
type Handler struct {
	verifier       *oidc.IDTokenVerifier
	redis          *redis.Client
	service        *Service
	issuer         string
	publicIssuer   string
	clientID       string
	clientSecret   string
	tenantSlug     string
	metrics        *metrics.Metrics
	sessionStore   *sharedauth.SessionStore
	tokenStore     *sharedauth.TokenStore
	tokenRefresher *sharedauth.TokenRefresher
}

// NewHandler создаёт Handler.
func NewHandler(verifier *oidc.IDTokenVerifier, redis *redis.Client, service *Service, issuer, publicIssuer, clientID, clientSecret string, m *metrics.Metrics) *Handler {
	if publicIssuer == "" {
		publicIssuer = issuer
	}

	tenantSlug := ""
	parts := strings.Split(issuer, "/")
	for _, p := range parts {
		if strings.HasPrefix(p, "lkfl-") {
			tenantSlug = strings.TrimPrefix(p, "lkfl-")
			break
		}
	}

	sessionStore := sharedauth.NewSessionStore(redis, 24*time.Hour)
	tokenStore := sharedauth.NewTokenStore(redis)
	tokenRefresher := sharedauth.NewTokenRefresher(issuer, clientID, clientSecret, tokenStore)

	return &Handler{
		verifier:       verifier,
		redis:          redis,
		service:        service,
		issuer:         issuer,
		publicIssuer:   publicIssuer,
		clientID:       clientID,
		clientSecret:   clientSecret,
		tenantSlug:     tenantSlug,
		metrics:        m,
		sessionStore:   sessionStore,
		tokenStore:     tokenStore,
		tokenRefresher: tokenRefresher,
	}
}

// LogStartupWarnings логирует предупреждения о конфигурации при старте сервера.
// Вызывается один раз после инициализации Handler.
func LogStartupWarnings(clientSecret, cookieDomain string) {
	// 1. Keycloak Admin REST API
	if clientSecret == "" {
		slog.Warn("keycloak admin REST API unavailable — client_secret not set",
			"impact", "SSO logout will use fallback (POST logout endpoint)",
			"fix", "set KEYCLOAK_CLIENT_SECRET env var",
		)
	} else {
		slog.Info("keycloak admin REST API available — SSO invalidation will use Admin API")
	}

	// 2. Cookie domain
	if cookieDomain == "" {
		slog.Info("cookie domain not set — SameSite=Lax mode (development)",
			"fix", "set COOKIE_DOMAIN for production (e.g. .lkfl.ru)",
		)
	} else {
		slog.Info("cookie domain configured — SameSite=None mode (production)",
			"domain", cookieDomain,
		)
	}
}

// SessionMiddleware возвращает SessionMiddleware, использующий те же
// SessionStore, TokenStore и TokenRefresher, что и этот Handler.
// Используется для настройки роутера (server.go).
func (h *Handler) SessionMiddleware() func(http.Handler) http.Handler {
	return sharedauth.SessionMiddleware(h.sessionStore, h.tokenStore, h.tokenRefresher)
}

// generateState генерирует криптографически безопасный state-параметр.
func generateState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// generatePKCEVerifier генерирует PKCE code_verifier.
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
func (h *Handler) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	if h.metrics != nil {
		h.metrics.AuthLoginTotal.WithLabelValues("success").Inc()
	}

	state := generateState()
	verifier := generatePKCEVerifier()
	challenge := pkceCodeChallenge(verifier)

	redirect := r.URL.Query().Get("redirect")
	if redirect == "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == protoHTTPS {
			scheme = protoHTTPS
		}
		redirect = fmt.Sprintf("%s://%s/api/v1/auth/callback", scheme, r.Host)
	}

	// Сохраняем state + code_verifier + redirect_uri в Redis (TTL 10 мин)
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
	redirectKey := fmt.Sprintf("auth:redirect:%s", state)
	if err := h.redis.Set(r.Context(), redirectKey, redirect, 10*time.Minute).Err(); err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "failed to store redirect URI")
		return
	}

	prompt := ""
	if r.URL.Query().Get("retry") == "1" {
		prompt = "&prompt=login"
	}

	authorizeURL := fmt.Sprintf(
		"%s/protocol/openid-connect/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid+profile+email&state=%s&code_challenge=%s&code_challenge_method=S256%s",
		h.issuer,
		h.clientID,
		url.QueryEscape(redirect),
		state,
		url.QueryEscape(challenge),
		prompt,
	)

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

// LoginCallback — Keycloak callback с authorization code.
func (h *Handler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	if isBrowserRequest(r) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		frontendOrigin := r.Header.Get("Origin")
		if frontendOrigin == "" {
			ref := r.Header.Get("Referer")
			if ref != "" {
				if u, err := url.Parse(ref); err == nil {
					frontendOrigin = u.Scheme + "://" + u.Host
				}
			}
		}
		if frontendOrigin == "" {
			frontendOrigin = "http://localhost:5173"
		}

		frontendURL := fmt.Sprintf("%s/callback?code=%s&state=%s", frontendOrigin, url.QueryEscape(code), url.QueryEscape(state))
		http.Redirect(w, r, frontendURL, http.StatusFound)
		return
	}

	// API запрос от фронтенда — обмен code на token
	state := r.URL.Query().Get("state")
	stateKey := fmt.Sprintf("auth:state:%s", state)

	_, err := h.redis.Get(r.Context(), stateKey).Result()
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthCallbackTotal.WithLabelValues("failure").Inc()
		}
		if isBrowserRequest(r) {
			http.Redirect(w, r, "/login?error=expired_session", http.StatusFound)
			return
		}
		shhttp.WriteJSONError(w, http.StatusGone, "session expired, please login again")
		return
	}
	h.redis.Del(r.Context(), stateKey)

	// PKCE code_verifier
	verifierKey := fmt.Sprintf("auth:pkce:%s", state)
	verifierStr, err := h.redis.Get(r.Context(), verifierKey).Result()
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "PKCE verifier not found")
		return
	}
	h.redis.Del(r.Context(), verifierKey)

	// redirect_uri
	redirectKey := fmt.Sprintf("auth:redirect:%s", state)
	savedRedirect, err := h.redis.Get(r.Context(), redirectKey).Result()
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusBadRequest, "redirect URI not found")
		return
	}
	h.redis.Del(r.Context(), redirectKey)

	// authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		if h.metrics != nil {
			h.metrics.AuthCallbackTotal.WithLabelValues("error").Inc()
		}
		shhttp.WriteJSONError(w, http.StatusBadRequest, "no authorization code provided")
		return
	}

	// Exchange authorization code for tokens
	tokenEndpoint := h.issuer + "/protocol/openid-connect/token"
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", h.clientID)
	form.Set("code", code)
	form.Set("redirect_uri", savedRedirect)
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
		slog.Error("token_exchange_failed", "status", resp.StatusCode, "body", string(body))
		shhttp.WriteJSONError(w, resp.StatusCode, string(body))
		return
	}

	var tokenSet struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &tokenSet); err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "parse token response: "+err.Error())
		return
	}

	// Верифицируем ID token (нужен verifier.go для первичной верификации)
	idToken, err := h.verifier.Verify(r.Context(), tokenSet.IDToken)
	if err != nil {
		if h.metrics != nil {
			h.metrics.AuthCallbackTotal.WithLabelValues("failure").Inc()
		}
		slog.Error("id_token_verify_failed", "error", err.Error())
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "invalid id_token: "+err.Error())
		return
	}

	// Извлекаем claims и роли из ID token
	claims, roles, err := sharedauth.ExtractClaims(idToken)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "claim parsing error")
		return
	}

	if len(roles) == 0 {
		roles = extractRolesFromAccessToken(tokenSet.AccessToken)
	}

	// Создаём/обновляем пользователя в БД
	user, err := h.service.CreateOrUpdateUser(r.Context(), claims, roles)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
		return
	}

	// Извлекаем tenant_id из issuer
	tenantID := sharedauth.ResolveTenantSlug(claims.Issuer)

	// Создаём server-side session (сохраняем email для Keycloak Admin API lookup)
	sessionToken, err := h.sessionStore.Create(r.Context(), sharedauth.SessionData{
		UserID:    claims.Subject,
		Email:     claims.Email,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Сохраняем Keycloak токены в TokenStore
	if tokenSet.AccessToken != "" {
		if err := h.tokenStore.SaveTokens(r.Context(), claims.Subject, tokenSet.AccessToken, tokenSet.RefreshToken); err != nil {
			slog.Warn("failed to save tokens", "error", err.Error())
			// Не критично — сессия создана, токены сохранятся при следующем запросе
		}
	}

	// Устанавливаем httpOnly cookie с session token
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == protoHTTPS
	setSessionCookie(w, sessionToken, isSecure)

	if h.metrics != nil {
		h.metrics.AuthCallbackTotal.WithLabelValues("success").Inc()
	}

	shhttp.WriteJSON(w, http.StatusOK, map[string]any{
		"user":  user.ToProfile(),
		"roles": roles,
	})
}

// setSessionCookie устанавливает cookie с session token.
//
// Production (isSecure=true): SameSite=None — требуется для subdomain multi-tenancy.
// Development (isSecure=false): SameSite=Lax — работает без HTTPS.
//
// Браузеры отклоняют cookie с SameSite=None если Secure=false,
// поэтому в dev используется Lax (безопасная деградация).
func setSessionCookie(w http.ResponseWriter, sessionToken string, isSecure bool) {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")

	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode, // default для dev (без HTTPS)
		Path:     "/",
		MaxAge:   86400, // 24 часа
	}

	if cookieDomain != "" {
		cookie.Domain = cookieDomain
	}

	// SameSite=None для production (HTTPS + домен задан)
	if isSecure {
		cookie.SameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, cookie)
}

// clearSessionCookie удаляет cookie сессии.
// Параметры SameSite/Secure должны совпадать с setSessionCookie,
// иначе браузер не удалит cookie (атрибуты должны совпадать).
func clearSessionCookie(w http.ResponseWriter, isSecure bool) {
	cookieDomain := os.Getenv("COOKIE_DOMAIN")

	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	}

	if cookieDomain != "" {
		cookie.Domain = cookieDomain
	}

	if isSecure {
		cookie.SameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, cookie)
}

// extractRolesFromAccessToken извлекает роли из access token Keycloak.
func extractRolesFromAccessToken(rawToken string) []string {
	return sharedauth.ExtractRolesFromJWT(rawToken)
}

// isBrowserRequest проверяет, является ли запрос от браузера.
func isBrowserRequest(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return true
	}
	for _, part := range strings.Split(accept, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "text/html") {
			return true
		}
	}
	return false
}

// extractRealmSlug извлекает realm имя из issuer URL.
// Примеры:
//
//	https://host/realms/lkfl-sdek → lkfl-sdek
//	http://keycloak:8080/realms/lkfl → lkfl
func extractRealmSlug(issuer string) string {
	parts := strings.Split(issuer, "/")
	for _, p := range parts {
		if p != "" {
			return p
		}
	}
	return ""
}

// Logout — инвалидация сессии + browser-based Keycloak SSO logout.
//
// Вместо server-side инвалидации (Admin API) используем browser-based redirect:
// браузер переходит на Keycloak logout endpoint, который инвалидирует SSO
// и удаляет KAUTH_SESSION_ID cookie. Это гарантирует, что повторный вход
// потребует ввода логина/пароля.
//
// Для всех запросов (AJAX и GET) — единый поток:
// 1. Очистка server-side state (Redis session + tokens)
// 2. Очистка session cookie
// 3. 302 redirect → Keycloak logout → frontend /login
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == protoHTTPS

	// 1. Очистка server-side state
	sessionToken := sharedauth.ExtractSessionCookie(r)
	var sessionData sharedauth.SessionData
	var hasSession bool

	if sessionToken != "" {
		var err error
		sessionData, err = h.sessionStore.Get(r.Context(), sessionToken)
		if err == nil {
			hasSession = true
		}
		_ = h.sessionStore.Delete(r.Context(), sessionToken)
	}

	if hasSession && sessionData.UserID != "" {
		_ = h.tokenStore.Delete(r.Context(), sessionData.UserID)
	}

	// 2. Очистка session cookie
	clearSessionCookie(w, isSecure)

	// 3. Browser-based Keycloak SSO logout (AJAX и GET — одинаково)
	//
	// Keycloak logout endpoint:
	// GET /protocol/openid-connect/logout?post_logout_redirect_uri=...
	//
	// Этот endpoint инвалидирует SSO сессию и удаляет KAUTH_SESSION_ID cookie
	// в браузере. После этого Keycloak redirect-ит на post_logout_redirect_uri.
	//
	// Для dev (localhost) post_logout_redirect_uri = http://localhost:5173/
	// Для production = домен frontend (из env FRONTEND_URL).
	var postLogoutRedirect string

	redirect := r.URL.Query().Get("post_logout_redirect_uri")
	if redirect != "" && isValidPostLogoutRedirect(redirect) {
		postLogoutRedirect = redirect
	} else {
		// Определяем frontend origin из заголовков запроса браузера.
		// publicIssuer — это URL Keycloak, а не frontend!
		frontendOrigin := r.Header.Get("Origin")
		if frontendOrigin == "" {
			ref := r.Header.Get("Referer")
			if ref != "" {
				if u, err := url.Parse(ref); err == nil {
					frontendOrigin = u.Scheme + "://" + u.Host
				}
			}
		}
		if frontendOrigin == "" {
			frontendOrigin = "http://localhost:5173"
		}
		postLogoutRedirect = frontendOrigin + "/login"
	}

	// Используем publicIssuer (URL видимый из браузера) + realm из issuer
	realmPath := ""
	if idx := strings.Index(h.issuer, "/realms/"); idx >= 0 {
		realmPath = h.issuer[idx:] // "/realms/lkfl-sdek"
	}
	logoutURL := fmt.Sprintf(
		"%s%s/protocol/openid-connect/logout?client_id=%s&post_logout_redirect_uri=%s",
		h.publicIssuer,
		realmPath,
		url.QueryEscape(h.clientID),
		url.QueryEscape(postLogoutRedirect),
	)

	slog.Debug("logout redirect to keycloak", "logout_url", logoutURL)
	http.Redirect(w, r, logoutURL, http.StatusFound)
}

// isValidPostLogoutRedirect проверяет redirect URI по allowlist.
func isValidPostLogoutRedirect(uri string) bool {
	allowed := os.Getenv("POST_LOGOUT_REDIRECT_WHITELIST")
	if allowed == "" {
		allowed = os.Getenv("FRONTEND_URL")
	}
	if allowed == "" {
		return false
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return false
	}

	for _, origin := range strings.Split(allowed, ",") {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if strings.HasPrefix(parsed.String(), origin) {
			return true
		}
	}
	return false
}

// Me — текущий пользователь с provisioning.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(sharedauth.ClaimsKey).(sharedauth.Claims)
	if !ok || claims.Subject == "" {
		shhttp.WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	roles := sharedauth.RolesFromContext(r.Context())

	user, err := h.service.CreateOrUpdateUser(r.Context(), &claims, roles)
	if err != nil {
		shhttp.WriteJSONError(w, http.StatusInternalServerError, "failed to resolve user: "+err.Error())
		return
	}

	shhttp.WriteJSON(w, http.StatusOK, user.ToProfile())
}

// isAjaxRequest проверяет, является ли запрос AJAX.
func isAjaxRequest(r *http.Request) bool {
	return r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
		strings.Contains(r.Header.Get("Accept"), "application/json") ||
		r.Header.Get("Content-Type") == "application/json"
}

// ExtractSessionCookie — публичный алиас для extractSessionCookie.
func ExtractSessionCookie(r *http.Request) string {
	return sharedauth.ExtractSessionCookie(r)
}
