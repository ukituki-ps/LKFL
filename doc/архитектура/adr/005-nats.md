# ADR-005: NATS JetStream — optional dependency (был message broker для интеграций)

**Статус:** ❌ Superseded by ADR-024 (Modular Monolith)
**Дата:** 2026-05-22, обновлено 2026-05-25 (M12)
**Контекст:** М01-создание-описания

> ⚠️ **Optional после M12 (ADR-024).** NATS JetStream заменён Go function calls в modular monolith. В mono-режиме не используется. Feature flag `nats.enabled=false` по умолчанию.

## Обновление M12

> **M12:** NATS JetStream оставлен как **optional dependency** для future microservice split. В mono-режиме интерфейсы не используется. См. [ADR-024](./024-modular-monolith.md).
> Все межмодульные вызовы в mono-режиме → Go interfaces through direct function calls. NATS consumer'и не запускаются. Feature flag `nats.enabled=false` по умолчанию.

## Контекст (исторический)

N внешних интеграций (провайдеры льгот, HR-система, 1С, пэйшлюз, SSO) с разными протоколами и частотами. Прямые вызовы от платформы к внешнему миру невозможны:
- Отказоустойчивость: падение одного провайдера не должно ломать платформу
- ПДн: данные провайдеров передаются только после проверки consent
- ФСТЭК: единое место хранения

## Решение (историческое)

**NATS JetStream** как message broker между platform и integrations сервисами:
- Persistent messages (JetStream)
- Native Go client
- Лёгкий deploy (один бинарник)
- Dead letter queue для failed messages

**Топики (основные):**

| Subject | Producer | Consumer |
|-----|--|-|
| `integration.activate` | platform | integrations |
| `integration.deactivate` | platform | integrations |
| `integration.status` | integrations | platform |
| `integration.catalog.sync` | integrations | platform |
| `integration.hr.pull` | platform | integrations |
| `integration.hr.synced` | integrations | platform |
| `integration.payment.*` | platform | integrations |
| `integration.payroll.*` | platform | integrations |
| `billing.credit` | platform | billing |
| `billing.debit` | platform | billing |

> **M12:** Эти subject'ы больше не используются — см. `модули.md` → Go interfaces.

## Альтернативы

| Вариант | Плюсы | Минусы |
|-----|--|-|
| Apache Kafka | Strict ordering, ecosystem | Heavy, сложный deploy, overkill для N провайдеров |
| RabbitMQ | DLX, plugin ecosystem | Erlang dependency, сложнее масштабировать |
| Direct API call | Проще | Нет изоляции, нет retry, нарушает ПДн |

## Следствия

- **M12 в mono-режиме:** NATS не используется. `internal/integrations/`, `internal/billing/`, `internal/payments/` вызываются через direct Go call
- **Future microservice split:** NATS consumer'и и publisher'и можно включить через feature flag, и интерфейсы для каждого сервиса заменяются на NATS publisher/client
- Platform хранит кэшированные данные локально
- Dead letter → alert в Sentry
- Circuit breaker: >10 ошибок за 5 мин → использовать кэш
