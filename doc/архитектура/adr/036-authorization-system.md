# ADR-036: Система авторизации — адаптация April Ecosystem → LKFL

**Статус:** Accepted
**Дата:** 2026-05-30
**Контекст:** M17 — полная реализация авторизации (задачи T1701–T1705)
**Связанные:** ADR-003 (Keycloak), ADR-009 (Multi-tenancy), ADR-035 (Integration Proxy)
**Источник:** April Ecosystem AUTHORIZATION_REFERENCE.md (адаптировано)

---

## Контекст

LKFL — white-label multi-tenant платформа, требующая:
- Realm per tenant (ADR-003) — каждый tenant имеет отдельный Keycloak realm
- Динамический JWKS — tenant resolver ДО JWT-валидации
- Stateless backend — бэкенд не эмитит JWT, только валидирует
- ФСТЭК + 152-ФЗ — audit trail, structured logging, tenant isolation

Из проекта April Ecosystem взята проверенная система авторизации (Keycloak OIDC, JWT middleware, RBAC, фронтенд state machine) и адаптирована под multi-tenant модель LKFL.

---

## Ключевые отличия от April

| Аспект | April Ecosystem | LKFL |
|--------|----------------|------|
| **Multi-tenancy** | Один realm, `tenant_id` user attribute | **Realm per tenant** — динамический JWKS |
| **Tenant resolution** | Не требуется (один realm) | **Tenant Resolver middleware** (subdomain/host → tenant_id) ДО JWT валидации |
| **Роли** | `user`, `manager`, `admin` (3) | `employee`, `hr`, `catalog_manager`, `admin` (4+) |
| **ABAC** | Field-level document segments по ролям | **CEL Rule Engine** (ADR-021) — eligibility, billing, flow, gamification |
| **Frontend data fetching** | `authorizedFetch()` — кастомный fetch | **React Query** (ADR-031) + fetch wrapper с token rotation |
| **Integration** | Profile API — прямой HTTP | **Integration Proxy** (ADR-035) — gRPC localhost |

---

## Решение

### Архитектура

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Nginx (:80)                                   │
│  / → lkfl-frontend:4173                                             │
│  /auth → keycloak:8080                                              │
│  /api/ → lkfl-server:8080                                           │
│  /healthz → lkfl-server:8080                                        │
│  /metrics → lkfl-server:8080 (IP whitelist)                         │
└──────┬──────────────────────┬──────────────────────────┬────────────┘
       │                      │                          │
       ▼                      ▼                          ▼
┌──────────────┐    ┌──────────────────┐     ┌──────────────────────┐
│ LKFL Frontend│    │ lkfl-server      │     │ Keycloak             │
│ React+Vite   │───►│ :8080            │     │ OIDC IdP             │
│ keycloak-js  │    │ TenantResolver   │     │ Realm per tenant     │
│ React Query  │    │ JWT Validate     │     │ JWKS per realm       │
│ Zustand      │    │ RBAC Guard       │     │                      │
└──────────────┘    └──────┬───────────┘     └──────────────────────┘
                           │ gRPC localhost
                           ▼
                  ┌──────────────────┐
                  │ Integration Proxy│
                  │ :8090 gRPC       │
                  │ :8091 webhooks   │
                  └──────────────────┘
```

### Три слоя авторизации

| Слой | Механизм | Реализация |
|------|----------|------------|
| **Tenant Resolution** | subdomain/host → tenant_id → realm | `TenantResolver` middleware |
| **Authentication** | JWT signature via JWKS + claims | `shared/pkg/auth/verifier.go` |
| **Authorization (RBAC)** | Realm roles из Keycloak | `RBACGuard` middleware |

> **ABAC отсутствует** — вместо него CEL Engine (ADR-021) для бизнес-правил (eligibility, billing, flow, gamification).

---

## Backend — Go

### Tenant Resolver

**Первый middleware в цепи.** Определяет tenant из запроса:

```go
// shared/pkg/auth/tenantresolver.go

type TenantResolver struct {
    repo TenantRepository  // tenants → realm_name, keycloak_url
    cache *RedisCache      // tenant resolution cache (TTL 5min)
}

func (r *TenantResolver) Resolve(ctx context.Context, req *http.Request) (*Tenant, error) {
    // 1. Host/subdomain → slug: "sdek.lkfl.ru" → "sdek"
    // 2. X-Tenant header (internal requests)
    // 3. Cache lookup: tenant:{slug}
    // 4. DB query: SELECT * FROM tenants WHERE slug = $1
    // 5. Store in context
}
```

Request lifecycle:
```
Request → TenantResolver → JWT Validate (tenant-specific JWKS) → RBAC Guard → Handler
```

### JWT Validation Middleware

**File:** `shared/pkg/auth/verifier.go`

```go
type Verifier struct {
    // Per-tenant JWKS cache
    jwksCache map[string]*keyfunc.JWKS  // keyed by realm name
    // Tenant → JWKS URL builder
    jwksURLBuilder func(tenant *Tenant) string
}

func (v *Verifier) Validate(ctx context.Context, token string, tenant *Tenant) (*Claims, error) {
    // 1. Extract Bearer token
    // 2. Build JWKS URL: tenant.KeycloakURL + "/realms/" + tenant.RealmName + "/protocol/openid-connect/certs"
    // 3. Load/cache JWKS for realm
    // 4. Parse JWT: RS256/384/512, check iss, exp, azp
    // 5. Return Claims with tenant context
}
```

### Claims structure

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
    jwt.RegisteredClaims  // sub, iss, aud, exp, nbf, iat, jti
}

type realmAccess struct {
    Roles []string `json:"roles"`
}
```

### RBAC Middleware

```go
func RBACGuard(requiredRoles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := ClaimsFromContext(r.Context())
            
            for _, role := range claims.RealmAccess.Roles {
                if slices.Contains(requiredRoles, role) {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            
            writeError(w, r, 403, "forbidden", "insufficient role")
        })
    }
}
```

### Middleware chain

```
Request →
  TenantResolver →      // subdomain → tenant_id → realm
  JWT Validate →        // RS256 JWKS validation per realm
  RBAC Guard →          // realm roles check
  Handler               // business logic
```

### Route registration pattern

```go
mux := http.NewServeMux()

// PUBLIC — no auth
mux.Handle("GET /healthz", health.Healthz)
mux.Handle("GET /livez", health.Livez)
mux.Handle("GET /readyz", health.Readyz)
mux.Handle("GET /metrics", promhttp.Handler())

// PROTECTED — JWT + RBAC (employee+admin)
mux.Handle("GET /api/v1/me",
    v.Middleware(RBACGuard("employee", "hr", "catalog_manager", "admin")(handlers.Me)),
)
mux.Handle("GET /api/v1/catalog",
    v.Middleware(RBACGuard("employee", "hr", "catalog_manager", "admin")(handlers.Catalog)),
)

// ADMIN HR — JWT + admin role
mux.Handle("GET /api/v1/admin/users",
    v.Middleware(RBACGuard("hr", "admin")(handlers.AdminUsers)),
)

// ADMIN CATALOG — JWT + catalog_manager role
mux.Handle("GET /api/v1/admin/catalog",
    v.Middleware(RBACGuard("catalog_manager", "admin")(handlers.AdminCatalog)),
)
```

### Error handling

Единый формат ответа:

```json
{
  "code": "unauthorized",
  "message": "invalid token",
  "metadata": {
    "correlationId": "abc-123",
    "requestId": "def-456",
    "sourceService": "lkfl-server"
  }
}
```

| Условие | HTTP | Code | Message |
|---------|------|------|---------|
| Нет `Authorization` header | 401 | `missing_bearer` | `authentication failed` |
| Не найден tenant | 404 | `tenant_not_found` | `tenant resolution failed` |
| Неверная подпись JWT | 401 | `invalid_token` | `authentication failed` |
| Токен истёк | 401 | `token_expired` | `authentication failed` |
| Неправильный issuer | 401 | `invalid_issuer` | `authentication failed` |
| Недостаточно роли | 403 | `forbidden` | `insufficient role` |
| Неправильный HTTP метод | 405 | `method_not_allowed` | `method not allowed` |

---

## Frontend — React

### Tenant resolution (frontend)

Frontend определяет tenant realm ДО инициализации keycloak-js. Механизм:

```
Browser → https://sdek.lkfl.ru
  ↓
window.location.hostname = "sdek.lkfl.ru"
  ↓
extractTenantSlug("sdek.lkfl.ru") → "sdek"
  ↓
realm = extractTenantSlug(hostname)  // subdomain = realm name
  ↓
keycloak.init({ realm: "sdek" })
```

```typescript
// src/utils/tenant.ts

/**
 * Извлекает slug tenant из subdomain.
 * "sdek.lkfl.ru" → "sdek"
 * "lkfl.ru" → "demo" (fallback для dev/localhost)
 * "localhost:5173" → "demo" (fallback для dev)
 */
export function extractTenantSlug(hostname: string): string {
    // Dev fallback: localhost / root domain → "demo"
    if (hostname === 'localhost' || hostname === 'localhost:5173' || hostname.endsWith('.lkfl.ru')) {
        const parts = hostname.split('.');
        // "sdek.lkfl.ru" → parts[0] = "sdek"
        if (parts.length >= 3 && parts[0] !== 'www') {
            return parts[0];
        }
    }
    return 'demo';
}

/**
 * Строит Keycloak config из tenant slug.
 * В dev: относительный путь "/auth" (Nginx proxy → keycloak:8080)
 * В prod: абсолютный URL из env
 */
export function buildKeycloakConfig(slug: string): { url: string; realm: string; clientId: string } {
    const url = import.meta.env.VITE_KEYCLOAK_URL || '/auth';
    // Realm name = tenant slug (realm per tenant)
    const realm = slug;
    const clientId = 'lkfl-frontend';

    return { url, realm, clientId };
}
```

```typescript
// src/keycloak.ts
import { extractTenantSlug } from './utils/tenant';
import { buildKeycloakConfig } from './utils/tenant';

let keycloak: KeycloakInstance;

export function initKeycloak(): KeycloakInstance {
    const slug = extractTenantSlug(window.location.hostname);
    const config = buildKeycloakConfig(slug);

    keycloak = new Keycloak({
        url: config.url,
        realm: config.realm,
        clientId: config.clientId,
    });

    // Token rotation notification for React Query
    keycloak.onAuthRefreshSuccess = () => {
        notifyTokenRotated();
    };

    // Auto-refresh on expiry
    keycloak.onTokenExpired = () => {
        void keycloak.updateToken(70).catch(() => void keycloak.login());
    };

    return keycloak;
}
```

### Keycloak initialization

```typescript
// src/keycloak.ts (упрощено — config берётся из tenant resolution)
import { KeycloakInstance, Keycloak } from 'keycloak-js';

let keycloak: KeycloakInstance;

// initKeycloak() без параметров — realm определяется из window.location.hostname
// → extractTenantSlug() → buildKeycloakConfig()
// Полный код см. выше в §Tenant resolution
```

### Auth state machine

```
┌─────────┐    login()     ┌──────────────┐  GET /api/v1/me ┌──────────────┐
│  guest  │───────────────►│  transition  │────────────────►│  authorized  │
└─────────┘                └──────────────┘                 └──────────────┘
                                    │                              │
                                    │ 403                          │ me == null
                                    ▼                              ▼
                               ┌──────────────┐             ┌──────────────┐
                               │  forbidden   │             │  no-context  │
                               └──────────────┘             └──────────────┘
```

| Зона | Условие | UI |
|------|---------|-----|
| `guest` | `!keycloak.authenticated` | Landing + кнопка входа |
| `transition` | `keycloak.authenticated`, ждём `/api/v1/me` | Loading |
| `authorized` | `/api/v1/me` → 200 + profile loaded | Полная навигация + role filtering |
| `forbidden` | `/api/v1/me` → 403 | «Доступ запрещён» + logout |
| `no-context` | `authorized` но `me == null` | «Профиль не загружен» + logout |

### API layer — authorized fetch

```typescript
// src/api/platform.ts
export async function authorizedFetch(input: RequestInfo, init?: RequestInit): Promise<Response> {
    // Step 1: Ensure we have a token
    if (!keycloak.authenticated || !keycloak.token) {
        await keycloak.login();
        throw new Error('Unauthenticated session');
    }

    // Step 2: Attach Bearer token + telemetry headers
    const headers = new Headers(init?.headers);
    headers.set('Authorization', `Bearer ${keycloak.token}`);
    
    const response = await fetch(input, { ...init, headers });

    // Step 3: Handle 401 — automatic token refresh + one retry
    if (response.status === 401) {
        const refreshed = await keycloak.updateToken(30);
        if (!refreshed) {
            await keycloak.login();
            throw new Error('Token refresh failed');
        }
        notifyTokenRotated();
        // Retry once with new token
        headers.set('Authorization', `Bearer ${keycloak.token}`);
        return fetch(input, { ...init, headers });
    }

    return response;
}
```

### React Query integration (ADR-031)

```typescript
// Token rotation → invalidate React Query cache
import { invalidateQueries } from '../api/queryClient';

export function notifyTokenRotated(): void {
    // Revalidate all queries that use Authorization header
    invalidateQueries();
}
```

### Zustand auth store

```typescript
// store/auth.ts
interface AuthState {
    user: User | null;
    roles: string[];
    tenant: Tenant | null;
    status: 'guest' | 'transition' | 'authorized' | 'forbidden' | 'no-context';
    fetchMe: () => Promise<void>;
    logout: () => void;
}
```

### Role-based navigation

```typescript
// Filter nav items by user roles
const navigationItems = useMemo(
    () => filterNavByRoles(buildNavigation(), context.roles),
    [context.roles]
);
```

---

## Keycloak — tenant realm template

Каждый tenant получает свой realm. Шаблон:

```json
{
  "realm": "{{TENANT_SLUG}}",
  "displayName": "{{TENANT_NAME}}",
  "enabled": true,
  "roles": {
    "realm": [
      { "name": "employee", "description": "Сотрудник — базовый доступ" },
      { "name": "hr", "description": "HR — управление пользователями" },
      { "name": "catalog_manager", "description": "Менеджер каталога льгот" },
      { "name": "admin", "description": "Полный доступ" }
    ]
  },
  "clients": [
    {
      "clientId": "lkfl-frontend",
      "publicClient": true,
      "protocol": "openid-connect",
      "standardFlowEnabled": true,
      "directAccessGrantsEnabled": true,
      "redirectUris": ["{{FRONTEND_URL}}/*"],
      "attributes": {
        "pkce.code.challenge.method": "S256"
      }
    },
    {
      "clientId": "lkfl-server",
      "publicClient": false,
      "serviceAccountsEnabled": true,
      "secret": "{{SERVER_CLIENT_SECRET}}",
      "attributes": {
        "access.token.lifespan": "300"
      }
    }
  ],
  "protocolMappers": [
    {
      "name": "tenant_id",
      "protocolMapper": "oidc-usermodel-attribute-mapper",
      "config": {
        "user.attribute": "tenant_id",
        "claim.name": "tenant_id",
        "jsonType.label": "String",
        "access.token.claim": "true",
        "id.token.claim": "true"
      }
    }
  ]
}
```

---

## Multi-tenant isolation

### How tenant ID flows

```
Keycloak realm (per tenant)
    ↓ (login → JWT with tenant_id from mapper)
JWT claim "tenant_id"
    ↓ (frontend → keycloak-js → memory)
Authorization: Bearer <JWT>
    ↓ (Go backend → TenantResolver)
tenant_id from context (subdomain → realm → tenant)
    ↓ (JWT Validate with tenant-specific JWKS)
Claims validated
    ↓ (RBAC Guard)
Roles checked against realm_access.roles
    ↓ (Handler)
DB queries: WHERE tenant_id = $1
    ↓ (Cache)
Redis key: {tenant_id}:{path}:{user_id}:{query}
    ↓ (Rate limit)
Redis key: {tenant_id}:{identifier}
```

### Isolation by layer

| Layer | Mechanism | Status |
|-------|-----------|--------|
| **Keycloak** | Realm per tenant | ✅ |
| **Database** | `tenant_id` FK constraint + indexes | ✅ |
| **JWT** | Tenant-specific JWKS validation | ✅ |
| **Redis cache** | `{tenant_id}:` prefix in keys | ✅ |
| **Rate limiter** | `{tenant_id}:` prefix in keys | ✅ |
| **Logs** | `tenant_id` in structured JSON | ✅ |

### Cross-tenant leakage prevention

1. **Keycloak:** separate realm → separate users, separate JWKS
2. **DB:** `tenant_id` FK on every table + indexed queries
3. **Cache:** `tenant_id` in Redis key prefix
4. **Rate limit:** `tenant_id` in rate limit key
5. **API:** TenantResolver rejects unknown tenants with 404

---

## Request lifecycle

### Полный цикл авторизации

```
STEP 1: USER OPENS APPLICATION
  Browser → Nginx (:80) → lkfl-frontend:4173
  Result: React SPA loaded, keycloak-js initializes

STEP 2: TENANT RESOLUTION (frontend)
  Subdomain → tenant slug: "sdek.lkfl.ru" → "sdek"
  Keycloak URL: "/auth" (via Nginx → keycloak:8080)
  Realm: "sdek"
  Client: "lkfl-frontend"

STEP 3: KEYCLOAK SSO CHECK
  keycloak.init({ onLoad: "check-sso" })
  If active session → access token (JWT) with tenant_id claim
  If no session → guest zone, show login button

STEP 4: LOGIN FLOW (if needed)
  User clicks "Login"
  keycloak.login() → Keycloak login page for realm "sdek"
  User enters credentials → Keycloak validates → redirect back
  JWT issued with tenant_id claim

STEP 5: BOOTSTRAP — LOAD USER PROFILE
  frontend → authorizedFetch("/api/v1/me")
    Authorization: Bearer <JWT>
    Nginx (/api/) → lkfl-server:8080
    lkfl-server processes:
      1. TenantResolver: subdomain → tenant "sdek"
      2. JWT Validate: RS256 via "sdek" realm JWKS
      3. RBAC Guard: check roles ["employee", "hr", ...]
      4. Me handler → 200 { sub, username, email, name, roles, tenant }
  Response: 200 → authorized zone

STEP 6: RENDER AUTHORIZED SHELL
  → filterNavByRoles(navItems, roles)
  → Render full navigation

STEP 7: SUBSEQUENT API REQUESTS
  → authorizedFetch with Bearer token
  → Same middleware chain: TenantResolver → JWT → RBAC → Handler
  → If 401: keycloak.updateToken(30) → retry once
  → If refresh fails: full re-login
```

---

## Nginx configuration

```nginx
# Route table
location = / {
    proxy_pass http://lkfl-frontend:4173;
    # SPA fallback
    try_files $uri $uri/ /index.html;
}

location /api/ {
    proxy_pass http://lkfl-server:8080;
    # Authorization header passed through
    proxy_set_header Authorization $http_authorization;
}

location /auth {
    proxy_pass http://keycloak:8080;
    # Keycloak handles auth
}

location = /healthz {
    proxy_pass http://lkfl-server:8080;
}

location = /metrics {
    allow 172.16.0.0/12;   # Docker network
    allow 127.0.0.1/32;
    allow ::1/128;
    deny all;
    proxy_pass http://lkfl-server:8080;
}

# Security headers
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-Frame-Options "DENY" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;

# Frontend CSP
location / {
    # $base_domain — из env: NGINX_BASE_DOMAIN (lkfl.ru, sdek.ru, etc.)
    # В dev: "localhost" (без https)
    # Пример: base64 $base_domain "lkfl.ru";
    add_header Content-Security-Policy
        "default-src 'self';
         script-src 'self' 'unsafe-inline' 'unsafe-eval';
         style-src 'self' 'unsafe-inline';
         img-src 'self' data:;
         font-src 'self';
         connect-src 'self' https://*.$base_domain;
         frame-src 'self' https://*.$base_domain/auth;" always;
}

# API CSP
location /api/ {
    add_header Content-Security-Policy "default-src 'none'" always;
}
```

---

## Observability

### Auth logs (structured JSON)

```json
{
  "level": "DEBUG",
  "msg": "auth_success",
  "path": "/api/v1/me",
  "user": "subject-123",
  "tenant_id": "sdek",
  "roles": ["employee"]
}
```

Failed auth:
```json
{
  "level": "WARN",
  "msg": "auth_failed",
  "path": "/api/v1/me",
  "method": "GET",
  "status": 401,
  "reason": "invalid_token",
  "tenant_id": "sdek",
  "correlationId": "abc",
  "requestId": "def"
}
```

### Prometheus metrics

| Metric | Labels | Purpose |
|--------|--------|---------|
| `auth_errors_total` | `path`, `reason`, `status_code`, `tenant_id` | Auth failure tracking |
| `http_request_duration_seconds` | `method`, `path`, `status_code`, `tenant_id` | Request latency |
| `cache_hit_total` | `path`, `tenant_id` | Cache effectiveness |
| `tenant_resolution_duration_seconds` | `tenant_id` | Tenant resolver performance |

---

## Security

### Threat model (STRIDE for auth)

| Threat | STRIDE | Mitigation |
|--------|--------|------------|
| JWT подделка | Spoofing | RS256 signature check, realm-specific JWKS |
| Меж-tenant доступ | Spoofing | Tenant Resolver + realm isolation + DB FK |
| Token theft | Spoofing | Short TTL (15min), PKCE, HTTPS only |
| Brute force | Denial of Service | Keycloak account lockout, rate limiting |
| Session fixation | Repudiation | Keycloak session management |
| ПДн в логах | Information Disclosure | Только user_id, не ПДн в audit log |

### JWT TTL

| Token | TTL | Обоснование |
|-------|-----|-------------|
| Access token | 15 min | Баланс безопасности и UX |
| Refresh token | 7 days | Keycloak default |
| ID token | 15 min | Sync with access token |

---

## Файловая структура

### Backend

```
shared/pkg/auth/
├── tenantresolver.go   # TenantResolver — subdomain → tenant
├── verifier.go         # JWT verifier — RS256, JWKS per realm
├── middleware.go       # JWT middleware wrapper
├── rbac.go             # RBACGuard middleware
├── claims.go           # Claims type + context helpers
├── errors.go           # Auth errors (401/403)
└── cache.go            # JWKS cache per realm

internal/auth/
├── auth.go             # OIDCVerifier — thin wrapper over shared/pkg/auth
└── config.go           # Tenant-specific Keycloak config builder
```

### Frontend

```
src/
├── utils/
│   └── tenant.ts                   # extractTenantSlug + buildKeycloakConfig
├── keycloak.ts                     # Keycloak instance + init (realm из tenant resolution)
├── keycloak-token-subscribers.ts   # Token rotation notification
├── api/
│   └── platform.ts                 # authorizedFetch + React Query
├── store/
│   └── auth.ts                     # Zustand auth store
├── components/
│   └── auth/
│       ├── RequireAuth.tsx          # Auth guard
│       ├── GuestZone.tsx            # Landing + login
│       ├── ForbiddenZone.tsx        # 403 screen
│       └── AuthorizedShell.tsx      # Full nav + role filtering
└── App.tsx                         # Auth state machine
```

### Infra

```
infra/
├── keycloak/
│   ├── realm/
│   │   └── tenant-template.json    # Realm template for new tenants
│   ├── scripts/
│   │   └── create-tenant-realm.sh  # Script for tenant onboarding
│   └── seed-demo.sh                # Dev: create demo realm + users + clients
└── nginx/
    └── lkfl.conf                   # Nginx config ($base_domain configurable)
```

---

## Quick Reference Checklist

При изменении auth-кода проверять:

- [ ] Tenant Resolver работает ДО JWT валидации
- [ ] JWT валидация: signature (JWKS), `iss`, `azp`, `exp`
- [ ] RBAC guard соответствует required roles эндпоинта
- [ ] `tenant_id` корректно проходит через context
- [ ] Error responses содержат `code`, `message`, `metadata`
- [ ] Auth errors логируются structured JSON (без токенов!)
- [ ] CORS позволяет `Authorization` header
- [ ] Nginx security headers present
- [ ] Frontend: extractTenantSlug() корректно извлекает realm из hostname
- [ ] Frontend: token rotation уведомляет React Query
- [ ] Frontend: 401 → automatic refresh + retry once

---

## Альтернативы (рассмотренные)

| Вариант | Почему отклонён |
|---------|----------------|
| Single realm с `tenant_id` attribute (как April) | Нарушает realm isolation — FСТЭК требует разделения |
| JWT эмитировать из бэкенда | Нарушает single source of truth — Keycloak как IdP |
| Auth0 / Okta | SaaS — нарушает ФСТЭК (данные за пределами РФ) |
| Self-hosted Zitadel | Меньше mature, нет SAML broker из коробки |

---

## Следствия

1. `shared/pkg/auth` — единый пакет для всех auth-компонентов
2. `internal/auth` — thin wrapper для tenant-specific конфигурации
3. Frontend: Keycloak JS + React Query + Zustand
4. Keycloak: realm template as code (`infra/keycloak/realm/tenant-template.json`)
5. Nginx: transparent proxy для `/auth` и `/api/`
6. Audit log: все auth события → structured JSON → Loki
7. Prometheus: `tenant_id` label на всех auth метриках
8. Tenant onboarding: автоматическое создание realm из шаблона
