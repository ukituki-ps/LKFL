package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc" // v2.3.0+incompatible
)

// contextKey — тип ключей для context, чтобы избежать коллизий.
type contextKey string

const (
	// UserIDKey — ключ user ID (subject) в context.
	UserIDKey contextKey = "auth_user_id"
	// ClaimsKey — ключ Claims в context.
	// RolesKey — ключ ролей пользователя в context.
	RolesKey contextKey = "auth_roles"
)

// JWTMiddleware создаёт HTTP-мидлвэр для верификации JWT Bearer токенов.
//
// Алгоритм работы:
//  1. Извлекает токен из Authorization header (Bearer) или cookie (lkfl_session)
//  2. Верифицирует токен через OIDC verifier
//  3. Извлекает claims и roles из ID Token
//  4. Добавляет claims и roles в context запроса
//  5. Извлекает tenant slug из issuer и ставит X-Tenant-ID header
//
// D2: токен ищется сначала в Authorization: Bearer, затем в cookie lkfl_session
// (backward compat — оба источника работают).
//
// При любой ошибке возвращает JSON-ответ с соответствующим статусом.
func JWTMiddleware(verifier *oidc.IDTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// D2: сначала проверяем Authorization: Bearer header, затем cookie
			tokenString := extractToken(r)
			if tokenString == "" {
				WriteAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			idToken, err := verifier.Verify(r.Context(), tokenString)
			if err != nil {
				WriteAuthError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, roles, err := ExtractClaims(idToken)
			if err != nil {
				WriteAuthError(w, http.StatusUnauthorized, "claim parsing error")
				return
			}

			// Храним Claims как value (не pointer) — UserIDFromContext/RolesFromContext
			// ожидают именно Claims, не *Claims. Type assertion на pointer упадёт.
			ctx := context.WithValue(r.Context(), ClaimsKey{}, *claims)
			ctx = context.WithValue(ctx, RolesKey, roles)

			// Извлекаем tenant slug из issuer и ставим X-Tenant-ID.
			// Это заменяет JWTClaimsTenantMiddleware — slug доступен
			// для tenant middleware даже если nginx не поставил header.
			if r.Header.Get("X-Tenant-ID") == "" && claims.Issuer != "" {
				if slug := ResolveTenantSlug(claims.Issuer); slug != "" {
					r = r.Clone(ctx)
					r.Header.Set("X-Tenant-ID", slug)
					next.ServeHTTP(w, r)
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// sessionCookieName — имя httpOnly cookie для сессионного токена (D2).

// extractToken извлекает токен из запроса.
// Сначала проверяет Authorization: Bearer <token>, затем cookie lkfl_session.
// Это обеспечивает backward compatibility — оба источника работают.
func extractToken(r *http.Request) string {
	// Приоритет 1: Authorization: Bearer header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString != authHeader {
			return tokenString
		}
	}

	// Приоритет 2: httpOnly cookie (D2: 152-ФЗ compliance)
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// UserIDFromContext извлекает user ID (subject) из context.
//
// Возвращает пустую строку, если claims отсутствуют в context.
func UserIDFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(ClaimsKey{}).(Claims)
	if !ok {
		return ""
	}
	return claims.Subject
}

// RolesFromContext извлекает роли пользователя из context.
//
// Возвращает nil, если роли отсутствуют в context.
func RolesFromContext(ctx context.Context) []string {
	roles, ok := ctx.Value(RolesKey).([]string)
	if !ok {
		return nil
	}
	return roles
}

// SessionMiddleware — middleware для session-based аутентификации.
//
// Алгоритм:
//  1. Извлекает session token из cookie (lkfl_session)
//  2. Проверяет сессию в Redis (SessionStore)
//  3. Проверяет/обновляет access token в Redis (TokenStore)
//  4. Если token истёк — делает refresh через TokenRefresher
//  5. Извлекает claims из access token и добавляет в context
//
// При ошибке сессии (отсутствует/истёкла) — удаляет cookie и возвращает 401.
func SessionMiddleware(sessionStore *SessionStore, tokenStore *TokenStore, tokenRefresher *TokenRefresher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionToken := ExtractSessionCookie(r)
			if sessionToken == "" {
				WriteAuthError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			sessionData, err := sessionStore.Get(r.Context(), sessionToken)
			if err != nil {
				clearSessionCookie(w, r)
				WriteAuthError(w, http.StatusUnauthorized, "session expired")
				return
			}

			accessToken, err := tokenStore.GetAccessToken(r.Context(), sessionData.UserID)
			if err != nil {
				accessToken, _, err = tokenRefresher.Refresh(r.Context(), sessionData.UserID)
				if err != nil {
					clearSessionCookie(w, r)
					WriteAuthError(w, http.StatusUnauthorized, "token refresh failed")
					return
				}
			}

			claims, roles, err := extractClaimsFromAccessToken(accessToken)
			if err != nil {
				WriteAuthError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey{}, *claims)
			ctx = context.WithValue(ctx, RolesKey, roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractClaimsFromAccessToken парсит access token и извлекает claims + roles.
func extractClaimsFromAccessToken(token string) (*Claims, []string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid token format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, err
	}

	var rawClaims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &rawClaims); err != nil {
		return nil, nil, err
	}

	claims := &Claims{}
	if sub, ok := rawClaims["sub"].(string); ok {
		claims.Subject = sub
	}
	if email, ok := rawClaims["email"].(string); ok {
		claims.Email = email
	}
	if name, ok := rawClaims["name"].(string); ok {
		claims.Name = name
	}
	if iss, ok := rawClaims["iss"].(string); ok {
		claims.Issuer = iss
	}

	roles := extractKeycloakRolesFromClaims(rawClaims)

	return claims, roles, nil
}
