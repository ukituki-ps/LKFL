# T1904 — RBAC Middleware (интеграция)

## Веха

M19-auth-rbac

## Тип

code

## Контекст

Интеграция RBAC middleware в router. Определение ролей для каждого route.

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

```go
// Public routes (без auth)
r.Get("/healthz", healthzHandler)
r.Mount("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

// Auth routes (без JWT middleware)
r.Route("/api/v1/auth/", func(r chi.Router) {
    r.Get("/login", authHandler.LoginRedirect)
    r.Get("/callback", authHandler.LoginCallback)
    r.Post("/logout", authHandler.Logout)
})

// Employee routes (JWT + tenant)
r.Route("/api/v1/", func(r chi.Router) {
    r.Use(sharedauth.JWTMiddleware(verifier))
    r.Use(tenant.Middleware(hostResolver))

    r.Get("/users/me", userHandler.Me)
    r.Get("/engagements", catalogHandler.List)
    r.Get("/engagements/{id}", catalogHandler.Get)
    r.Get("/user-engagements", userEngagementHandler.List)
    // ...
})

// Admin routes (JWT + RBAC admin)
r.Route("/admin/", func(r chi.Router) {
    r.Use(sharedauth.JWTMiddleware(verifier))
    // Без tenant middleware (глобальное управление)

    r.Group(func(r chi.Router) {
        r.Use(sharedauth.RBACMiddleware([]string{"admin"}))
        r.Route("/tenants", func(r chi.Router) {
            r.Post("/", tenantHandler.Create)
            r.Get("/", tenantHandler.List)
            r.Get("/{id}", tenantHandler.Get)
            r.Put("/{id}", tenantHandler.Update)
            r.Delete("/{id}", tenantHandler.Delete)
            r.Get("/{id}/brand", tenantHandler.GetBrand)
            r.Put("/{id}/brand", tenantHandler.UpdateBrand)
        })
    })

    r.Group(func(r chi.Router) {
        r.Use(sharedauth.RBACMiddleware([]string{"hr", "admin"}))
        r.Route("/users", func(r chi.Router) {
            r.Get("/", userHandler.AdminList)
            // ...
        })
        r.Route("/billing", func(r chi.Router) {
            // ...
        })
    })

    r.Group(func(r chi.Router) {
        r.Use(sharedauth.RBACMiddleware([]string{"catalog_manager", "admin"}))
        r.Route("/engagements", func(r chi.Router) {
            // ...
        })
    })
})
```

## Требования

- RBAC middleware из `shared/pkg/auth/`
- Route groups по ролям
- Admin routes — без tenant middleware
- Employee routes — с tenant middleware
- Auth routes — без middleware (public)

## Критерии приёмки

- [ ] Router groups по ролям
- [ ] Admin routes без tenant middleware
- [ ] Employee routes с tenant middleware
- [ ] Auth routes public
- [ ] RBAC deny → 403 Forbidden
- [ ] Unit tests: role allow/deny
