# T1601 — ADR-035: Integration Proxy — Отчёт

## Что сделано

Создан ADR-035: `архитектура/adr/035-integration-proxy.md` (~350 строк).

## Содержание

1. **Контекст:** 5 рисков прямых интеграций (goroutine blocking, fault isolation, credential blast radius, webhook surface, circuit breaker только в документации)
2. **Варианты:** 4 рассмотрены — оставить как есть (❌), Asynq worker (❌), Integration Proxy (✅), Sidecar (❌)
3. **Решение:**
   - Два бинарника, один go.mod, один репозиторий
   - gRPC contract между монолитом и proxy
   - Асинхронная активация (job_id + callback)
   - Circuit breaker per provider (10 failures/60s, 30s recovery)
   - Worker pool per provider (5 concurrent, 100 queue)
   - Credential isolation (ключи только в proxy)
   - Database: 1 PG, 2 schemas (lkfl_platform + lkfl_integration)
   - 6 метрик мониторинга
   - Nginx routing: /api/* → mono, /webhooks/* → proxy
4. **Последствия:** fault isolation, credential isolation, goroutine safety vs дополнительный бинарник, gRPC contract
5. **ADR-024:** не отменяется, добавлено исключение для I/O boundary

## Проблемы

Нет. ADR согласован с существующей архитектурой.
