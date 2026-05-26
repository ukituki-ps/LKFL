# ADR-028: NATS namespace rename — `integration.*` → `provider.*` + consumer remapping

**Статус:** ❌ Superseded by ADR-024 (Modular Monolith)
**Дата:** 2026-05-25
**Контекст:** M11 T1104 — NATS consumer update после split integrations

## Ситуация

`архитектура/nats-subjects.md` документирует consumer mapping per subject. После M11:
- `integration.payroll.submit` — consumer становится Billing (было Integrations, T1103)
- `integration.hr.pull` / `integration.hr.synced` — deleted (internal call внутри Platform, T1102)
- `integration.engagement.*` — consumer остаётся provider-gateway

**Проблемы:**
- Namespace `integration.*` вводит в заблуждение: provider-gateway ≠ universal integration hub
- Consumer mapping устарел: billing должен consumer'ить payroll subject
- HR subjects не нужны (direct call, T1102)

## Решение

### Namespace rename

| Было | Стало | Причина |
|-----|---|-|
| `integration.engagement.activate` | `provider.engagement.activate` | Provider Gateway identity (T1101) |
| `integration.engagement.deactivate` | `provider.engagement.deactivate` | Provider Gateway identity (T1101) |
| `integration.status` | `provider.status` | Provider Gateway identity (T1101) |
| `integration.catalog.sync` | `provider.catalog.sync` | Provider Gateway identity (T1101) |
| `integration.payroll.submit` | `finance.payroll.submit` | Payroll → Billing (T1103) |

### Consumer remapping

| Subject | Был Consumer | Стал Consumer |
|---|-|---|
| `provider.engagement.*` | integrations | provider-gateway (same binary, conceptual) |
| `provider.status` | integrations | provider-gateway |
| `provider.catalog.sync` | integrations | provider-gateway |
| `finance.payroll.submit` | integrations | billing |
| *(deleted)* `integration.hr.pull` | integrations | — (Asynq worker) |
| *(deleted)* `integration.hr.synced` | integrations | — (Asynq worker) |

### Namespace Rules обновление

| Namespace | Владелец | Назначение |
|---|-|---|
| `provider.*` | Platform ↔ Provider Gateway | Провайдеры льгот |
| `finance.payroll.*` | Platform ↔ Billing | 1C бухгалтерия |
| `billing.*` | Platform ↔ Billing | Финансовые операции |
| `payment.*` | Platform ↔ Payment Gateway | Платежи (PCI DSS) |

## Альтернативы

| Вариант | Вердикт | Причина |
|-------|-|-----|
| Оставить `integration.*` | ❌ | Namespace не отражает реальность после split |
| Два rename | ❌ | Breaks все существующие consumers |
| Rename + consumer remapping (выбрано) | ✅ | Clear mapping, reflects new domain boundaries |

## Последствия

- ✅ Namespace `provider.*` отражает Provider Gateway identity
- ✅ `finance.payroll.*` отражает finance domain ownership
- ⚠️ Все consumers нужно переключить на новые subject'ы (migration window)
- ⚠️ NATS JetStream: старые streams `integration.*` можно archive после cut-over
- ⚠️ NATS JetStream: новые streams `provider.*`, `finance.payroll.*` создать до deploy