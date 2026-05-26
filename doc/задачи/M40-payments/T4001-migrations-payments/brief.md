# T4001-T4007 — Платежи + HR Sync

## Веха

M40-payments

## T4001 — Migrations: Payments
```sql
CREATE TABLE lkfl_platform.payment_methods (
    id UUID PK, user_id UUID FK UNIQUE,
    token VARCHAR(255) NOT NULL,  -- tokenized card (never raw data)
    last4 VARCHAR(4), brand VARCHAR(20), expiry_month INT, expiry_year INT,
    created_at, updated_at
);
CREATE TABLE lkfl_platform.payment_transactions (
    id UUID PK, user_id UUID FK, method_id UUID FK,
    amount_cents BIGINT, status VARCHAR(20) CHECK (status IN ('pending','frozen','confirmed','cancelled','refunded')),
    provider_transaction_id VARCHAR(255), created_at, updated_at
);
CREATE TABLE lkfl_platform.payroll_statements (
    id UUID PK, user_id UUID FK, period_id UUID FK,
    amount_cents BIGINT, status VARCHAR(20) CHECK (status IN ('pending','approved','rejected','processed')),
    approved_by UUID, approved_at TIMESTAMPTZ, created_at
);
CREATE TABLE lkfl_platform.approval_requests (
    id UUID PK, user_id UUID FK, type VARCHAR(50),
    status VARCHAR(20) CHECK (status IN ('pending','approved','rejected')),
    details JSONB, approved_by UUID, approved_at TIMESTAMPTZ, rejection_reason TEXT, created_at
);
```

## T4002 — internal/payments/ (Engine)
- Card payment (PCI DSS: tokenization, no raw card data)
- Payroll deduction

## T4003 — Card payment flow
- Форма карты → заморозка → подтверждение → debit
- Отмена → разморозка → возврат

## T4004 — Payroll flow
- Заявление → согласие → апрув HR → вычет → credit
- Отклонение → возврат баллов

## T4005 — API: Payments
```
POST /api/v1/payment-methods         — добавить карту
POST /api/v1/payments/card           — оплата картой
POST /api/v1/payments/payroll        — заявление на удержание
GET  /api/v1/payments/:id            — статус платежа
```

## T4006 — Admin API
```
GET  /admin/payments/approvals       — список заявок
POST /admin/payments/approvals/:id/approve
POST /admin/payments/approvals/:id/reject
GET  /admin/payments/statements      — payroll statements
```

## T4007 — HR Sync (user/HRSync)
- REST API pull от кадровой системы
- XLSX парсинг (excelize)
- Валидация реестра
- Email сотрудника при добавлении

## Критерии приёмки
- [ ] Все 7 задач
- [ ] PCI DSS compliance (no raw card data)
- [ ] Card → freeze → confirm → debit
- [ ] Payroll → HR approve → deduction → credit
- [ ] HR Sync → валидация → добавление → email
