# T1102 — HR Sync → Platform user/

## Веха

M11-split-integrations

## Контекст

`integrations/hr-sync/` документируется как:
- `PullRegistry()` — ежедневный pull кадрового реестра
- `SyncStatus()` — статус последней синхронизации

NATS flow:
```
platform publishes: integration.hr.pull     → integrations subscribes
integrations publishes: integration.hr.synced  → platform subscribes
```

**Проблема:**
HR-реестр — это данные о пользователях (импортирует → Platform user/). Он находится в Integrations = wrong domain. Если Integrations упал:
- Провайдеры льгот недоступны (это OK, expected)
- HR реестр не обновляется (это bad) — employee данные ≠ vendor данные

**Решение — перенести hr-sync в Platform user/:**
```
platform/
  internal/
    user/
      models.go
      repository.go
      import.go         ← HR registry import logic (было integrations/hr-sync/)
      hr_sync.go        ← PullRegistry, SyncStatus (было integrations/hr-sync/)
```

Platform делает outgoing HTTP call к HR-системе напрямую (без NATS промежуточного сервиса).
Раньше: Platform → NATS `integration.hr.pull` → Integrations → HR API → NATS `integration.hr.synced` → Platform
Теперь: Platform Asynq hr-sync-daily → user.HRSync.PullRegistry(ctx) → HR API (direct REST)

NATS subjects `integration.hr.pull` и `integration.hr.synced` → убраны (не нужны, direct call).

### Файлы-мишени

| Действие | Файл |
|---|-|-|-|-|-|
| hr-sync → user/ | `архитектура/пакеты-platform.md` — user/ добавлен hr_sync.go |
| Убрать из Integrations | `архитектура/модули.md` — hr-sync/ удалён из integrations |
| NATS subjects | `архитектура/nats-subjects.md` — integration.hr.pull/synced удалены |
| Asynq workers | `архитектура/пакеты-platform.md` — добавить hr-sync-daily worker |
| Обновить DI граф | `архитектура/модули.md` — Platform теперь consumer HR |
| Создать ADR | `архитектура/adr/026-hr-sync-platform.md` |
| Обновить README архитектуры | `архитектура/README.md` — ADR-026 |

### Критерии приёмки

- [ ] `архитектура/пакеты-platform.md` — user/ включает hr_sync.go (PullRegistry, SyncStatus)
- [ ] `архитектура/модули.md` — hr-sync/ удалён из Integrations tables
- [ ] `архитектура/nats-subjects.md` — integration.hr.pull и integration.hr.synced удалены
- [ ] Asynq worker hr-sync-daily → user.HRSync.PullRegistry documentирован в `пакеты-platform.md`
- [ ] HR-система REST client documentирован как dependency user/ (не через NATS)
- [ ] Создан ADR-026: обоснование переноса (HR data belongs to user domain, decoupling from vendor failures)
- [ ] `архитектура/README.md` — ADR-026 в таблицу
