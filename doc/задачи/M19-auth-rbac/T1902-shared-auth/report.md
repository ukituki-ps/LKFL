# Отчёт T1902 — shared/pkg/auth/ (OIDC Verifier + Middleware)

## Что сделано

Создан пакет `lkfl/shared/pkg/auth` — общая функциональность аутентификации для LKFL.

### Созданные файлы

| Файл | Описание |
|------|----------|
| `backend/shared/pkg/auth/verifier.go` | OIDC verifier (`NewVerifier`) — создание `*oidc.IDTokenVerifier` из Keycloak issuer URL |
| `backend/shared/pkg/auth/claims.go` | Claims struct + `ExtractClaims()` — парсинг стандартных OIDC claims + извлечение Keycloak roles из `resource_access.{clientID}.roles` |
| `backend/shared/pkg/auth/middleware.go` | JWT middleware (`JWTMiddleware`) — Bearer extraction → verify → context injection; helpers `UserIDFromContext()`, `RolesFromContext()`, `writeJSONError()` |
| `backend/shared/pkg/auth/rbac.go` | RBAC middleware (`RBACMiddleware`) — проверка ролей пользователя из context; helper `withRoles()` для тестов |
| `backend/shared/pkg/auth/rbac_test.go` | Unit-тесты: RBAC allow (4 кейса), RBAC deny (3 кейса), RolesFromContext (2 кейса), UserIDFromContext (2 кейса), extractKeycloakRoles (4 кейса), writeJSONError (1 кейс) |

### Изменённые файлы

| Файл | Изменение |
|------|----------|
| `backend/internal/app/wire.go` | Заменён вызов `newOIDCVerifier(cfg.Keycloak)` на `auth.NewVerifier(context.Background(), cfg.Keycloak.Issuer, cfg.Keycloak.ClientID)`; добавлен import `lkfl/shared/pkg/auth` |

## Результаты проверки

- `go build ./...` — чистая компиляция ✅
- `go vet ./...` — без замечаний ✅
- `go test ./shared/pkg/auth/...` — 6 тестов, все зелёные ✅
  - TestRBACMiddleware_Authorized (4 subtests) ✅
  - TestRBACMiddleware_Unauthorized (3 subtests) ✅
  - TestRolesFromContext (2 subtests) ✅
  - TestUserIDFromContext (2 subtests) ✅
  - TestExtractKeycloakRoles (4 subtests) ✅
  - TestJSONError (1 test) ✅

## Замечания

1. Локальные функции `newOIDCVerifier()` и `newOIDCProvider()` оставлены в wire.go — будут удалены в T1904 при полном переходе на shared-пакет.
2. `go-oidc` import оставлен в wire.go (используется локальными функциями).
3. Тесты адаптированы под Go 1.24 (nil parent context запрещён — использован `context.Background()`).
4. go.mod не изменялся.

## Время

~30 минут
