# T2603-T2607 — Flow (оставшиеся задачи)

## Веха

M26-engagement-flow

## T2603 — Flow + Billing integration
- Flow completion → billing.CreateTransaction(debit)
- Rollback on billing failure
- Unit tests: debit success, debit failure → rollback

## T2604 — Flow + Eligibility integration
- StartFlow → eligibility.Check()
- Not eligible → error with reason
- Unit tests

## T2605 — API: Activate
```
POST /api/v1/user-engagements        — начать flow
POST /api/v1/user-engagements/:id/steps/:stepKey/complete — завершить шаг
GET  /api/v1/user-engagements        — мои активные/в процессе
```

## T2606 — API: Revoke
```
POST /api/v1/user-engagements/:id/revoke — отмена → возврат баллов
```
- Revoke → billing.CreateTransaction(credit, amount=remaining)
- Status → revoked
- Unit tests

## T2607 — JSON-driven wizard config
- Flow.steps JSON schema
- Step types: info, confirm, form, external_redirect
- Validation schema per step type
- Admin API для управления flow config

## Критерии приёмки

- [ ] Все 5 задач реализованы
- [ ] Flow completion → debit
- [ ] Revoke → credit
- [ ] API endpoints
- [ ] JSON wizard config
