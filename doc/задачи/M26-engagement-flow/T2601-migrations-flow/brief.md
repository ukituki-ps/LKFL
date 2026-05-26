# T2601 — Migrations: Flow

## Веха

M26-engagement-flow

## Тип

code

## Что сделать

```sql
-- Engagement flows (step-by-step activation)
CREATE TABLE lkfl_platform.engagement_flows (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    engagement_type_id  UUID NOT NULL REFERENCES lkfl_platform.engagement_types(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    steps               JSONB NOT NULL,                   -- step definitions
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User engagements (activation instances)
CREATE TABLE lkfl_platform.user_engagements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    engagement_type_id  UUID NOT NULL REFERENCES lkfl_platform.engagement_types(id) ON DELETE CASCADE,
    flow_id             UUID REFERENCES lkfl_platform.engagement_flows(id),
    offer_id            UUID REFERENCES lkfl_platform.engagement_offers(id),
    status              VARCHAR(20) NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress', 'active', 'revoked', 'expired')),
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    revoked_at          TIMESTAMPTZ,
    metadata            JSONB DEFAULT '{}'
);

CREATE UNIQUE INDEX idx_user_engagement_type ON lkfl_platform.user_engagements (user_id, engagement_type_id, status) WHERE status IN ('in_progress', 'active');
CREATE INDEX idx_user_engagements_user ON lkfl_platform.user_engagements (user_id);

-- User engagement steps
CREATE TABLE lkfl_platform.user_engagement_steps (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_engagement_id  UUID NOT NULL REFERENCES lkfl_platform.user_engagements(id) ON DELETE CASCADE,
    step_key            VARCHAR(100) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'skipped')),
    data                JSONB DEFAULT '{}',
    completed_at        TIMESTAMPTZ
);

CREATE INDEX idx_ues_user_engagement ON lkfl_platform.user_engagement_steps (user_engagement_id);
```

## Критерии приёмки

- [ ] `engagement_flows` таблица
- [ ] `user_engagements` таблица
- [ ] `user_engagement_steps` таблица
- [ ] UNIQUE partial index (one active per user+type)
- [ ] CHECK constraints
- [ ] Migration apply + rollback OK
