# T0601 — Разбиение Platform на 3 внутренних пакета

## Контекст

После M05 (унификация Engagement) модуль `engagement/` внутри Platform стал **God Object**:

| Обязанность | Где сейчас | Проблема |
|---|---|---|
| Каталог энгейджментов (CRUD, фильтры, поиск) | `engagement/` | Смешана с flow-логикой |
| Eligibility engine (AND/OR/groups) | `engagement/` | Самая сложная бизнес-логика — нет изоляции для тестов |
| Flow execution (activation/completion/revert) | `engagement/` | 3 worker'а перемешаны с остальными Asynq |
| Collections (наборы benefit-офферов) | `engagement/` | ~40% всей поверхности Platform |
| Email/Push/In-app уведомления | нет выделенного пакета | Логика рассыпок в worker `notification-send` |
| Шаблонизатор уведомлений | нет | Нет abstractions над каналами доставки |
| Рекомендательная система (контекст + сегмент) | нет выделенного пакета | М06 — отдельная веха, нет изоляции для тестов |
| Debug рекомендаций по userId | нет | `/admin/recommendations/debug/:userId` перемешан |

**Документация `модули.md` описывает Platform как:**
```
app/        — инициализация, DI
auth/       — Keycloak OIDC
user/       — CRUD пользователей
engagement/ — каталог, eligibility, flow, collections, billing events (5 обязанностей)
```

**8 Asynq workers:**
```
notification-send      ← уведомление — нет своего пакета
document-generate      ← документ — ок
registry-import-xlsx   ← реестр — ок
registry-import-api    ← реестр — ок
consent-revoke         ← ПДн — ок
engagement-activate    ← engagement — ok
engagement-complete    ← engagement — ok
engagement-revert      ← engagement — ok
```

## Проблема

1. **SRP violation (Single Responsibility Principle):** `engagement/` — 5 обязанностей в одном пакете. Нарушение прямо противоречит принципу SOLID, заявленному в философии проекта.
2. **Нет изоляции для тестирования:** eligibility engine (AND/OR/groups/segments) — самая сложная бизнес-логика. Без выделенного пакета невозможно написать unit-тесты без запуска всего Platform.
3. **Notification scattered:** уведомления распределены между worker и api-handlers. Нет `Channel` interface. Нет шаблонов.
4. **Recommendations невидимы:** 5 admin endpoints + 1 user endpoint + 1 debug endpoint — всё перемешано в `api/`. М06 — отдельная веха с собственным lifecycle.
5. **Документация не отражает внутреннюю структуру:** `модули.md` описывает 4 модуля Platform, но реальная бизнес-логика распределена иначе.

## Решение

Разбить Platform на **3 новых внутренних пакета** (не сервисы, не отдельные бинарники):

```
platform/
├── cmd/server/          # HTTP entry — один бинарник
├── cmd/worker/          # Asynq entry — один бинарник
├── internal/
│   ├── auth/           — JWT, tenant resolver (без изменений)
│   ├── user/           — CRUD profile (без изменений)
│   ├── consent/        — ПДн lifecycle (без изменений)
│   │
│   ├── engagement/     ← НОВЫЙ — каталог, eligibility, flow execution, collections
│   ├── notification/   ← НОВЫЙ — шаблоны, каналы (email/push/in-app), очередь
│   ├── recommendations/← НОВЫЙ — правила контекст+сегмент, debug, агрегация
│   │
│   └── api/            — thin HTTP handlers → делегируют в internal/*
├── pkg/                — общие типы (Claims, Tenant, Pagination)
└── go.mod
```

**Почему пакеты, а не сервисы:**

| Аргумент | Пакеты (internal/*) | Отдельные сервисы |
|---|---|---|
| Бинарников | 2 (server + worker) — без изменений | 5 ( engagement-svc, notification-svc, ...) |
| Nginx routes | нет новых правил | 5+ новых upstream → сложность |
| Общие зависимости (auth, user, consent) | один import | dублирование или RPC |
| Тестируемость | unit-тесты пакета изолированы | integration-тесты нужны |
| Масштаб (40 endpoints) | один handler package | overkill |
| Масштаб-когда-смысленно | 300+ endpoints, 3+ команды | сейчас не применимо |

### Пакет `engagement/`

**Обязанности:** каталог, eligibility, flow execution, collections.

**Перемещается из `engagement/` (переименовывается в `internal/engagement/`):**

| Что сейчас | Что будет |
|---|---|
| `engagement/` (god object) | `internal/engagement/catalog/` — CRUD, фильтры, поиск |
| eligibility в `engagement/` | `internal/eligibility/` — CEL-based engine (вынесен M07 T0701) |
| flow execution в `engagement/` | `internal/engagement/flow/` — step-by-step executor |
| collections в `engagement/` | `internal/engagement/collections/` — bundle management |

**Asynq worker'ы остаются:**
- `engagement-activate` → запускает `flow/flow_engine.go`.Activate(ctx, offerId, userId)
- `engagement-complete` → запускает `flow/flow_engine.go`.Complete(ctx, engagementId)
- `engagement-revert` → запускает `flow/flow_engine.go`.Revert(ctx, engagementId)

**Публичный API модуля (был God Object → разделён на подпакеты):**
> ⚠️ *Исторический контекст M06.* `EngagementEngine` заменён на `CatalogService` (catalog/), `FlowEngine` (flow/), `CollectionsEngine` (collections/).

```go
// engagement/catalog/
type CatalogService struct{ /* DI: db, cache, logger */ }
func (s *CatalogService) List(ctx, q) ([]Engagement, Pagination, error)
func (s *CatalogService) Get(ctx, id) (*Engagement, error)

// engagement/flow/
type FlowEngine struct{ /* DI: db, billing, integrations, payments, asynq, logger, eligibility */ }
func (f *FlowEngine) Activate(ctx, offerId, userId) error
func (f *FlowEngine) Complete(ctx, engagementId) error
func (f *FlowEngine) Revert(ctx, engagementId) error
func (f *FlowEngine) ExecuteStep(ctx, engagementId, stepId, data) error

// engagement/collections/
type CollectionsEngine struct{ /* DI: db, flowEngine, logger */ }
func (c *CollectionsEngine) List(ctx) ([]Collection, error)
func (c *CollectionsEngine) Engage(ctx, collectionId, userId) error
```

### Пакет `notification/`

**Обязанности:** шаблоны, каналы доставки, очередь уведомлений.

**Создаётся с нуля — логика вытаскивается из worker `notification-send` и api-handlers.**

**Структура:**
```
internal/notification/
├── notification.go        — main API: Send(ctx, userId, template, data)
├── channels.go            — Channel interface, implementations
├── email.go               — email channel (SMTP/template)
├── push.go                — push channel (FCM/APNs)
├── inapp.go               — in-app channel (в БД, через /notifications)
├── templates.go           — template engine (GOTemplates)
├── store.go               — persistence (sent, read, pending)
└── handler.go             — HTTP handlers для /notifications
```

**Публичный API пакета:**
```go
type NotificationEngine struct{ /* DI: db, templates, channels */ }

type Channel interface {
    Send(ctx Context, recipient string, payload *Payload) error
    Name() string
}

func (n *NotificationEngine) Send(ctx Context, userId UUID, template string, data map[string]string) error
func (n *NotificationEngine) ListUnread(ctx Context, userId UUID) ([]Notification, error)
func (n *NotificationEngine) MarkRead(ctx Context, userId UUID, notificationId UUID) error
func (n *NotificationEngine) MarkAllRead(ctx Context, userId UUID) error
```

**Asynq worker:**
- `notification-send` → notification.go.SendFromEvent(ctx, event) — единая точка входа

### Пакет `recommendations/`

**Обязанности:** правила контекст+сегмент, evaluation, debug.

**Создаётся с нуля — логика вытаскивается из api-handlers и M06 endpoints.**

**Структура:**
```
internal/recommendations/
├── engine.go              — main API: Recommend(ctx, userId, context)
├── rules.go               — CRUD правил (context + segment rules)
├── evaluation.go          — evaluation engine (сегмент matching)
├── debug.go               — debug по userId (какие правила сработали)
├── storage.go             — persistence (rules, hit counts)
└── handler.go             — HTTP handlers для /recommendations + admin
```

**Публичный API пакета:**
```go
type RecommendationsEngine struct{ /* DI: db, cache */ }

func (r *RecommendationsEngine) Recommend(ctx Context, userId UUID, context *EngagementContext) ([]Recommendation, error)
func (r *RecommendationsEngine) Debug(ctx Context, userId UUID) (*DebugResult, error)
```

## Что НЕ меняется

- `auth/` — JWT, tenant resolver — тонкий, без изменений
- `user/` — CRUD profile — тонкий, без изменений
- `consent/` — ПДн lifecycle — тонкий, без изменений
- Количество бинарников: 2 (server + worker) — без изменений
- NATS subjects — без изменений
- Redis структура — без изменений

## Зависимости

- **T0501** (M05-унификация-энгейджмента) — предшествующая задача. Без M05 нет унифицированной модели Engagement для разбиения.

## Файлы-мишени

Каждый файл требует обновления для согласованности с новой структурой пакетов.

### Архитектура (3 файла)
- `архитектура/модули.md` → полная перерисовка структуры Platform: 4 → 7 внутренних модулей
- `архитектура/README.md` → nav-ссылки: новый документ `пакеты-platform.md`
- `архитектура/engagement.md` → ссылки на пакеты (не на плоский engagement/)

### Спецификация (2 файла)
- `спецификация/api.md` → модули-таблица: engagement/ → 3 пакета
- `спецификация/критерии-приёмки.md` → критерии по пакетам (test isolation)

### Контекст (2 файла)
- `контекст/акторы.md` → «Платформа льгот» раздел: обновление структуры сервисов
- `контекст/настраиваемость.md` → нет прямых изменений (бизнес-логика та же)

### Документация (3 файла)
- `архитектура/пакеты-platform.md` — **НОВЫЙ** — детальное описание 7 пакетов Platform
- `README.md` (doc root) → обновление «Сервисов: 4» → уточнение структуры
- `задачи/README.md` → добавить веху M06

### Задачи (2 файла)
- `задачи/README.md` → M06
- `задачи/M06-разбиение-platform/T0601/plan.yaml` + `report.md` + `brief.md` (этот файл)

### АDR (1 файл)
- `архитектура/adr/013-пакеты-platform.md` — **НОВЫЙ** — ХАДД: почему пакеты, не сервисы

## Критерии приёмки

1. [ ] Документ `архитектура/пакеты-platform.md` создан (~300 строк): детальное описание 7 пакетов, package dependencies, public API каждого пакета, DI граф
2. [ ] ADR-013 создан: «Почему internal пакеты, не gRPC микросервисы» (ХАДД формат)
3. [ ] `архитектура/модули.md` полностью перерисован:
4. [ ] `архитектура/README.md` → новая ссылка в nav-table
5. [ ] `архитектура/engagement.md` → все ссылки обновлены (engagement/ → internal/engagement/)
6. [ ] `спецификация/api.md` → таблица модулей: старая секция "Platform — основной API-сервис → Модули" → обновлена на 7 пакетов
7. [ ] `спецификация/критерии-приёмки.md` → добавлены критерии test isolation по пакетам
8. [ ] `контекст/акторы.md` → «Платформа льгот» раздел обновлен (7 пакетов вместо 4)
9. [ ] `doc/README.md` → метрическая таблица не меняется (сервисов всё ещё 4) — проверить согласованность
10. [ ] `архитектура/интеграции.md` → нет прямых изменений (Integrations — отдельный сервис) — проверить на согласованность ссылок
11. [ ] `архитектура/безопасность.md` → нет прямых изменений — проверить согласованность ссылок
12. [ ] `архитектура/стек.md` → нет прямых изменений — проверить согласованность
13. [ ] `архитектура/биллинг-движок.md` → нет прямых изменений — проверить согласованность (Billing — отдельный сервис)
14. [ ] `спецификация/артефакты.md` → 27 артефактов: проверить что все ссылки на Platform корректны
15. [ ] Journeys: проверить что все journeys согласованы (разбиение на пакеты не меняет user-facing поведения)
16. [ ] `задачи/README.md` → M06 добавлена в таблицу вех
