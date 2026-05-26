# T2302 — Migrations: CEL

## Веха

M23-cel-engine

## Тип

code

## Что сделать

```sql
CREATE TABLE lkfl_platform.cel_rules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    domain     VARCHAR(30) NOT NULL CHECK (domain IN ('billing', 'eligibility', 'flow', 'game', 'recommendations')),
    name       VARCHAR(255) NOT NULL,
    expression TEXT NOT NULL,                              -- CEL expression
    status     VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'disabled')),
    version    INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cel_rules_domain ON lkfl_platform.cel_rules (domain);
CREATE INDEX idx_cel_rules_status ON lkfl_platform.cel_rules (status);
```

## Критерии приёмки

- [ ] `cel_rules` таблица
- [ ] CHECK на domain (5 доменов)
- [ ] CHECK на status
- [ ] Indexes
- [ ] Migration apply + rollback OK
