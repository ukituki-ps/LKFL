# M12 — Переход документации на модульный монолит

## Описание

Архитектурный аудит показал: план описывает 4 Go-сервиса с NATS JetStream между ними, однако проект разрабатывается агентами (LLM), а не командой из 20 человек. При агентной разработке микросервисная архитектура создаёт overhead без benefits:

- Каждый NATS contract — не типизирован, не проверяется компилятором → тихие баги
- 16 925 строк документации содержат рассинхроны (ADR-022 vs M10 T1002)
- Агент тратит 8.7K токенов контекста на 1K токенов кода (читает doc + чужой сервис)
- `go build ./...` → 4 модуля, 4 результата, 4 ошибки вместо одной

**Решение:** перейти на модульный монолит — один бинарник, один `go.mod`, модули в `internal/` с Go типизированными интерфейсами.

## Что меняется

| Было (в документации) | Станет |
|---|-|
| 4 Go-сервиса, 4 go.mod, 4 Dockerfile | 1 бинарник, 1 go.mod, 1 Dockerfile (`lkfl-server`) |
| NATS между platform ↔ billing ↔ integrations | Go-функции: `billing.Debit(ctx, ...)`, `integration.Activate(ctx, ...)` |
| Platform публикует `billing.credit` в NATS | `billing.Credit(db, tx, ...)` — прямой вызов, одна PostgreSQL transaction |
| LLM `internal/llm/` подписывается на NATS `llm.generate` | `llm.GenerateCel(ctx, ...)` — прямой вызов из `cel/generator.go` |
| Billing separate service | `internal/billing/` — пакет в Platform |
| Integrations Hub separate service | `internal/integrations/` — пакет в Platform |
| Payment Gateway separate service | `internal/payments/` — пакет в Platform |
| 5 Redis DB (0,1,2,3,4) | Один Redis, key prefixes: `jwt:`, `asynq:`, `catalog:`, `cel:` |
| Docker compose: 14+ контейнеров | Docker compose: 4-5 (pg, redis, keycloak, nginx, lkfl-server) |

## Что НЕ меняется

- Модульные границы **сохраняются** — те же пакеты, та же ответственность
- RBAC, multi-tenancy, Keycloak — без изменений
- PostgreSQL schemas per module → остаются
- Security isolation через DB roles → усиливается (была второстепенной)
- Возможность будущего split на сервисы → контракт interface → replace на gRPC

## Структура после M12

```
backend/
├── cmd/server/main.go          # единая точка входа
├── cmd/worker/main.go          # Asynq worker (тот же)
├── internal/
│   ├── auth/                   # JWT, tenant resolver, RBAC
│   ├── user/                   # CRUD пользователей
│   ├── consent/                # ПДн lifecycle
│   ├── cel/                    # CEL Rule Engine
│   ├── llm/                    # LLM (in-process, было M10 T1002)
│   ├── eligibility/            # Eligibility engine
│   ├── compliance/             # Cascade revoke, audit, retention
│   ├── engagement/             # Каталог, flow, collections
│   ├── notification/           # Email/push/in-app
│   ├── gamification/           # Ачивки, loyalty
│   ├── billing/                # ← бЫЛ отдельный сервис → теперь пакет
│   ├── integrations/           # ← БЫЛ отдельный сервис → теперь пакет
│   ├── payments/               # ← БЫЛ отдельный сервис → теперь пакет
│   └── api/                    # Public + Admin router'а
│       ├── public_router.go
│       ├── admin_router.go
│       └── ...handlers...
├── shared/pkg/                 # общие типы
│   ├── auth/                   # verifier, middleware, rbac
│   └── celcontext/             # CELContext type
├── pkg/                        # общие типы Platform
├── app/                        # DI wiring
├── migrations/                 # Atlas SQL → объединённые
├── go.mod                      # один go.mod
└── Dockerfile                  # один Dockerfile
```

## Fайлы-конфликты

Все задачи M12 правят `архитектура/модули.md` и `архитектура/пакеты-platform.md`.
**Правило:** T1201 создаёт новую ADR → T1202-T1206 правят doc по отдельным файлам → T1207 финальная проверкa.

| Файл | Кто редактирует | Порядок |
|---|--|--|
| `архитектура/adr/024-modular-monolith.md` | T1201 | Новая ADR — основа для всех |
| `архитектура/модули.md` | T1202 | Полная переделка: 4 сервиса → один бинарник |
| `архитектура/пакеты-platform.md` | T1203 | billing/, integrations/, payments/ → internal пакеты |
| `архитектура/nats-subjects.md` | T1204 | Удалено — заменено на DI через interfaces |
| `спецификация/api.md` | T1205 | Убрал `/billing/v1/`, `/payments/v1/` — теперь `/api/v1/` |
| `архитектура/стек.md` | T1206 | NATS JetStream → optional, Docker → 1 backend container |
| Все ADR mention NATS/services | T1207 | Append note "[M12: merged into monolith]" |
| `архитектура/README.md` | T1207 | Обновил диаграмму системы |
| `задачи/README.md` | T1207 | Добавил M12 в таблицу |

## Волны выполнения

### Wave A (новая ADR + main files — параллельно невозможно, sequential)

```
T1201 ──► T1202 ──► T1203
```

T1201 создаёт ADR-024 → T1202 читает ADR и правит модули.md → T1203 читает модули.md и правит пакеты-platform.md.

### Wave B (удаление NATS doc + API spec — параллельно)

```
T1204    T1205    T1206
```

Три задачи правят независимые файлы.

### Wave C (финальная консистентность)

```
T1207
```

Проходит по всем 23 ADR → добавляет note про M12. Обновляет README архитектуры. Обновляет таблицу вех.

## Веху можно закрывать когда

- [ ] T1201 — ADR-024 создана: Modular Monolith — rationale, структура, future split guidance
- [ ] T1202 — `модули.md`: 1 бинарник, 13 internal пакетов, DI через interfaces
- [ ] T1203 — `пакеты-platform.md`: billing/, integrations/, payments/ описаны как internal пакеты
- [ ] T1204 — `nats-subjects.md` удалён, NATS упоминания убраны из doc (оставлен как dev-only)
- [ ] T1205 — `api.md`: все endpoints под `/api/v1/`, убраны `/billing/v1/`, `/payments/v1/`
- [ ] T1206 — `стек.md`: NATS → optional, 1 backend container, убрал :8081-:8085
- [ ] T1207 — Все ADR проверены, README обновлён, диаграмма системы переделана

## Задачи вехи

| Задача | Описание | Волна | Статус |
|---|--|-|---|
| T1201 | ADR-024: Modular Monolith rationale | A | ⬜ не начата |
| T1202 | Переделка `архитектура/модули.md` — 1 бинарник | A | ⬜ не начата |
| T1203 | Переделка `архитектура/пакеты-platform.md` — billing/integrations/payments как internal | A | ⬜ не начата |
| T1204 | Удаление NATS doc, замена на DI interfaces | B | ⬜ не начата |
| T1205 | Переделка `спецификация/api.md` — unified `/api/v1/` | B | ⬜ не начата |
| T1206 | Обновление `стек.md` — NATS optional, 1 container | B | ⬜ не начата |
| T1207 | Финальная консистентность: ADR notes + README + диаграмма | C | ⬜ не начата |
