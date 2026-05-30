# T1904 — RBAC Middleware (интеграция) — Отчёт

## Выполнено

Полная интеграция RBAC middleware в router. Все admin и employee routes защищены JWT + RBAC.

### `backend/internal/app/server.go`

Реструктуризация router'а на 4 группы маршрутов:

| Группа | Путь | Middleware | Роли |
|--------|------|-----------|------|
| Public | `/healthz`, `/metrics` | — | none |
| Auth | `/api/v1/auth/*` | — | none |
| Employee | `/api/v1/*` | JWT + Tenant | employee, hr, catalog_manager, admin |
| Admin: tenant | `/admin/tenants/*` | JWT + RBAC(admin) + AdminTenant | admin |
| Admin: users | `/admin/users/*` | JWT + RBAC(hr,admin) + AdminTenant | hr, admin |
| Admin: catalog | `/admin/catalog/*` | JWT + RBAC(catalog_manager,admin) + AdminTenant | catalog_manager, admin |

### `backend/internal/app/wire.go`

DI-цепочка для auth + user модулей:
- `user.NewRepository(dbPool)` → `user.NewService(repo)` → `user.NewHandler(service)`
- `auth.NewService(userRepo)` → `auth.NewHandler(verifier, redis, service, issuer, clientID)`
- Оба handler'а передаются в `NewServer()`
- Удалён мёртвый код: `newOIDCVerifier()`, `newOIDCProvider()`

### `backend/internal/tenant/isolation.go`

Новый `AdminTenantMiddleware()` — лёгкий middleware для admin routes:
- Извлекает tenant из X-Tenant-ID header через HostResolver
- Если header пуст — устанавливает uuid.Nil (глобальные admin запросы)
- Не блокирует запрос при ошибке resolution (отличие от TenantMiddlewareWithService)

### `backend/shared/pkg/http/response.go`

Общий пакет для JSON-ответов:
- `WriteJSON(w, status, data)` — успешные ответы
- `WriteJSONError(w, status, message)` — ошибки в формате `{"error": "..."}`

Убраны дубликаты `writeJSON`/`writeJSONError` из:
- `internal/auth/handler.go`
- `internal/user/handler.go`
- `internal/tenant/handler.go`
- `internal/tenant/middleware.go`

### `backend/migrations/20260526120000_users.sql`

- `accounts.user_id` ON DELETE: CASCADE → RESTRICT (соответствует schema.md, защита финансовых данных)

## Проверки

- ✅ `go build ./...` — чистая компиляция
- ✅ `go vet ./...` — без замечаний
- ✅ `go test ./...` — все тесты зелёные (26 тестов, 4 пакета)
- ✅ Три нуля: код не привязан к конкретному бренду, провайдеру или модели начислений
- ✅ RBAC deny → 403 Forbidden (JSON)
- ✅ Auth routes public (без JWT)
- ✅ Admin routes protected (JWT + RBAC)
- ✅ Employee routes protected (JWT + Tenant)

## Время

~45 минут

## Замечания

- PKCE code challenge не реализован (TODO в LoginRedirect) — требует frontend code_verifier
- Authorization code flow не реализован (TODO в LoginCallback) — используется implicit flow для dev
- Роли из Keycloak не назначаются автоматически (TODO в CreateOrUpdateUser) — требуется синхронизация через user.RoleRepository
- `shared/pkg/http` — пакет использует имя `http`, конфликтующее со stdlib; все импорты используют alias `shhttp`
