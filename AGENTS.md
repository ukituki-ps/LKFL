# AGENTS.md — LKFL Agent Registry

Этот файл — точка входа для агентов проекта **LKFL** (Платформа гибких льгот).
Определения агентов расположены в `.kilo/agent/`.

## Доступные агенты

| Агент | Файл | Роль | Субагенты |
|-------|------|------|-----------|
| **architect-lkfl** | `.kilo/agent/architect-lkfl.md` | Архитектор — проектирует, планирует, раздаёт задачи | `sde-lkfl` |
| **sde-lkfl** | `.kilo/agent/sde-lkfl.md` | Senior Developer — реализация задач из doc/задачи/ | — |

## Цепочка работы

```
Пользователь → architect-lkfl → sde-lkfl (реализация кода)
```

- **architect-lkfl** — основной агент для взаимодействия с пользователем. Делегирует реализацию кода через `task` tool → `sde-lkfl`
- **sde-lkfl** — субагент для реализации Go backend (modular monolith) + React frontend задач

## Контекст проекта — LKFL

**White-label multi-tenant платформа корпоративных льгот.** One codebase, any brand.
Go modular monolith backend (`lkfl-server`) + React SPA frontend.

### Философия — «Три нуля»

| Принцип | Что значит |
|---------|-----------|
| **Нулевая привязка к бренду** | Новый tenant — только CSS + конфиг, без изменения кода |
| **Нулевая привязка к льготам** | Новый провайдер — конфигурация (YAML), не код |
| **Нулевая привязка к модели начислений** | Правила биллинга — CRUD через ЛК-2, не программирование |

### Архитектура — Modular Monolith

**Два бинарника (`lkfl-server` + `lkfl-integration-proxy`), один `go.mod`, 16 internal-пакетов монолита.**

```
lkfl-server (:8080)
├── internal/
│   ├── tenant/          # Tenant CRUD, brand config (системный)
│   ├── auth/            # JWT, tenant resolver, RBAC
│   ├── user/            # CRUD пользователей + HR sync
│   ├── consent/         # ПДн lifecycle
│   ├── cel/             # CEL Rule Engine (5 доменов)
│   ├── llm/             # LLM: agent routing, prompt mgmt, cost tracking
│   ├── eligibility/     # Eligibility engine (CEL-based)
│   ├── compliance/      # Cascade revoke, audit, retention
│   ├── engagement/      # Каталог, flow, collections, survey/ (подпакет)
│   ├── notification/    # Email/push/in-app
│   ├── gamification/    # Ачивки, loyalty
│   ├── billing/         # ← БЫЛ отдельный сервис (M12)
│   ├── integrationclient/ # gRPC client к proxy (M16, заменил integrations/)
│   ├── payments/        # ← БЫЛ отдельный сервис (M12)
│   ├── content/         # FAQ, баннеры, описания карточек
│   ├── recommendations/ # Stub (Phase 2, M15)
│   └── api/             # Public + Admin router'а
├── shared/pkg/
│   ├── auth/            # verifier, middleware, rbac
│   └── celcontext/      # CELContext type
├── cmd/server/          # HTTP entry point
└── cmd/worker/          # Asynq background jobs

lkfl-integration-proxy (:8090 gRPC + :8091 HTTP webhooks) [M16, ADR-035]
├── cmd/integration-proxy/main.go
├── integration-proxy/
│   ├── adapters/        # 11 провайдеров (9 YAML + 2 hard-coded)
│   ├── circuitbreaker/  # Circuit breaker per provider
│   ├── webhook/         # Webhook receiver + verifier
│   ├── grpc/            # gRPC server + generated code
│   └── config/          # YAML config загрузка
└── proto/integration/v1/ # gRPC proto definition
```

**Инфраструктура:** PostgreSQL 17 (2 schemas: `lkfl_platform` + `lkfl_integration`), Redis (key prefixes), Keycloak (OIDC IdP, realm per tenant).

### Frontend — React SPA

Vite + React 18 + `@ukituki-ps/april-ui` + `@ukituki-ps/april-tokens` + Mantine.
State management: Zustand. API: `fetch` через Nginx `/api/v1/`.

### Документация

> **🗺️ Навигация:** [`doc/NAVIGATION.md`](./doc/NAVIGATION.md) — карта «вопрос → файл:строка». Всегда читай первым.

- `doc/README.md` — карта всей документации
- `doc/NAVIGATION.md` — навигация для агентов (вопрос → файл:строка)
- `doc/контекст/` — 6 файлов: философия, проблема, акторы, ограничения, negative-criteria, настраиваемость
- `doc/архитектура/` — 9 файлов + 35 ADR: модули, стек, интеграции, безопасность, schema.md (47 таблиц), cel-engine, engagement, теги, пакеты-platform
- `doc/архитектура/adr/` — 35 ADR. Индекс: [`adr/README.md`](./doc/архитектура/adr/README.md)
- `doc/спецификация/` — артефакты (30), journeys (57), API (118 endpoints), критерии приёмки (66)
- `doc/план/` — вехи M00→M16 (doc) + M17→M44 (code), задачи T{MM}{NN}, зависимости, exit criteria
- `doc/задачи/` — brief.md, plan.yaml, report.md по вехам M01→M16 (doc) + M17→M44 (code)
- `doc/глоссарий.md` — термины, аббревиатуры, коды артефактов/journeys/задач

> **Никогда НЕ читать все файлы документации целиком** — брать только нужный раздел по NAVIGATION.md.

### Система задач

- `doc/план/задачи.md` — реестр 307 задач: 67 doc (M00→M16) + 240 code (M17→M44)
- Каждая задача: `doc/задачи/M{MM}-{slug}/T{MM}{NN}-{name}/`
- **Структура:** `brief.md` (context + plan-ref) + `plan.yaml` (checklist) + `report.md` (отчёт)
- **Номинклатура:** `T{MM}{NN}` — MM = номер вехи, NN = порядковый номер
- **Никогда НЕ удалять** — даже отменённые задачи остаются

### Фазы разработки кода

| Фаза | Вехи | Задач | Результат |
|------|------|------:|-----------|
| **F1** Рабочий каталог | M17–M22 | 40 | Каталог, login, multi-tenancy |
| **F2** Активация и баланс | M23–M29 | 52 | Flow, биллинг, ПДн, периоды |
| **F3** Полный продукт | M30–M38 | 85 | Activity, survey, gamification, proxy |
| **F4** Polish | M39–M44 | 63 | LLM, payments, mobile, v1.0.0 |

Каждая фаза завершается **Hardening вехой** (🛡️): рефакторинг, unit/integration/E2E/load тесты, мониторинг, CI/CD, деплой на стенд, security audit, release package.

### Текущее состояние

- **Документация:** 63/67 задач ✅ (94.0%). M01→M13 + M15 + M16 завершены. M14 отменена.
- **Код:** 0/240 задач. Go-код не начат.
- **Следующая задача:** M17 T1701 — go.mod инициализация
- **M14:** ⛔ отменена (Survey Implementation). Архитектура (M13, ADR-025) сохранена.

### Ключевые метрики

| Метрика | Значение |
|---------|----------|
| Бизнес-проблем | 12 |
| Акторов | 10 |
| Internal-пакетов монолита | 16 (14 business + tenant + api + integrationclient) |
| ADR | 35 (26 Accepted, 4 Superseded, 5 Note) |
| Таблиц БД | 47 (41 lkfl_platform + 6 lkfl_integration) |
| API endpoints | 118 |
| User journeys | 57 |
| Артефактов | 30 |
| Критериев приёмки | 66 |
| Doc задач (M00→M16) | 67 (63 doc ✅, 1 code ⛔ отменена, 3 не учитываются M00) |
| Code задач (M17→M44) | 240 (0 выполнено, 4 фазы) |

### Что НЕ делать без спроса

- Менять `go.mod` или добавлять зависимости
- Давать монолитные коммиты
- Менять систему номенклатуры путей/файлов
- Менять структуру документации (Контекст → Архитектура → Спецификация → План → Задачи)
- Игнорировать раздел «СДЭК в документации» — примеры СДЭК = иллюстрации, не ограничения
