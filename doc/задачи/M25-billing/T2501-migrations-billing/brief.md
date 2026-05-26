# T2501 — Migrations: Billing

## Веха

M25-billing

## Тип

code

## Что сделать

```sql
-- Billing rules (начисления/списания)
CREATE TABLE lkfl_platform.billing_rules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    direction  VARCHAR(10) NOT NULL CHECK (direction IN ('credit', 'debit')),
    amount_cents BIGINT NOT NULL,
    category   VARCHAR(50),
    frequency  VARCHAR(20) NOT NULL DEFAULT 'one-time' CHECK (frequency IN ('one-time', 'monthly', 'yearly')),
    trigger    VARCHAR(50),                              -- activation, period_start, manual
    cel_rule_id UUID REFERENCES lkfl_platform.cel_rules(id),
    status     VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Billing periods
CREATE TABLE lkfl_platform.billing_periods (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'open', 'closed', 'burned')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Transactions
CREATE TABLE lkfl_platform.transactions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES lkfl_platform.tenants(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES lkfl_platform.accounts(id) ON DELETE CASCADE,
    direction  VARCHAR(10) NOT NULL CHECK (direction IN ('credit', 'debit')),
    amount_cents BIGINT NOT NULL,
    category   VARCHAR(50),
    status     VARCHAR(20) NOT NULL DEFAULT 'confirmed' CHECK (status IN ('frozen', 'confirmed', 'cancelled', 'burned')),
    reference_type VARCHAR(30),                          -- engagement, period, burn, payment, achievement
    reference_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user ON lkfl_platform.transactions (user_id);
CREATE INDEX idx_transactions_period ON lkfl_platform.transactions (created_at);
CREATE INDEX idx_transactions_status ON lkfl_platform.transactions (status);

-- User account balances by category
CREATE TABLE lkfl_platform.user_balances (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL UNIQUE REFERENCES lkfl_platform.users(id) ON DELETE CASCADE,
    category   VARCHAR(50) NOT NULL DEFAULT 'general',
    balance_cents BIGINT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_user_balances_user_category ON lkfl_platform.user_balances (user_id, category);
```

## Критерии приёмки

- [ ] `billing_rules` таблица
- [ ] `billing_periods` таблица
- [ ] `transactions` таблица
- [ ] `user_balances` таблица
- [ ] CHECK constraints
- [ ] Indexes
- [ ] Migration apply + rollback OK
