# T2504 — Admin API: Billing

## Веха

M25-billing

## Тип

code

## Что сделать

```
POST /admin/billing/rules        — создать правило
GET  /admin/billing/rules        — список правил
PUT  /admin/billing/rules/:id    — обновить
DELETE /admin/billing/rules/:id  — удалить

POST /admin/periods              — создать период
GET  /admin/periods              — список периодов
PUT  /admin/periods/:id          — обновить
POST /admin/periods/:id/open     — открыть период (mass notify)
POST /admin/periods/:id/distribute — массовое начисление (Asynq job)
POST /admin/periods/:id/close    — закрыть период
```

## Критерии приёмки

- [ ] Billing rules CRUD
- [ ] Periods CRUD + open/distribute/close
- [ ] RBAC: hr, admin
- [ ] Unit tests
