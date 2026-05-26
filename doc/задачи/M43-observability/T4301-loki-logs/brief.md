# T4301-T4305 — Observability финализация

## Веха

M43-observability

## T4301 — Loki logs
- Structured JSON logs (svc, tenant_id, user_id, msg)
- Loki config в docker-compose
- Grafana Explore query test
- Log retention policy

## T4302 — Prometheus финальный
- Все metrics из стек.md active
- Custom counters, histograms, gauges
- Scrape config

## T4303 — Alerting финальный
- Alert rules для всех метрик
- PagerDuty/OpsGenie integration (stub)
- Escalation policy

## T4304 — CI/CD финальный
- GitHub Actions: lint → unit → integration → E2E → build → docker push → deploy staging
- Coverage gate > 80% overall, > 90% для billing+payments
- Security scan (Trivy)

## T4305 — Нагрузочное тестирование финальное
- 10000 concurrent users (catalog + balance)
- 1000 RPS activation
- 500 RPS proxy gRPC
- Payment processing 100 TPS
- LLM generation 50 RPS
- P95 < 300ms для всех P0 endpoints

## Критерии приёмки
- [ ] Все 5 задач
- [ ] Observability полный стек
- [ ] Load test production target
