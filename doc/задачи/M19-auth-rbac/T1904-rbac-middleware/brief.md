# T1904 — RBAC Middleware (интеграция)

## Веха

M19-auth-rbac

## Тип

code

## Контекст

Интеграция RBAC middleware в router. Определение ролей для каждого route.

**Текущее состояние `app/server.go` после M18:**
- `NewServer()` принимает `tenantService *tenant.Service`
- Admin tenant routes зарегистрированы без JWT+RBAC (TODO M19)
- `/healthz`, `/metrics` — публичные
- CORS middleware разрешает `X-Tenant-ID` header

**Что нужно изменить в `server.go`:**
- Добавить JWT middleware (`sharedauth.JWTMiddleware`) для защищённых route groups
- Добавить RBAC middleware (`sharedauth.RBACMiddleware`) для admin routes
- Добавить tenant middleware (`tenant.TenantMiddlewareWithService`) для employee routes
- Auth routes (`/api/v1/auth/`) — публичные, без middleware

## Что сделать

### Роли и доступ

| Route | Роли | Описание |
|-------|------|----------|
| `/api/v1/auth/*` | none (public) | Auth endpoints |
| `/api/v1/users/me` | employee, hr, catalog_manager, admin | Профиль |
| `/api/v1/engagements` | employee, hr, catalog_manager, admin | Каталог (чтение) |
| `/api/v1/user-engagements` | employee, hr | Мои льготы |
| `/admin/tenants/*` | admin | Tenant management |
| `/admin/users/*` | hr, admin | User management |
| `/admin/engagements/*` | catalog_manager, admin | Catalog admin |
| `/admin/billing/*` | hr, admin | Billing admin |

### Integration в `app/server.go`

> **Внимание:** `NewServer()` после M18 принимает `tenantService *tenant.Service`.
> T1904 добавляет `authHandler *auth.Handler` и `userHandler *user.Handler` как параметры.

```go
func NewServer(
    cfg ServerConfig,
    db *pgxpool.Pool,
    redis *redis.Client,
    verifier *oidc.IDTokenVerifier,
    logger Logger,
    reg *prometheus.Registry,
    tenantService *tenant.Service,
    authHandler *auth.Handler,     // ← NEW T1903
    userHandler *user.Handler,     // ← NEW T1905
) *Server {
    r := chi.NewRouter()

    // Global middleware (без auth)
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))

    metrics := newHTTPMetrics(reg)
    r.Use(PrometheusMiddleware(metrics))
    r.Use(corsMiddleware())

    // ─── Public routes (без auth) ───
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    r.Mount("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

    // ─── Auth routes (публичные, без JWT) ───
    r.Route("/api/v1/auth/", func(r chi.Router) {
        r.Get("/login", authHandler.LoginRedirect)
        r.Get("/callback", authHandler.LoginCallback)
        r.Post("/logout", authHandler.Logout)
    })

    // ─── Employee routes (JWT + tenant middleware) ───
    r.Route("/api/v1/", func(r chi.Router) {
        r.Use(sharedauth.JWTMiddleware(verifier))
        r.Use(tenant.TenantMiddlewareWithService(tenantService, redis))

        // User profile
        r.Get("/users/me", userHandler.Me)
        r.Put("/users/me", userHandler.UpdateMe)

        // Future: catalog, engagements, etc.
        // r.Get("/engagements", catalogHandler.List)
    })

    // ─── Admin routes (JWT + RBAC, без tenant middleware) ───
    r.Route("/admin/", func(r chi.Router) {
        r.Use(sharedauth.JWTMiddleware(verifier))

        // Admin-only
        r.Group(func(r chi.Router) {
            r.Use(sharedauth.RBACMiddleware([]string{"admin"}))
            r.Route("/tenants", func(r chi.Router) {
                th := tenant.NewHandler(tenantService)
                r.Post("/", th.Create)
                r.Get("/", th.List)
                r.Get("/{id}", th.GetByID)
                r.Put("/{id}", th.Update)
                r.Delete("/{id}", th.Delete)
                r.Get("/{id}/brand", th.GetBrandConfig)
                r.Put("/{id}/brand", th.UpsertBrandConfig)
            })
        })

        // HR + Admin
        r.Group(func(r chi.Router) {
            r.Use(sharedauth.RBACMiddleware([]string{"hr", "admin"}))
            r.Route("/users", func(r chi.Router) {
                r.Get("/", userHandler.AdminList)
                r.Get("/{id}", userHandler.AdminGet)
                r.Put("/{id}", userHandler.AdminUpdate)
                r.Post("/{id}/deactivate", userHandler.AdminDeactivate)
            })
        })

        // Catalog Manager + Admin
        r.Group(func(r chi.Router) {
            r.Use(sharedauth.RBACMiddleware([]string{"catalog_manager", "admin"}))
            // Future: catalog admin routes
        })
    })

    // ...
}
```

Также нужно обновить `wire.go` — добавить создание `authHandler` и `userHandler`
и передать их в `NewServer()`.

## Требования

- RBAC middleware из `shared/pkg/auth/`
- Route groups по ролям
- Admin routes — без tenant middleware
- Employee routes — с tenant middleware
- Auth routes — без middleware (public)

### Источники ролей

**RBAC в middleware (T1902) использует роли из JWT claims Keycloak** — это быстрый путь:
роль проверяется при каждом запросе без обращения к БД.

**Таблица `user_roles` (T1901) — source of truth в БД** — используется для:
- Админ-панели: назначение/снятие ролей
- Аудит: кто, кому, когда назначил роль (`granted_by`, `granted_at`)
- Отчёты и аналитика

**Синхронизация:** при изменении роли в БД (через admin API) — роль должна обновляться
в Keycloak через Admin API. Это обеспечивает консистентность между БД и JWT claims.
Реализация синхронизации — часть T1905 (admin user operations).

## Критерии приёмки

- [ ] Router groups по ролям
- [ ] Admin routes без tenant middleware
- [ ] Employee routes с tenant middleware
- [ ] Auth routes public
- [ ] RBAC deny → 403 Forbidden
- [ ] Unit tests: role allow/deny
