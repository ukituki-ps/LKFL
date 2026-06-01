# T2309 — Пересборка Logout: гибридный подход

## Веха

M23 — Frontend Polish

## Тип

code

## Контекст

### Проблема

Logout не работает в реальном браузере. T2308 реализовал browser-based logout (`window.location.href → 302 Keycloak logout`), но Keycloak показывает страницу подтверждения (`#kc-logout`) которую **никто не кликает**. SSO сессия остаётся живой → при редиректе на `/login` → `RequireAuth` → `/api/v1/auth/login` → Keycloak → silent SSO login → callback → dashboard. Пользователь попадает обратно в систему.

Тест Playwright проходил за 7.6s потому что скрипт явным образом кликал `#kc-logout`. В реальном браузере — провал.

### Состояние handler.go (КОРРУПИРОВАНО)

Файл `backend/internal/auth/handler.go` (784 строки) находится в повреждённом состоянии после цикла inline-редактирований в сессии `ses_17b6a131dffeAmY5b6NKxTUVdd`:

| Строки | Проблема |
|--------|----------|
| 482-492 | `invalidateKeycloakSSO()` вызывается **ДВАЖДЫ** (дубликат блока) |
| 495-503 | `tokenStore.Delete()` вызывается **ДВАЖДЫ** (дубликат блока) |
| 488 | Комментарий `// 2.` — на самом деле дубликат шага 1 |
| 505 | Комментарий `// 3.` — нумерация нарушена |

### Что работает и НЕ трогать

| Компонент | Файл | Статус |
|-----------|------|--------|
| SessionStore | `shared/pkg/auth/session.go` | ✅ Идеален |
| TokenStore | `shared/pkg/auth/tokenstore.go` | ✅ Идеален |
| TokenRefresher | `shared/pkg/auth/refresher.go` | ✅ Идеален |
| SessionMiddleware | `shared/pkg/auth/middleware.go` | ✅ Идеален |
| Claims extraction | `shared/pkg/auth/claims.go` | ✅ Идеален |
| OIDC Verifier | `shared/pkg/auth/verifier.go` | ✅ Идеален |
| Error responses | `shared/pkg/auth/errors.go` | ✅ Идеален |
| RBAC | `shared/pkg/auth/rbac.go` | ✅ Идеален |
| Auth Service | `internal/auth/service.go` | ✅ Идеален |
| LoginRedirect | `handler.go:135-186` | ✅ Идеален |
| LoginCallback | `handler.go:189-363` | ✅ Идеален |
| Cookie helpers | `handler.go:365-422` | ✅ Идеален |
| `extractRealmSlug` | `handler.go:449-457` | ✅ Идеален |
| `isBrowserRequest` | `handler.go:430-442` | ✅ Идеален |
| `Me` handler | `handler.go:756-772` | ✅ Идеален |
| `isValidPostLogoutRedirect` | `handler.go:729-753` | ✅ Идеален |
| Frontend authStore (кроме logout) | `authStore.ts` | ✅ Идеален |
| Frontend Callback.tsx | `Callback.tsx` | ✅ Идеален |
| Frontend Login.tsx | `Login.tsx` | ✅ Идеален |
| Frontend RequireAuth | `RequireAuth.tsx` | ✅ Идеален |

### Решение — Гибридный logout

```
1. Server-side: Admin REST API → удалить Keycloak SSO сессию в БД Keycloak
   Fallback: POST /protocol/openid-connect/logout с Bearer token
2. Server-side: Redis → удалить session + token + refresh
3. Browser: 302 → Keycloak logout → очистить KAUTH_SESSION_ID cookie → redirect на /login
```

Порядок КРИТИЧЕСКИЙ:
- invalidateKeycloakSSO() — ДО удаления токенов (fallback POST logout нужен access_token)
- tokenStore.Delete() — после инвалидации SSO
- Redirect на Keycloak — последним

## План

### Шаг 1: Rewrite `Logout()` в handler.go

**Файл:** `backend/internal/auth/handler.go`, строки 459-545

**Что удалить:**
- Строки 459-545 целиком (текущий коррумпированный `Logout`)

**Что добавить:**

```go
// Logout — гибридная инвалидация сессии.
//
// Поток:
//   1. Server-side: Keycloak SSO invalidation (Admin REST API → fallback POST logout)
//   2. Server-side: Redis cleanup (session + tokens)
//   3. Browser-based: 302 → Keycloak logout → очистка KAUTH_SESSION_ID cookie
//
// Шаг 1 вызывается ДО шага 2, потому что fallback POST logout
// требует access_token из TokenStore.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == protoHTTPS

	// ── Извлечение сессии ──
	sessionToken := sharedauth.ExtractSessionCookie(r)
	var sessionData sharedauth.SessionData
	var hasSession bool

	if sessionToken != "" {
		var err error
		sessionData, err = h.sessionStore.Get(r.Context(), sessionToken)
		if err == nil {
			hasSession = true
		}
	}

	// ── 1. Server-side Keycloak SSO invalidation (ДО удаления токенов) ──
	if hasSession && sessionData.UserID != "" {
		h.invalidateKeycloakSSO(r.Context(), sessionData)
	}

	// ── 2. Очистка server-side state (Redis) ──
	if sessionToken != "" {
		_ = h.sessionStore.Delete(r.Context(), sessionToken)
	}
	if hasSession && sessionData.UserID != "" {
		_ = h.tokenStore.Delete(r.Context(), sessionData.UserID)
	}

	// ── 3. Очистка session cookie ──
	clearSessionCookie(w, isSecure)

	// ── 4. Browser-based Keycloak SSO logout ──
	postLogoutRedirect := h.resolvePostLogoutRedirect(r)
	logoutURL := h.buildKeycloakLogoutURL(postLogoutRedirect)

	slog.Debug("logout redirect to keycloak", "logout_url", logoutURL)
	http.Redirect(w, r, logoutURL, http.StatusFound)
}
```

**Обоснование:** Один чистый метод, без дубликатов. Логика вынесена в хелперы для читаемости. Порядок строго: SSO invalidation → Redis cleanup → cookie cleanup → redirect.

### Шаг 2: Добавить хелпер `resolvePostLogoutRedirect`

**Файл:** `backend/internal/auth/handler.go`, после `Logout()`

**Что добавить:**

```go
// resolvePostLogoutRedirect определяет URL для redirect после logout.
//
// Приоритет:
//   1. Query param: ?post_logout_redirect_uri=...
//   2. Origin header
//   3. Referer header (parsed to origin)
//   4. Default: http://localhost:5173/login
func (h *Handler) resolvePostLogoutRedirect(r *http.Request) string {
	// 1. Query param (validирован по allowlist)
	if redirect := r.URL.Query().Get("post_logout_redirect_uri"); redirect != "" {
		if isValidPostLogoutRedirect(redirect) {
			return redirect
		}
	}

	// 2. Origin header
	if origin := r.Header.Get("Origin"); origin != "" {
		return origin + "/login"
	}

	// 3. Referer header
	if ref := r.Header.Get("Referer"); ref != "" {
		if u, err := url.Parse(ref); err == nil {
			return u.Scheme + "://" + u.Host + "/login"
		}
	}

	// 4. Default
	return "http://localhost:5173/login"
}
```

**Обоснование:** Выделяет логику определения frontend URL из Logout. Упрощает тестирование. Приоритет query param → Origin → Referer → default обеспечивает гибкость (прямой вызов из браузера, SPA, E2E тест).

### Шаг 3: Добавить хелпер `buildKeycloakLogoutURL`

**Файл:** `backend/internal/auth/handler.go`, после `resolvePostLogoutRedirect`

**Что добавить:**

```go
// buildKeycloakLogoutURL формирует Keycloak logout URL.
//
// Формула: publicIssuer + realmPath + /protocol/openid-connect/logout
//
// publicIssuer = http://localhost:8081 (публичный URL Keycloak)
// realmPath = /realms/lkfl-sdek (извлечён из issuer)
//
// Результат: http://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/logout
func (h *Handler) buildKeycloakLogoutURL(postLogoutRedirect string) string {
	realmPath := ""
	if idx := strings.Index(h.issuer, "/realms/"); idx >= 0 {
		realmPath = h.issuer[idx:]
	}

	return fmt.Sprintf(
		"%s%s/protocol/openid-connect/logout?client_id=%s&post_logout_redirect_uri=%s",
		h.publicIssuer,
		realmPath,
		url.QueryEscape(h.clientID),
		url.QueryEscape(postLogoutRedirect),
	)
}
```

**Обоснование:** Выделяет логику сборки Keycloak URL. Критично: использует `publicIssuer` (публичный URL, доступный из браузера), а не `issuer` (внутренний docker URL `keycloak:8080`).

### Шаг 4: Удалить дубликаты из handler.go

**Файл:** `backend/internal/auth/handler.go`

**Что удалить:**
- Строка 488: `// 2. Server-side Keycloak SSO invalidation — ДО удаления токенов!` (дубликат)
- Строки 489-492: дублирующий блок `invalidateKeycloakSSO` вызова
- Строки 499-503: дублирующий блок `tokenStore.Delete` вызова

**Обоснование:** Код в коррумпированном состоянии — два вызова invalidateKeycloakSSO и два вызова tokenStore.Delete. После rewrite Logout() (Шаг 1) эти строки исчезнут автоматически.

### Шаг 5: Проверить `invalidateKeycloakSSO`

**Файл:** `backend/internal/auth/handler.go`, строки 547-650

**Текущее состояние:** Функция существует (восстановлена агентом в цикле rebuild), но может содержать артефакты inline-редактирования.

**Что проверить:**
- [ ] Функция вызывает `getAdminToken()` → Admin REST API lookup по email
- [ ] Fallback: POST logout с `access_token` из `tokenStore`
- [ ] Функция НЕ удаляет токены из Redis (это делает Logout())
- [ ] Функция НЕ redirect-ит (это делает Logout())

**Если функция повреждена — переписать целиком (см. Шаг 6).**

### Шаг 6: Переписать `invalidateKeycloakSSO` (если повреждена)

**Файл:** `backend/internal/auth/handler.go`

**Целевая версия:**

```go
// invalidateKeycloakSSO программно инвалидирует Keycloak SSO сессию.
//
// Двухуровневый подход:
//   1. Admin REST API — точечное удаление сессий по email пользователя
//   2. Fallback — POST logout с access_token hint (если Admin API недоступен)
//
// Важно: вызывается ДО удаления токенов из TokenStore.
func (h *Handler) invalidateKeycloakSSO(ctx context.Context, sd sharedauth.SessionData) {
	realm := extractRealmSlug(h.issuer)
	if realm == "" {
		slog.Warn("keycloak SSO invalidation skipped", "reason", "cannot extract realm from issuer")
		return
	}

	adminBase := strings.Split(h.issuer, "/realms/")[0]
	if adminBase == "" {
		adminBase = h.issuer
	}

	// ── Попытка 1: Admin REST API ──
	adminToken, err := h.getAdminToken(ctx)
	if err != nil {
		slog.Debug("admin REST API unavailable, trying fallback", "error", err.Error())
	} else if sd.Email != "" {
		userUUID := h.findUserByUUID(ctx, adminToken, adminBase, realm, sd.Email)
		if userUUID != "" {
			deleted := h.deleteUserSessions(ctx, adminToken, adminBase, realm, userUUID)
			if deleted > 0 {
				slog.Info("keycloak SSO invalidated via Admin REST API", "deleted", deleted)
				return
			}
		}
	}

	// ── Попытка 2: Fallback — POST logout с access_token ──
	accessToken, err := h.tokenStore.GetAccessToken(ctx, sd.UserID)
	if err != nil {
		slog.Warn("keycloak SSO invalidation failed",
			"admin_api", "unavailable", "access_token", "not found")
		return
	}

	logoutURL := fmt.Sprintf(
		"%s/protocol/openid-connect/logout?client_id=%s&post_logout_redirect_uri=%s",
		h.issuer,
		url.QueryEscape(h.clientID),
		url.QueryEscape(h.publicIssuer+"/"),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logoutURL, nil)
	if err != nil {
		slog.Warn("failed to build keycloak logout request", "error", err.Error())
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("keycloak POST logout failed", "error", err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.ReadAll(resp.Body)

	slog.Info("keycloak SSO invalidated via fallback POST logout", "status", resp.StatusCode)
}
```

**Обоснование:** Чистая перепись функции. Два уровня: Admin API (точный) → POST fallback (надёжный). Не трогает Redis (это делает Logout). Не делает redirect (это делает Logout).

### Шаг 7: Проверить `getAdminToken` и `fetchAdmin`

**Файл:** `backend/internal/auth/handler.go`

**Что проверить:**
- [ ] `getAdminToken()` — client_credentials grant, возвращает admin access token
- [ ] `fetchAdmin()` — GET запрос к Admin REST API

Если функции существуют и не повреждены — оставить. Если повреждены — переписать.

### Шаг 8: `go build` + `go vet`

**Команда:**
```bash
go build ./...
go vet ./...
```

**Ожидаемый результат:** Без ошибок.

**Обоснование:** Проверка компиляции и статического анализа после изменения handler.go.

### Шаг 9: Frontend — `authStore.logout()` — без изменений

**Файл:** `frontend/src/stores/authStore.ts`

**Текущий код (строки 100-123):**

```typescript
logout: async () => {
    localStorage.removeItem(LS_USER)
    localStorage.removeItem(LS_ROLES)
    set({ user: null, userRoles: [], isAuthenticated: false })
    sessionStorage.removeItem('lkfl_login_redirecting')
    sessionStorage.removeItem('lkfl_login_attempts')
    sessionStorage.setItem('lkfl_just_logged_out', 'true')
    window.location.href = '/api/v1/auth/logout'
}
```

**Не менять.** Код правильный:
- Очищает frontend state ДО redirect (защита от сбоев backend)
- Устанавливает `lkfl_just_logged_out` → Login.tsx покажет кнопку «Войти»
- `window.location.href` → backend Logout() → server-side SSO invalidation → 302 Keycloak → /login

**Обоснование:** Фронтенд делает то, что должен. Проблема была на бэкенде.

### Шаг 10: Unit-тесты frontend

**Команда:**
```bash
cd frontend && npm test -- --run
```

**Ожидаемый результат:** 111/111 passed

**Обоснование:** Frontend не изменён, тесты должны пройти.

### Шаг 11: Интеграционный тест logout

**Файл:** `backend/internal/auth/integration_test.go`, строки 180-198

**Текущий тест:**
```go
func TestLogout(t *testing.T) {
    // ...
    resp, err := ts.GetWithToken("/api/v1/auth/logout", adminToken)
    // ...
    if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
        t.Logf("Logout returned status %d (may vary)", resp.StatusCode)
    }
}
```

**Что изменить:** Усилить тест:
- [ ] Проверить `Location` header содержит `protocol/openid-connect/logout`
- [ ] Проверить `post_logout_redirect_uri` в Location header

**Обоснование:** Текущий тест почти ничего не проверяет. После исправления logout нужно убедиться что 302 + правильный URL.

### Шаг 12: E2E тест

**Файл:** `login-flow.spec.js`

**Что изменить:**
- [ ] Заменить жёсткий клик `#kc-logout` на устойчивую обработку (строки 72-75):
  - Case A: `#kc-logout` видим (timeout 3s) → кликнуть → confirmation page обработана
  - Case B: `#kc-logout` не видим → OK, SSO уже инвалидирована server-side, Keycloak redirect-ит автоматически
- [ ] Добавить проверку: `KAUTH_SESSION_ID` cookie отсутствует после logout

**Почему это сработает:**
- Backend инвалидирует SSO через Admin REST API / POST fallback ДО browser redirect
- Keycloak logout endpoint получает request → два сценария:
  - SSO уже мертва в БД → Keycloak может сразу redirect без confirmation page (case B)
  - Keycloak всё равно показывает confirmation page (cookie ещё в браузере) → кликаем (case A)
- В обоих случаях после logout SSO инвалидирована и auto-login невозможен

**Обоснование:** T2308-версия теста работала ТОЛЬКО с ручным кликом `#kc-logout`. Новая версия работает в обоих сценариях — server-side инвалидация делает SSO мёртвой, и Keycloak не сможет auto-login при повторном входе.

### Шаг 13: Обновить риски

**Файл:** `doc/архитектура/риски.md`

**R-042:** Обновить статус:
- Было: `✅ Закрыт` (T2308 browser-based logout)
- Стало: `✅ Закрыт` (T2309 гибридный logout: Admin REST API + POST fallback + browser redirect)

**Обоснование:** R-042 был закрыт T2308, но на самом деле не решён. T2309 закрывает его окончательно гибридным подходом.

### Шаг 14: Создать `report.md`

**Файл:** `doc/задачи/M23-frontend-polish/T2309-auth-rebuild/report.md`

**Содержание:**
- Что сделано
- Изменённые файлы
- Результаты проверок
- Поток logout после исправления

### Шаг 15: Создать `plan.yaml`

**Файл:** `doc/задачи/M23-frontend-polish/T2309-auth-rebuild/plan.yaml`

**Содержание:** Checklist всех шагов (1-14), 100% progress.

## Критерии приёмки

- [ ] `handler.go` — нет дубликатов, `go build ./...` без ошибок
- [ ] `invalidateKeycloakSSO()` вызывается ОДИН раз, ДО удаления токенов
- [ ] `Logout()` возвращает 302 на правильный Keycloak logout URL
- [ ] `go vet ./...` без ошибок
- [ ] `tsc --noEmit` без ошибок
- [ ] `npm test -- --run` — 111/111 passed
- [ ] Интеграционный тест: `TestLogout` проверяет Location header (302 + `protocol/openid-connect/logout` + params)
- [ ] E2E: login → logout → login — PASS (устойчив к case A/B confirmation page)
- [ ] `KAUTH_SESSION_ID` cookie удалена после logout
- [ ] R-042 обновлён в `doc/архитектура/риски.md`

## Зависимости

- **depends_on:** T2307 (auth overhaul — session layer), T2308 (browser-based logout — частичная реализация)
- **extends:** T2308 (добавляет server-side SSO invalidation поверх browser-based redirect)
- **touches:** `backend/internal/auth/handler.go` (основной), `backend/internal/auth/integration_test.go`, `login-flow.spec.js`, `doc/архитектура/риски.md`

## Риски

| Риск | Митигация |
|------|-----------|
| Admin REST API недоступен (нет `KEYCLOAK_CLIENT_SECRET`) | Fallback POST logout с access_token |
| Access token уже истёк (TTL 15 мин в Redis) | Fallback может не сработать → Keycloak logout page всё равно очистит cookie при ручном клике |
| `post_logout_redirect_uri` не в allowed list Keycloak | Добавить `http://localhost:5173/login` в Keycloak client config |
| Keycloak недоступен при logout | Backend всё равно очистит Redis + cookie; redirect на Keycloak покажет ошибку, но пользователь уже выведен |
