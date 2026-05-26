# T1601 — ADR-035: Integration Proxy

## Веха

M16-integration-proxy

## Контекст

ADR-024 (M12) принял modular monolith с прямыми Go вызовами `integrations.ProviderGateway.*()` из `engagement/flow/`. Прямые HTTP calls из монолита к внешним провайдерам создают риски: goroutine blocking, отсутствие fault isolation, credential blast radius, webhook surface area.

## Что сделать

Написать ADR-035: обоснование выноса внешних интеграций в отдельный бинарник `lkfl-integration-proxy`.

### Содержание ADR

1. Контекст — риски прямых интеграций (таблица 5 рисков)
2. Термины — монолит, proxy, провайдер, адаптер, hot path
3. Рассмотренные варианты:
   - Вариант 1: оставить как есть (❌)
   - Вариант 2: Asynq worker без нового бинарника (❌)
   - Вариант 3: Integration Proxy (✅ ВЫБРАН)
   - Вариант 4: Sidecar pattern (❌)
4. Решение:
   - Архитектура (diagram: Nginx → монолит + proxy → провайдеры)
   - Структура проекта (cmd/server, cmd/integration-proxy, integration-proxy/, proto/)
   - gRPC contract (protobuf definition)
   - Sync vs Async операции (таблица)
   - Flow активации (7 шагов, асинхронный)
   - Webhook handling (provider → proxy → mono callback)
   - Circuit breaker (threshold, states, recovery)
   - Worker pool per provider
   - Credential storage (proxy only, mono knows nothing)
   - Database (1 PG, 2 schemas: lkfl_platform + lkfl_integration)
   - Fault tolerance (таблица сценариев)
   - Мониторинг (6 метрик)
   - Nginx config
   - Ports (8080 mono, 8090 gRPC, 8091 HTTP webhooks)
   - Docker Compose (6 контейнеров)
5. Последствия (positive + negative)
6. Влияние на ADR-024 (exception, не отмена)
7. Связь с другими ADR

### Файлы-мишени

| Действие | Файл |
|----------|------|
| Создать | `архитектура/adr/035-integration-proxy.md` |

### Критерии приёмки

- [ ] ADR-035 создан
- [ ] 4 варианта рассмотрены с плюсами/минусами
- [ ] Архитектура описана: diagram, структура проекта, gRPC contract
- [ ] Sync vs Async операции определены
- [ ] Flow активации описан (7 шагов)
- [ ] Webhook handling описан
- [ ] Circuit breaker параметры определены
- [ ] Database schema определена (lkfl_integration)
- [ ] Fault tolerance сценарии описаны
- [ ] Мониторинг метрики определены
- [ ] Влияние на ADR-024 описано
