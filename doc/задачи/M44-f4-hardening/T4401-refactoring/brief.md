# T4401-T4416 — F4 Hardening — Production Ready

## Веха

M44-f4-hardening

## T4401 — Рефакторинг финальный
- Полный audit всех 22 пакетов
- Remove dead code, unused imports
- Godoc coverage
- Interface satisfaction check

## T4402 — Unit тесты: финальные edge cases
- LLM: rate limit, invalid response, cost overflow
- Payments: PCI DSS (raw card data never stored, tokenization leak)
- Recommendations: empty segment, conflicting rules, priority tie
- Mobile: touch event vs click conflict

## T4403 — Integration тесты: полный стек
- testcontainers (PG + Redis + mock Keycloak + mock LLM + mock payment gateway)
- Все критические пути
- 66/66 критериев приёмки

## T4404 — E2E тесты: полный набор
- Playwright: все 57 journeys (критические J01-J21)
- Payment flow (card add → freeze → confirm → debit)
- Payroll flow (request → consent → HR approve → deduction)
- LLM CEL generation (русский → CEL → validate)
- Mobile viewport

## T4405 — Нагрузочное тестирование: production target
- 10000 concurrent users
- 1000 RPS activation
- 500 RPS proxy gRPC
- 100 TPS payments
- 50 RPS LLM
- P95 < 300ms для P0 endpoints

## T4406 — Мониторинг финальный
- Все 6 dashboards
- All metrics из стек.md active
- Grafana variables

## T4407 — Alerting финальный
- Alert rules для всех метрик
- PagerDuty/OpsGenie integration
- Escalation policy

## T4408 — CI/CD pipeline
- lint → unit → integration → E2E → build → docker push → deploy staging
- Coverage gate > 80% overall, > 90% billing+payments
- Security scan (Trivy)

## T4409 — Docker production hardening
- Multi-stage (Go build → distroless)
- Non-root user
- Read-only filesystem
- Healthcheck
- Image signing (cosign)

## T4410 — Деплой на staging
- Full production-like environment
- 2 бинарника + 6 infra containers
- Nginx (TLS, rate limiting, caching)
- Keycloak realms (tenant isolation test)

## T4411 — Security audit финальный
- OWASP Top 10 полный
- PCI DSS check (payments)
- 152-ФЗ (ПДн storage, encryption at rest)
- ФСТЭК baseline
- Dependency audit (govulncheck + npm audit + Trivy)
- Penetration test (критические endpoints)

## T4412 — Performance audit
- pprof profiling (CPU + memory)
- DB query analysis (EXPLAIN ANALYZE для всех queries)
- Redis memory usage
- gRPC connection pooling optimization

## T4413 — Disaster recovery — тестирование
- DB restore from backup (RTO < 1h)
- Redis loss (rebuild from DB)
- Provider credential rotation (zero downtime)
- Proxy failover (mono degrades gracefully)

## T4414 — Runbooks
- Incident response playbook
- Billing operations guide
- Provider onboarding guide
- Tenant onboarding guide
- Scaling guide (read replicas, Redis cluster)

## T4415 — Production checklist
- Env vars inventory
- Secrets management (Vault setup)
- Domain/TLS setup
- Monitoring access
- Backup schedule
- Log retention policy

## T4416 — Пакет F4 — Production Release
- git tag `v1.0.0`
- Changelog (полный)
- API docs (Redoc)
- ADR update
- Known limitations doc
- Migration guide F3→F4

## Критерии приёмки
- [ ] Все 16 задач
- [ ] 66/66 критериев приёмки
- [ ] Load test 10000 concurrent users
- [ ] Security audit (OWASP + PCI DSS + 152-ФЗ + ФСТЭК)
- [ ] CI/CD pipeline зелёный
- [ ] DR tested
- [ ] Runbooks написаны
- [ ] Production release package готов
