# T2207 — Logging (Loki)

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Настройка структурированного логирования с отправкой в Loki.
Интеграция с Grafana Explore для поиска и анализа логов.

## Что сделать

### Structured JSON logs

Формат лога:
```json
{
  "ts": "2025-01-01T12:00:00Z",
  "level": "info",
  "svc": "lkfl-server",
  "tenant_id": "sdek",
  "user_id": "uuid",
  "msg": "catalog query executed",
  "duration_ms": 42,
  "trace_id": "uuid"
}
```

### Log levels по пакету

- `internal/tenant/` — info
- `internal/auth/` — info (security events — warn)
- `internal/user/` — info
- `internal/engagement/catalog/` — info (queries — debug in dev)
- `internal/api/` — info (HTTP requests — debug in dev)
- `cmd/` — info

### Loki integration

- **loki-promtail** — log collector в docker-compose
- Конфиг: `promtail.yml` — сбор логов из Docker containers
- Loki labels: `job=lkfl-server`, `tenant_id`, `level`
- Retention: 30 days

### Grafana Explore

- Query examples для common patterns
- Dashboard panel: recent errors, slow queries, auth failures
- Log context: ±5 lines around matched entry

### Конфигурация

- `promtail.yml` — promtail config
- `docker-compose.yml` — +loki, +promtail services
- Logger injection во все пакеты (slog или zap)
- Log level через env: `LOG_LEVEL=debug`

## Критерии приёмки

- [ ] Structured JSON logs формат
- [ ] Logger injection во все пакеты
- [ ] Log levels по пакету настроены
- [ ] Loki config в docker-compose
- [ ] Promtail config
- [ ] Grafana Explore query test
- [ ] Логи содержат tenant_id, user_id, trace_id
