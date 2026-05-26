# T3001-T3005 — Activity (опросы + события)

## Веха

M30-activity

## T3001 — Migrations: Activity
```sql
CREATE TABLE lkfl_platform.activity_templates (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255), description TEXT,
    type VARCHAR(20) CHECK (type IN ('survey', 'event', 'custom')),
    reward_cents BIGINT, status VARCHAR(20) CHECK (status IN ('draft','active','completed')),
    created_at, updated_at
);
CREATE TABLE lkfl_platform.activity_instances (
    id UUID PK, tenant_id UUID FK, template_id UUID FK, start_date, end_date,
    max_participants INT, status VARCHAR(20), created_at
);
CREATE TABLE lkfl_platform.activity_registrations (
    id UUID PK, user_id UUID FK, instance_id UUID FK,
    status VARCHAR(20) CHECK (status IN ('registered','completed','cancelled')),
    registered_at, completed_at
);
```

## T3002 — internal/engagement/ (Activity Engine)
- Event registration + completion
- Survey completion (via survey engine, M31)
- Custom activity completion
- Unit tests

## T3003 — Activity + Billing integration
- Completion → CEL check → credit на баланс
- billing.CreateTransaction(credit, activity.reward_cents)
- Unit tests

## T3004 — API: Activity
```
GET  /api/v1/activities              — доступные активности
POST /api/v1/activities/:id/register — регистрация
POST /api/v1/activities/:id/complete — завершение → credit
GET  /api/v1/activity-registrations  — мои регистрации
```

## T3005 — Admin API: Activity
```
POST /admin/activities/templates     — создать шаблон
GET  /admin/activities/templates     — список
POST /admin/activities/instances     — создать инстанс
PUT  /admin/activities/instances/:id — обновить
```

## Критерии приёмки
- [ ] Все 5 задач
- [ ] Activity → credit flow
- [ ] API + Admin API
