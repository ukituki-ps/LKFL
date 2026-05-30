# T1709 — Закрытие гэпов M17 (Gap Closure)

> **Тип:** code. Реактивная задача, порождённая аудитом реализации M17 (T1700–T1708).
> **Цель:** устранить все критические и средние дефекты, закрытые в T1707,
> добавить отсутствующие файлы из plan.yaml и привести документацию в соответствие с кодом.

## Контекст

Аудит реализации M17-authorization выявил следующие категории проблем:

1. **Критические (безопасность + production bugs):** D1–D5
2. **Средние (качество кода + missing файлы из plan.yaml):** D6–D12
3. **Низкие (стиль + cleanup):** D13–D16
4. **Документация:** все отчёты T1700–T1705 показывают «⏳ Не начато» при полном коде

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1707 (fixes), T1708 (CI/CD Foundation)
**ADR:** ADR-036 (Authorization System), ADR-003 (Keycloak), ADR-009 (Multi-tenancy)

---

## Критические дефекты

### D1 — Hard-coded realm `lkfl-sdek` в Callback.tsx

```
frontend/src/pages/Callback.tsx:48-49
window.location.href =
    `${currentOrigin}/realms/lkfl-sdek/protocol/openid-connect/logout`
```

Нарушает «Нулевая привязка к бренду». Realm должен определяться из tenant resolution
или `VITE_KEYCLOAK_REALM`.

### D2 — Token в localStorage (XSS уязвимость, 152-ФЗ)

```
frontend/src/stores/authStore.ts:36-54
localStorage.setItem(LS_TOKEN, token)
```

Token в `localStorage` уязвим к XSS. Для платформы с ПДн и 152-ФЗ
должен использоваться httpOnly cookie от backend.

**Решение:** backend `LoginCallback` → `Set-Cookie: lkfl_session=...; HttpOnly; SameSite=Lax; Path=/`,
frontend `fetch({credentials: 'include'})`, убрать `LS_TOKEN` из authStore.

### D3 — Bug: неверная переменная ошибки в `runMigrate`/`runSeed`

```go
// backend/cmd/server/main.go:110
if mErr := runMigrationsDB(ctx, conn); mErr != nil {
    fmt.Fprintf(os.Stderr, "migrations error: %v\n", err) // ← err от pgx.Connect(), не mErr!
    os.Exit(1)
}
// то же на строке 243 в runSeed
```

При ошибке миграции пользователю показывается сообщение об ошибке подключения,
а не реальной причине сбоя миграции.

### D4 — `MicahParks/keyfunc/v3` не в go.mod

T1707 report заявляет:
> ✅ `shared/pkg/auth/verifier.go`: полноценная RSA JWKS через keyfunc.NewDefault

Но `go.mod` содержит только `go-oidc`. Фактическая реализация в `verifier.go`
использует `go-oidc.NewProvider` (которая сама загружает JWKS), а не keyfunc.

**Решение:** либо добавить keyfunc и переделать verifier.go, либо исправить T1707 report.

### D5 — `atlas.hcl` отсутствует

T1707 report заявляет:
> ✅ `migrations/atlas.hcl` — Atlas config dev environment

Файл `backend/migrations/atlas.hcl` не существует.

---

## Средние дефекты

### D6 — Unused `roles` в `CreateOrUpdateUser`

`backend/internal/auth/service.go:97`: `_ = roles` с TODO. Нужно реализовать
или убрать параметр.

### D7 — Dev credentials в docker-compose.yml

`changeme_dev_password` и `admin` — добавить комментарий и/или использовать
`.env.example`.

### D8 — Hard-coded retry (30×2s=60s) в verifier

Нет конфигурируемости timeout/startup-retry.

### D9 — `ExtractRolesFromJWT` без верификации подписи

`backend/shared/pkg/auth/claims.go:116-133`: decode base64 payload без верификации.
Может использоваться с поддельным токеном для extraction ролей.
**Решение:** сделать private или добавить verif.

### D10 — `KEYCLOAK_PUBLIC_URL` отсутствует в docker-compose.yml

CI проверяет наличие в staging/prod compose, но dev compose не имеет этого env var.

### D11 — Callback.tsx fallback `'token-from-backend'`

`frontend/src/pages/Callback.tsx:73`: `?? 'token-from-backend'` — fallback строка
не должна использоваться в production.

### D12 — `atlas.hcl` absent (см. D5)

---

## Отсутствующие файлы из plan.yaml

| Плановый файл (T1702 plan.yaml) | Статус |
|----------------------------------|--------|
| `shared/pkg/auth/tenantresolver.go` | ❌ отсутствует (функционал inline в middleware.go) |
| `shared/pkg/auth/errors.go` | ❌ отсутствует (writeJSONError inline в middleware.go) |
| `shared/pkg/auth/cache.go` | ❌ отсутствует (JWKS cache не реализован) |
| `shared/pkg/auth/verifier_test.go` | ❌ отсутствует (упоминается в T1707 report) |
| `internal/api/` | ⚠️ отсутствует (consolidated в internal/app/server.go — допустимо) |
| `migrations/atlas.hcl` | ❌ отсутствует |

---

## Отсутствующая документация

| Отчёт | Фактический статус | Требуется |
|-------|-------------------|-----------|
| T1700 report | «⏳ Не начато» | Переписать ✅ |
| T1701 report | «⏳ Не начато» | Переписать ✅ |
| T1702 report | «⏳ Не начато» | Переписать ✅ |
| T1703 report | «⏳ Не начато» | Переписать ✅ |
| T1704 report | «⏳ Не начато» | Переписать ✅ |
| T1705 report | «⏳ Не начато (отложено)» | Переписать ⚠️ частично реализовано |

---

## Что НЕ входит

- Рефакторинг всего auth flow — только точечные исправления
- Добавление Playwright E2E (отдельная задача M19-T01)
- Добавление OpenAPI spec (отдельная задача M19-T02)
- Настройка Grafana dashboards (T1705 отложено до M18)

---

## Критерии приёмки

### Критические
- [ ] Callback.tsx не содержит hard-coded realm
- [ ] Token не хранится в localStorage (httpOnly cookie или другое решение)
- [ ] `runMigrate` / `runSeed` логируют правильную ошибку
- [ ] go.mod согласован с кодом verifier.go
- [ ] `atlas.hcl` существует или удалён из plan/report

### Средние
- [ ] `shared/pkg/auth/tenantresolver.go` создан (вынесен из middleware.go)
- [ ] `shared/pkg/auth/errors.go` создан (вынесен из middleware.go)
- [ ] `shared/pkg/auth/cache.go` создан или обоснованно удалён из плана
- [ ] `shared/pkg/auth/verifier_test.go` создан
- [ ] `ExtractRolesFromJWT` — private или с verif подписи
- [ ] `KEYCLOAK_PUBLIC_URL` добавлен в docker-compose.yml
- [ ] Дублирующийся код миграций убран

### Документация
- [ ] T1700–T1705 report.md обновлены (статус ✅ или ⚠️)
- [ ] `go build ./...` проходит
- [ ] `go test ./...` проходит (53+ PASS)
- [ ] CI pipeline зелёный (build.yml все job'ы PASS)

---

## Порядок реализации

1. **D3** — quick fix ошибки в cmd/server/main.go
2. **D4+D5** — reconcile go.mod + verifier.go + atlas.hcl
3. **D9** — fix `ExtractRolesFromJWT` security issue
4. **D1+D2** — frontend security fixes (hard-coded realm + localStorage token)
5. **D6–D8, D10–D12** — medium fixes
6. **Missing files** — tenantresolver.go, errors.go, cache.go, verifier_test.go
7. **Documentation** — обновить T1700–T1705 report.md
8. **CI** — verify build.yml green
