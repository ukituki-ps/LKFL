# T1702 — Отчёт: Backend Auth

## Статус

✅ Завершено

## Что сделано

### shared/pkg/auth (7 файлов)
1. `verifier.go` — OIDC verifier (go-oidc), configurable retry (WithMaxRetries, WithRetryDelay)
2. `claims.go` — Claims struct, ExtractClaims, extractKeycloakRoles, ExtractRolesFromJWT
3. `middleware.go` — JWTMiddleware (Bearer header + cookie fallback), extractToken, context helpers
4. `rbac.go` — RBACMiddleware, withRoles helper
5. `tenantresolver.go` — ResolveTenantSlug из issuer URL
6. `errors.go` — WriteAuthError, AuthError struct, WriteUnauthorizedError, WriteForbiddenError
7. `cache.go` — stub (go-oidc внутреннее кэширование JWKS)

### Тесты
- `verifier_test.go` — TestNewVerifier_Options, TestNewVerifier_FailsOnBadIssuer, TestExtractToken, TestResolveTenantSlug, TestWriteAuthError
- `middleware_test.go` — 800+ строк, все сценарии JWTMiddleware
- `rbac_test.go` — RBACMiddleware, UserIDFromContext, RolesFromContext, extractKeycloakRoles

### internal/auth
- `handler.go` — LoginRedirect (PKCE, state в Redis), LoginCallback (token exchange, PKCE verification, session), Logout, Me
- `service.go` — CreateOrUpdateUser (first login → create, subsequent → update + sync roles)

### internal/tenant
- `middleware.go` — TenantMiddleware, TenantIDFromContext, JSONB type
- `repository.go` — Tenant CRUD

### internal/user
- `model.go` — User, UserProfile, Account, UserRole, UserFilter
- `repository.go` — Repository interface + pgx impl (CRUD, roles, accounts)
- `service.go` — бизнес-логика (валидация, status transitions)
- `handler.go` — HTTP handlers

### Миграции
- `migrations/20260526110000_tenants.sql` — tenants, tenant_brand_config
- `migrations/20260526120000_users.sql` — users, accounts, user_roles
- `migrations/20260526130000_engagement.sql` — engagement tables
- `shared/pkg/migrate/` — общая логика (D13 deduplication)

## Проблемы

- writeJSONError был дублирован в middleware.go → вынесен в errors.go (T1709)
- extractTenantSlug был inline → вынесен в tenantresolver.go (T1709)
- roles не назначались (_ = roles) → реализована синхронизация (T1709 D6)

## Следующие шаги

Н/Д — задача завершена.
