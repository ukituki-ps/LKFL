# T1804 — Отчёт о выполнении

## Статус

✅ выполнено

## Что сделано

### `backend/internal/tenant/isolation.go`

- **WithTenantID(ctx, query)** — добавляет tenant_id WHERE clause к SQL query:
  - Если tenant_id = uuid.Nil — не добавляет (admin queries)
  - Если tenant_id отсутствует в context — не добавляет
  - Если query содержит WHERE — добавляет `AND tenant_id = $1`
  - Если WHERE отсутствует — добавляет `WHERE tenant_id = $1`
  - Case-insensitive поиск WHERE
  - Trim whitespace от query

- **TenantContext(ctx, tid)** — создаёт context с tenant ID (для тестов)
- **WithAdminTenant(ctx)** — создаёт context без tenant фильтрации (uuid.Nil)

### Unit тесты (`isolation_test.go`)

- WithTenantID: adds WHERE clause, adds AND clause, nil tenant_id (no filter), no tenant in context, admin context, case-insensitive WHERE, trims whitespace, multiple WHERE in query
- TenantContext helper
- WithAdminTenant helper

## Критерии приёмки

- [x] `isolation.go` — WithTenantID() функция
- [x] Работает с существующим WHERE
- [x] uuid.Nil → без фильтра
- [x] Usage pattern документация (package comment)
- [x] Unit tests: с WHERE, без WHERE, nil tenant_id

## Время

~15 мин
