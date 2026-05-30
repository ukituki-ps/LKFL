package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/coreos/go-oidc"
)

// Claims — стандартные OIDC claims.
// Keycloak-specific данные (roles, tenant_slug) извлекаются через
// idToken.Claims(map[string]interface{}) для гибкости.
type Claims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	// Issuer — OIDC issuer (например, https://host/realms/lkfl-sdek).
	// Используется tenant middleware для извлечения slug tenant'а из JWT.
	Issuer string `json:"iss"`
}

// ExtractClaims извлекает стандартные claims и Keycloak роли из ID Token.
//
// Возвращает Claims (стандартные OIDC поля) и roles (Keycloak resource_access).
// Роли извлекаются из claim `resource_access.{clientID}.roles` —
// обходятся все клиенты для поиска первых найденных ролей.
func ExtractClaims(idToken *oidc.IDToken) (*Claims, []string, error) {
	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, nil, err
	}

	// Извлекаем issuer напрямую из IDToken (не через JSON unmarshal)
	claims.Issuer = idToken.Issuer

	// Keycloak roles — из resource_access.{clientID}.roles
	var rawClaims map[string]interface{}
	if err := idToken.Claims(&rawClaims); err != nil {
		return nil, nil, err
	}

	roles := extractKeycloakRoles(rawClaims)
	return &claims, roles, nil
}

// extractKeycloakRoles извлекает роли из Keycloak claims.
// Проверяет два источника:
//  1. resource_access.{clientID}.roles — client roles
//  2. realm_access.roles — realm roles
//
// Формат resource_access:
//
//	{
//	  "resource_access": {
//	    "lkfl-spa": {
//	      "roles": ["employee", "admin"]
//	    }
//	  }
//	}
//
// Формат realm_access:
//
//	{
//	  "realm_access": {
//	    "roles": ["admin", "employee"]
//	  }
//	}
//
// Функция объединяет роли из обоих источников.
func extractKeycloakRoles(raw map[string]interface{}) []string {
	// Собираем роли из resource_access (client roles)
	var roles []string
	ra, ok := raw["resource_access"].(map[string]interface{})
	if ok {
		for _, clientObj := range ra {
			client, ok := clientObj.(map[string]interface{})
			if !ok {
				continue
			}

			rolesObj, ok := client["roles"].([]interface{})
			if !ok {
				continue
			}

			for _, r := range rolesObj {
				if s, ok := r.(string); ok {
					roles = append(roles, s)
				}
			}
		}
	}

	// Если client roles не найдены, проверяем realm_access (realm roles)
	if len(roles) == 0 {
		if realmAccess, ok := raw["realm_access"].(map[string]interface{}); ok {
			if realmRolesObj, ok := realmAccess["roles"].([]interface{}); ok {
				for _, r := range realmRolesObj {
					if s, ok := r.(string); ok {
						roles = append(roles, s)
					}
				}
			}
		}
	}

	return roles
}

// ExtractRolesFromJWT извлекает роли из сырого JWT токена (access token).
//
// ВНИМАНИЕ: декодирует payload БЕЗ верификации подписи.
// Использовать ТОЛЬКО с токенами, полученными напрямую от Keycloak token endpoint
// (внутри LoginCallback). Никогда не использовать с токенами от внешних клиентов.
//
// Для верифицированного извлечения ролей используйте ExtractClaims с OIDC ID Token.
func ExtractRolesFromJWT(rawToken string) []string {
	parts := splitJWT(rawToken)
	if len(parts) < 2 {
		return []string{}
	}

	payload, err := decodeBase64(parts[1])
	if err != nil {
		return []string{}
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return []string{}
	}

	return extractKeycloakRoles(claims)
}

// splitJWT разбивает JWT на части (header.payload.signature).
func splitJWT(token string) []string {
	return strings.Split(token, ".")
}

// decodeBase64 декодирует base64url строку с padding.
func decodeBase64(s string) ([]byte, error) {
	// Добавляем padding если нужно
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
