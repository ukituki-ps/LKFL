# T3801-T3815 — F3 Hardening

## Веха

M38-f3-hardening

## T3801 — Рефакторинг
- Audit proxy ↔ mono gRPC integration
- Notification fan-out optimization
- Compliance cascade ordering

## T3802 — Unit тесты: Edge cases
- Survey: branching loop detection, empty response, concurrent submit
- Gamification: achievement double grant, loyalty level boundary
- Notification: template injection, mass notify partial failure
- Compliance: cascade order violation, audit log loss
- Proxy: provider timeout, circuit breaker state machine, webhook replay

## T3803 — Integration тесты
- testcontainers + mock providers
- Cascade revoke полный путь
- Survey → activity → credit
- Mass notification → CEL segment

## T3804 — E2E тесты
- Survey flow (branching, submit, credit)
- Gamification (achievement earn, loyalty upgrade)
- Cascade revoke (admin dismiss → employee blocked → benefits deactivated)
- Collections (create → activate all → batch debit)
- Admin full cycle

## T3805 — Нагрузочное тестирование
- Mass рассылка 50000 recipients (Asynq)
- Survey submit 300 RPS
- Proxy gRPC 500 RPS per provider
- Cascade revoke 100 concurrent users

## T3806 — Мониторинг
- Provider Health dashboard
- User Activity dashboard
- Security dashboard

## T3807 — Alerting
- Circuit breaker open > 5min
- Provider error rate > 10%
- Cascade revoke failure
- Dead letter queue > 100
- Notification delivery failure > 5%

## T3808 — CI pipeline
- Proxy build stage
- gRPC contract validation (proto lint)
- Provider adapter smoke tests

## T3809 — Docker production profile
- Separate compose для staging/production
- Resource limits
- Restart policies

## T3810 — Деплой на стенд
- 2 бинарника (server + proxy)
- Provider configs (YAML)
- Webhook endpoint exposure
- Staging Keycloak realms

## T3811 — Security audit
- Proxy credential isolation (ADR-035)
- Webhook signature verification
- Notification template injection
- Survey response ПДн compliance (152-ФЗ)
- Compliance data retention enforcement

## T3812 — Data validation
- Survey response integrity
- Gamification achievement uniqueness
- Notification delivery tracking
- Audit log completeness

## T3813 — Chaos testing
- Provider down simulation
- Redis down (cache miss → DB fallback)
- PG connection pool exhaustion

## T3814 — Disaster recovery
- DB backup/restore procedure
- Redis data loss tolerance test
- Proxy credential rotation

## T3815 — Пакет F3
- git tag `f3-complete`
- Changelog, API docs (all 118 endpoints)
- Runbook для provider operations
- Incident response playbook

## Критерии приёмки
- [ ] Все 15 задач
- [ ] 55+ критериев приёмки пройдено
- [ ] Chaos testing пройден
- [ ] DR procedure задокументирована
