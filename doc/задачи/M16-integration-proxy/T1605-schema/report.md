# T1605 — Отчёт: `архитектура/schema.md` — lkfl_integration

## Веха

M16-integration-proxy

## Выполнено

### Обновлён `doc/архитектура/schema.md`

1. **Помечена существующая таблица `providers` в §3.9** — добавлена пометка «переезжает в `lkfl_integration` в M16» для backward compatibility.

2. **Добавлен §3.10 Integration Proxy Schema** — 6 таблиц в схеме `lkfl_integration`:

   | Таблица | DDL | Индексы | Constraints |
   |---------|-----|---------|-------------|
   | `providers` | ✅ | idx_ip_prov_tenant, idx_ip_prov_endpoints_gin (GIN) | CHECK status, CHECK protocol |
   | `activation_jobs` | ✅ | idx_ip_aj_tenant, idx_ip_aj_status (partial) | CHECK status |
   | `provider_sync_log` | ✅ | idx_ip_psl_provider, idx_ip_psl_timestamp | — |
   | `webhook_events` | ✅ | idx_ip_we_provider, idx_ip_we_status (partial) | CHECK status |
   | `dead_letters` | ✅ | idx_ip_dl_provider, idx_ip_dl_next_retry (partial) | — |
   | `circuit_breaker_state` | ✅ | PK по provider_name | CHECK state |

3. **Go package → Table mapping** — добавлена строка `lkfl-integration-proxy` → 6 таблиц. Обновлена строка `internal/integrations/` — убрана `providers` (переехала в proxy).

4. **ER-диаграмма (Mermaid)** — добавлены связи: TENANTS → INTEGRATION_PROVIDERS, TENANTS → ACTIVATION_JOBS, USERS → ACTIVATION_JOBS.

5. **Ренумерация** — §3.10 Approvals → §3.11. TOC обновлён.

6. **Итоговая строка** — обновлена: 47 таблиц (41 `lkfl_platform` + 6 `lkfl_integration`).

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Обновлён | `doc/архитектура/schema.md` |
| Обновлён | `doc/задачи/M16-integration-proxy/T1605-schema/plan.yaml` |
| Создан | `doc/задачи/M16-integration-proxy/T1605-schema/report.md` (этот файл) |

## Время

~20 мин

## Замечания

- FK в `activation_jobs` и `providers` (lkfl_integration) ссылаются на `lkfl_platform.tenants` и `lkfl_platform.users` — cross-schema references. В production PostgreSQL это работает через search_path или явные schema-квалификаторы.
- Индексы с префиксом `idx_ip_` для избежания коллизий с индексами в `lkfl_platform`.
