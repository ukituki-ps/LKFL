# T2305 — Admin API: CEL Rules

## Веха

M23-cel-engine

## Тип

code

## Что сделать

```
POST /admin/cel/rules          — создать правило
GET  /admin/cel/rules          — список правил
GET  /admin/cel/rules/:id      — детали
PUT  /admin/cel/rules/:id      — обновить
DELETE /admin/cel/rules/:id    — удалить
POST /admin/cel/rules/:id/evaluate — тестовая оценка (admin debug)
```

### Evaluate endpoint

```json
{
  "expression": "balance.total >= 3500",
  "user_id": "uuid",
  "result": true
}
```

## Критерии приёмки

- [ ] CRUD CEL rules
- [ ] Evaluate endpoint (debug)
- [ ] Expression validation (compile check)
- [ ] RBAC: admin only
