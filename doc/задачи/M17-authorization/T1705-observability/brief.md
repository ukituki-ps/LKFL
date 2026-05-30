# T1705 — Observability

## Контекст

Настраиваем Observability в docker-compose: Prometheus + Grafana + Loki с автопроvisioned дашбордами.

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1701 (инфраструктура)
**ADR:** ADR-030 (CI/CD Pipeline), стек.md

**Можно отложить** — не влияет на работу auth. Отложить до M18.

## Что включено

### Observability (docker-compose)
- `docker-compose.yml` — +prometheus, +grafana, +loki
- `infra/prometheus/prometheus.yml` — scrape configs (lkfl-server, proxy, postgres, redis, keycloak)
- `infra/grafana/provisioning/` — auto-provisioned dashboards
- `infra/grafana/dashboards/` — JSON дашборды: Platform Overview, Backend Metrics, Infrastructure, Application Logs

### Loki (логи)
- Loki в docker-compose для логов
- Docker logs driver → Loki
- Grafana Explore: filter по tenant_id, level, svc

## Результат

- `make docker-up` включает Prometheus + Grafana + Loki
- Grafana показывает дашборды Platform Overview, Backend Metrics, Infrastructure, Application Logs
- Логи собираются в Loki, доступны через Grafana Explore
