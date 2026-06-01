# T2308 — Browser-based logout: инвалидация Keycloak SSO через redirect

## Веха

M23 — Frontend Polish

## Тип

code

## Контекст

После T2307 (auth overhaul) logout работает как AJAX-запрос:
1. Frontend → `POST /api/v1/auth/logout` (fetch)
2. Backend → `invalidateKeycloakSSO()` (server-to-server Admin API или POST Bearer)
3. Backend → 200 OK → frontend очищает localStorage → `navigate('/login')`

**Проблема:** server-side инвалидация (`invalidateKeycloakSSO()`) удаляет сессию в БД Keycloak, но **не трогает `KAUTH_SESSION_ID` cookie в браузере**. При повторном входе Keycloak видит живую cookie → silent redirect без формы логина.

E2E тест (`login-flow.spec.js`) подтверждает: после logout нужно вручную инвалидировать Keycloak SSO через `page.goto(kcLogoutURL)` и чистить cookies — иначе повторный вход происходит без пароля.

**Статистика:**
- Admin REST API подход (4 HTTP запроса к Keycloak): ненадёжен, зависит от service account, slow
- Fallback POST Bearer: не инвалидирует browser cookie
- Browser-based redirect (OIDC standard): удаляет `KAUTH_SESSION_ID` cookie → гарантирует инвалидацию

## Решение

Заменить AJAX logout на browser-based redirect через Keycloak logout endpoint.

### До (текущий)

```
Frontend                           Backend                           Keycloak
───────                            ───────                           ────────
  │                                  │                                 │
  │ POST /api/v1/auth/logout        │                                 │
  │ Accept: application/json ──────▶│                                 │
  │                                  │ invalidateKeycloakSSO() ──────▶│
  │                                  │ (Admin API / POST Bearer)       │
  │                                  │ ← server-side только            │
  │ ← 200 {"ok":true}               │                                 │
  │                                  │                                 │
  ❌ KAUTH_SESSION_ID cookie жив!   │                                 │
  │ clearAuth() + navigate('/login') │                                 │
  │                                  │                                 │
  │ (повторный login)                │                                 │
  │ GET /api/v1/auth/login ────────▶│ → 302 Keycloak auth ──────────▶│
  │                                  │                                 │  ❌ silent redirect (cookie жива!)
  │                                  │                                 │
```

### После (целевой)

```
Frontend                           Backend                           Keycloak
───────                            ───────                           ────────
  │                                  │                                 │
  │ window.location.href =          │                                 │
  │   '/api/v1/auth/logout' ──────▶│                                 │
  │                                  │ 1. SessionStore.Delete()        │
  │                                  │ 2. TokenStore.Delete()          │
  │                                  │ 3. clearSessionCookie()         │
  │ ← 302 Keycloak logout URL ─────│                                 │
  │                                  │                                 │
  │ GET Keycloak logout ───────────│────────────────────────────────▶│
  │                                  │                                 │
  │                                  │                                 │  ✅ Инвалидация SSO + удаление cookie
  │                                  │                                 │
  │ ← 302 /login ─────────────────│◀────────────────────────────────│
  │ (полная перезагрузка)           │                                 │
  │                                  │                                 │
  ✅ KAUTH_SESSION_ID удалена!      │                                 │
  │                                  │                                 │
  │ (повторный login)                │                                 │
  │ GET /api/v1/auth/login ────────▶│ → 302 Keycloak auth ──────────▶│
  │                                  │                                 │  ✅ Форма логина (нет cookie)
  │                                  │                                 │
```

## План

### Шаг 1: Backend — `Logout()` возвращает 302 на Keycloak logout для AJAX

**Файл:** `backend/internal/auth/handler.go`

**Что менять:**
- AJAX-запрос (`Accept: application/json`) → тоже 302 redirect на Keycloak logout
- Не возвращать 200 OK, потому что backend не знает как инвалидировать browser SSO cookie
- Session и token в Redis удалить ДО редиректа
- Session cookie очистить ДО редиректа
- `invalidateKeycloakSSO()` удалить — browser redirect делает это сам
- `getAdminToken()` и `fetchAdmin()` удалить — больше не используются

### Шаг 2: Frontend — `authStore.logout()` → `window.location.href`

**Файл:** `frontend/src/stores/authStore.ts`

**Что менять:**
- Вместо `fetch('/api/v1/auth/logout', ...)` → `window.location.href = '/api/v1/auth/logout'`
- Не нужен `try/catch` — `window.location.href` не может провалиться в catch
- Не нужно чистить localStorage/Zustand — страница перезагрузится
- Не нужен `navigateOutside('/login')` — Keycloak redirect-ит на `/login`

### Шаг 3: Убрать `invalidateKeycloakSSO()` и `getAdminToken()`

**Файл:** `backend/internal/auth/handler.go`

Удалить:
- `invalidateKeycloakSSO()` (строки ~506–614)
- `getAdminToken()` (строки ~459–505)
- `fetchAdmin()` (строки ~618–647)

### Шаг 4: Callback.tsx — исправить getRealm()

**Файл:** `frontend/src/pages/Callback.tsx`

Функция `getRealm()` использует fallback `'lkfl'` вместо `'lkfl-sdek'`.
Добавить `VITE_KEYCLOAK_REALM=lkfl-sdek` в `.env.dev`.

### Шаг 5: Обновить E2E тест

**Файл:** `login-flow.spec.js`

Убрать ручную инвалидацию Keycloak SSO — теперь logout сам делает redirect.

### Шаг 6: Обновить `doc/архитектура/риски.md`

R-042 (Keycloak SSO инвалидация) → статус: mitigated (browser-based logout)

## Критерии приёмки

- [ ] Backend: `Logout()` возвращает 302 на Keycloak logout URL (для AJAX и GET)
- [ ] Backend: `invalidateKeycloakSSO()` удалён
- [ ] Backend: `go build ./...` без ошибок
- [ ] Frontend: `authStore.logout()` использует `window.location.href`
- [ ] Frontend: `tsc --noEmit` без ошибок
- [ ] Frontend: `npm run build` без ошибок
- [ ] E2E: `npx playwright test login-flow.spec.js` — PASS (login → logout → login)
- [ ] Риск R-042 обновлён в `doc/архитектура/риски.md`

## Зависимости

- **depends_on:** T2307 (auth overhaul — session layer)
- **touches:** `internal/auth/handler.go`, `frontend/src/stores/authStore.ts`, `frontend/src/pages/Callback.tsx`, `login-flow.spec.js`
