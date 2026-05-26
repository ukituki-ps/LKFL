# T3301-T3306 — Уведомления

## Веха

M33-notification

## T3301 — Migrations
```sql
CREATE TABLE lkfl_platform.notification_templates (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255), channel VARCHAR(20) CHECK (channel IN ('email','push','in-app')),
    subject_template TEXT, body_template TEXT, settings JSONB, created_at
);
CREATE TABLE lkfl_platform.notifications (
    id UUID PK, tenant_id UUID FK, user_id UUID FK, template_id UUID FK,
    channel VARCHAR(20), status VARCHAR(20) CHECK (status IN ('pending','sent','failed','read')),
    payload JSONB, sent_at TIMESTAMPTZ, read_at TIMESTAMPTZ, created_at
);
CREATE INDEX idx_notifications_user ON lkfl_platform.notifications (user_id, status);
CREATE TABLE lkfl_platform.notification_preferences (
    id UUID PK, user_id UUID FK UNIQUE, channels JSONB DEFAULT '["in-app"]', created_at
);
CREATE TABLE lkfl_platform.mass_notifications (
    id UUID PK, tenant_id UUID FK, template_id UUID FK,
    cel_segment TEXT, status VARCHAR(20), total_recipients INT, sent_count INT,
    created_at
);
```

## T3302 — internal/notification/ (Engine)
- Email (SMTP + template rendering)
- In-app notifications (DB stored)
- Push (stub for F3)

## T3303 — Notification triggers
- Engagement activated, balance warning, achievement granted, period opened, activity available
- Event-driven (from other packages via interface)

## T3304 — Mass notification
- Admin → CEL segment → audience → email/in-app
- Asynq job для отправки

## T3305 — API
```
GET  /api/v1/notifications           — in-app уведомления
POST /api/v1/notifications/:id/read  — прочитано
GET  /api/v1/notification/preferences
PUT  /api/v1/notification/preferences
```

## T3306 — Admin API
```
CRUD templates
POST /admin/notifications/mass       — массовая рассылка
GET  /admin/notifications/mass/:id   — статус рассылки
```

## Критерии приёмки
- [ ] Все 6 задач
- [ ] Email + in-app работают
- [ ] Mass notification через CEL сегмент
- [ ] Triggers от других пакетов
