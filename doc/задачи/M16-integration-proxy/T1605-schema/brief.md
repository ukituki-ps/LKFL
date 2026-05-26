# T1605 — `архитектура/schema.md` — lkfl_integration

## Веха

M16-integration-proxy

## Контекст

Proxy имеет собственную schema `lkfl_integration` в PostgreSQL. Нужно описать 6 таблиц.

## Что сделать

Добавить §3.10 Integration Proxy Schema в `schema.md`:

| Таблица | Поля |
|---------|------|
| `providers` | id, tenant_id, name, category, protocol, endpoints (JSONB), auth_method, credentials_encrypted, status, health_last_check_at, health_status, error_rate_30d, latency_p95_ms |
| `activation_jobs` | id, tenant_id, user_id, offer_id, provider_name, status, result (JSONB), error, created_at, completed_at |
| `provider_sync_log` | id, provider_name, timestamp, count, errors, duration_ms |
| `webhook_events` | id, provider_name, payload (JSONB), signature, processed_at, status, mono_callback_status |
| `dead_letters` | id, provider_name, operation, request (JSONB), error, retry_count, next_retry_at, created_at |
| `circuit_breaker_state` | provider_name, state, last_failure_at, failure_count, opened_at |

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновить | `архитектура/schema.md` — +§3.10 |

### Критерии приёмки

- [ ] 6 таблиц описаны с DDL
- [ ] Индексы определены
- [ ] Constraints (CHECK) определены
- [ ] Go owner указан (lkfl-integration-proxy)
- [ ] Go package → Table mapping обновлён
