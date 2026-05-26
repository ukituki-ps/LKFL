# T2401 — Migrations: Eligibility

## Веха

M24-eligibility

## Тип

code

## Что сделать

```sql
CREATE TABLE lkfl_platform.eligibility_conditions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    engagement_type_id  UUID NOT NULL REFERENCES lkfl_platform.engagement_types(id) ON DELETE CASCADE,
    cel_rule_id         UUID NOT NULL REFERENCES lkfl_platform.cel_rules(id) ON DELETE CASCADE,
    priority            INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_eligibility_engagement ON lkfl_platform.eligibility_conditions (engagement_type_id);
```

## Критерии приёмки

- [ ] `eligibility_conditions` таблица
- [ ] FK → engagement_types, cel_rules
- [ ] Index на engagement_type_id
- [ ] Migration apply + rollback OK
