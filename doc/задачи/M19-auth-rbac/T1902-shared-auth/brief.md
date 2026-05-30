# T1902 — shared/pkg/auth/ (OIDC Verifier + Middleware)

## Веха

M19-auth-rbac

## Тип

code

## Контекст

`backend/shared/pkg/auth/` — общий пакет для аутентификации.
Используется монолитом (`lkfl-server`) и future пакетами.
Исходник: `doc/архитектура/пакеты-platform.md` (строка ~146 — auth/).

**Важно:** `app/wire.go` (M17) уже создаёт OIDC verifier через `newOIDCVerifier()`.
T1902 создаёт shared-пакет для переиспользования. После реализации M19 —
`wire.go` должен использовать `auth.NewVerifier()` вместо локальной функции.

**Зависимости go.mod:** `github.com/coreos/go-oidc v2.3.0+incompatible` (НЕ v3).

## Что сделать

### Структура

```
backend/shared/pkg/auth/
├── verifier.go      # OIDC IDTokenVerifier (go-oidc)
├── middleware.go    # JWT middleware для chi
├── rbac.go          # RBAC middleware
└── claims.go        # JWT claims extraction
```

### `verifier.go`

```go
package auth

import (
    "context"
    "fmt"

    "github.com/coreos/go-oidc"  // v2.3.0+incompatible (см. go.mod)
)

// NewVerifier создаёт OIDC verifier из Keycloak issuer URL.
func NewVerifier(ctx context.Context, issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
    provider, err := oidc.NewProvider(ctx, issuerURL)
    if err != nil {
        return nil, fmt.Errorf("oidc provider: %w", err)
    }

    return provider.Verifier(&oidc.Config{
        ClientID: clientID,
    }), nil
}
```

### `claims.go`

```go
package auth

// Claims — стандартные OIDC claims.
// Keycloak-specific данные (roles, tenant_slug) извлекаются через
// idToken.Claims(map[string]interface{}) для гибкости.
type Claims struct {
    Subject    string `json:"sub"`
    Email      string `json:"email"`
    PreferredUsername string `json:"preferred_username"`
    Name       string `json:"name"`
    GivenName  string `json:"given_name"`
    FamilyName string `json:"family_name"`
}

// ExtractClaims извлекает claims из ID Token.
func ExtractClaims(idToken *oidc.IDToken) (*Claims, []string, error) {
    var claims Claims
    if err := idToken.Claims(&claims); err != nil {
        return nil, nil, err
    }

    // Keycloak roles — из resource_access.{clientID}.roles
    var rawClaims map[string]interface{}
    if err := idToken.Claims(&rawClaims); err != nil {
        return nil, nil, err
    }

    roles := extractKeycloakRoles(rawClaims)
    return &claims, roles, nil
}

// extractKeycloakRoles — извлекает роли из Keycloak resource_access claim.
// Формат: {"resource_access": {"lkfl-spa": {"roles": ["employee", "admin"]}}}
func extractKeycloakRoles(raw map[string]interface{}) []string {
    ra, ok := raw["resource_access"].(map[string]interface{})
    if !ok {
        return nil
    }
    // Ищем roles в любом client (lkfl-spa, lkfl-service)
    for _, clientObj := range ra {
        client, ok := clientObj.(map[string]interface{})
        if !ok {
            continue
        }
        rolesObj, ok := client["roles"].([]interface{})
        if !ok {
            continue
        }
        var roles []string
        for _, r := range rolesObj {
            if s, ok := r.(string); ok {
                roles = append(roles, s)
            }
        }
        if len(roles) > 0 {
            return roles
        }
    }
    return nil
}
```

### `middleware.go`

```go
package auth

import (
    "context"
    "encoding/json"
    "net/http"
    "strings"

    "github.com/coreos/go-oidc"  // v2.3.0+incompatible
)

type contextKey string

const (
    UserIDKey   contextKey = "auth_user_id"
    ClaimsKey   contextKey = "auth_claims"
    RolesKey    contextKey = "auth_roles"
)

// JWTMiddleware — извлекает Bearer token, верифицирует, добавляет claims в context.
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

            ctx := context.WithValue(r.Context(), ClaimsKey, claims)
            ctx = context.WithValue(ctx, RolesKey, roles)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// UserIDFromContext — helper для извлечения user ID из context.
func UserIDFromContext(ctx context.Context) string {
    claims, ok := ctx.Value(ClaimsKey).(Claims)
    if !ok {
        return ""
    }
    return claims.Subject
}

// RolesFromContext — helper для извлечения ролей из context.
func RolesFromContext(ctx context.Context) []string {
    roles, ok := ctx.Value(RolesKey).([]string)
    if !ok {
        return nil
    }
    return roles
}

// writeJSONError — JSON error response (не plain text).
func writeJSONError(w http.ResponseWriter, status int, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]string{"error": message})
}
```

### `rbac.go`

```go
package auth

import "net/http"

// RBACMiddleware — проверяет, что пользователь имеет хотя бы одну из требуемых ролей.
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
                writeJSONError(w, http.StatusForbidden, "forbidden")
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Требования

- OIDC verifier — `github.com/coreos/go-oidc v2.3.0+incompatible` (НЕ v3, см. go.mod)
- JWT middleware — Bearer token extraction → verification → context injection
- RBAC middleware — role check в context roles vs required roles
- Claims — Keycloak custom claims mapping (`resource_access.{clientID}.roles`)
- Helper functions: `UserIDFromContext()`, `RolesFromContext()`
- Error responses — JSON (`writeJSONError`), не plain text (`http.Error` запрещён)
- Путь пакета: `backend/shared/pkg/auth/` (import path: `lkfl/shared/pkg/auth`)
- После реализации — `app/wire.go` заменить `newOIDCVerifier()` на `auth.NewVerifier()`

## Критерии приёмки

- [ ] `verifier.go` — OIDC verifier из Keycloak issuer
- [ ] `claims.go` — Claims struct + extraction
- [ ] `middleware.go` — JWT middleware (Bearer → verify → context)
- [ ] `rbac.go` — RBAC middleware (role check)
- [ ] Helper functions: UserIDFromContext, RolesFromContext
- [ ] JSON error responses
- [ ] Unit tests: valid token, expired token, invalid signature, missing Bearer, RBAC allow/deny
