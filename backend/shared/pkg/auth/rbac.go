package auth

import (
	"context"
	"net/http"
)

// RBACMiddleware создаёт HTTP-мидлвэр для проверки ролей пользователя.
//
// Проверяет, что пользователь имеет хотя бы одну из требуемых ролей.
// Роли берутся из context (должны быть установлены JWTMiddleware ранее).
//
// При отсутствии подходящей роли возвращает 403 Forbidden (JSON).
//
// Пример использования:
//
//	r.Use(auth.JWTMiddleware(verifier))
//	r.Group(func(r chi.Router) {
//	    r.Use(auth.RBACMiddleware([]string{"admin"}))
//	    r.Get("/admin/dashboard", adminHandler)
//	})
func RBACMiddleware(requiredRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles := RolesFromContext(r.Context())

			hasRole := false
			for _, ur := range userRoles {
				for _, rr := range requiredRoles {
					if ur == rr {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				WriteAuthError(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// withRoles — helper для тестирования: создаёт context с указанными ролями.
func withRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, RolesKey, roles)
}
