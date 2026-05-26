# T1902 — shared/pkg/auth/ (OIDC Verifier + Middleware)

## Веха

M19-auth-rbac

## Тип

code

## Контекст

`shared/pkg/auth/` — общий пакет для аутентификации.
Используется монолитом (`lkfl-server`) и future пакетами.
Исходник: `doc/архитектура/пакеты-platform.md` (строка ~146 — auth/).

## Что сделать

### Структура

```
shared/pkg/auth/
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
    "github.com/coreos/go-oidc/v3/oidc"
)

// NewVerifier создаёт OIDC verifier из Keycloak issuer URL
func NewVerifier(issuerURL string) (*oidc.IDTokenVerifier, error) {
    provider, err := oidc.NewProvider(context.Background(), issuerURL)
    if err != nil {
        return nil, fmt.Errorf("oidc provider: %w", err)
    }

    return provider.Verifier(&oidc.Config{
        ClientID: clientID, // из config
    }), nil
}
```

### `claims.go`

```go
package auth

import "github.com/google/uuid"

// Claims — структура JWT claims (Keycloak custom claims)
type Claims struct {
    Subject    string   `json:"sub"`            // Keycloak user ID
    Email      string   `json:"email"`
    PreferredUsername string `json:"preferred_username"`
    Name       string   `json:"name"`
    GivenName  string   `json:"given_name"`
    FamilyName string   `json:"family_name"`
    Roles      []string `json:"resource_access,lkfl-spa,roles"` // Keycloak roles
    TenantSlug string   `json:"tenant_slug"`    // custom claim
}

// ExtractClaims извлекает claims из ID Token
func ExtractClaims(tokenString string) (*Claims, error) {
    // Parse unverified header для получения audience
    // Verify token
    // Extract claims
}
```

### `middleware.go`

```go
package auth

import (
    "context"
    "net/http"
    "strings"

    "github.com/go-chi/chi/v5"
    "github.com/coreos/go-oidc/v3/oidc"
)

type contextKey string

const (
    UserIDKey   contextKey = "auth_user_id"
    ClaimsKey   contextKey = "auth_claims"
    RolesKey    contextKey = "auth_roles"
)

// JWTMiddleware — извлекает Bearer token, верифицирует, добавляет claims в context
func JWTMiddleware(verifier *oidc.IDTokenVerifier) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                http.Error(w, `{"error":"invalid token format"}`, http.StatusUnauthorized)
                return
            }

            idToken, err := verifier.Verify(r.Context(), tokenString)
            if err != nil {
                http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
                return
            }

            var claims Claims
            if err := idToken.Claims(&claims); err != nil {
                http.Error(w, `{"error":"claim parsing error"}`, http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), ClaimsKey, claims)
            ctx = context.WithValue(ctx, RolesKey, claims.Roles)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// UserIDFromContext — helper для извлечения user ID из context
func UserIDFromContext(ctx context.Context) string {
    claims, ok := ctx.Value(ClaimsKey).(Claims)
    if !ok {
        return ""
    }
    return claims.Subject
}

// RolesFromContext — helper для извлечения ролей из context
func RolesFromContext(ctx context.Context) []string {
    roles, ok := ctx.Value(RolesKey).([]string)
    if !ok {
        return nil
    }
    return roles
}
```

### `rbac.go`

```go
package auth

// RBACMiddleware — проверяет, что пользователь имеет хотя бы одну из требуемых ролей
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
                http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

## Требования

- OIDC verifier — `go-oidc` + `coreos/go-oidc`
- JWT middleware — Bearer token extraction → verification → context injection
- RBAC middleware — role check в context roles vs required roles
- Claims — Keycloak custom claims mapping (`resource_access.lkfl-spa.roles`)
- Helper functions: `UserIDFromContext()`, `RolesFromContext()`
- Error responses — JSON, не plain text
- Skip paths: `/healthz`, `/metrics` (без auth middleware)

## Критерии приёмки

- [ ] `verifier.go` — OIDC verifier из Keycloak issuer
- [ ] `claims.go` — Claims struct + extraction
- [ ] `middleware.go` — JWT middleware (Bearer → verify → context)
- [ ] `rbac.go` — RBAC middleware (role check)
- [ ] Helper functions: UserIDFromContext, RolesFromContext
- [ ] JSON error responses
- [ ] Unit tests: valid token, expired token, invalid signature, missing Bearer, RBAC allow/deny
