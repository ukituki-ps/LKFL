# T1803 — Tenant Middleware (resolver)

## Веха

M18-multi-tenancy

## Тип

code

## Контекст

Tenant middleware — резолвит tenant из request (Host header или path segment) и добавляет `tenant_id` в context.
Используется во всех handlers бизнес-пакетов для multi-tenant isolation.

## Что сделать

### `internal/tenant/middleware.go`

```go
package tenant

import (
    "context"
    "net/http"
    "strings"

    "github.com/go-chi/chi/v5"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"
const TenantSlugKey contextKey = "tenant_slug"

// TenantFromContext — извлечь tenant_id из context
func TenantIDFromContext(ctx context.Context) uuid.UUID {
    id, _ := ctx.Value(TenantIDKey).(uuid.UUID)
    return id
}

// TenantSlugFromContext — извлечь tenant slug из context
func TenantSlugFromContext(ctx context.Context) string {
    slug, _ := ctx.Value(TenantSlugKey).(string)
    return slug
}

// Resolver — интерфейс для разрешения tenant из request
type Resolver interface {
    Resolve(r *http.Request) (Tenant, error)
}

// HostResolver — резолвит tenant из subdomain (sdek.example.com → slug=sdek)
type HostResolver struct {
    service *Service
}

// PathResolver — резолвит tenant из path segment (/t/sdek/api/...)
type PathResolver struct {
    service *Service
}

// Middleware — chi middleware для tenant resolution
func Middleware(resolver Resolver) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            tenant, err := resolver.Resolve(r)
            if err != nil {
                http.Error(w, `{"error":"tenant not found"}`, http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), TenantIDKey, tenant.ID)
            ctx = context.WithValue(ctx, TenantSlugKey, tenant.Slug)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// SkipPaths — middleware не применяется к указанным path'ам
func SkipPaths(paths []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for _, p := range paths {
                if strings.HasPrefix(r.URL.Path, p) {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### HostResolver implementation

```go
func (r *HostResolver) Resolve(req *http.Request) (Tenant, error) {
    // sdek.example.com → sdek
    host := strings.Split(req.Host, ".")[0]

    if host == "www" || host == "localhost" {
        // Fallback: X-Tenant-ID header
        slug := req.Header.Get("X-Tenant-ID")
        if slug == "" {
            return Tenant{}, ErrNoTenantHeader
        }
        return r.service.GetBySlug(req.Context(), slug)
    }

    return r.service.GetBySlug(req.Context(), host)
}
```

### PathResolver implementation

```go
func (r *PathResolver) Resolve(req *http.Request) (Tenant, error) {
    // /t/sdek/api/v1/... → sdek
    parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/t/"), "/")
    if len(parts) == 0 {
        return Tenant{}, ErrNoTenantInPath
    }
    slug := parts[0]
    return r.service.GetBySlug(req.Context(), slug)
}
```

### Integration в `app/server.go`

```go
// After CORS middleware, before routes:
// Admin routes — без tenant middleware (глобальное управление)
r.Route("/admin/", func(r chi.Router) {
    r.Use(sharedauth.JWTMiddleware(verifier))
    r.Use(sharedauth.RBACMiddleware([]string{"admin"}))
    // Admin handlers без tenant middleware
})

// Public routes — с tenant middleware
r.Route("/api/v1/", func(r chi.Router) {
    r.Use(sharedauth.JWTMiddleware(verifier))
    r.Use(tenant.Middleware(hostResolver))
    // Business handlers с tenant_id в context
})
```

## Требования

- Два resolver'а: Host (subdomain) + Path (`/t/{slug}/`)
- Fallback: `X-Tenant-ID` header (для dev + API clients)
- Context keys — typed (не string constant напрямую)
- Helper функции: `TenantIDFromContext()`, `TenantSlugFromContext()`
- Skip paths: `/healthz`, `/metrics`, `/admin/` (без tenant middleware)
- Error: 401 Unauthorized при отсутствии tenant
- Redis cache: `tenant:resolve:{slug}` TTL 5min (избегаем DB hit на каждый request)

## Критерии приёмки

- [ ] `middleware.go` — chi middleware + context helpers
- [ ] HostResolver — subdomain → slug → tenant
- [ ] PathResolver — `/t/{slug}/` → tenant
- [ ] Fallback: `X-Tenant-ID` header
- [ ] Context keys typed
- [ ] Skip paths: `/healthz`, `/metrics`, `/admin/`
- [ ] Redis cache для resolution
- [ ] 401 при отсутствии tenant
- [ ] Unit tests: HostResolver, PathResolver, header fallback, cache
