# T2307 — Ревизия системы аутентификации: token refresh, logout, session management

## Веха

M23 — Frontend Polish (кросс-веховая задача по безопасности)

## Тип

code

## Контекст

Текущая система аутентификации LKFL имеет ряд критических архитектурных проблем, выявленных при сравнении с референсом April Ecosystem (`AUTHORIZATION_REFERENCE.md`) и анализом кодовой базы.

### Выявленные проблемы

| # | Проблема | Приоритет | Влияние |
|---|----------|-----------|---------|
| 1 | **Нет token refresh** | 🔴 Критично | Пользователь видит мерцание login→callback при каждом истечении ID Token (5 мин) |
| 2 | **Logout не убивает Keycloak SSO** | 🔴 Критично | После logout SSO сессия Keycloak жива → повторный вход без пароля |
| 3 | **ID Token как сессионный токен** | 🔴 Критично | ID Token TTL (5 мин) ≠ Redis session TTL (24 ч) → сессия «живая», но токен мёртв |
| 4 | **Нет проверки tenant_id claim** | 🟠 Серьёзно | JWT без tenant_id проходит валидацию → риск меж-tenant доступа |
| 5 | **Формат ошибок не соответствует ADR-036** | 🟡 Умеренно | `{"error":"..."}` вместо `{"code":"...","message":"...","metadata":{...}}` |
| 6 | **Open redirect на logout** | 🟡 Умеренно | `post_logout_redirect_uri` без allowlist проверки |
| 7 | **Cookie SameSite=Lax** | 🟡 Умеренно | Несовместимо с будущим subdomain multi-tenancy |
| 8 | **invalidateKeycloakSession использует ID token вместо access token** | 🟠 Серьёзно | Keycloak logout не может корректно идентифицировать сессию |

## Целевая архитектура

### Принцип: Backend Stateful Session + Keycloak SSO

Вместо хранения ID Token в cookie/Redis, вводим **server-side session token** (случайная строка 32 байта), привязанную к Keycloak subject ID. Backend остаётся stateful только для маппинга session_token → user, что соответствует практике FСТЭК-систем (audit trail, принудительный logout).

```
┌──────────┐    ┌──────────────┐    ┌───────────────────┐    ┌──────────┐
│  Browser  │    │   Nginx      │    │  lkfl-server      │    │ Keycloak │
│  (SPA)    │◄──►│  :80         │◄──►│  :8080            │◄──►│  IdP     │
│           │    │              │    │                   │    │          │
│ Session   │    │ /api/ → :8080│    │ JWT Middleware    │    │ JWKS     │
│ Cookie    │    │ /auth → KC   │    │ RBAC Middleware   │    │ Logout   │
│ (s_token) │    │              │    │ Session Store     │    │          │
└──────────┘    └──────────────┘    └───────────────────┘    └──────────┘
                                   │
                                   ▼
                              ┌──────────┐
                              │  Redis   │
                              │          │
                              │ session: │
                              │ {s_token} │ → {user_sub, tenant_id}
                              │          │
                              │ kc:token:│
                              │ {user_sub}│ → {access_token} (for KC logout)
                              └──────────┘
```

### Токены и их роли

| Токен | Где хранится | TTL | Назначение |
|-------|-------------|-----|------------|
| **server_session** | httpOnly cookie `lkfl_session` | 24 часа (настраиваемо) | Идентификация пользователя для backend API |
| **Keycloak access_token** | Redis `kc:token:{user_sub}` | 15 минут | Инвалидация Keycloak сессии при logout |
| **Keycloak refresh_token** | Redis `kc:refresh:{user_sub}` | 7 дней (Keycloak default) | Server-side silent token refresh |
| **User profile + roles** | localStorage `lkfl_user`, `lkfl_roles` | до logout | UI-состояние frontend (без секретов) |

### Поток Login (целевой)

```
1. Браузер → /login → window.location.href = /api/v1/auth/login
2. Backend:
   a. Генерирует state (32 bytes hex) + PKCE (S256)
   b. Сохраняет в Redis: auth:state:{state} = {code_verifier, redirect_uri, ts}
   c. 302 → Keycloak authorize endpoint
3. Keycloak: логин → redirect на callback
4. Backend LoginCallback:
   a. Проверка state + PKCE
   b. Обмен code на tokens (access_token + id_token + refresh_token)
   c. Верификация ID Token → claims (sub, email, roles, tenant_id)
   d. Создание server_session: random 32 bytes → Redis session:{s_token}
   e. Сохранение access_token в Redis: kc:token:{sub} (TTL 15 мин)
   f. Сохранение refresh_token в Redis: kc:refresh:{sub} (TTL 7 дней)
   g. Provisioning: CreateOrUpdateUser в БД
   h. Set-Cookie: lkfl_session={s_token}; HttpOnly; Secure; SameSite=None; Path=/; Max-Age=86400
   i. Ответ: { user, roles }
5. Frontend Callback.tsx:
   a. setAuth(null, user, roles) — token = null (cookie-only)
   b. navigate('/') → авторизованная оболочка
```

### Поток API Request (целевой)

```
1. Браузер → /api/v1/... (с cookie lkfl_session)
2. JWTMiddleware:
   a. Извлекает session_token из cookie lkfl_session
   b. Lookup в Redis: session:{s_token} → {user_sub, tenant_id, ts}
   c. Если session не найдена / истёкла → 401
   d. Получает access_token из Redis: kc:token:{user_sub}
   e. Проверяет expiration access_token:
      - Если не истёк → OK, кладём claims в context
      - Если истёк → Server-side refresh (см. ниже)
   f. Извлекает tenant_id, проверяем presence
   g. Извлекает roles из access_token claims
   h. Устанавливаем в context: Claims + Roles
3. RBACMiddleware: проверка ролей из context
4. Handler → бизнес-логика
```

### Server-side Token Refresh

```
При 401 или истечении access_token:
1. Middleware пытается refresh через Keycloak:
   POST /protocol/openid-connect/token
   grant_type=refresh_token
   refresh_token=<из Redis>
   client_id=lkfl-server
   client_secret=<env>
2. Успех:
   - Обновить kc:token:{sub} (новый access_token)
   - Обновить kc:refresh:{sub} (новый refresh_token)
   - Продолжить запрос с новыми claims
3. Неудача (refresh_token expired/revoked):
   - Удалить session из Redis
   - Вернуть 401
   - Frontend: clearAuth → redirect /login
   - Keycloak SSO session жива → silent re-login (без ввода пароля)
```

### Поток Logout (целевой)

```
1. Frontend UserMenu → authStore.logout()
   POST /api/v1/auth/logout (AJAX)
2. Backend:
   a. Извлечь session_token из cookie
   b. Удалить session из Redis: session:{s_token}
   c. Извлечь user_sub
   d. Удалить kc:token:{sub} и kc:refresh:{sub}
   e. Инвалидация Keycloak SSO:
      - Использовать client_credentials grant (lkfl-server → kc)
      - REST Admin API Keycloak: DELETE /admin/realms/{realm}/sessions/{session_uuid}
      - Фallback: POST /protocol/openid-connect/logout с access_token hint
   f. Set-Cookie: lkfl_session=; MaxAge=0 (удаление cookie)
   g. Ответ: 200 { ok: true }
3. Frontend:
   - clearAuth()
   - sessionStorage.setItem('lkfl_just_logged_out', 'true')
   - navigate('/login')
```

**Важно:** Для программной инвалидации Keycloak SSO используем **Keycloak Admin REST API** (не logout endpoint), авторизуясь как service account `lkfl-server` с `realm-management` ролью.

### Проверка tenant_id

```go
// JWTMiddleware (расширенный):
// После верификации session → extraction of claims:
if claims.TenantID == "" {
    WriteAuthError(w, http.StatusForbidden, "tenant_required")
    return 403
}
```

### Формат ошибок (унифицированный)

```json
{
    "code": "unauthorized",
    "message": "authentication failed",
    "metadata": {
        "requestId": "req-abc-123",
        "sourceService": "lkfl-server"
    }
}
```

## Что делать

### Фаза 1: Backend — Session Layer

#### Шаг 1: Session Store (`shared/pkg/auth/session.go`)

Создать абстракцию для хранения серверных сессий.

**Обоснование:** ID Token — не сессионный токен. Его TTL (5 мин) не подходит для управления пользовательской сессией (24 ч). Server-side session даёт: принудительный logout, audit trail, контроль TTL.

```go
type SessionStore struct {
    redis *redis.Client
    ttl   time.Duration
}

type SessionData struct {
    UserID    string    // Keycloak sub
    TenantID  string    // resolved from issuer
    CreatedAt time.Time
}

func (s *SessionStore) Create(ctx context.Context, data SessionData) (string, error)
func (s *SessionStore) Get(ctx context.Context, token string) (SessionData, error)
func (s *SessionStore) Delete(ctx context.Context, token string) error
func (s *SessionStore) DeleteByUserID(ctx context.Context, userID string) error
```

Redis keys:
- `lkfl:sess:{session_token}` → JSON{user_sub, tenant_id, created_at}
- TTL: 24 часа

#### Шаг 2: Token Store (`shared/pkg/auth/tokenstore.go`)

Хранение Keycloak токенов для server-side refresh и logout.

**Обоснование:** Нужен access_token для инвалидации Keycloak SSO. Нужен refresh_token для server-side silent refresh.

```go
type TokenStore struct {
    redis *redis.Client
}

func (t *TokenStore) SaveTokens(ctx context.Context, userID string, accessToken, refreshToken string) error
func (t *TokenStore) GetAccessToken(ctx context.Context, userID string) (string, error)
func (t *TokenStore) GetRefreshToken(ctx context.Context, userID string) (string, error)
func (t *TokenStore) Delete(ctx context.Context, userID string) error
```

Redis keys:
- `lkfl:kc:token:{user_sub}` → access_token (TTL 15 мин)
- `lkfl:kc:refresh:{user_sub}` → refresh_token (TTL 7 дней)

#### Шаг 3: Рефакторинг JWTMiddleware (`shared/pkg/auth/middleware.go`)

Заменить OIDC ID Token verification на session-based auth.

**Обоснование:** OIDC verifier проверяет только подпись и expiration ID Token. Не подходит для сессий длиннее TTL токена. Session-based middleware даёт контроль над жизненным циклом сессии.

Новый поток middleware:
```
1. extractSessionCookie(r) → session_token
2. sessionStore.Get(session_token) → {user_sub, tenant_id}
3. tokenStore.GetAccessToken(user_sub) → access_token
4. decodeJWT(access_token) → claims (без верификации подписи — токен наш)
5. checkExpiration(access_token):
   - OK → claims в context
   - EXPIRED → server-side refresh → новые claims
   - REFRESH_FAILED → 401
6. checkTenantID(claims) → если пусто → 403
7. claims + roles в context
```

#### Шаг 4: Server-side refresh (`shared/pkg/auth/refresher.go`)

```go
type TokenRefresher struct {
    issuer       string
    clientID     string
    clientSecret string
    tokenStore   *TokenStore
}

func (r *TokenRefresher) Refresh(ctx context.Context, userID string) (Claims, error)
```

**Обоснование:** Позволяет продлевать сессию без участия frontend. Пользователь не видит мерцания при истечении access token.

#### Шаг 5: Рефакторинг LoginCallback (`internal/auth/handler.go`)

Изменить ответ callback:
- Генерировать server_session token
- Сохранять access_token + refresh_token в TokenStore
- Cookie: session_token (не ID Token)
- Ответ: `{ user, roles }` — без token в body

**Обоснование:** Frontend больше не получает токен → нет XSS-риска, нет токена в localStorage.

#### Шаг 6: Рефакторинг Logout (`internal/auth/handler.go`)

```go
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    // 1. Session cleanup
    sessionToken := extractSessionCookie(r)
    session, _ := h.sessionStore.Get(r.Context(), sessionToken)
    
    // 2. Token cleanup
    h.tokenStore.Delete(r.Context(), session.UserID)
    
    // 3. Session cleanup
    h.sessionStore.Delete(r.Context(), sessionToken)
    
    // 4. Keycloak SSO invalidation via Admin REST API
    h.invalidateKeycloakSSO(r.Context(), session.UserID)
    
    // 5. Cookie cleanup
    http.SetCookie(w, &http.Cookie{
        Name: "lkfl_session", Value: "", MaxAge: -1,
        HttpOnly: true, Secure: true, SameSite: http.SameSiteNoneMode, Path: "/",
    })
    
    w.WriteHeader(http.StatusOK)
    shhttp.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) invalidateKeycloakSSO(ctx context.Context, userID string) {
    // Option A: Keycloak Admin REST API (preferred)
    // GET /admin/realms/{realm}/users?username={username} → session UUID
    // DELETE /admin/realms/{realm}/sessions/{session_uuid}
    // Auth: service account client_credentials grant
    
    // Option B: POST /protocol/openid-connect/logout with access_token hint
    // (fallback, less reliable)
}
```

**Обоснование:** Текущий logout использует ID Token hint → ненадёжно. Admin REST API гарантирует инвалидацию.

### Фаза 2: Backend — Error format + Tenant validation

#### Шаг 7: Унификация формата ошибок (`shared/pkg/auth/errors.go`)

```go
type AuthErrorResponse struct {
    Code     string                 `json:"code"`
    Message  string                 `json:"message"`
    Metadata map[string]string      `json:"metadata,omitempty"`
}

func WriteAuthError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
    metadata := map[string]string{}
    if reqID := chi.RouteContext(r.Context()); reqID != nil {
        metadata["requestId"] = reqID.Value("request-id").(string)
    }
    metadata["sourceService"] = "lkfl-server"
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(AuthErrorResponse{Code: code, Message: message, Metadata: metadata})
}
```

**Обоснование:** Соответствие ADR-036, observability, frontend может различать типы ошибок.

#### Шаг 8: Tenant ID validation в middleware

```go
// После extraction claims:
if claims.TenantID == "" {
    WriteAuthError(w, http.StatusForbidden, "tenant_required", 
        "JWT missing tenant_id claim", r)
    return
}
```

**Обоснование:** Критическое требование multi-tenancy isolation (ADR-036). Без tenant_id — риск cross-tenant access.

#### Шаг 9: Open redirect защита на logout

```go
var allowedPostLogoutOrigins = []string{
    os.Getenv("FRONTEND_URL"), // http://localhost:5173, https://sdek.lkfl.ru
}

func isValidPostLogoutRedirect(uri string) bool {
    parsed, err := url.Parse(uri)
    if err != nil {
        return false
    }
    for _, origin := range allowedPostLogoutOrigins {
        if strings.HasPrefix(parsed.String(), origin) {
            return true
        }
    }
    return false
}
```

**Обоснование:** OWASP A01 — open redirect после logout может перенаправить на фишинговый сайт.

### Фаза 3: Frontend

#### Шаг 10: Рефакторинг authStore (`frontend/src/stores/authStore.ts`)

```typescript
interface AuthState {
    // Убрано: token (больше не нужно — cookie-only)
    user: UserProfile | null
    userRoles: UserRole[]
    isAuthenticated: boolean
    isLoading: boolean

    setAuth: (user: UserProfile, roles: UserRole[]) => void
    logout: () => Promise<void>
    clearAuth: () => void
}
```

**Изменения:**
- Убрать `token` поле (не используется, cookie-only архитектура)
- `setAuth()` принимает только user + roles
- `logout()` — без изменений (AJAX POST + clearAuth + redirect)
- `checkAuthSession()` — без изменений (credentials: include)

**Обоснование:** Упрощение API store, удаление устаревшего поля.

#### Шаг 11: Рефакторинг api/client.ts

```typescript
// При 401:
if (response.status === 401) {
    // Session истекла ИЛИ refresh_token невалиден
    // → clearAuth + redirect на login
    // → Keycloak SSO session может быть жива → silent re-login
    useAuthStore.getState().clearAuth()
    window.location.href = '/login'
    throw new Error('Session expired')
}
```

**Обоснование:** Без client-side refresh (токен в cookie, backend управляет сессией) — 401 означает конец сессии.

#### Шаг 12: Удалить Callback.tsx проверку `token` из body

```typescript
// Убрать:
// if (!token) { setError('...'); return }

// Заместить на:
setAuth(data.user, data.roles ?? [])
navigate('/', { replace: true })
```

**Обоснование:** Backend больше не возвращает token в body.

### Фаза 4: Cookie configuration

#### Шаг 13: Cookie SameSite + Secure

```go
http.SetCookie(w, &http.Cookie{
    Name:     "lkfl_session",
    Value:    sessionToken,
    HttpOnly: true,
    Secure:   true,       // Всегда true (требует HTTPS)
    SameSite: http.SameSiteNoneMode,  // Для subdomain multi-tenancy
    Path:     "/",
    MaxAge:   86400,      // 24 часа
    Domain:   ".lkfl.ru", // Production; пустая строка для localhost
})
```

**Обоснование:** SameSite=None необходим для subdomain-based multi-tenancy (sdek.lkfl.ru, acme.lkfl.ru). Secure=true требуется для SameSite=None.

### Фаза 5: Тестирование

#### Шаг 14: Интеграционные тесты

- Login flow: state + PKCE + callback → session created
- API request with valid session → 200
- API request with invalid session → 401
- Server-side refresh: expired access_token → refresh → 200
- Logout: session deleted + Keycloak SSO invalidated
- Logout: cookie cleared
- Open redirect protection
- Tenant ID missing → 403

#### Шаг 15: E2E тесты (Playwright)

- Full login flow: browser → Keycloak → callback → dashboard
- Session expiration → server-side refresh (transparent)
- Logout → redirect to /login → no automatic re-login
- API request after logout → 401

## Требования

- **Безопасность:**
  - Все токены в httpOnly cookie (не localStorage, не memory)
  - Server-side session management (удалённая инвалидация)
  - Keycloak SSO инвалидация при logout
  - tenant_id обязательный claim (403 без него)
  - Open redirect защита на logout

- **Совместимость:**
  - Cookie: SameSite=None; Secure (для subdomain multi-tenancy)
  - Формат ошибок: ADR-036 (`{code, message, metadata}`)
  - Frontend API: без изменений для consumers (credentials: include)

- **Надёжность:**
  - Server-side refresh token (прозрачный для пользователя)
  - Fallback при неудачном refresh → 401 → re-login
  - Graceful degradation при недоступности Keycloak Admin API

## Критерии приёмки

### Backend

- [ ] `shared/pkg/auth/session.go` — SessionStore реализован
- [ ] `shared/pkg/auth/tokenstore.go` — TokenStore реализован
- [ ] `shared/pkg/auth/refresher.go` — server-side refresh реализован
- [ ] `JWTMiddleware` использует session-based auth (не OIDC ID Token)
- [ ] LoginCallback генерирует session token + сохраняет KC tokens
- [ ] Logout удаляет session + tokens + инвалидирует Keycloak SSO
- [ ] tenant_id claim валидация (403 без него)
- [ ] Формат ошибок: `{code, message, metadata}`
- [ ] Open redirect защита на logout
- [ ] Cookie: SameSite=None; Secure; HttpOnly

### Frontend

- [ ] `authStore.ts`: убрано `token` поле
- [ ] `setAuth()` принимает только user + roles
- [ ] `api/client.ts`: 401 → clearAuth → redirect /login
- [ ] `Callback.tsx`: убрана проверка `token` из body
- [ ] `tsc --noEmit` без ошибок
- [ ] `npm run build` без ошибок

### Тесты

- [ ] Интеграционные тесты: login, session, refresh, logout
- [ ] E2E тест: full login flow
- [ ] E2E тест: logout → Keycloak SSO invalidated
- [ ] Unit тесты: SessionStore, TokenStore, Refresher

## Зависимости

- **depends_on:** M17-authorization (T1700-T1714), M19-auth-rbac (T1901-T1906)
- **touches:** `shared/pkg/auth/`, `internal/auth/`, `frontend/src/stores/`, `frontend/src/api/`, `frontend/src/pages/`
- **infra:** Keycloak realm config — service account `lkfl-server` с `realm-management` ролью
