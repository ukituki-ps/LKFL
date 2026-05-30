package auth

import (
	"context"
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
	ClaimsKey contextKey = "auth_claims"
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
			ctx := context.WithValue(r.Context(), ClaimsKey, *claims)
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
const sessionCookieName = "lkfl_session"

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
	claims, ok := ctx.Value(ClaimsKey).(Claims)
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
