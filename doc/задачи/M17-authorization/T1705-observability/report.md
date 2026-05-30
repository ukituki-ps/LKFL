# T1705 — Отчёт: Observability

## Статус

⚠️ Частично реализовано (без dashboards)

## Что сделано

### Docker Compose
- `docker-compose.yml` — добавлены Prometheus, Loki, Grafana сервисы
- Health checks для всех трёх сервисов
- Volumes для persistence данных

### Prometheus
- `infra/prometheus/prometheus.yml` — scrape configs для lkfl-server, Grafana, Prometheus self-monitoring
- Retention: 15d

### Grafana
- `infra/grafana/provisioning/datasources/prometheus.yml` — auto-provisioned Prometheus datasource
- `infra/grafana/provisioning/datasources/loki.yml` — auto-provisioned Loki datasource
- `infra/grafana/provisioning/dashboards.yaml` — dashboard provisioning config
- `infra/grafana/dashboards/` — пустая директория (dashboards отложены до M18)

### Loki
- `infra/loki/loki.yml` — local config

### Backend metrics
- `internal/metrics/metrics.go` — Prometheus metrics для auth flow:
  - `auth_login_total` — login attempts
  - `auth_callback_total` — callback success/failure
  - `http_request_duration_seconds` — request latency
  - `http_requests_total` — request counter

## Что НЕ сделано

- Grafana dashboards (отложено до M18)
- Docker logs driver → Loki integration
- Alert rules

## Проблемы

Нет.

## Следующие шаги

1. M18: Grafana dashboards для auth flow
2. Loki: Docker logs driver configuration
3. Alert rules для Prometheus
