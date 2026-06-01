# T2308 — Browser-based logout: инвалидация Keycloak SSO через redirect

## Статус: ✅ ВЫПОЛНЕНО

## Что сделано

### Backend (`internal/auth/handler.go`)

1. **Logout() → 302 Keycloak logout** — вместо AJAX `200 OK` возвращает 302 redirect на Keycloak logout endpoint
2. **Удалён `invalidateKeycloakSSO()`** (~185 строк) — больше не нужен, browser-based logout делает это сам
3. **Удалён `getAdminToken()`** и `fetchAdmin()` — использовались только для server-side SSO invalidation
4. **GET вместо POST** — `window.location.href` шлёт GET, router изменён с `Post` на `Get`
5. **post_logout_redirect_uri** — определяется из `Origin`/`Referer` заголовков (не из `publicIssuer`!)
6. **realmPath** — извлекается из `issuer` (`/realms/lkfl-sdek`) + `publicIssuer` = корректный logout URL

### Frontend (`frontend/src/stores/authStore.ts`)

1. **`logout()` → `window.location.href`** — browser-based redirect вместо AJAX fetch
2. **Очистка state ДО redirect** — localStorage/sessionStorage чистятся на случай сбоев backend/Keycloak
3. **Убран `navigateOutside`** — больше не нужен

### Frontend (`frontend/.env.dev`)

1. **`VITE_KEYCLOAK_REALM=lkfl-sdek`** — исправлен fallback realm в `Callback.tsx`

### Интеграционные тесты

1. **`integration_test.go`** — `PostWithToken` → `GetWithToken`
2. **`testcontainers.go`** — `r.Post` → `r.Get` для `/logout`

### E2E тест (`login-flow.spec.js`)

```
✓ login-flow.spec.js:26:3 › Login Flow (browser-based logout) (7.6s)
```

Полный поток:
1. Login → dashboard ✅
2. Logout (avatar → Выйти → Keycloak logout confirmation → #kc-logout) ✅
3. SSO инвалидирована (KAUTH_SESSION_ID cookie удалена) ✅
4. Второй login → Keycloak показывает форму логина ✅
5. Dashboard → avatar visible ✅

### Unit-тесты (`authStore.test.ts`)

- [x] Все 111 тестов прошли

## Результаты

| Проверка | Результат |
|----------|-----------|
| `go build ./...` | ✅ OK |
| `tsc --noEmit` | ✅ OK |
| `npm run build` | ✅ OK |
| `npm test -- --run` | ✅ OK (111 passed) |
| E2E: login → logout → login | ✅ OK (7.6s) |

## Изменённые файлы

| Файл | Изменение |
|------|-----------|
| `internal/auth/handler.go` | Logout() → 302 Keycloak; удалён invalidateKeycloakSSO/getAdminToken/fetchAdmin |
| `internal/app/server.go` | `Post("/logout")` → `Get("/logout")` |
| `internal/testutil/testcontainers.go` | `Post` → `Get` для `/logout` |
| `internal/auth/integration_test.go` | `PostWithToken` → `GetWithToken` |
| `frontend/src/stores/authStore.ts` | logout() → window.location.href; убран navigateOutside |
| `frontend/.env.dev` | Добавлен VITE_KEYCLOAK_REALM=lkfl-sdek |
| `frontend/src/stores/authStore.test.ts` | Обновлены logout-тесты |
| `doc/архитектура/риски.md` | R-042 → mitigated |
