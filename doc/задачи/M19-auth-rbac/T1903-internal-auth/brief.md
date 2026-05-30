# T1903 — internal/auth/ (Auth handlers)

## Веха

M19-auth-rbac

## Тип

code

## Контекст

`backend/internal/auth/` — handlers для auth flow: login redirect, callback, logout.
SPA auth flow через Keycloak (PKCE, realm `lkfl-sdek`).

**Зависимости от M18:**
- `tenant.Service` — для резолва tenant при CreateOrUpdateUser (tenant из Keycloak claims → tenant в БД)
- `tenant.TenantIDFromContext()` — используется в handler'ах после tenant middleware
- `internal/user/` repository (T1905) — для CRUD пользователей
- Redis — для хранения auth state (CSRF) и сессий (key prefix: `auth:state:`, `auth:session:`)

## Что сделать

### Структура

```
internal/auth/
├── handler.go   # HTTP handlers
└── service.go   # Auth service (user creation on first login)
```

### `handler.go`

```go
package auth

// LoginRedirect — редирект на Keycloak login page
// GET /api/v1/auth/login
// Response: 302 → Keycloak authorize endpoint
func (h *Handler) LoginRedirect(w http.ResponseWriter, r *http.Request) {
    // Generate state parameter (CSRF protection)
    state := generateState()
    // Store state in Redis (jwt:auth:state:{state})
    // Redirect to Keycloak authorize endpoint
}

// LoginCallback — Keycloak callback с code/token
// GET /api/v1/auth/callback
// Response: 200 с user data (для API clients)
// Для SPA: редирект обратно с hash token
func (h *Handler) LoginCallback(w http.ResponseWriter, r *http.Request) {
    // Exchange code for tokens (если authorization code flow)
    // Verify ID token
    // Extract claims
    // Create/update user in DB (first login → create)
    // Set session in Redis
    // Return user data
}

// Logout — инвалидация сессии
// POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    // Remove session from Redis
    // Redirect to Keycloak logout
}

// Me — текущий пользователь
// GET /api/v1/auth/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
    // Get user from DB by keycloak_sub
    // Return user profile
}
```

### `service.go`

```go
package auth

// CreateOrUpdateUser — first login → create, subsequent → update
func (s *Service) CreateOrUpdateUser(ctx context.Context, claims Claims) (User, error) {
    // Find by keycloak_sub
    // If not found → create (with tenant from claims.TenantSlug)
    // If found → update (name, email from Keycloak)
    // Return user
}
```

## Требования

- Login redirect — Keycloak authorize endpoint с state parameter
- Callback — token verification → user create/update → session
- Logout — Redis session removal + Keycloak logout redirect
- First login → create user in DB
- Subsequent login → update user data from Keycloak
- State parameter — CSRF protection (Redis storage, TTL 10min)

## Критерии приёмки

- [ ] `handler.go` — LoginRedirect, LoginCallback, Logout, Me
- [ ] `service.go` — CreateOrUpdateUser
- [ ] Login redirect → Keycloak authorize
- [ ] Callback → verify → create/update → session
- [ ] Logout → session removal
- [ ] State parameter CSRF protection
- [ ] Unit tests
