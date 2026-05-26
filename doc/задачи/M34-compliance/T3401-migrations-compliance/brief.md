# T3401-T3406 — Compliance

## Веха

M34-compliance

## T3401 — Migrations
```sql
CREATE TABLE lkfl_platform.audit_logs (
    id UUID PK, tenant_id UUID FK, user_id UUID FK,
    action VARCHAR(100), entity_type VARCHAR(50), entity_id UUID,
    old_values JSONB, new_values JSONB, ip_address INET,
    created_at TIMESTAMPTZ
);
CREATE INDEX idx_audit_logs_tenant ON lkfl_platform.audit_logs (tenant_id, created_at);
CREATE INDEX idx_audit_logs_user ON lkfl_platform.audit_logs (user_id, created_at);
CREATE TABLE lkfl_platform.data_retention_policies (
    id UUID PK, tenant_id UUID FK, entity_type VARCHAR(50),
    retention_days INT, anonymize BOOLEAN DEFAULT true,
    created_at
);
CREATE TABLE lkfl_platform.compliance_events (
    id UUID PK, tenant_id UUID FK, user_id UUID FK,
    event_type VARCHAR(50), status VARCHAR(20),
    details JSONB, created_at
);
```

## T3402 — internal/compliance/ (Engine)
- Cascade revoke engine
- Audit trail logger (every write → audit entry)
- Data retention scheduler (Asynq)

## T3403 — Cascade revoke flow
Увольнение → блокировка ЛК → открепление ДМС → остановка billing rules → запрос баланса партнёрам

## T3404 — Consent revoke cascade
Consent revoke → деактивация льгот → удаление ПДн → уведомление провайдеров

## T3405 — Data retention
Asynq scheduled job, anonymization по политике

## T3406 — Admin API
```
GET /admin/compliance/audit-logs    — аудит-логи (фильтры)
GET /admin/compliance/retention     — политики хранения
```

## Критерии приёмки
- [ ] Все 6 задач
- [ ] Cascade revoke детерминирован
- [ ] Audit trail полный
- [ ] Data retention работает
