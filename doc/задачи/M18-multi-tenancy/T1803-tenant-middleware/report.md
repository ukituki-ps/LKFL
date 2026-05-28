# T1803 — Отчёт о выполнении

## Статус

✅ выполнено

## Что сделано

### `backend/internal/tenant/middleware.go`

- **Context keys**: `TenantIDKey`, `TenantSlugKey` (typed как `type contextKey string`)
- **Helpers**: `TenantIDFromContext()`, `TenantSlugFromContext()`
- **Resolver interface**: `Resolve(r *http.Request) (Tenant, error)`
- **HostResolver**: subdomain → slug (sdek.example.com → sdek). Fallback на X-Tenant-ID header для www/localhost. Обработка порта в host (localhost:8080 → localhost).
- **PathResolver**: /t/{slug}/ → tenant
- **Middleware(resolver)**: chi middleware, добавляет tenant в context. 401 при ошибке.
- **SkipPaths**: middleware для пропуска tenant resolution на указанных путях
- **Redis cache**: `tenant:resolve:{slug}` TTL 5min (HostResolver), `tenant:path:{slug}` (PathResolver)
- **TenantMiddlewareWithService**: удобный конструктор middleware

### Интеграция

- `wire.go`: создание TenantService (через tenant.NewRepository(pool) → tenant.NewService(repo))
- `server.go`: регистрация admin tenant routes в `/admin/tenants`

### Unit тесты (`middleware_test.go`)

- HostResolver: subdomain resolution, www fallback header, localhost fallback header, no header error, tenant not found
- PathResolver: valid path, no prefix, empty slug
- Middleware: success (tenant в context), not found (401)
- Context helpers: TenantIDFromContext, TenantSlugFromContext (filled/empty)
- SkipPaths
- Redis cache: serialization/deserialization, fallback without Redis

## Критерии приёмки

- [x] `middleware.go` — chi middleware + context helpers
- [x] HostResolver — subdomain → slug
- [x] PathResolver — `/t/{slug}/`
- [x] Fallback: X-Tenant-ID header
- [x] Context keys typed
- [x] Skip paths: /healthz, /metrics, /admin/
- [x] Redis cache (tenant:resolve:{slug}, TTL 5min)
- [x] 401 при отсутствии tenant
- [x] Unit tests: HostResolver, PathResolver, header fallback, cache

## Время

~30 мин
