# T2901-T2913 — F2 Hardening

## Веха

M29-f2-hardening

## T2901 — Рефакторинг F2
- Audit CEL+billing+flow интеграции
- Race conditions (concurrent debit)
- Transaction isolation review
- golangci-lint + ESLint strict

## T2902 — Unit тесты: Edge cases
- CEL: invalid expression, sandbox escape, type mismatch
- Billing: double debit, negative balance, concurrent period distribute
- Flow: step skip attempt, parallel activation same engagement
- Consent: revoke during active engagement

## T2903 — Integration тесты
- testcontainers: полный цикл catalog → eligibility → flow → debit → balance → revoke → credit
- Multi-tenant isolation для billing
- Asynq job integration (distribute, burn)

## T2904 — E2E тесты
- Playwright: activation flow (success/eligibility fail/balance insufficient)
- Balance page real-time update
- Consent sign/revoke
- Admin period distribute

## T2905 — Нагрузочное тестирование
- k6: balance query 1000 RPS (period start peak)
- Activation 200 RPS
- Period distribute 10000 пользователей (Asynq worker)
- P95 < 300ms для P0 endpoints

## T2906 — Мониторинг
- Grafana dashboards: Billing Operations, CEL evaluation
- All metrics from стек.md active

## T2907 — Alerting
- Billing transaction failure rate > 1%
- CEL evaluation P95 > 500ms
- Asynq job queue depth > 1000

## T2908 — CI pipeline
- Integration tests stage
- Coverage gate > 70% для billing+cel
- Docker build + push

## T2909 — DB миграции: rollback
- Atlas rollback strategy
- Test rollback каждой миграции F2

## T2910 — Деплой на стенд
- Blue-green deploy script
- DB migration run на staging

## T2911 — Security audit
- Billing: amount manipulation attempt
- Flow: step bypass
- Consent: unauthorized revoke
- RBAC: role escalation

## T2912 — Data validation
- Billing invariant: sum transactions = balance
- Period consistency
- Flow state machine validation

## T2913 — Пакет F2
- git tag `f2-complete`
- Changelog, API docs, runbook для billing operations

## Критерии приёмки

- [ ] Все 13 задач реализованы
- [ ] Financial accuracy guaranteed
- [ ] Load test passed
- [ ] Security audit passed
- [ ] Release package ready
