package auth

import (
	"context"
	"encoding/json"
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
//  1. Извлекает Authorization header (Bearer token)
//  2. Верифицирует токен через OIDC verifier
//  3. Извлекает claims и roles из ID Token
//  4. Добавляет claims и roles в context запроса
//  5. Извлекает tenant slug из issuer и ставит X-Tenant-ID header
//
// При любой ошибке возвращает JSON-ответ с соответствующим статусом.
func JWTMiddleware(verifier *oidc.IDTokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			idToken, err := verifier.Verify(r.Context(), tokenString)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, roles, err := ExtractClaims(idToken)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "claim parsing error")
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
				if slug := extractTenantSlug(claims.Issuer); slug != "" {
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

// extractTenantSlug извлекает slug tenant'а из issuer URL.
// Формат: https://host/realms/lkfl-{slug} → slug
func extractTenantSlug(issuer string) string {
	parts := strings.Split(issuer, "/")
	for _, p := range parts {
		if strings.HasPrefix(p, "lkfl-") {
			return strings.TrimPrefix(p, "lkfl-")
		}
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

// writeJSONError пишет JSON-ответ с ошибкой.
//
// Формат ответа: {"error": "<message>"}
// Content-Type: application/json
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
