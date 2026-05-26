# M16 — Integration Proxy

## Веха

M16-integration-proxy

## Контекст

ADR-024 (M12) принял modular monolith: один бинарник `lkfl-server`, 17 internal-пакетов, Go interfaces вместо NATS. Пакет `internal/integrations/` содержит ProviderGateway — gateway к внешним провайдерам льгот.

Прямые HTTP вызовы из монолита к внешним провайдерам создают критические риски:
- **Блокировка горутин** в hot path активации льготы
- **Отсутствие fault isolation** — panic в адаптере → crash всего бинарника
- **Credential blast radius** — ключи всех провайдеров в одном процессе
- **Webhook endpoint'ы** на основном бинарнике

## Решение

**Integration Proxy** — отдельный бинарник `lkfl-integration-proxy` в том же `go.mod`, communicating с монолитом через gRPC на localhost.

- Все внешние HTTP calls → через proxy
- Активация/деактивация → асинхронные (job_id + callback)
- Webhook handling → только на proxy
- Circuit breaker + worker pool per provider → на уровне процесса proxy
- Credential isolation → ключи только в proxy

## Файлы-мишени

| Файл | Изменение |
|------|-----------|
| `архитектура/adr/035-integration-proxy.md` | Новый ADR |
| `архитектура/интеграции.md` | Rewrite: proxy architecture, gRPC contract |
| `архитектура/пакеты-platform.md` | +integrationclient/, proxy structure |
| `архитектура/schema.md` | +lkfl_integration schema (6 таблиц) |
| `спецификация/api.md` | Admin endpoints → proxy delegation |
| `архитектура/модули.md` | +lkfl-integration-proxy binary |
| `архитектура/стек.md` | +gRPC, 2 binaries |
| `архитектура/безопасность.md` | Credential isolation |
| `контекст/акторы.md` | Админ интеграций → proxy |
| `контекст/negative-criteria.md` | Прямые вызовы → только через proxy |
| `NAVIGATION.md` | +Integration Proxy навигация |

## Критерии приёмки

- [ ] ADR-035 создан и принят
- [ ] Все файлы документации согласованы с архитектурой proxy
- [ ] gRPC contract определён (proto definition)
- [ ] Schema `lkfl_integration` описана (6 таблиц)
- [ ] ADR-024 обновлён (exception для proxy)
- [ ] 0 рассинхронов между файлами
