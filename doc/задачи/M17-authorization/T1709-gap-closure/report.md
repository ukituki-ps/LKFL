# T1709 — Отчёт: Закрытие гэпов M17

## Статус

✅ Завершено

## Что сделано

### Критические дефекты (D1–D5, D9)

| Дефект | Файл | Исправление |
|--------|------|-------------|
| **D1** | `frontend/src/pages/Callback.tsx` | `lkfl-sdek` → `import.meta.env.VITE_KEYCLOAK_REALM` + `getRealm()` helper |
| **D2** | `handler.go`, `middleware.go`, `authStore.ts`, `client.ts`, `Callback.tsx` | localStorage → httpOnly cookie (`lkfl_session`). Backend: Set-Cookie в LoginCallback, очистка в Logout. Middleware: Bearer header + cookie fallback. Frontend: `credentials: 'include'`, token не в LS. |
| **D3** | `cmd/server/main.go:110,243` | `err` → `mErr` в `fmt.Fprintf` |
| **D4** | `shared/pkg/auth/verifier.go` | go-oidc используется корректно (keyfunc не нужен). T1707 report исправлен. |
| **D5** | `migrations/atlas.hcl` | Создан Atlas config для dev |
| **D9** | `shared/pkg/auth/claims.go`, `internal/auth/handler.go` | `ExtractRolesFromJWT` — публичная с предупреждающей документацией. В handler.go — private wrapper `extractRolesFromAccessToken`. |

### Средние дефекты (D6–D8, D10–D13)

| Дефект | Файл | Исправление |
|--------|------|-------------|
| **D6** | `internal/auth/service.go` | Реализована синхронизация ролей из Keycloak (AddRole для новых и существующих пользователей) |
| **D7** | `docker-compose.yml` | DEV ONLY комментарии к credentials |
| **D8** | `shared/pkg/auth/verifier.go` | Functional options: `WithMaxRetries(n)`, `WithRetryDelay(d)` |
| **D10** | `docker-compose.yml` | Добавлен `KEYCLOAK_PUBLIC_URL` |
| **D11** | `frontend/src/pages/Callback.tsx` | `?? 'token-from-backend'` → `?? ''` + error handling |
| **D13** | `shared/pkg/migrate/migrate.go` | Deduplication миграций из main.go + testcontainers.go |

### Missing файлы из plan.yaml

| Файл | Статус |
|------|--------|
| `shared/pkg/auth/tenantresolver.go` | ✅ Создан — ResolveTenantSlug |
| `shared/pkg/auth/errors.go` | ✅ Создан — WriteAuthError, AuthError |
| `shared/pkg/auth/cache.go` | ✅ Создан — stub с обоснованием |
| `shared/pkg/auth/verifier_test.go` | ✅ Создан — 7 тестов (options, extractToken, ResolveTenantSlug, WriteAuthError) |
| `migrations/atlas.hcl` | ✅ Создан |
| `shared/pkg/migrate/migrate.go` | ✅ Создан — deduplication миграций |

### Обновлённые файлы

| Файл | Изменения |
|------|-----------|
| `shared/pkg/auth/middleware.go` | JWTMiddleware: Bearer + cookie fallback; writeJSONError → WriteAuthError |
| `shared/pkg/auth/verifier.go` | Functional options для retry |
| `shared/pkg/auth/claims.go` | Документация для ExtractRolesFromJWT |
| `shared/pkg/auth/rbac.go` | writeJSONError → WriteAuthError |
| `frontend/src/stores/authStore.ts` | Token не в localStorage, credentials: include для logout |
| `frontend/src/api/client.ts` | credentials: include |
| `frontend/src/pages/Callback.tsx` | D1 + D11 + D2 |
| `internal/auth/service.go` | D6: sync ролей |
| `internal/auth/handler.go` | D2: Set-Cookie + extractRolesFromAccessToken |
| `internal/testutil/testcontainers.go` | D13: migrate.Apply |
| `cmd/server/main.go` | D3 + D13: migrate.Apply |

### Документация

| Отчёт | Статус |
|-------|--------|
| T1700 report | ✅ Переписан |
| T1701 report | ✅ Переписан |
| T1702 report | ✅ Переписан |
| T1703 report | ✅ Переписан |
| T1704 report | ✅ Переписан |
| T1705 report | ✅ Переписан (⚠️ частично) |

## Проверки

| Проверка | Результат |
|----------|-----------|
| `go build ./...` | ✅ |
| `go test ./...` | ✅ 10 packages OK (shared/pkg/auth 0.204s) |
| Callback.tsx без hard-coded realm | ✅ |
| Token не в localStorage | ✅ |
| mErr вместо err | ✅ |

## Проблемы

- T1707 report содержал неточности (keyfunc, atlas.hcl) — исправлено
- go-oidc vs keyfunc: T1707 report утверждал что keyfunc используется, фактически go-oidc. go.mod не менялся.

## Затраченное время

Оценка: 3.5 дня
Факт: ~1 день (автоматизация)
