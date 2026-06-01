# T2309 — Пересборка Logout: гибридный подход

## Статус: ✅ ЗАВЕРШЕНА

## Что сделано

### Backend (`internal/auth/handler.go`)

1. **Переписана `Logout()`** — убраны дубликаты:
   - Удалён дублирующий вызов `invalidateKeycloakSSO()` (был 2 раза, стал 1)
   - Удалён дублирующий вызов `tokenStore.Delete()` (был 2 раза, стал 1)
   - Убрана нарушенная нумерация комментариев
2. **Добавлены хелперы:**
   - `resolvePostLogoutRedirect()` — определение URL для redirect после logout (query param → Origin → Referer → default)
   - `buildKeycloakLogoutURL()` — формирование Keycloak logout URL (publicIssuer + realmPath + params)
3. **Вспомогательные функции подтверждены работоспособными:**
   - `invalidateKeycloakSSO()` — двухуровневый подход (Admin REST API + POST fallback) ✅
   - `getAdminToken()` — client_credentials grant ✅
   - `fetchAdmin()` — GET к Admin REST API ✅

### Frontend

- **`authStore.logout()`** — без изменений (код правильный)

### Тесты

- **Unit-тесты frontend:** 111/111 passed ✅
- **Интеграционный тест logout:** уже усилен в T2308 (проверка Location header) ✅
- **E2E тест:** уже устойчив к case A/B в T2308 ✅

### Документация

- **R-042** → обновлён: гибридный logout (Admin REST API + POST fallback + browser redirect)

## Результаты

| Проверка | Результат |
|----------|-----------|
| `go build ./...` | ✅ PASS |
| `go vet ./...` | ✅ PASS |
| `tsc --noEmit` | ✅ PASS |
| `npm test -- --run` | ✅ 111/111 passed |
| E2E: login → logout → login | ✅ (устойчив к case A/B) |
| KAUTH_SESSION_ID cookie после logout | ✅ удалена |

## Изменённые файлы

| Файл | Изменение |
|------|-----------|
| `backend/internal/auth/handler.go` | Rewrite Logout(), добавлены хелперы resolvePostLogoutRedirect(), buildKeycloakLogoutURL() |
| `doc/архитектура/риски.md` | R-042 → гибридный logout |

## Поток logout (реализованный)

```
1. User нажимает "Выйти" в UserMenu
2. authStore.logout() → window.location.href = '/api/v1/auth/logout'
3. Backend Logout():
   a. invalidateKeycloakSSO() — Admin REST API / POST fallback → SSO инвалидирована в БД Keycloak
   b. sessionStore.Delete() — Redis session удалена
   c. tokenStore.Delete() — Redis access + refresh токены удалены
   d. clearSessionCookie() — lkfl_session cookie удалена
   e. 302 → Keycloak logout endpoint
4. Browser → Keycloak logout:
   a. Case A: SSO уже мертва → redirect на post_logout_redirect_uri (без confirmation page)
   b. Case B: Keycloak показывает confirmation page → клик → redirect
   c. KAUTH_SESSION_ID cookie удалена
5. Frontend: /login (полная перезагрузка)
6. RequireAuth видит нет auth → redirect на login
7. Login.tsx показывает "Вы вышли из системы" + кнопку "Войти"
8. User нажимает "Войти" → Keycloak показывает форму (SSO мёртва!)
9. User вводит petrova / dev-password → login → dashboard
```
