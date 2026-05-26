# T1004 — Split api/ → public router + admin router

## Веха

M10-рефакторинг-по-результатам-аудита

## Контекст

`HandlerDeps` сейчас содержит 12 полей — knows о каждом business-пакете Platform:
```go
type HandlerDeps struct {
    Auth            *auth.OIDCVerifier
    User            *user.UserRepository
    Consent         *consent.ConsentEngine
    Eligibility     *eligibility.EligibilityEngine
    Engagement      *engagement.EngagementEngine  // → заменён на catalog.CatalogService + flow.FlowEngine + collections.CollectionsEngine
    Notification    *notification.NotificationEngine
    Recommendations  *recommendations.RecommendationsEngine  // T1001 → stub
    Gamification    *gamification.GamificationEngine
    DB              *sql.DB
    NATS            *nats.Conn
    Asynq           *asynq.Server
    Redis           *redis.Client
    Logger          Logger
}
```

17 handler файлов смешивают public API (`/api/v1/...`) и admin API (`/admin/...`) в одном роутере с одинаковой middleware chain:

```
Request → Recovery → Logger → RateLimiter → JWTMiddleware → TenantResolver → RBAC → Handler
```

**Проблема:**
Admin endpoints требуют different middleware:
- Rate limiting: admin API needs lower limits (admin is fewer users)
- RBAC: admin должен быть `hr | catalog_manager | admin`, public API = любой authenticated
- Audit trail depth: admin actions need full logging, public API = minimal

**Решение — 2 router'а:**
```go
// Public API: /api/v1/...
type PublicHandlerDeps struct {
    Auth          *auth.OIDCVerifier
    User          *user.UserRepository
    Consent       *consent.ConsentEngine
    Engagement    *engagement.EngagementEngine  // → заменён на catalog.CatalogService + flow.FlowEngine + collections.CollectionsEngine
    Notification  *notification.NotificationEngine
    Gamification  *gamification.GamificationEngine
    DB            *sql.DB
    NATS          *nats.Conn  // → удалён M12
    Redis         *redis.Client
    Logger        Logger
}

// Admin API: /admin/...
type AdminHandlerDeps struct {
    User          *user.UserRepository
    Consent       *consent.ConsentEngine
    Engagement    *engagement.EngagementEngine  // → заменён на catalog.CatalogService + flow.FlowEngine + collections.CollectionsEngine
    Notification  *notification.NotificationEngine
    Gamification  *gamification.GamificationEngine
    DB            *sql.DB
    Logger        Logger
}
```

Middleware chains разделяются:
```
Public:  Recovery → Logger → RateLimiter(HIGH) → JWT → Tenant → RBAC(any) → Handler
Admin:   Recovery → Logger → RateLimiter(LOW) → JWT → Tenant → RBAC(admin-only) → Audit → Handler
```

### Файлы-мишени

| Действие | Файл |
|---|-|- |
| 2 router'а | `архитектура/пакеты-platform.md` — api/ public + admin |
| HandlerDeps → 2 struct | `архитектура/пакеты-platform.md` — таблица, DI граф |
| Handler table | `архитектура/пакеты-platform.md` — group by router |
| Middleware chain | `архитектура/пакеты-platform.md` — 2 chains |
| Модули | `архитектура/модули.md` — comment о разделении |

### Критерии приёмки

- [ ] `архитектура/пакеты-platform.md` — 2 HandlerDeps struct (PublicHandlerDeps, AdminHandlerDeps)
- [ ] 2 middleware chains документированы (different rate limits + RBAC)
- [ ] Handler table разделена: public handlers (10) vs admin handlers (5)
- [ ] `архитектура/модули.md` — comment о разделении
- [ ] T1001 (stub recommendations) совместим: admin/recommendations handler → stub notice
