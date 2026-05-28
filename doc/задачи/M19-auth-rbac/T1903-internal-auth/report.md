# T1903 — internal/auth/ (Auth handlers) — Отчёт

## Что сделано

Создан пакет `lkfl/internal/auth/` с двумя файлами реализации и тестами:

### `handler.go` — HTTP handlers для auth flow

- **LoginRedirect** (`GET /api/v1/auth/login`) — генерирует криптографически безопасный state-параметр (crypto/rand, 32 байта), сохраняет в Redis (TTL 10 мин), редиректит на Keycloak authorize endpoint
- **LoginCallback** (`GET /api/v1/auth/callback`) — проверяет state (CSRF protection), верифицирует ID token через OIDC verifier, извлекает claims через `sharedauth.ExtractClaims()`, создаёт/обновляет пользователя через Service, устанавливает сессию в Redis (TTL 24 часа)
- **Logout** (`POST /api/v1/auth/logout`) — удаляет сессию из Redis по userID из context, редиректит на Keycloak logout endpoint
- **Me** (`GET /api/v1/auth/me`) — возвращает профиль текущего пользователя по keycloak_sub из context (требует JWT middleware)

### `service.go` — Auth service

- **CreateOrUpdateUser** — first login → create user в БД, subsequent → update email/имя/фамилия из Keycloak. Tenant ID берётся из context (tenant middleware)
- **GetUserByKeycloakSub** — получение пользователя по Keycloak subject ID

### `service_test.go` — Unit тесты

8 тестов, все зелёные:
- `TestService_CreateOrUpdateUser_Create` — создание нового пользователя с tenant из context
- `TestService_CreateOrUpdateUser_Update` — обновление существующего пользователя
- `TestService_CreateOrUpdateUser_NoTenant` — создание без tenant в context
- `TestService_CreateOrUpdateUser_CreateError` — обработка ошибки создания
- `TestService_CreateOrUpdateUser_UpdateError` — обработка ошибки обновления
- `TestService_GetUserByKeycloakSub` — поиск пользователя (найден / не найден)
- `TestGenerateState` — уникальность и длина state-параметра

## Технические решения

- **Import alias**: пакет `lkfl/shared/pkg/auth` импортирован как `sharedauth` во избежание конфликта имён с `package auth`
- **State generation**: `crypto/rand` (32 байта → 64 hex) вместо uuid — более безопасный CSRF token
- **Redis key prefixes**: `auth:state:{state}` (10 мин TTL), `auth:session:{userID}` (24 часа TTL)
- **User profile**: callback и Me возвращают `user.ToProfile()` (без keycloak_sub)
- **PKCE**: TODO marker для code_challenge (не реализовано в этой итерации)
- **Role assignment**: TODO marker — роли из Keycloak пока не назначаются автоматически

## Проверки

- ✅ `go build ./...` — чистая компиляция
- ✅ `go vet ./...` — без замечаний
- ✅ `go test ./internal/auth/...` — 8/8 тестов пройдено
- ✅ `go test ./...` — все тесты проекта зелёные
- ✅ Три нуля: код не привязан к конкретному бренду, провайдеру или модели начислений

## Время

~40 минут

## Замечания

- PKCE code challenge не реализован (TODO в LoginRedirect) — требует генерации code_verifier на frontend стороне
- Authorization code flow не реализован (TODO в LoginCallback) — используется implicit flow с id_token в query param для dev
- Роли из Keycloak не назначаются автоматически (TODO в CreateOrUpdateUser) — требуется дополнительный шаг через user.RoleRepository
