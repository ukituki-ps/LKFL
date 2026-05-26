# T2204-T2212 — Hardening F1 (оставшиеся задачи)

## Веха

M22-f1-hardening

## Тип

code

## Краткое описание

### T2204 — E2E тесты (Playwright)
- Login flow: success, failed credentials, password reset
- Catalog: load, filter by category, filter by type, search, pagination
- Multi-tenant: switch tenant, verify isolation
- Admin: create category, create type, update status, delete protection
- Config: `playwright.config.ts`, browsers: Chromium, Firefox

### T2205 — Нагрузочное тестирование (k6)
- Catalog query: 500 RPS, P95 < 200ms
- Auth callback: 100 RPS, P95 < 500ms
- User profile: 200 RPS, P95 < 100ms
- Script: `loadtest/catalog.js`, `loadtest/auth.js`
- Report: HTML + JSON

### T2206 — Мониторинг (Prometheus + Grafana)
- Prometheus metrics: `http_requests_total`, `http_request_duration_seconds`
- Custom metrics: `catalog_query_total`, `tenant_resolve_total`
- Grafana dashboard JSON: Platform Overview F1
- Alerting rules: error rate > 1%, P95 > 500ms

### T2207 — Logging (Loki)
- Structured JSON logs: `{"ts", "level", "svc", "tenant_id", "user_id", "msg"}`
- Log levels по пакету
- Loki config в docker-compose
- Grafana Explore query test

### T2208 — CI Pipeline (GitHub Actions)
- Workflow: `.github/workflows/ci.yml`
- Stages: lint → unit test → integration test → build → docker push
- Coverage gate: > 60%
- Caching: go modules, npm dependencies, docker layers

### T2209 — Docker Production
- Multi-stage Dockerfile: Go build → distroless
- Non-root user
- Read-only filesystem
- Healthcheck endpoint
- Image signing (cosign)

### T2210 — Деплой на стенд
- Staging docker-compose
- Nginx config (TLS self-signed)
- Environment variables
- Healthcheck verification

### T2211 — Security Audit
- OWASP Top 10 check
- Rate limiting на auth endpoints
- CORS policy verification
- Dependency audit: govulncheck + npm audit
- SQL injection test
- XSS test

### T2212 — Пакет F1
- git tag `f1-complete`
- Changelog
- API documentation (Redocly)
- Release notes

## Критерии приёмки

- [ ] Все 9 задач реализованы
- [ ] E2E тесты зелёные (Chromium + Firefox)
- [ ] Load test: 500 RPS catalog, P95 < 200ms
- [ ] Prometheus metrics active
- [ ] Grafana dashboard
- [ ] CI pipeline зелёный
- [ ] Docker production image
- [ ] Staging deployed
- [ ] Security audit passed
- [ ] Release package готов
