# T2701 — Migrations: Consent

## Веха

M27-consent

## Тип

code

## Что сделать

```sql
CREATE TABLE lkfl_platform.consent_documents (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    content    TEXT NOT NULL,                              -- HTML текст согласия
    version    INT NOT NULL DEFAULT 1,
    status     VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('draft', 'active', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE lkfl_platform.consents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    document_id     UUID NOT NULL REFERENCES lkfl_platform.consent_documents(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'signed' CHECK (status IN ('signed', 'revoked')),
    signed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    signed_ip       INET,
    revoked_at      TIMESTAMPTZ,
    revoked_reason  TEXT
);

CREATE INDEX idx_consents_user ON lkfl_platform.consents (user_id);
CREATE UNIQUE INDEX idx_consents_user_doc ON lkfl_platform.consents (user_id, document_id) WHERE status = 'signed';
```

## Критерии приёмки

- [ ] `consent_documents` таблица
- [ ] `consents` таблица
- [ ] CHECK constraints
- [ ] Unique partial index (one active consent per user+document)
- [ ] Migration apply + rollback OK
