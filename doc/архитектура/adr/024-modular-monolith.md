# ADR-024: Modular Monolith — переход от микросервисов к одному бинарнику

**Статус:** Accepted
**Дата:** 2026-05-25
**Контекст:** M12-переход-на-модульный-монолит, T1201

---

## Контекст

Документация проекта описывает 4 Go-сервиса (`platform`, `billing`, `integrations`, `payment-gateway`) с NATS JetStream для межсервисной коммуникации, каждый со своим `go.mod`, Dockerfile и lifecycle. Однако проект разрабатывается с использованием LLM-агентов, а не распределённой командой из 20+ разработчиков.

Микросервисная архитектура создаёт overhead без соответствующих benefits при агентной разработке:

| Проблема | Влияние |
|---|---|
| NATS contract'и не типизированы, не проверяются компилятором | Тихие баги: typo в subject name → message в вакууме |
| Агент тратит 8.7K токенов контекста на 1K токенов кода | Нужно читать doc другого сервиса + его код + NATS contract |
| 4 `go.mod`, 4 Dockerfile, 14+ контейнеров | Infra noise: `docker-compose up` → 14 сервисов вместо 4 |
| `go build ./...` → 4 модуля, 4 результата, 4 возможных ошибки | Slow feedback loop для agent: итерация = 4 билда |
| 16 925 строк документации содержат рассинхроны | ADR-022 vs M10 T1002: LLM Proxy описан и как сервис, и как in-process |
| Distributed transactions (NATS publish → consumer) | 2PC/saga вместо одной PostgreSQL transaction |

Кроме того, после M10 T1002 LLM Proxy уже слит в Platform как `internal/llm/`, а после M11 hr-sync и 1C перенесены в соответствующие пакеты в Platform и Billing. Архитектура уже движется к монолиту на практике — документация должна это отразить.

---

## Рассмотренные варианты

### Вариант 1: Микросервисы (текущее состояние документации)

4 Go-сервиса с NATS JetStream, 1 React SPA.

| Плюсы | Минусы |
|---|---|
| Независимый deploy каждого сервиса | NATS contract'и — runtime-only validation, нет compile-time safety |
| Независимое масштабирование | Агент тратит 8.7K токенов на чтение контекста другого сервиса |
| Физическая изоляция (PCI DSS для Payment Gateway) | 4 go.mod, 4 Dockerfile, 14+ контейнеров |
| Familiar pattern для команд >10 человек | Distributed transactions — 2PC/saga вместо ACID одной DB |

**Вердикт:** ❌ Отказ. Team = 1 агент, масштабирование не требуется. Overhead > benefit.

### Вариант 2: Монолит без модульности

Один бинарник, один `internal/` с одним God Object пакетом.

| Плюсы | Минусы |
|---|---|
| Простота, минимум overhead | SRP violation: один пакет = все бизнес-домены |
| Один go.mod, один Dockerfile | Нет тестовой изоляции — unit-тест требует весь God Object |
| Быстрая итерация | Maintenance cost растёт линейно с количеством endpoint'ов |

**Вердикт:** ❌ Отказ. 146+ endpoint'ов, growing feature set — нужна модульность.

### Вариант 3: Modular Monolith (ВЫБРАН)

Один бинарник `lkfl-server`, один `go.mod`, модули в `internal/` с Go типизированными интерфейсами.

| Плюсы | Минусы |
|---|---|
| Type-safe контракты: Go interface → compile-time validation | Нельзя деплоить модуль отдельно (не нужен при 1 агенте) |
| Один `go build ./...` → один результат | Future split требует рефакторинг (mitigated: interface → gRPC) |
| Unit-тесты пакета изолированы через DI | — |
| Агент читает один `go.mod` + один `internal/` tree | |
| ACID гарантии: одна PostgreSQL transaction | |
| 0 NATS overhead (или optional для future split) | |

**Вердикт:** ✅ Выбран. Best fit для агентной разработки: compile-time contracts, single build, тестовая изоляция, минимальный context overhead.

---

## Решение

Перейти на **modular monolith** — один бинарник, один `go.mod`, 17 internal-пакетов (15 business + tenant + api), Go-интерфейсы для контрактов между модулями.

### Структура

```
backend/
├── cmd/server/main.go          # единая точка входа, HTTP API
├── cmd/worker/main.go          # Asynq worker (тот же)
├── internal/
│   ├── tenant/                 # Tenant CRUD, brand config (системный)
│   ├── auth/                   # JWT, tenant resolver, RBAC
│   ├── user/                   # CRUD пользователей
│   ├── consent/               # ПДн lifecycle
│   ├── cel/                    # CEL Rule Engine
│   ├── llm/                    # LLM (in-process, было M10 T1002)
│   ├── eligibility/            # Eligibility engine
│   ├── compliance/             # Cascade revoke, audit, retention
│   ├── engagement/             # Каталог, flow, collections, survey/
│   ├── notification/           # Email/push/in-app
│   ├── gamification/           # Ачивки, loyalty
│   ├── recommendations/        # Stub (Phase 2, M15)
│   ├── billing/                # ← БЫЛ отдельный сервис → теперь пакет
│   ├── integrations/           # ← БЫЛ отдельный сервис → теперь пакет
│   ├── payments/               # ← БЫЛ отдельный сервис → теперь пакет
│   ├── content/                # FAQ, баннеры, описания
│   └── api/                    # Public + Admin router'а
├── shared/pkg/                 # общие типы
│   ├── auth/                   # verifier, middleware, rbac
│   └── celcontext/             # CELContext type
├── pkg/                        # общие типы Platform
├── app/                        # DI wiring
├── migrations/                 # Atlas SQL → объединённые
├── go.mod                      # один go.mod
└── Dockerfile                  # один Dockerfile
```

### Контракты между модулями

Было (NATS — runtime, нет type safety):
```
platform → NATS "billing.credit" → billing
platform → NATS "payment.authorize" → payment-gateway
platform → NATS "provider.engagement.activate" → integrations
```

Стало (Go interface — compile-time safety):
```go
billing.Debit(ctx, userId, amount, category, offerId)
payments.Authorize(ctx, userId, amount, method)
integrations.Activate(ctx, userId, offerId, provider)
```

### Замещение инфраструктуры

| Было | Стало |
|---|---|
| 4 Go-сервиса, 4 go.mod, 4 Dockerfile | 1 бинарник, 1 go.mod, 1 Dockerfile |
| NATS между сервисами | Go function calls через internal interface |
| 5 Redis DB (0,1,2,3,4) | Один Redis, key prefixes: `jwt:`, `asynq:`, `catalog:`, `cel:`, `rate:` |
| Docker compose: 14+ контейнеров | Docker compose: 5 (pg, redis, keycloak, nginx, lkfl-server) |
| Ports :8080, :8081, :8082, :8084, :8085 | Port :8080 + :8083 (asynq dashboard) |
| 4 health endpoints | Один `/api/healthz` |
| `/billing/v1/` + `/payments/v1/` | Всё под `/api/v1/` |

### NATS — optional dependency

NATS JetStream не удаляется из кодовой базы, но помечается как **optional**:
- В mono-режиме NATS не используется (все вызовы — Go function calls)
- Future microservice split: interface → replace implementation на NATS/gRPC publisher
- NATS remains as dev-only dependency для feature flag / split readiness testing

### Модульные границы сохраняются

Важно: модульные границы (ответственность каждого пакета) **не меняются**. Меняется только deployment boundary:
- `billing/` всё ещё отвечает за финансовые операции с ACID-гарантиями
- `payments/` всё ещё изолирован для PCI DSS
- `integrations/` всё ещё gateway к провайдерам льгот
- Тестовая изоляция сохраняется: каждый пакет тестируется через DI + mock

### Future split guidance

Если в будущем понадобится split на отдельные сервисы:
1. Go interface в `internal/` уже определяет контракт
2. Замени implementation на gRPC/NATS publisher без изменения API поверхности
3. Вынеси `internal/billing/` → отдельный go module → новый сервис
4. Модульная дисциплина уже обеспечена — split не требует рефакторинга

---

## Альтернативы (кратко)

| Вариант | Приёмлемо? | Причина |
|---|:---:|---|
| Микросервисы (текущее) | ❌ | Overhead при агентной разработке |
| Монолит без модульности | ❌ | Нет SRP, нет тестовой изоляции |
| **Modular Monolith** | ✅ | Type-safe, один build, тестовая изоляция |
| Plugin-based architecture | ❌ | Сложнее чем нужно, нет Go plugin ecosystem |

---

## Последствия

### Положительные

1. **Compile-time contracts** — Go interfaces вместо NATS string subjects
2. **Один `go build`** — feedback loop ускорен в 4 раза
3. **Тестовая изоляция** — unit-тест каждого internal пакета через DI mock
4. **ACID транзакции** — одна PostgreSQL transaction вместо distributed saga
5. **Меньше контекста для агента** — один `go.mod`, один tree
6. **Инфраструктура** — 5 контейнеров вместо 14+; один Redis вместо 5 DB
7. **Консистентность документации** — все ADR, модули, API spec согласованы

### Отрицательные

1. **Независимый deploy** — нельзя деплоить только billing → не критично при 1 агенте
2. **Совместное масштабирование** — billing растёт вместе с platform → не критично, peak нагрузки не совпадают в текущем масштабе
3. **PCI DSS** — `internal/payments/` логически изолирован, credentials не share'ятся → физическая изоляция заменяется на separation of concerns на уровне кода

---

## Связь с другими ADR

| ADR | Статус | Изменение |
|-----|---|---|
| ADR-005 (NATS JetStream) | ⚠️ Optional | NATS → optional dependency |
| ADR-006 (Billing отдельно) | ⚠️ Note | Billing → `internal/billing/` |
| ADR-010 (Nginx) | ⚠️ Note | Routes обновлены |
| ADR-011 (Monorepo) | ✅ Valid | Monorepo, но 1 go.mod вместо 4 |
| ADR-013 (пакеты Platform) | ✅ Valid | Теперь 17 пакетов вместо 11 |
| ADR-018 (Payment Gateway) | ⚠️ Note | Payment → `internal/payments/` |
| ADR-020 (NATS registry) | ❌ Superseded | NATS registry заменён DI interfaces |

---

## M16: Exception — Integration Proxy

> **M16 (ADR-035):** Integration Proxy — отдельный бинарник для I/O isolation.
> Это исключение из "1 бинарник" правила, аналогичное `cmd/worker/` как второму entry point.
> Обоснование: внешние HTTP calls — это не бизнес-логика, это I/O boundary.
> Один go.mod сохраняется. ADR-024 остаётся valid для бизнес-логики (16 internal-пакетов монолита).

---

## Статус

✅ Accepted (M12, T1201) — с исключением M16 (Integration Proxy)