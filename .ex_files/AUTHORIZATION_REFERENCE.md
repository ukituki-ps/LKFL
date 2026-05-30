# April Ecosystem — Authorization Reference Guide

> **Purpose:** Single source of truth for how authorization works across the entire April ecosystem.
> Use this file to understand every auth layer before modifying auth-related code.
>
> **Scope:** Keycloak (IdP), Nginx (ingress), Frontend (hub-shell), Hub-BFF (Go), Profile API (Go).

---

## Table of Contents

1. [Architecture at a Glance](#1-architecture-at-a-glance)
2. [Identity Provider — Keycloak](#2-identity-provider--keycloak)
3. [Ingress Layer — Nginx](#3-ingress-layer--nginx)
4. [Frontend Authorization — Hub-Shell](#4-frontend-authorization--hub-shell)
5. [Backend Authorization — Hub-BFF](#5-backend-authorization--hub-bff)
6. [Backend Authorization — Profile API](#6-backend-authorization--profile-api)
7. [Multi-Tenancy Isolation](#7-multi-tenancy-isolation)
8. [Request Lifecycle](#8-request-lifecycle)
9. [Error Handling Matrix](#9-error-handling-matrix)
10. [Security Headers and Policies](#10-security-headers-and-policies)
11. [Observability](#11-observability)
12. [File Index](#12-file-index)

---

## 1. Architecture at a Glance

### Core Principle

April uses an **external IdP model via Keycloak (OIDC)**. The backends are **stateless**: they never issue JWT, never store user sessions, never implement `/auth/login` or `/auth/refresh`. They only **validate** incoming JWT tokens by verifying their signature against Keycloak's JWKS endpoint.

```
                         ┌──────────────────────────────────────┐
                         │            Nginx (:80)               │
                         │  /          → hub-shell:4173         │
                         │  /auth     → keycloak:8080           │
                         │  /api/     → hub-bff:8081            │
                         │  /healthz  → hub-bff:8081            │
                         │  /metrics  → hub-bff:8081 (IP-only)  │
                         └──────────┬───────────────────────────┘
                                    │
         ┌──────────────────────────┼──────────────────────────┐
         │                          │                          │
         ▼                          ▼                          ▼
┌────────────────┐      ┌──────────────────┐      ┌─────────────────┐
│  Hub-Shell     │      │  Hub-BFF (Go)    │      │  Keycloak (26.1)│
│  React+Vite    │─────►│  Gin-less mux    │      │  OIDC IdP       │
│  keycloak-js   │      │  JWT validate    │      │  Realm: april   │
│  Bearer JWT    │      │  RBAC + ABAC     │      │  JWKS endpoint  │
└────────────────┘      └────────┬─────────┘      └─────────────────┘
                                 │ reverse proxy
                                 ▼
                        ┌──────────────────┐
                        │  Profile API (Go)│
                        │  JWT validate    │
                        │  RBAC + ABAC     │
                        │  Tenant isolation│
                        └──────────────────┘
```

### Three Authorization Layers

| Layer | Mechanism | Enforced By |
|---|---|---|
| **Authentication** | JWT signature via JWKS + claims validation | Go middleware (`auth.Middleware.Validate`) |
| **Authorization (RBAC)** | Realm roles from Keycloak (`realm_access.roles`) | `RequireAnyRole` / `RequireRealmRole` |
| **Authorization (ABAC)** | Document segment access by role | Profile API `abac.Policy` |

---

## 2. Identity Provider — Keycloak

### Realm Configuration

**File:** `april-worker-kilo/infra/keycloak/realm/april-realm.json`
**Image:** `quay.io/keycloak/keycloak:26.1`
**Command:** `start-dev --import-realm --http-relative-path /auth`

#### Realm Settings
```json
{
  "realm": "april",
  "displayName": "April Dev Realm",
  "loginTheme": "aprilhub",
  "accountTheme": "aprilhub",
  "internationalizationEnabled": true,
  "supportedLocales": ["ru", "en"],
  "defaultLocale": "ru"
}
```

#### RBAC Roles

Three realm roles, used across all services:

| Role | Access Level | Used For |
|---|---|---|
| `user` | Read + basic write | All authenticated endpoints |
| `manager` | Extended operations | Reserved for future expansion |
| `admin` | Admin endpoints, profile management | `/api/v1/admin/*`, `/v1/admin/*` |

#### OIDC Clients

**Three clients are registered in the realm:**

##### Client 1: `aprilhub-shell` (Frontend)

```json
{
  "clientId": "aprilhub-shell",
  "publicClient": true,
  "protocol": "openid-connect",
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": true,
  "redirectUris": [
    "http://localhost:4173/*",
    "http://localhost:8080/*",
    "http://localhost:18080/*",
    "http://dev.april.ukituki.tech/*",
    "https://dev.april.ukituki.tech/*"
  ],
  "attributes": {
    "pkce.code.challenge.method": "S256"
  }
}
```

- **Public client** — no secret, uses PKCE for security
- **Flow:** Authorization Code + PKCE S256
- **Mapper:** `tenant_id` user attribute → JWT claim

##### Client 2: `aprilhub-bff` (Service Account)

```json
{
  "clientId": "aprilhub-bff",
  "publicClient": false,
  "secret": "dev-only-secret",
  "serviceAccountsEnabled": true,
  "attributes": {
    "access.token.lifespan": "300"
  }
}
```

- **Confidential client** — has a client secret
- **Service Accounts enabled** — can use `client_credentials` grant
- **Short-lived tokens** — 300s access token lifespan
- Used for machine-to-machine communication between services

##### Client 3: `april-profile-api` (API Audience)

```json
{
  "clientId": "april-profile-api",
  "publicClient": true,
  "standardFlowEnabled": false,
  "directAccessGrantsEnabled": true
}
```

- **Public client** — not used for browser login
- **Purpose:** Defines the `aud` / `azp` claim value for Profile API token validation
- **Mapper:** `tenant_id` user attribute → JWT claim

#### Protocol Mapper — `tenant_id`

Both `aprilhub-shell` and `april-profile-api` have this mapper:

```json
{
  "name": "tenant_id",
  "protocolMapper": "oidc-usermodel-attribute-mapper",
  "config": {
    "user.attribute": "tenant_id",
    "claim.name": "tenant_id",
    "jsonType.label": "String",
    "access.token.claim": "true",
    "id.token.claim": "true",
    "userinfo.token.claim": "true"
  }
}
```

This reads the `tenant_id` attribute from the Keycloak user and injects it into:
- **Access token** — used by Go backends for tenant isolation
- **ID token** — available to frontend via `keycloak.idToken`
- **Userinfo endpoint** — available via `/userinfo`

#### Demo Users

```json
{
  "username": "april-dev",
  "email": "april-dev@example.local",
  "attributes": { "tenant_id": ["00000000-0000-0000-0000-000000000001"] },
  "realmRoles": ["user", "admin"],
  "credentials": [{ "type": "password", "value": "april-dev-pass" }]
}
{
  "username": "april-user",
  "email": "april-user@example.local",
  "attributes": { "tenant_id": ["00000000-0000-0000-0000-000000000001"] },
  "realmRoles": ["user"],
  "credentials": [{ "type": "password", "value": "april-user-pass" }]
}
```

### Environment Variables for Keycloak

| Variable | Default | Purpose |
|---|---|---|
| `KEYCLOAK_ADMIN` | `admin` | Admin username |
| `KEYCLOAK_ADMIN_PASSWORD` | `admin` | Admin password |
| `KC_HOSTNAME` | `https://dev.april.ukituki.tech/auth` | Public hostname (determines `iss` in JWT) |
| `KC_DB` | `postgres` | Database driver |
| `KC_DB_URL_HOST` | `keycloak-db` | DB host (Docker service name) |
| `KC_DB_URL_DATABASE` | `keycloak` | DB name |
| `KC_DB_USERNAME` | `keycloak` | DB username |
| `KC_DB_PASSWORD` | `keycloak` | DB password |

### Setup Scripts

| Script | Purpose |
|---|---|
| `scripts/keycloak-april-user-profile-allow-tenant-attr.sh` | Sets `unmanagedAttributePolicy=ENABLED` so custom user attributes (like `tenant_id`) are allowed in Keycloak 24+ |
| `scripts/keycloak-ensure-april-theme.sh` | Applies `aprilhub` login/account theme to the realm via `kcadm` |

---

## 3. Ingress Layer — Nginx

### Two Configuration Files

| File | Purpose |
|---|---|
| `infra/nginx/aprilhub.conf` | Production config — strict security headers, no HMR |
| `infra/nginx/default.conf` | Dev config — includes HMR WebSocket upgrade, showcase routing |

### Route Table

| Path | Upstream | Auth Handling | Notes |
|---|---|---|---|
| `/` | `hub-shell:4173` | None (transparent proxy) | SPA fallback |
| `/api/` | `hub-bff:8081` | None (transparent proxy) | `Authorization` header passed through |
| `/auth` | `keycloak:8080` | None (Keycloak handles auth) | OIDC endpoints |
| `/metrics` | `hub-bff:8081` | **IP whitelist** | Only `172.21.0.0/16`, `127.0.0.1`, `::1` |
| `/healthz` | `hub-bff:8081` | None | Public health check |
| `/livez` | `hub-bff:8081` | None | Public liveness check |
| `/readyz` | `hub-bff:8081` | None | Public readiness check |
| `/showcase/` | `april-showcase:4174` | None | Dev only (default.conf) |
| `/docs/` | Static files | None | Dev only (default.conf) |
| `/swagger/` | `swagger-ui:8080` | None | Dev only (default.conf) |

### Nginx Does NOT Perform Authentication

Nginx is a **transparent reverse proxy** for auth. It forwards the `Authorization: Bearer` header from the client to the upstream Go backend. The Go backend performs JWT validation.

### Security Headers (aprilhub.conf)

```nginx
# Global headers (all locations)
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "0" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;

# Hub Shell (frontend) location
add_header Content-Security-Policy
  "default-src 'self';
   script-src 'self' 'unsafe-inline' 'unsafe-eval';
   style-src 'self' 'unsafe-inline';
   img-src 'self' data:;
   font-src 'self';
   connect-src 'self' https://dev.april.ukituki.tech https://sentry.io;
   frame-src 'self' https://dev.april.ukituki.tech/auth;
   report-uri /api/v1/csp-report" always;
add_header X-Frame-Options "SAMEORIGIN" always;

# API location
add_header Content-Security-Policy "default-src 'none'" always;
add_header X-Frame-Options "DENY" always;

# Keycloak location
add_header X-Frame-Options "DENY" always;
```

### Metrics IP Whitelist

```nginx
location = /metrics {
    allow 172.21.0.0/16;    # Docker network
    allow 127.0.0.1/32;     # Localhost
    allow ::1/128;          # IPv6 localhost
    deny all;
    proxy_pass http://hub-bff:8081/metrics;
}
```

---

## 4. Frontend Authorization — Hub-Shell

**Location:** `april-worker-kilo/hub-shell/src/`
**Tech:** React + TypeScript + Vite + Mantine + `keycloak-js`

### Initialization Flow

**File:** `src/main.tsx`
```
1. initializeSentry()
2. installGlobalRuntimeHandlers()
3. await initializeAuth()    // keycloak.init(...)
   └─ If error → captured as authInitError
4. Render <App authInitError={...} />
```

**File:** `src/keycloak.ts`
```typescript
export const keycloak = new Keycloak({
  url: authConfig.keycloakUrl,       // "/auth" (via Nginx)
  realm: authConfig.realm,           // "april"
  clientId: authConfig.clientId,     // "aprilhub-shell"
});

export async function initializeAuth(): Promise<boolean> {
  keycloak.onAuthRefreshSuccess = () => notifyKeycloakTokenRotated();
  keycloak.onTokenExpired = () => {
    void keycloak.updateToken(70).catch(() => void keycloak.login());
  };
  return keycloak.init({
    onLoad: "check-sso",        // Check existing SSO session
    pkceMethod: "S256",         // PKCE for security
    checkLoginIframe: false,    // Disable legacy iframe check
  });
}
```

### Auth Configuration

**File:** `src/auth.ts`
```typescript
export const authConfig: AuthConfig = {
  keycloakUrl: VITE_KEYCLOAK_URL ?? "/auth",
  realm: VITE_KEYCLOAK_REALM ?? "april",
  clientId: VITE_KEYCLOAK_CLIENT_ID ?? "aprilhub-shell",
  apiBaseUrl: VITE_API_BASE_URL ?? "/api",
};
```

### Application Auth State Machine

**File:** `src/App.tsx`

The app manages four auth zones:

```
┌───────┐    login()    ┌────────────┐  GET /api/v1/me  ┌────────────┐
│ guest │──────────────►│ transition │──────────────────►│ authorized │
└───────┘               └────────────┘                   └────────────┘
                                │                             │
                                │ 403                         │ context == null
                                ▼                             ▼
                         ┌────────────┐                ┌──────────────┐
                         │ forbidden  │                │ no-context   │
                         └────────────┘                └──────────────┘
```

| Zone | Condition | UI Shown |
|---|---|---|
| `guest` | `!keycloak.authenticated` | `GuestB2BLanding` — login button |
| `transition` | `keycloak.authenticated`, waiting for `/v1/me` | Loading state with message |
| `authorized` | `/v1/me` returned 200 + user profile loaded | `AuthorizedShellGate` with full navigation |
| `forbidden` | `/v1/me` returned 403 | "Доступ ограничен" + logout button |
| `no-context` | `authorized` but `me == null` | "Профиль не загружен" + logout button |

### API Client — Token Handling

**File:** `src/api.ts`

Every API request goes through `authorizedFetch()`:

```typescript
export async function authorizedFetch(input, init, options) {
  // Step 1: Ensure we have a token
  if (!keycloak.authenticated || !keycloak.token) {
    await keycloak.login();
    throw new Error("Unauthenticated session");
  }

  // Step 2: Attach Bearer token + telemetry headers
  headers.set("Authorization", `Bearer ${keycloak.token}`);
  if (telemetry?.requestId) headers.set("X-Request-Id", telemetry.requestId);
  if (telemetry?.correlationId) headers.set("X-Correlation-Id", telemetry.correlationId);

  // Step 3: Make the request
  const response = await fetch(input, { headers });

  // Step 4: Handle 401 — automatic token refresh + one retry
  if (response.status === 401 && options.canRetry) {
    const refreshed = await keycloak.updateToken(30);  // 30s grace period
    if (!refreshed) {
      await keycloak.login();                           // Full re-login
      throw new Error("Token refresh failed");
    }
    notifyKeycloakTokenRotated();                       // Notify React tree
    return authorizedFetch(input, retryInit, { canRetry: false });  // Retry once
  }

  // Step 5: Capture HTTP errors for Sentry (400, 404, 503, 5xx)
  return response;
}
```

### Token Rotation Notification

**File:** `src/keycloak-token-subscribers.ts`

When Keycloak refreshes the access token, React components need to know so they can update their `Authorization` headers:

```typescript
const listeners = new Set<() => void>();

export function subscribeKeycloakTokenRotation(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function notifyKeycloakTokenRotated(): void {
  for (const listener of listeners) listener();
}
```

Connected in `src/keycloak.ts`:
```typescript
keycloak.onAuthRefreshSuccess = () => notifyKeycloakTokenRotated();
```

### Role-Based Navigation

**File:** `src/shell/AuthorizedShellGate.tsx`
```typescript
const navigationItems = useMemo(
  () => filterNavByRoles(buildPrimaryShellNav(), context.roles),
  [context.roles]
);
```

Navigation items are filtered by the user's roles from the JWT. Users without the required role won't see certain nav items.

---

## 5. Backend Authorization — Hub-BFF

**Location:** `april-worker-kilo/hub-bff/`
**Tech:** Go (stdlib `net/http`, no framework)

### Global Middleware Chain

**File:** `cmd/hub-bff/main.go` (line 127-130)

```
Request arrives →
  Rate Limiter →        // Redis-backed, tenant-aware
  Cache →               // Redis-backed, tenant-scoped keys
  CORS →                // Origin whitelist, preflight caching
  Metadata →            // X-Request-Id, X-Correlation-Id generation
  Access Log →          // Structured JSON logging
  ServeMux              // Route to handler
```

### Per-Endpoint Middleware Chain

For protected routes:

```
ServeMux →
  AllowedMethods →     // HTTP method restriction (GET only, etc.)
  am.Validate →        // JWT validation (signature, iss, aud, exp)
  RequireAnyRole →     // RBAC check (user/admin)
  Handler              // Business logic
```

### JWT Validation Middleware

**File:** `internal/auth/middleware.go`

```go
type Middleware struct {
    issuer   string           // KEYCLOAK_ISSUER
    audience string           // KEYCLOAK_AUDIENCE
    jwks     keyfunc.Keyfunc  // Loaded from KEYCLOAK_JWKS_URL
}

func NewMiddleware(issuer, audience, jwksURL string) (*Middleware, error) {
    // Retries 60 times (2s intervals) to wait for Keycloak to start
    jwks, err = keyfunc.NewDefaultCtx(context.Background(), []string{jwksURL})
}

func (m *Middleware) Validate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract Bearer token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            writeError(w, r, 401, "unauthorized", "missing bearer token")
            return
        }

        // 2. Parse JWT with claims validation
        claims := &Claims{}
        _, err := jwt.ParseWithClaims(tokenValue, claims, m.jwks.Keyfunc,
            jwt.WithIssuer(m.issuer),                        // Check iss
            jwt.WithValidMethods([]string{"RS256","RS384","RS512"}),
            jwt.WithExpirationRequired(),                    // Check exp
        )
        if err != nil {
            writeError(w, r, 401, "unauthorized", "invalid token")
            return
        }

        // 3. Check audience (aud OR azp claim)
        if m.audience != "" && !slices.Contains(claims.Audience, m.audience) && claims.AZP != m.audience {
            writeError(w, r, 401, "unauthorized", "invalid audience")
            return
        }

        // 4. Store claims in context for downstream middleware/handlers
        ctx := context.WithValue(r.Context(), claimsContextKey, claims)

        // 5. Debug log
        slog.Debug("hub-bff auth", "user", claims.Subject, "tenant_id", claims.TenantID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Claims Structure

```go
type Claims struct {
    AZP               string      `json:"azp"`
    TenantID          string      `json:"tenant_id"`
    PreferredUsername string      `json:"preferred_username"`
    Email             string      `json:"email"`
    GivenName         string      `json:"given_name"`
    FamilyName        string      `json:"family_name"`
    Name              string      `json:"name"`
    RealmAccess       realmAccess `json:"realm_access"`
    jwt.RegisteredClaims                        // sub, iss, aud, exp, nbf, iat, jti
}

type realmAccess struct {
    Roles []string `json:"roles"`
}
```

### RBAC Middleware

```go
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := ClaimsFromContext(r.Context())
            if !ok {
                writeError(w, r, 401, "unauthorized", "missing auth context")
                return
            }

            // Check if ANY role in claims matches ANY required role
            for _, role := range claims.RealmAccess.Roles {
                if slices.Contains(roles, role) {
                    next.ServeHTTP(w, r)
                    return
                }
            }

            writeError(w, r, 403, "forbidden", "insufficient role")
        })
    }
}
```

### HTTP Method Restriction

**File:** `internal/http/methodrestriction.go`

```go
func AllowedMethods(allowed []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !slices.Contains(allowed, r.Method) {
                w.Header().Set("Allow", strings.Join(allowed, ", "))
                w.WriteHeader(405)
                // Returns: {"code":"method_not_allowed","message":"метод не разрешён"}
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### Route Registration

**File:** `cmd/hub-bff/main.go`

```go
mux := http.NewServeMux()

// PUBLIC — no auth
mux.Handle("/healthz", AllowedMethods([]string{"GET"})(health.Healthz))
mux.Handle("/livez", AllowedMethods([]string{"GET"})(health.Livez))
mux.Handle("/readyz", AllowedMethods([]string{"GET"})(health.Readyz))
mux.Handle("/metrics", AllowedMethods([]string{"GET"})(promhttp.Handler()))

// PROTECTED — JWT + RBAC
mux.Handle("/api/v1/overview",
    AllowedMethods([]string{"GET"})(
        am.Validate(auth.RequireAnyRole("user", "admin")(handlers.Overview)),
    ),
)
mux.Handle("/api/v1/me",
    AllowedMethods([]string{"GET"})(
        am.Validate(auth.RequireAnyRole("user", "admin")(httpapi.Me)),
    ),
)

// ADMIN ONLY
mux.Handle("/api/v1/admin/ping",
    AllowedMethods([]string{"GET"})(
        am.Validate(auth.RequireAnyRole("admin")(httpapi.AdminPing)),
    ),
)
mux.Handle("/api/v1/admin/profile/",
    AllowedMethods([]string{"GET","PUT","PATCH","DELETE"})(
        am.Validate(auth.RequireAnyRole("admin")(profileAdminProxy)),
    ),
)

// PUBLIC (CSP reports)
mux.Handle("/api/v1/csp-report", AllowedMethods([]string{"POST"})(httpapi.CSReport))
```

### Full Route Table

| Route | Methods | Auth | Role | Notes |
|---|---|---|---|---|
| `/healthz` | GET | None | — | Liveness probe |
| `/livez` | GET | None | — | Liveness probe |
| `/readyz` | GET | None | — | Readiness probe (checks Redis) |
| `/metrics` | GET | None (IP whitelist on Nginx) | — | Prometheus metrics |
| `/api/v1/overview` | GET | JWT | user, admin | Dashboard overview |
| `/api/v1/aggregation/dashboard` | GET | JWT | user, admin | Full dashboard data |
| `/api/v1/aggregation/home` | GET | JWT | user, admin | Home page data |
| `/api/v1/aggregation/summary` | GET | JWT | user, admin | Summary data |
| `/api/v1/me` | GET | JWT | user, admin | Bootstrap: returns user profile |
| `/api/v1/admin/ping` | GET | JWT | admin | Admin health check |
| `/api/v1/admin/profile/` | GET,PUT,PATCH,DELETE | JWT | admin | Reverse proxy to Profile API |
| `/api/v1/csp-report` | POST | None | — | CSP violation reports |

### Profile Admin Proxy

**File:** `internal/http/profile_proxy.go`

Admin routes `/api/v1/admin/profile/*` are reverse-proxied to the Profile API:

```go
proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
proxy.Director = func(req *http.Request) {
    // Strip /api/v1/admin/profile prefix, normalize path
    // Forward: Authorization (transparent), X-Correlation-Id, X-Request-Id, X-Source-Service
}
```

The JWT `Authorization` header passes through transparently. The Profile API validates it independently.

### CORS Middleware

**File:** `internal/http/cors.go`

```go
func CORS(origins []string, next http.Handler) http.Handler {
    origin := r.Header.Get("Origin")
    if slices.Contains(origins, origin) {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Max-Age", "86400")     // Preflight cache 24h
        w.Header().Set("Access-Control-Expose-Headers", "X-Correlation-Id, X-Request-Id")
    }
    if r.Method == "OPTIONS" {
        w.WriteHeader(204)
        return
    }
}
```

### Environment Variables

| Variable | Default | Purpose |
|---|---|---|
| `KEYCLOAK_ISSUER` | `https://dev.april.ukituki.tech/auth/realms/april` | Expected `iss` claim in JWT |
| `KEYCLOAK_AUDIENCE` | `aprilhub-shell` | Expected `aud` or `azp` claim |
| `KEYCLOAK_JWKS_URL` | `http://keycloak:8080/auth/realms/april/protocol/openid-connect/certs` | JWKS endpoint URL |
| `HUB_BFF_PORT` | `8081` | Listen port |
| `REDIS_PASSWORD` | *(empty)* | Redis password |
| `APRIL_PROFILE_ADMIN_URL` | *(empty)* | Profile API URL for admin proxy |

---

## 6. Backend Authorization — Profile API

**Location:** `april-profile-kilo/internal/`
**Tech:** Go (stdlib `net/http`)

### Key Difference from Hub-BFF

Profile API has **stricter** JWT validation:
- Only accepts **RS256** (not RS384/RS512)
- **Requires `tenant_id` claim** — returns 403 if missing
- Uses `jwt.MapClaims` instead of typed struct (more flexible)

### JWT Validator

**File:** `internal/auth/jwt.go`

```go
type Validator struct {
    keyfunc keyfunc.Keyfunc
    issuer  string
    aud     string
    tenant  string  // claim name, default: "tenant_id"
}

func (v *Validator) ValidateBearer(ctx, authorizationHeader) (Principal, error) {
    // 1. Extract Bearer token
    raw, ok := bearerToken(authorizationHeader)
    if !ok → ErrMissingBearer

    // 2. Parse JWT (RS256 only!)
    parser := jwt.NewParser(
        jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
        jwt.WithExpirationRequired(),
    )
    _, err := parser.ParseWithClaims(raw, &claims, v.keyfunc.KeyfuncCtx(ctx))
    if err → ErrInvalidToken

    // 3. Check issuer
    if claims["iss"] != v.issuer → ErrInvalidToken

    // 4. Check audience (aud OR azp)
    if !audienceMatches(claims, v.aud) → ErrInvalidToken

    // 5. Check subject
    if claims["sub"] == "" → ErrInvalidToken

    // 6. CRITICAL: Check tenant_id (MUST be present!)
    if claims[v.tenant] == "" → ErrMissingTenantClaim  (→ 403)

    // 7. Extract realm roles
    roles := extractRealmRoles(claims)

    // 8. Return Principal
    return Principal{Subject: sub, TenantID: tenantID, RealmRoles: roles}, nil
}
```

### Principal Context

**File:** `internal/auth/principal.go`

```go
type Principal struct {
    Subject    string   // JWT sub claim
    TenantID   string   // JWT tenant_id claim (REQUIRED)
    RealmRoles []string // JWT realm_access.roles
}

// Context helpers:
func ContextWithPrincipal(parent, p) context.Context
func PrincipalFromContext(ctx) (Principal, bool)
func SubjectFromContext(ctx) string
func TenantIDFromContext(ctx) string      // Used by ALL handlers
func RealmRolesFromContext(ctx) []string   // Used by ABAC
```

**Every handler calls `auth.TenantIDFromContext(r.Context())` to get the tenant for DB queries.**

### Middleware

**File:** `internal/auth/middleware.go`

```go
func (v *Validator) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w, r) {
        p, err := v.ValidateBearer(r.Context(), r.Header.Get("Authorization"))
        if err != nil → writeAuthError(w, err)   // 401 or 403
        next.ServeHTTP(w, r.WithContext(ContextWithPrincipal(r.Context(), p)))
    })
}

func writeAuthError(w, err) {
    switch {
    case errors.Is(err, ErrMissingTenantClaim):
        403 {"code":"tenant_required","message":"authentication failed"}
    case errors.Is(err, ErrMissingBearer):
        401 {"code":"missing_bearer","message":"authentication failed"}
    default:
        401 {"code":"invalid_token","message":"authentication failed"}
    }
}
```

### RBAC Middleware

```go
func RequireRealmRole(role string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w, r) {
            if role == "" → next.ServeHTTP(w, r)  // Empty = skip (dev mode)
            for _, got := range RealmRolesFromContext(r.Context()) {
                if got == role → next.ServeHTTP(w, r)
            }
            // 403 {"code":"forbidden","message":"insufficient realm role"}
        })
    }
}
```

### ABAC — Attribute-Based Access Control

**File:** `internal/abac/policy.go`

An **additional layer** on top of RBAC, specific to Profile API. Filters document fields based on the user's roles:

```go
// Config format (JSON env var):
// {"default":["reader"],"hr":["reader","hr-manager"],"security":["security-admin"]}

type Policy struct {
    segmentRoles map[string][]string  // segment name → required roles
}

func (p *Policy) FilterSnapshot(snap, realmRoles) Snapshot {
    // Only keeps document segments where user has at least one required role
    // E.g., user with role "reader" sees "default" and "hr" segments,
    // but NOT "security" segment
}
```

Applied to these handlers:
- `GET /v1/entities/{entityID}` (current version)
- `GET /v1/entities/{entityID}/versions/{version}` (specific version)
- `GET /v1/external-mappings/{system}/{id}/entity` (external lookup)

### Route Registration

**File:** `internal/httpapi/server.go`

```go
func NewMux(v *auth.Validator, ...) http.Handler {
    mux := http.NewServeMux()

    // PUBLIC — no auth
    mux.HandleFunc("GET /healthz", handleHealthz)
    mux.Handle("GET /readyz", handleReadyz(readiness))
    mux.HandleFunc("GET /v1/system/ping", handlePing)
    mux.Handle("GET /metrics", prometheusMetricsHandler())

    // AUTHENTICATED — JWT required (tenant_id mandatory)
    mux.Handle("GET /v1/auth/whoami", v.Middleware(handleWhoAmI))
    mux.Handle("POST /v1/entity-types", v.Middleware(handleCreateEntityTypeDraft))
    mux.Handle("GET /v1/entity-types", v.Middleware(handleListEntityTypes))
    mux.Handle("GET /v1/entities", v.Middleware(handleListEntities))
    mux.Handle("GET /v1/entities/{id}", v.Middleware(handleGetEntityCurrent(service, abacPolicy)))
    // ... all CRUD endpoints ...

    // ADMIN — JWT + specific realm role
    adminChain := func(h http.Handler) http.Handler {
        return v.Middleware(auth.RequireRealmRole(adminRealmRole)(h))
    }
    mux.Handle("GET /v1/admin/profile-conflicts", adminChain(handleListProfileConflicts))
    mux.Handle("POST /v1/admin/profile-conflicts/{id}/resolve", adminChain(...))
    mux.Handle("POST /v1/admin/entities/merge", adminChain(handleMergeEntities))

    // Global middleware (applied to ALL routes)
    return withRequestLogging(withPrometheusHTTPMetrics(withCorrelationID(withRequestID(mux))))
}
```

### Global Middleware Chain

```
Request → RequestID → CorrelationID → Prometheus Metrics → RequestLogging → ServeMux
```

- **RequestID** — generates/forwards `X-Request-Id`
- **CorrelationID** — generates/forwards `X-Correlation-Id`
- **Prometheus Metrics** — HTTP duration/status histograms
- **RequestLogging** — structured JSON log with `tenant_id` when authenticated

### Environment Variables

| Variable | Purpose |
|---|---|
| `KEYCLOAK_JWKS_URL` | JWKS endpoint URL |
| `KEYCLOAK_ISSUER` | Expected `iss` in JWT |
| `KEYCLOAK_AUDIENCE` | Expected `aud`/`azp` (e.g., `april-profile-api`) |
| `KEYCLOAK_TENANT_CLAIM` | Claim name for tenant (default: `tenant_id`) |
| `KEYCLOAK_ADMIN_REALM_ROLE` | Role required for `/v1/admin/*` (empty = skip in dev) |
| `ABAC_SEGMENT_ACCESS_JSON` | JSON policy for field-level access control |

---

## 7. Multi-Tenancy Isolation

### How Tenant ID Flows

```
Keycloak user attribute "tenant_id"
    ↓ (protocol mapper)
JWT claim "tenant_id"
    ↓ (frontend receives token)
keycloak-js stores token in memory
    ↓ (frontend sends to backend)
Authorization: Bearer <JWT with tenant_id>
    ↓ (Go backend validates)
Claims.TenantID / Principal.TenantID
    ↓ (placed in context)
auth.TenantIDFromContext(r.Context())
    ↓ (used by handlers)
DB queries: WHERE tenant_id = $1
    ↓ (used by middleware)
Redis cache key: {tenantID}:{path}:{userID}:{query}
    ↓ (used by middleware)
Rate limit key: {tenantID}:{identifier}
```

### Isolation by Layer

| Layer | Mechanism | Status | File |
|---|---|---|---|
| **Database** | All tables have `tenant_id uuid NOT NULL REFERENCES tenants(id)` with indexes | ✅ Full | `vendor/april-profile/atlas/migrations/` |
| **Hub-BFF JWT** | `Claims.TenantID` extracted from JWT | ✅ Works | `hub-bff/internal/auth/middleware.go` |
| **Profile JWT** | `Principal.TenantID` required, 403 if missing | ✅ Mandatory | `april-profile/internal/auth/jwt.go` |
| **Redis Cache** | Key format: `{tenantID}:{path}:{userID}:{query}` | ✅ Tenant-scoped | `hub-bff/internal/middleware/cache.go` |
| **Rate Limiter** | Key format: `{tenantID}:{identifier}` | ✅ Tenant-scoped | `hub-bff/internal/middleware/ratelimiter.go` |
| **Logs** | `tenant_id` in structured logs after auth | ✅ Present | `hub-bff/internal/auth/middleware.go`, `april-profile/internal/httpapi/server.go` |
| **ABAC** | Segment filtering by realm roles | ✅ Present | `april-profile/internal/abac/policy.go` |

### Cross-Tenant Leakage Prevention

1. **DB-level:** PostgreSQL FK constraints + `tenant_id` on every table + indexed queries
2. **Cache-level:** `tenant_id` in Redis key prefix prevents cross-tenant cache hits
3. **Rate-limit-level:** `tenant_id` in rate limit key prevents cross-tenant limit sharing
4. **API-level:** Profile API rejects tokens without `tenant_id` (403)

---

## 8. Request Lifecycle

### Full Login + Access Flow

```
┌──────────────────────────────────────────────────────────────────────────┐
│  STEP 1: USER OPENS APPLICATION                                          │
│  Browser → Nginx (:80) → Hub-Shell (:4173)                               │
│  Result: React SPA loaded, keycloak-js initializes                       │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│  STEP 2: KEYCLOAK SSO CHECK                                             │
│  Hub-Shell → Nginx (/auth) → Keycloak (:8080)                           │
│  keycloak.init({ onLoad: "check-sso" })                                  │
│                                                                          │
│  If user has active Keycloak session:                                    │
│    → Keycloak returns access token (JWT)                                 │
│    → keycloak-js stores token in memory                                  │
│                                                                          │
│  If user has NO session:                                                 │
│    → keycloak.authenticated = false                                      │
│    → App renders GuestB2BLanding with "Login" button                     │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│  STEP 3: LOGIN FLOW (if needed)                                         │
│  User clicks "Login"                                                     │
│  keycloak.login() → redirect to Keycloak login page                      │
│  User enters credentials → Keycloak validates → redirects back           │
│  Keycloak issues: access token (JWT) + refreshes SSO cookie              │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│  STEP 4: BOOTSTRAP — LOAD USER PROFILE                                  │
│  Hub-Shell → authorizedFetch("/v1/me")                                   │
│    → Authorization: Bearer <JWT>                                         │
│    → Nginx (/api/) → Hub-BFF (:8081)                                     │
│                                                                          │
│  Hub-BFF processes:                                                      │
│    1. AllowedMethods("GET") — check method                               │
│    2. am.Validate():                                                     │
│       a. Extract Bearer token                                            │
│       b. Verify signature via JWKS (RS256/384/512)                       │
│       c. Check iss matches KEYCLOAK_ISSUER                               │
│       d. Check aud/azp matches KEYCLOAK_AUDIENCE                         │
│       e. Check exp (not expired)                                         │
│       f. Store Claims in context                                         │
│    3. RequireAnyRole("user","admin"):                                     │
│       a. Read claims.RealmAccess.Roles from context                      │
│       b. Check if ANY role matches ["user", "admin"]                     │
│       c. If no match → 403                                               │
│    4. Me() handler → returns {sub, username, email, name, roles}         │
│                                                                          │
│  Response: 200 { sub, username, email, name, roles }                     │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│  STEP 5: RENDER AUTHORIZED SHELL                                        │
│  Hub-Shell receives user profile                                         │
│  → buildShellUserContext(me) → {user, roles, orgScope, correlationId}    │
│  → filterNavByRoles(navItems, roles)                                     │
│  → Render AuthorizedShellGate with full navigation                       │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│  STEP 6: SUBSEQUENT API REQUESTS                                        │
│  Hub-Shell → authorizedFetch("/api/v1/aggregation/dashboard")            │
│    → Authorization: Bearer <JWT> (current token from keycloak-js)        │
│    → Same middleware chain as Step 4                                     │
│                                                                          │
│  If token expired during request:                                        │
│    → Hub-BFF returns 401                                                 │
│    → authorizedFetch catches 401 → keycloak.updateToken(30)              │
│    → If refresh succeeds → retry request with new token (once)           │
│    → If refresh fails → keycloak.login() → full re-auth                  │
└──────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────┐
│  STEP 7: PROFILE API REQUESTS (via Hub-BFF admin proxy)                 │
│  Hub-BFF admin proxy → Profile API                                       │
│    → Authorization header passes through transparently                   │
│    → Profile API re-validates JWT independently:                         │
│       1. ValidateBearer():                                               │
│          a. RS256 signature via JWKS                                     │
│          b. Check iss, aud/azp, exp, sub                                 │
│          c. CHECK tenant_id present → 403 if missing                     │
│          d. Extract Principal{Subject, TenantID, RealmRoles}             │
│       2. v.Middleware → stores Principal in context                      │
│       3. RequireRealmRole("admin") for admin routes                      │
│       4. Handler uses TenantIDFromContext() for ALL DB queries           │
│       5. ABAC policy filters document segments by roles (if active)      │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 9. Error Handling Matrix

### Hub-BFF Errors

| Condition | HTTP Status | App Code | Message | Frontend Response |
|---|---|---|---|---|
| No `Authorization` header | 401 | `unauthorized` | `missing bearer token` | Trigger login flow |
| Invalid JWT signature | 401 | `unauthorized` | `invalid token` | Silent refresh → re-login |
| Token expired | 401 | `unauthorized` | `invalid token` | Silent refresh → re-login |
| Wrong issuer | 401 | `unauthorized` | `invalid token` | Re-login |
| Wrong audience | 401 | `unauthorized` | `invalid audience` | Re-login |
| Missing auth context | 401 | `unauthorized` | `missing auth context` | Re-login |
| Insufficient role | 403 | `forbidden` | `insufficient role` | Show forbidden zone |
| Wrong HTTP method | 405 | `method_not_allowed` | `метод не разрешён` | Show error |
| Profile proxy error | 502 | `proxy_error` | `profile admin upstream is unavailable` | Show error |

Error response format:
```json
{
  "code": "unauthorized",
  "message": "invalid token",
  "metadata": {
    "correlationId": "abc-123",
    "requestId": "def-456",
    "sourceService": "hub-bff"
  }
}
```

### Profile API Errors

| Condition | HTTP Status | App Code | Message |
|---|---|---|---|
| No `Authorization` header | 401 | `missing_bearer` | `authentication failed` |
| Invalid JWT | 401 | `invalid_token` | `authentication failed` |
| Missing `tenant_id` claim | 403 | `tenant_required` | `authentication failed` |
| Insufficient realm role | 403 | `forbidden` | `insufficient realm role` |
| Entity not found | 404 | `entity_not_found` | *(domain error)* |
| Version not found | 404 | `version_not_found` | *(domain error)* |
| Invalid request body | 400 | `invalid_request` | *(domain error)* |
| Schema validation failed | 422 | `schema_validation_failed` | *(domain error with issues array)* |
| Entity type not published | 409 | `entity_type_not_published` | *(domain error)* |
| Draft version conflict | 409 | `draft_version_conflict` | *(domain error)* |

Error response format:
```json
{
  "code": "entity_not_found",
  "message": "entity not found",
  "request_id": "abc-123"
}
```

---

## 10. Security Headers and Policies

### Nginx-Level

| Header | Value | Applied To |
|---|---|---|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | All locations |
| `X-Content-Type-Options` | `nosniff` | All locations |
| `X-XSS-Protection` | `0` | All locations |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | All locations |
| `Permissions-Policy` | `geolocation=(), microphone=(), camera=()` | All locations |

### Content Security Policy

| Location | CSP | Purpose |
|---|---|---|
| Hub Shell | `default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self' https://dev.april.ukituki.tech https://sentry.io; frame-src 'self' https://dev.april.ukituki.tech/auth; report-uri /api/v1/csp-report` | Allow SPA execution, Keycloak iframe, Sentry |
| API | `default-src 'none'` | Block all content (JSON only) |
| Keycloak | (inherits global) + `X-Frame-Options: DENY` | Prevent Keycloak pages from being framed |

### Frame Protection

| Location | X-Frame-Options | Purpose |
|---|---|---|
| Hub Shell | `SAMEORIGIN` | Allow self-embedding |
| API | `DENY` | Never frame JSON responses |
| Keycloak | `DENY` | Prevent clickjacking on login |

---

## 11. Observability

### Auth Logs (structured JSON)

**Hub-BFF — successful auth:**
```json
{
  "level": "DEBUG",
  "msg": "hub-bff auth",
  "event": "auth_success",
  "path": "/api/v1/me",
  "user": "subject-123",
  "tenant_id": "00000000-..."
}
```

**Hub-BFF — failed auth:**
```json
{
  "level": "WARN",
  "msg": "hub-bff auth failed",
  "event": "auth_error",
  "path": "/api/v1/me",
  "method": "GET",
  "status": 401,
  "reason": "invalid token",
  "correlationId": "abc",
  "requestId": "def",
  "sourceService": "hub-bff"
}
```

**Profile API — request log (includes tenant):**
```json
{
  "level": "INFO",
  "msg": "http request",
  "method": "GET",
  "path": "/v1/entities/123",
  "status": 200,
  "duration_ms": 15,
  "requestId": "abc",
  "correlationId": "def",
  "tenant_id": "00000000-..."
}
```

### Prometheus Metrics

| Metric | Labels | Purpose |
|---|---|---|
| Auth error counter | `path`, `reason`, `status_code` | Track auth failures |
| HTTP duration histogram | `method`, `path`, `status_code` | Request latency |
| Cache hit/miss | `path` | Cache effectiveness |

### Log Collection

```
Go service (JSON stdout) → Docker → Promtail → Loki → Grafana
```

---

## 12. File Index

### Configuration

| File | Purpose |
|---|---|
| `april-worker-kilo/infra/keycloak/realm/april-realm.json` | Keycloak realm config (roles, clients, users, mappers) |
| `april-worker-kilo/infra/nginx/aprilhub.conf` | Production Nginx config |
| `april-worker-kilo/infra/nginx/default.conf` | Dev Nginx config |
| `april-worker-kilo/docker-compose.yml` | Service orchestration (Keycloak, BFF, Shell) |

### Keycloak Setup Scripts

| File | Purpose |
|---|---|
| `april-worker-kilo/scripts/keycloak-april-user-profile-allow-tenant-attr.sh` | Enable custom user attributes |
| `april-worker-kilo/scripts/keycloak-ensure-april-theme.sh` | Apply login/account theme |

### Frontend Auth (Hub-Shell)

| File | Purpose |
|---|---|
| `april-worker-kilo/hub-shell/src/main.tsx` | App entry, auth initialization |
| `april-worker-kilo/hub-shell/src/keycloak.ts` | Keycloak instance, init config |
| `april-worker-kilo/hub-shell/src/auth.ts` | Auth config from env vars |
| `april-worker-kilo/hub-shell/src/api.ts` | API client with token handling, 401 retry |
| `april-worker-kilo/hub-shell/src/keycloak-token-subscribers.ts` | Token rotation notification |
| `april-worker-kilo/hub-shell/src/App.tsx` | Auth state machine (4 zones) |
| `april-worker-kilo/hub-shell/src/user-context.ts` | ShellUserContext builder |
| `april-worker-kilo/hub-shell/src/shell/AuthorizedShellGate.tsx` | Authorized shell with role-filtered nav |

### Hub-BFF (Go Backend)

| File | Purpose |
|---|---|
| `april-worker-kilo/hub-bff/cmd/hub-bff/main.go` | Server setup, middleware chain, route registration |
| `april-worker-kilo/hub-bff/internal/auth/middleware.go` | JWT validation, Claims, RequireAnyRole |
| `april-worker-kilo/hub-bff/internal/http/methodrestriction.go` | HTTP method restriction middleware |
| `april-worker-kilo/hub-bff/internal/http/cors.go` | CORS middleware with origin whitelist |
| `april-worker-kilo/hub-bff/internal/http/handlers.go` | Business logic handlers, Me() endpoint |
| `april-worker-kilo/hub-bff/internal/http/profile_proxy.go` | Reverse proxy to Profile API |

### Profile API (Go Backend)

| File | Purpose |
|---|---|
| `april-profile-kilo/internal/app/run.go` | Server setup, validator init |
| `april-profile-kilo/internal/auth/jwt.go` | JWT validator (strict, requires tenant_id) |
| `april-profile-kilo/internal/auth/middleware.go` | Auth middleware, RequireRealmRole |
| `april-profile-kilo/internal/auth/principal.go` | Principal struct, context helpers |
| `april-profile-kilo/internal/abac/policy.go` | ABAC policy (field-level access by role) |
| `april-profile-kilo/internal/httpapi/server.go` | Route registration, global middleware, handlers |

### Documentation

| File | Purpose |
|---|---|
| `april-worker-kilo/docs/auth-jwt-keycloak-adapted.md` | Auth architecture doc (flow diagrams, state machine) |
| `april-profile-kilo/docs/keycloak-stand-coordinates.md` | Keycloak integration guide |
| `april-space/docs/adr/ADR-0005-tenant-isolation-audit.md` | Tenant isolation audit |

---

## Quick Reference Checklist

When modifying auth-related code, verify:

- [ ] JWT validation checks: signature (JWKS), `iss`, `aud`/`azp`, `exp`
- [ ] RBAC check matches the endpoint's required roles
- [ ] `tenant_id` flows correctly through context
- [ ] Error responses include `correlationId` and `requestId`
- [ ] Auth errors are logged with structured JSON (no token leakage)
- [ ] CORS allows `Authorization` header
- [ ] Nginx security headers are present
- [ ] ABAC policy is applied to GET profile endpoints (Profile API)
- [ ] Rate limiter and cache keys include `tenant_id` prefix
