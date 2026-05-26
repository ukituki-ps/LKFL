# T1102 — HR Sync → Platform user/ — отчёт

## Статус

✅ выполнено (документация)

## Что сделано

### `архитектура/пакеты-platform.md` (основной артефакт):
- user/ секция обновлена: назначение расширено "M11 T1102: добавлен HR Sync"
- Структура +fайл: `hr_sync.go` — `HRSync.PullRegistry()`, `SyncStatus()`
- Публичный API +HRSync struct: `PullRegistry(ctx)`, `SyncStatus(ctx)` с Go код-примером
- Зависимости user/ обновлены: +HR-система REST API client (direct call, no NATS)
- Asynq workers mapping: +`hr-sync-daily` worker → `user.HRSync.PullRegistry(ctx)`

### `архитектура/модули.md` (cross-cutting):
- Provider Gateway: hr-sync/ удалён из таблиц модулей и системных адаптеров
- Dependencies: "Platform → HR-система напрямую (Asynq worker `hr-sync-daily`, ADR-026)"
- NATS: integration.hr.* subjects удалены из секции NATS JetStream

### ADR:
- `adr/026-hr-sync-platform.md` — ХАДД: HR-реестр ≠ vendor данные, разные SLA, 2 NATS hop'а → direct call

## Проблемы

- HR-система REST client — новая зависимость Platform (было: через NATS → Integrations). Нужен circuit breaker на будущую реализацию.

## Следующие шаги

N/A — задача выполнена.
