# ADR-026: HR Sync → Platform user/ (перенос из Integrations)

**Статус:** ❌ Superseded by ADR-024 (Modular Monolith)
**Дата:** 2026-05-25
**Контекст:** M11 T1102 — split integrations по доменам

## Ситуация

`integrations/hr-sync/` — кадровый реестр — находится в Integrations Hub = wrong domain:
```
Platform → NATS `integration.hr.pull` → Integrations → HR API
Integrations → NATS `integration.hr.synced` → Platform
```

**Проблемы:**
- HR-данные ≠ vendor данные. HR-реестр — это user domain.
- Если Provider Gateway падает → HR-sync тоже падает. Разные SLA.
- 2 NATS hop'а для одного pull-запроса = latency overhead.

## Решение

Перенести hr-sync в Platform `internal/user/`:
```
Platform Asynq `hr-sync-daily` → user.HRSync.PullRegistry(ctx) → HR-система REST (direct)
```

**Изменения:**
- `архитектура/пакеты-platform.md` — user/ +hr_sync.go (PullRegistry, SyncStatus)
- `архитектура/пакеты-platform.md` — Asynq workers +hr-sync-daily
- `архитектура/модули.md` — hr-sync/ удалён из Provider Gateway tables
- `архитектура/nats-subjects.md` — integration.hr.pull/synced удалены

**Почему не отдельный сервис:** 1 worker/daily, low traffic, direct REST call. Overhead отдельного бинарника не оправдан.

**Зависимости:**
- HR-система REST API client (новый dependency для Platform user/)
- Asynq cron scheduler (Redis DB 1) — существующая инфраструктура

**Последствия:**
- ✅ HR-data belongs к user domain (domain correctness)
- ✅ Decoupling: HR-down ≠ provider-down
- ⚠️ Platform теперь зависит от HR-системы напрямую (было: через NATS → Integrations)
- ⚠️ Нужен circuit breaker в HRSync для fallback на кэшированные данные