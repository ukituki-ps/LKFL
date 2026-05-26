# ADR-006: Биллинг как модуль (был отдельный сервис)

**Статус:** ⚠️ Note: M12 merged
**Дата:** 2026-05-22, обновлено 2026-05-25 (M12)
**Контекст:** М01-создание-описания

## Обновление M12

> **M12:** Billing слит в lkfl-server как `internal/billing/` ([ADR-024](./024-modular-monolith.md)). Один бинарник, один go.mod. ACID гарантии сохраняются — одна PostgreSQL transaction между `engagement/flow/` и `billing/`. Контракты: Go interface вместо NATS.

## Контекст (исторический)

Биллинг отвечает за баланс пользователя, транзакции (credit/debit), периоды распределения, сгорание баллов. 4+ источника начислений, N категорий списаний (конфигурируемо), peak-нагрузка при начале периода.

## Решение (историческое)

**billing/** — отдельный Go-сервис с own БД (`lkfl_billing`), own `go.mod`, own Dockerfile. Операционная коммуникация через NATS JetStream. Управленческий API через Nginx (HR, менеджер каталога).

**NATS subject'ы (операционные):**
```
billing.credit           # начисление
billing.debit            # списание
billing.balance.query    # запрос баланса
billing.balance.result   # ответ с балансом
billing.transactions.*   # история транзакций
```

**Управленческий REST (через Nginx):**
```
GET/POST/PUT/DELETE /billing/v1/rules       # CRUD правил
POST /billing/v1/periods/:id/activate       # открыть период
POST /billing/v1/periods/:id/expire         # закрыть + сгорание
```

> **M12:** NATS subjects больше не используются. `/billing/v1/` → `/api/v1/billing/`.

## Аргументы «за» (исторические)

- Финансовая точность: ACID на каждую транзакцию
- Peak isolation: биллинг не падает из-за проблем в каталоге
- ФСТЭК: финансовые данные — отдельный audit trail
- Independent release: новая категория списаний → deploy только billing

## Аргументы «против» (исторические)

- Extra service = extra infra overhead
- RPC latency вместо in-process call
- Нужно синхронизировать API контракт при изменении

## Вердикт

**За.** Финансовые операции критичны — separation of concerns оправдана.
> **M12:** separation of concerns сохраняется на уровне `internal/billing/` пакета, не отдельного бинарника.

## Следствия

- **M12:** Billing общается с engagement/ через direct Go call `billing.BillingService.*()` — compile-time safety, одна PostgreSQL transaction
- **M12:** Platform имеет прямой доступ к `internal/billing/` — без NATS
- **M12:** Изменение API контракта → compile-time error (Go interface), не runtime NATS mismatch
