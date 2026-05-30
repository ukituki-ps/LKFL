# T1700 — Отчёт: Полная система авторизации

## Статус

✅ Завершено

## Что сделано

### Реализовано (фактическое содержание)

**Backend (Go):**
- `go.mod` — инициализация с 20+ зависимостями (go-oidc, chi/v5, pgx/v5, go-redis, CEL, Prometheus)
- `cmd/server/main.go` — сервер с подкомандами migrate/seed, graceful shutdown
- `cmd/seed/main.go` — seed data для tenant, пользователей, категорий, engagement types/offers
- `cmd/worker/main.go` — stub Asynq worker
- `cmd/deploy-worker/main.go` — deploy worker
- `cmd/integration-proxy/main.go` — proxy entry point

**Auth (shared/pkg/auth):**
- `verifier.go` — OIDC verifier (go-oidc) с retry logic + functional options
- `claims.go` — ExtractClaims, extractKeycloakRoles, ExtractRolesFromJWT
- `middleware.go` — JWTMiddleware (Bearer header + httpOnly cookie fallback)
- `rbac.go` — RBACMiddleware для проверки ролей
- `tenantresolver.go` — ResolveTenantSlug из issuer URL
- `errors.go` — WriteAuthError, AuthError struct
- `cache.go` — stub (go-oidc кэширует JWKS внутренно)
- `verifier_test.go` — unit тесты (options, extractToken, ResolveTenantSlug, WriteAuthError)

**Internal пакеты:**
- `internal/auth/handler.go` — LoginRedirect, LoginCallback, Logout, Me (PKCE, state, Redis sessions)
- `internal/auth/service.go` — CreateOrUpdateUser с sync ролей из Keycloak
- `internal/tenant/` — tenant middleware, context helpers
- `internal/user/` — CRUD пользователей, ролей, аккаунтов
- `internal/engagement/catalog/` — каталог engagement types/offers
- `internal/metrics/` — Prometheus metrics для auth flow

**Migrations:**
- `migrations/` — 3 SQL миграции (tenants, users, engagement) + down файлы
- `migrations/atlas.hcl` — Atlas config для dev
- `shared/pkg/migrate/` — общая логика миграций (deduplication из main.go + testcontainers)

**Frontend (React):**
- Vite project setup с April UI + Mantine
- `src/pages/Login.tsx` — login page с redirect на Keycloak
- `src/pages/Callback.tsx` — callback handler (realm из env, cookie-based auth)
- `src/stores/authStore.ts` — Zustand store (user + roles, без token в LS)
- `src/api/client.ts` — API client с `credentials: 'include'`
- `src/App.tsx` — router с protected routes

**Infra:**
- `docker-compose.yml` — PostgreSQL, Redis, Keycloak, Backend, Proxy, Nginx, Prometheus, Loki, Grafana
- `Dockerfile.server`, `Dockerfile.proxy` — Docker build
- `infra/nginx/` — Nginx конфиг с proxy к backend + frontend + Keycloak
- `infra/prometheus/`, `infra/loki/`, `infra/grafana/` — observability stack
- `.github/workflows/build.yml` — CI pipeline

### Исправления (T1709 gap closure):
- D3: фикс err → mErr в main.go
- D1: hard-coded realm → VITE_KEYCLOAK_REALM
- D2: localStorage → httpOnly cookie (152-ФЗ compliance)
- D9: ExtractRolesFromJWT → private wrapper с документацией
- D6: реализация синхронизации ролей в CreateOrUpdateUser
- D8: configurable retry в verifier (functional options)
- D10: KEYCLOAK_PUBLIC_URL в docker-compose
- D11: fallback token → error handling
- D13: deduplication миграций → shared/pkg/migrate/

## Проблемы

- T1707 report содержал неточности (keyfunc вместо go-oidc, atlas.hcl отсутствовал)
- Initial reports T1700-T1705 были пустыми (исправлено в T1709)
- Token в localStorage (исправлено → httpOnly cookie)
- Hard-coded realm (исправлено → env var)

## Следующие шаги

1. M18: Grafana dashboards (T1705 отложено)
2. M19: E2E тесты (Playwright), OpenAPI spec
3. Production hardening: rate limiting, audit logging
