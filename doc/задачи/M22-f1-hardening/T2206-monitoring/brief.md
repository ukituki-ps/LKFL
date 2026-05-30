# T2206 — Мониторинг (Prometheus + Grafana)

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Настройка мониторинга с Prometheus и Grafana для production observability.
Сбор метрик, дашборды, алертинг.

## Что сделать

### Prometheus metrics

**HTTP middleware metrics:**
- `http_requests_total` — counter (method, path, status, tenant_id)
- `http_request_duration_seconds` — histogram (method, path, tenant_id)

**Custom metrics:**
- `catalog_query_total` — counter (type: list|filter|search, status)
- `catalog_query_duration_seconds` — histogram
- `tenant_resolve_total` — counter (method: host|path|header, status)
- `tenant_resolve_duration_seconds` — histogram
- `auth_login_total` — counter (status: success|failure)
- `auth_callback_total` — counter (status: success|failure|error)
- `redis_cache_hits_total` — counter
- `redis_cache_misses_total` — counter
- `redis_cache_evictions_total` — counter

### Grafana dashboard

- **Platform Overview F1** — JSON dashboard
- Panels: RPS, latency (P50/P95/P99), error rate, tenant distribution, cache hit ratio
- Variables: tenant_id, path, method
- Time range presets: 1h, 6h, 24h, 7d

### Alerting rules

- Error rate > 1% (5min window) → PAGE
- P95 latency > 500ms (5min window) → WARNING
- P99 latency > 1000ms (5min window) → PAGE
- Redis connection errors > 0 (1min window) → WARNING
- DB connection pool exhausted → PAGE

### Конфигурация

- `prometheus.yml` — scrape config
- `grafana/dashboards/platform-f1.json` — dashboard
- `grafana/alerting/rules.yml` — alerting rules
- docker-compose: +prometheus, +grafana services

## Критерии приёмки

- [ ] Prometheus metrics endpoint (`/metrics`)
- [ ] HTTP middleware metrics (requests_total, duration)
- [ ] Custom metrics (catalog, tenant, auth, redis)
- [ ] Grafana dashboard JSON (Platform Overview F1)
- [ ] Alerting rules настроены
- [ ] docker-compose: prometheus + grafana
- [ ] Метрики видны в Grafana
