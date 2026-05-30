// Package auth — Tenant resolver для извлечения tenant slug из OIDC issuer.
//
// Tenant slug определяется из issuer URL Keycloak:
//
//	https://host/realms/lkfl-{slug} → slug
//
// Используется JWTMiddleware для установки X-Tenant-ID header
// и auth handler для fallback tenant resolution.
package auth

import "strings"

// ResolveTenantSlug извлекает slug tenant'а из OIDC issuer URL.
//
// Формат issuer Keycloak: https://host/realms/lkfl-{slug}
// Функция ищет сегмент URL, начинающийся с "lkfl-" и возвращает
// часть после префикса.
//
// Примеры:
//
//	ResolveTenantSlug("https://keycloak.example.com/realms/lkfl-sdek") → "sdek"
//	ResolveTenantSlug("http://keycloak:8080/realms/lkfl-acme") → "acme"
//	ResolveTenantSlug("unknown") → ""
func ResolveTenantSlug(issuer string) string {
	parts := strings.Split(issuer, "/")
	for _, p := range parts {
		if strings.HasPrefix(p, "lkfl-") {
			return strings.TrimPrefix(p, "lkfl-")
		}
	}
	return ""
}
