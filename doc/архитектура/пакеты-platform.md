# Внутренние пакеты lkfl-server — детальное описание

## TL;DR (для агентов)

> Этот файл — **полный публичный API всех 16 internal-пакетов** lkfl-server + структура integration-proxy. 1559 строк.
> - **Общий обзор структуры проекта** → `модули.md`
> - **Сводная таблица всех пакетов (один взгляд)** → строка 1547
> - **DI граф** → строка 1356
> - **Asynq workers** → строка 1424
> - **Пакет `billing/`** → строка 1021 | **`cel/`** → строка 318 | **`engagement/`** → строка 553
> - **Пакет `user/` + HR Sync** → строка 197 | **`integrationclient/`** → строка 1133 | **`payments/`** → строка 1203
> - **`recommendations/` — stub, M15 отложен** → строка 890. Не читать для новой разработки.

> **M12:** Platform → lkfl-server (единый бинарник). 11 пакетов → 15 бизнес-пакетов + api/ router. Добавили billing/, integrations/, payments/ (бывшие отдельные сервисы) + content/ (FAQ, баннеры) + recommendations/ (stub, M15). [ADR-024](./adr/024-modular-monolith.md)

> **M16:** `internal/integrations/` → вынесен в отдельный бинарник `lkfl-integration-proxy`. В монолите остался `internal/integrationclient/` — typed gRPC client к proxy. [ADR-035](./adr/035-integration-proxy.md)

> **Связь:** `архитектура/модули.md` → lkfl-server (16 пакетов после M16: 14 business + tenant + api + integrationclient). `ADR-013` → почему пакеты, а не микросервисы. `ADR-014` → eligibility → собственный пакет. `ADR-015` → compliance → собственный пакет. `ADR-035` → integration proxy (I/O isolation).
> `задачи/M06` → T0601. `задачи/M07` → T0701–T0709. `задачи/M12` → T1201–T1207. `задачи/M16` → T1601–T1604.

---

## Содержание

| Раздел | Строка |
|--------|--------|
| История: до M06 → M16 | 26–183 |
| Пакет `auth/` | 191 |
| Пакет `user/` + HR Sync | 242 |
| Пакет `consent/` | 290 |
| Пакет `compliance/` | 319 |
| Пакет `cel/` | 363 |
| `shared/pkg/celcontext` | 429 |
| Пакет `llm/` | 502 |
| Пакет `eligibility/` | 555 |
| Модуль `engagement/` (4 подпакета) | 598 |
| Пакет `notification/` | 794 |
| Пакет `recommendations/` (stub) | 935 |
| Пакет `gamification/` | 981 |
| Пакет `billing/` | 1066 |
| Пакет `integrationclient/` | 1133 |
| `integration-proxy/` (отдельный бинарник) | 1203 |
| Пакет `payments/` | 1273 |
| Пакет `content/` | 1298 |
| Пакет `api/` | 1331 |
| DI граф | 1442 |
| Asynq workers mapping | 1510 |
| Что НЕ меняется | 1530 |
| Почему пакеты, а не gRPC | 1551 |
| Миграционный план | 1589 |
| Сводная таблица пакетов | 1627 |

---

> **M10 T1002:** LLM Proxy слит как `internal/llm/`.
> **M11:** hr-sync → `internal/user/`, 1C → `internal/billing/payroll/`.

---

## Контекст до M06

До вехи M06 Platform-сервис содержал 4 внутренних модуля:

| Модуль | Обязанности | Проблема |
|--------|-----------|-|--|
| `app/` | Инициализация, DI, graceful shutdown | Ок — тонкий модуль |
| `auth/` | Keycloak OIDC, JWT verification | Ок — тонкий модуль |
| `user/` | CRUD пользователей | Ок — тонкий модуль |
| `engagement/` | Каталог + eligibility + flow execution + collections + billing events | **God Object** — 5 обязанностей в одном пакете → разделено на catalog/, flow/, collections/ |

Помимо этого:
- **notification-логика** рассыпана между worker `notification-send` и api-handlers — нет выделенного пакета
- **recommendations** — 5 admin endpoints + 1 user endpoint + 1 debug endpoint — всё перемешано в `api/`
- **consent** — упоминается в зависимостях, но не выделен как отдельный пакет

---

## Контекст M07 — 9 пакетов

После M07 eligibility engine вынесен из `engagement/`, а compliance выделен из `consent/`:

| Изменение | Описание |
|---|-|-|-|
| `eligibility/` | **M07 НОВЫЙ (T0701)** — EligibilityEngine: Check, EvaluateRule, EvaluateGroup |
| `compliance/` | **M07 НОВЫЙ (T0702)** — CascadeRevoke, AuditTrail, EnforceRetention |
| `engagement/` | Уменьшен: eligibility.go удалён, оставлен catalog/ + flow/ + collections/ + billing_events (разделено на подпакеты) |
| `consent/` | Уменьшен: CascadeRevoke удалён, оставлен lifecycle только (Grant, Revoke, List, Check) |
| Count | 7 → 9 internal packages |

## АDR-021 — 10 пакетов (CEL Engine)

Добавлен пакет `cel/` — единый движок бизнес-логики на базе Google CEL. Используется `eligibility/`, `engagement/flow/`, `recommendations/` и Billing Rule Engine.

| Изменение | Описание |
|---|-|-|-|
| `cel/` | **ADR-021 НОВЫЙ** — CELGenerator, CELEvaluator, CELValidator, CELContext |
| `eligibility/` | EvaluateRule, EvaluateGroup → EvaluateCEL (CEL expression вместо AND/OR/Groups) |
| `engagement/flow/` | flow condition_check → CEL evaluation через cel/ |
| `recommendations/` | segment matching → CEL evaluation через cel/ |
| Count | 9 → 10 internal packages |

## M09 — 11 пакетов (Gamification)

Добавлен пакет `gamification/` — система достижений, уровней лояльности, триггеров присвоения, XLSX-импорта (ADR-023).

| Изменение | Описание |
|---|-|-|-|
| `gamification/` | **M09 НОВЫЙ (T0901)** — AchievementEngine, GrantEngine, LoyaltyEngine, TriggerHandler, XLSX Import |
| `cel/` | CELContext расширен блоком `game.*` (achievements, engagement_count, loyalty_level и др.) |
| `engagement/flow/` | FlowEngine.Complete() вызывает TriggerHandler.OnEngagementCompleted() как Go-callback |
| Count | 10 → 11 internal packages |

---

## Архитектура после M06+M07+ADR021+M09 — 11 пакетов

```
platform/
├── cmd/
│   ├── server/main.go         # HTTP entry — один бинарник
│   └── worker/main.go         # Asynq entry — один бинарник
├── internal/
│   ├── auth/                  # JWT, tenant resolver, RBAC
│   ├── user/                  # CRUD пользователей, профиль
│   ├── consent/               # ПДн lifecycle, шаблоны согласий
│   ├── cel/                   # ← ADR-021: CEL Rule Engine (LLM generation + cel-go evaluation)
│   ├── eligibility/           # ← M07 T0701: EligibilityEngine (CEL-based, ADR-021)
│   ├── compliance/            # ← M07 T0702: CascadeRevoke, AuditTrail, Retention (был в consent/)
│   ├── engagement/            # ← 4 подпакета: catalog/, flow/, collections/, survey/ (condition_expr → CEL, ADR-021)
│   ├── notification/          # ← НОВЫЙ: шаблоны, каналы, очередь, persistence
│   ├── recommendations/       # ← НОВЫЙ: правила, evaluation, debug (segments → CEL, ADR-021)
│   ├── gamification/          # ← M09: Ачивки, loyalty (ADR-023)
│   └── api/                   # thin HTTP handlers → делегируют в internal/* + CEL generation endpoint
├── pkg/                       # общие типы (Claims, Tenant, Pagination)
├── migrations/                # Atlas SQL migrations
├── go.mod
└── Dockerfile
```

> **ADR-021:** `cel/` — кросс-функциональный пакет. Зависимости: `eligibility/` → `cel/`, `engagement/flow/` → `cel/`, `recommendations/` → `cel/`, `api/` → `cel/` (CEL generation endpoint).
> **ADR-023:** `gamification/` — 5-й CEL домен.

---

## M12 — 17 пакетов (Modular Monolith)

> **M12:** Переход на modular monolith ([ADR-024](./adr/024-modular-monolith.md)). Бывшие отдельные сервисы billing, integrations (Provider Gateway), payment-gateway → internal пакеты. NATS → Go interfaces.

| Изменение | Описание |
|---|-|-|-|
| `billing/` | **M12 НОВЫЙ** — Баланс, транзакции (Credit/Debit), периоды, правила (CEL), payroll/1C (был отдельный сервис) |
| `integrations/` | **M12 НОВЫЙ** — Provider Gateway (benefit-providers), webhook handler (был отдельный сервис) |
| `payments/` | **M12 НОВЫЙ** — PCI DSS платежи: авторизация, подтверждение, отмена (был отдельный сервис) |
| `engagement/flow/` | Вместо NATS → direct call `billing.BillingService.Credit(ctx, ...)`, `integrations.ProviderGateway.Activate(ctx, ...)` (**M16:** → `integrationclient.IntegrationClient.Activate(ctx, ...)`) |
| Count | 11 → 14 internal business packages (+ api/ router) |

```
backend/
├── cmd/
│   ├── server/main.go         # HTTP entry — один бинарник lkfl-server
│   └── worker/main.go         # Asynq entry
├── internal/
│   ├── auth/                  # JWT, tenant resolver, RBAC
│   ├── user/                  # CRUD пользователей, профиль, HR sync
│   ├── consent/               # ПДн lifecycle
│   ├── cel/                   # CEL Rule Engine
│   ├── llm/                   # LLM engine (in-process)
│   ├── eligibility/           # EligibilityEngine (CEL-based)
│   ├── compliance/            # CascadeRevoke, AuditTrail, Retention
│   ├── engagement/            # 4 подпакета: catalog/, flow/, collections/, survey/
│   ├── notification/          # Шаблоны, каналы, persistence
│   ├── gamification/          # Ачивки, loyalty
│   ├── billing/               # ← M12: баланс, транзакции, правила, периоды
│   ├── integrations/          # ← M12: provider gateway, вебхуки
│   ├── payments/              # ← M12: PCI DSS авторизация, подтверждение
│   ├── content/               # FAQ, баннеры, описания карточек
│   └── api/                   # Public + Admin router'а
├── shared/pkg/               # общие типы
│   ├── auth/                 # verifier, middleware, rbac
│   └── celcontext/           # CELContext type
├── pkg/                      # общие типы Platform
├── migrations/               # Atlas SQL → объединённые
└── go.mod                    # один go.mod
```

---

## M16 — 16 пакетов (Integration Proxy)

> **M16:** `internal/integrations/` вынесен в отдельный бинарник `lkfl-integration-proxy` ([ADR-035](./adr/035-integration-proxy.md)). В монолите остался `internal/integrationclient/` — typed gRPC client к proxy. Fault isolation, credential isolation, goroutine safety.

| Изменение | Описание |
|---|-|-|-|
| `integrationclient/` | **M16 НОВЫЙ** — typed gRPC client к `lkfl-integration-proxy` (localhost:8090). Interface `IntegrationService` для mock в тестах |
| `integrations/` | **M16 УДАЛЁН** — вынесен в `integration-proxy/` (отдельный бинарник) |
| `engagement/flow/` | Вместо direct call `integrations.ProviderGateway.Activate()` → gRPC `integrationclient.IntegrationClient.Activate()` |
| `api/` (Public) | `PublicHandlerDeps.Integrations` → `PublicHandlerDeps.IntegrationClient` |
| Count | 17 → 16 internal packages (монолит) + integration-proxy (отдельный бинарник) |

```
lkfl/
├── cmd/
│   ├── server/main.go              # lkfl-server (:8080) — бизнес-логика
│   ├── worker/main.go              # Asynq worker (тот же бинарник server)
│   └── integration-proxy/main.go   # lkfl-integration-proxy (:8090 gRPC)
├── internal/
│   ├── auth/                       # JWT, tenant resolver, RBAC
│   ├── user/                       # CRUD пользователей, профиль, HR sync
│   ├── consent/                    # ПДн lifecycle
│   ├── cel/                        # CEL Rule Engine
│   ├── llm/                        # LLM engine (in-process)
│   ├── eligibility/                # EligibilityEngine (CEL-based)
│   ├── compliance/                 # CascadeRevoke, AuditTrail, Retention
│   ├── engagement/                 # 4 подпакета: catalog/, flow/, collections/, survey/
│   ├── notification/              # Шаблоны, каналы, persistence
│   ├── gamification/              # Ачивки, loyalty
│   ├── billing/                   # ← M12: баланс, транзакции, правила, периоды
│   ├── integrationclient/         # ← M16: gRPC client к proxy (замена integrations/)
│   ├── payments/                  # ← M12: PCI DSS авторизация, подтверждение
│   ├── content/                   # FAQ, баннеры, описания карточек
│   └── api/                       # Public + Admin router'а
├── integration-proxy/             # ← M16: отдельный бинарник
│   ├── adapters/                  # ProviderAdapter реализации (11 провайдеров)
│   ├── circuitbreaker/            # Circuit breaker per provider
│   ├── webhook/                   # Webhook receiver + verifier
│   ├── grpc/                      # gRPC server + generated code
│   └── config/                    # YAML config загрузка
├── proto/
│   └── integration/v1/
│       └── integration.proto      # gRPC service definition
├── provider-configs/              # YAML конфиги провайдеров
├── shared/pkg/
│   ├── auth/                      # verifier, middleware, rbac
│   └── celcontext/                # CELContext type
├── pkg/                           # общие типы Platform
├── migrations/                    # Atlas SQL → объединённые
└── go.mod                         # один go.mod (два бинарника)
```

> **Важно:** один `go.mod` сохраняется (ADR-024 требование). Два бинарника — два `cmd/`, один модуль.

---

## Пакет `internal/auth/` — M10 T1005: thin tenant wrapper over shared/pkg/auth

**Назначение.** OIDC-клиент Keycloak: верификация JWT, tenant resolution, RBAC middleware. Core logic (verifier, middleware, rbac) вынесен в `shared/pkg/auth` (M10 T1005).

> **M10 T1005:** `shared/pkg/auth` — общий Go-пакет (verifier.go, middleware.go, rbac.go). payment-gateway/internal/auth/ и platform/internal/auth/ используют shared/pkg/auth напрямую. Platform auth/ остаётся как thin wrapper для tenant-specific config.

### Структура

| Файл | Назначение |
|--------|---------|
| `auth.go` | `OIDCVerifier` — `VerifyToken(ctx, token) (*Claims, error)` → делегирует в `shared/pkg/auth` |
| `middleware.go` | `JWTMiddleware`, `TenantResolver`, `RBACGuard` → делегирует в `shared/pkg/auth` |
| `cache.go` | Redis cache для JWT результатов (`jwt:` prefix, TTL = lifetime токена) |

### Публичный API

```go
type OIDCVerifier struct { /* DI: keycloak URL, Redis, logger, shared/pkg/auth */ }

func (v *OIDCVerifier) VerifyToken(ctx context.Context, token string) (*Claims, error)

func JWTMiddleware(v *OIDCVerifier) func(http.Handler) http.Handler
func TenantResolver() func(http.Handler) http.Handler
func RBACGuard(requiredRoles ...string) func(http.Handler) http.Handler
```

### shared/pkg/auth (M10 T1005)

```
shared/
└── pkg/
    └── auth/
        ├── verifier.go     # OIDC verifier, JWT validation
        ├── middleware.go   # JWTMiddleware, TenantResolver
        └── rbac.go         # RBACGuard role check
```

| Файл | Назначение | Используется |
|------|--|--|--|
| `verifier.go` | OIDC verification + JWT validation | platform + payment-gateway |
| `middleware.go` | JWTMiddleware + TenantResolver | platform + payment-gateway |
| `rbac.go` | RBACGuard role check | platform + payment-gateway |

### Зависимости

- `shared/pkg/auth` — core OIDC verification + middleware + RBAC (M10 T1005)
- `pkg/` — типы Claims, Tenant
- Redis (`jwt:` prefix) — JWT verification cache

---

## Пакет `internal/user/` — M11 T1102: добавлен HR Sync

**Назначение.** CRUD пользователей, расширенный профиль (грейд, стаж, отдел), статусы. **M11 T1102:** HR-sync перенесён из Integrations Hub в этот пакет — HR-реестр ближе к user domain, независим от отказов провайдеров льгот.

### Структура

| Файл | Назначение |
|--|--|-|
| `user.go` | `UserRepository` — Create, Get, Update, List, Activate, Deactivate |
| `profile.go` | Расширенный профиль: grade, yearsOfService, department, hasChildren |
| `registry.go` | Импорт реестра (XLSX + API), валидация, дедупликация |
| `hr_sync.go` | **M11 T1102 НОВЫЙ:** `HRSync.PullRegistry()`, `SyncStatus()` — daily HR pull (был `integrations/hr-sync/`) |

### Публичный API

```go
type UserRepository struct { /* DI: db, consent, logger */ }

func (r *UserRepository) Create(ctx context.Context, in *UserInput) (*User, error)
func (r *UserRepository) Get(ctx context.Context, id uuid.UUID) (*User, error)
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, in *UserInput) error
func (r *UserRepository) List(ctx context.Context, q *ListQuery) ([]User, Pagination, error)
func (r *UserRepository) Activate(ctx context.Context, id uuid.UUID) error
func (r *UserRepository) Deactivate(ctx context.Context, id uuid.UUID) error

type RegistryImporter struct { /* DI: userRepo, validator, logger */ }

func (i *RegistryImporter) ImportXLSX(ctx context.Context, file io.Reader) (*ImportResult, error)
func (i *RegistryImporter) ImportAPI(ctx context.Context, hrClient HRClient) (*ImportResult, error)

type HRSync struct { /* DI: userRepo, hrClient (REST), logger */ }

// **M11 T1102 НОВЫЙ:** PullRegistry — ежедневный pull кадрового реестра из HR-системы.
// Раньше: Platform → NATS `integration.hr.pull` → Integrations → HR API → NATS `integration.hr.synced` → Platform.
// Теперь: Platform Asynq `hr-sync-daily` → user.HRSync.PullRegistry(ctx) → HR-система REST (direct).
func (s *HRSync) PullRegistry(ctx context.Context) (*SyncResult, error)
func (s *HRSync) SyncStatus(ctx context.Context) (*SyncStatus, error)
```

### Зависимости

- `db/` — PostgreSQL
- `consent/` — каскадный отзыв при деактивации
- `pkg/` — общие типы
- **M11 T1102:** HR-система REST API client (прямой вызов, через NATS не идёт)

---

## Пакет `internal/consent/` — M07 обновлён

**Назначение.** Lifecycle согласий ПДн: создание → действие → продление → отзыв. CascadeRevoke вынесен в `compliance/` (T0702).

### Структура

| Файл | Назначение |
|--|--|
| `consent.go` | `ConsentEngine` — Grant, Revoke, List, CheckGranted |
| `templates.go` | Шаблоны согласий (Go templates) |

### Публичный API

```go
type ConsentEngine struct { /* DI: db, templates, logger */ }

func (e *ConsentEngine) Grant(ctx context.Context, userId uuid.UUID, provider string, scope string, method string) (*Consent, error)
func (e *ConsentEngine) Revoke(ctx context.Context, userId uuid.UUID, consentId uuid.UUID) error
func (e *ConsentEngine) List(ctx context.Context, userId uuid.UUID) ([]Consent, error)
func (e *ConsentEngine) CheckGranted(ctx context.Context, userId uuid.UUID, provider string) (bool, error)
```

### Зависимости

- `db/` — PostgreSQL
- `pkg/` — общие типы

---

## Пакет `internal/compliance/` — M07 НОВЫЙ

**Назначение.** Compliance enforcement: каскадный отзыв при удалении ПДн, audit trail для ФСТЭК, data retention policies. Вынесен из `consent/` в M07 (T0702, ADR-015).

**Почему отдельный пакет:** CascadeRevoke затрагивает 3 домена — engagement (деактивация льгот), notification (информирование), audit trail (ФСТЭК-логирование). Смешивать grant/revoke с cascade delete нарушает SRP.

### Структура

| Файл | Назначение |
|--|--|
| `cascade.go` | `ComplianceEngine.CascadeRevoke` — delete all user data |
| `audit.go` | Audit trail для ФСТЭК: кто/когда/что согласился, удалил |
| `retention.go` | Data retention enforcement (3 года, 5 лет политики) |

### Публичный API

```go
type ComplianceEngine struct {
    userRepo    *user.UserRepository
    flowEngine  *engagement.FlowEngine
    notification *notification.NotificationEngine
    db          *sql.DB
    logger      Logger
}

func (e *ComplianceEngine) CascadeRevoke(ctx context.Context, userId uuid.UUID) error
// Каскадный отзыв: все consent → деактивация льгот → уведомление → audit log

func (e *ComplianceEngine) AuditTrail(ctx context.Context, userId uuid.UUID) ([]ComplianceEvent, error)
// Получить все compliance-события для пользователя

func (e *ComplianceEngine) EnforceRetention(ctx context.Context) error
// Применить политики retention: что архивировать, что удалять
```

### Зависимости

- `user/` — для деактивации пользователя
- `engagement/flow/` — для деактивации всех льгот
- `notification/` — для информирования о результате каскадного удаления
- `db/` — PostgreSQL (audit trail storage)

---

## Пакет `internal/cel/` — ADR-021 + M10 T1003 shared CELContext

**Назначение.** CEL (Common Expression Language) Rule Engine — единый движок бизнес-логии. Заменяет 4 независимых механизма условий: billing YAML-array, eligibility AND/OR/Groups, flow condition_expr, recommendations JSON segments. LLM генерирует CEL из русского текста через `internal/llm/` (M10 T1002). **P1:** TagResolver — вычисление тегов пользователя из профиля.

> **M10 T1003:** `CELContext` type вынесен в `shared/pkg/celcontext` — common Go package. Platform и Billing импортируют одну и ту же CELContext struct → нет дублирования типов, нет deserialization mismatch между сервисами. Platform `internal/cel/context.go` теперь import-ит shared.

> **Детальное описание:** [`cel-engine.md`](./cel-engine.md) — schema, LLM integration, 3 фазы миграции, **P1:** dry-run + audit trail.
> **Детальное описание тегов:** [`теги.md`](./теги.md) — P1: TagResolver, каталог тегов, кэширование.

**Почему отдельный пакет:** CEL evaluation вызывается из 4 доменов: `eligibility/` (проверка доступа), `engagement/flow/` (flow condition_check), `recommendations/` (segment matching), `billing/rule_engine/` (filter rules). CEL-движок — чистая логика: парсит expression, компилирует через cel-go, кэширует в Redis (`cel:` prefix), evaluates в sandbox.

### Структура

| Файл | Назначение |
|------|-----------|
| `generator.go` | `CELGenerator.Generate(ctx, sourceText)` → `llm/` (in-process, M10 T1002) → CEL string |
| `evaluator.go` | `CELEvaluator.Evaluate(ctx, celExpr, context)` → cel-go interpreter |
| `validator.go` | `CELValidator.Validate(celExpr)`, `Compile(celExpr) → *cel.Program` |
| `context.go` | `CELContext` — единая схема контекста (nested: user.*, benefit.*, date.*, context.*) |
| `tag_resolver.go` | **P1:** `TagResolver.Resolve(ctx, user)` → вычисление тегов из профиля + Redis cache |
| `schema.go` | `CELSchema` — определение доступных полей и типов для cel-go environment |
| `functions.go` | Custom CEL functions: date_diff_days, str_contains, now_iso |

### Публичный API

```go
type CELGenerator struct {
    llm       *llm.LLMClient  // ← M10 T1002: in-process, был LLMProxyClient
    validator *CELValidator
    logger    Logging
}

func (g *CELGenerator) Generate(ctx context.Context, sourceText string) (cel string, err error)

type CELEvaluator struct {
    celEnv  *cel.Env
    cache   *redis.Client // Redis `cel:` prefix — compiled programs (TTL 24h)
}

// Evaluate CEL expression в контексте. Результат кэширует по hash(celExpr).
func (e *CELEvaluator) Evaluate(ctx context.Context, celExpr string, c *CELContext) (bool, error)

// Compile CEL expression → reusable program (для batch evaluation)
func (e *CELEvaluator) Compile(celExpr string) (*cel.Program, error)

type CELValidator struct {
    celEnv *cel.Env
}

func (v *CELValidator) Validate(celExpr string) error

// CELContext — M10 T1003: import из shared/pkg/celcontext (был inline)
import "lkfl/shared/pkg/celcontext"

type CELContext = celcontext.CELContext
```

### Зависимости

- `pkg/` — общие типы (Claims, Tenant)
- `internal/llm/` — M10 T1002: in-process LLM engine (вместо HTTP к LLM Proxy :8085)
- `shared/pkg/celcontext` — **M10 T1003:** CELContext type (common Go package, platform ↔ billing)
- Redis (`cel:` prefix) — кэш compiled CEL programs (TTL 24h) + TagResolver cache (TTL 1h) + LLM agent configs

---

## shared/pkg/celcontext — M10 T1003: CELContext type (platform ↔ billing)

**Назначение.** Common Go package с определением `CELContext` struct. Устраняет дублирование между `platform/internal/cel/context.go` и `billing/rule_engine/celcontext.go`.

> **M10 T1003:** Проблема: Platform и Billing имеют свои CELContext types → deserialization mismatch (Platform сериализует user.grade="Senior", Billing десериализует → nil field). Shared package → одна struct, 0 mismatch.

### Структура

```
shared/
└── pkg/
    ├── auth/            ← M10 T1005
    │   ├── verifier.go
    │   ├── middleware.go
    │   └── rbac.go
    └── celcontext/      ← M10 T1003
        ├── context.go   # CELContext struct (user, benefit, date, tags, ...)
        └── types.go     # CELField, CELSchema types
```

### Публичный API

```go
// context.go — единственное определение CELContext
package celcontext

type CELContext struct {
    User struct {
        Grade          string // user.grade
        YearsOfService int    // user.years_of_service
        HasChildren    bool   // user.has_children
        Department     string // user.department
        Status         string // user.status
        TenantID       string // user.tenant_id
        UserID         string // user.user_id
    }
    Tags    map[string]any
    Benefit struct {
        Category string
        Cost     float64
    }
    Date    struct{ Today string }
    Context map[string]any
    Answers map[string]any
    Events  map[string]any
    Period struct {
        Start string
        End   string
    }
    Balance struct{ Total float64 }
    // M09 ADR-023: Gamification domain (5-й CEL домен)
    Game struct {
        Achievements           []string         // ключи имеющихся ачивок
        AchievementCount       int              // количество ачивок
        EngagementCount        int              // всего завершённых энгейджментов
        EngagementByCategory   map[string]int   // по категориям
        BenefitCategoriesCount int              // кол-во категорий льгот
        LoyaltyLevel           string           // текущий уровень
        LoyaltyPoints          float64          // cumulative engagement points
        DaysSinceActive        int              // дней с последней активности
        EnpsSubmitted          bool
        HasFamily              bool             // есть родственники в ДМС
    }
}
```

### Зависимости (shared/pkg/celcontext)

- **Ноль зависимостей** — чистый Go-пакет, только стандартная библиотека
- Используется: `platform/internal/cel/`, `billing/rule_engine/`

---

## Пакет `internal/llm/` — M10 T1002 (бывший LLM Proxy separate service)

**Назначение.** In-process LLM engine: agent routing, prompt management, LLM provider clients (ollama, openai), cost tracking, audit trail. Был отдельный сервис `:8085` (ADR-022) → слит в Platform как `internal/llm/` (M10 T1002).

**Почему merge:** 1 активный агент (`cel-generator`). 3 future agents (moderation, analytics, personalization) — Phase 2. Overhead отдельного сервиса не оправдан при 1 агенте.

### Структура

| Файл | Назначение |
|------|-----------|
| `client.go` | `LLMClient` — in-process client (direct call от cel/generator.go) |
| `router.go` | `AgentRouter` — agent → model + prompt config (YAML) |
| `providers.go` | OllamaClient, OpenAIClient — адаптеры LLM провайдеров |
| `audit.go` | LLMAudit — request logging, cost tracking, token usage |

### Публичный API

```go
type LLMClient struct {
    router  *AgentRouter
    audit   *LLMAudit
    logger  Logger
}

// GenerateCEL — direct in-process call от cel/generator.go
func (c *LLMClient) GenerateCel(ctx context.Context, source string, agent string) (cel string, err error)

type AgentRouter struct {
    configs map[string]*AgentConfig
}

// Lookup agent config по имени агента
func (r *AgentRouter) Route(agent string) (*AgentConfig, error)

type LLMAudit struct {
    db    *sql.DB
    redis *redis.Client
}

// Log request + response для audit trail (ФСТЭК) + cost tracking
func (a *LLMAudit) Log(ctx context.Context, agent string, tenant string, tokensIn int, tokensOut int, latencyMs int) error
```

### Зависимости

- `pkg/` — общие типы
- Redis — agent config cache + rate limiting (`cel:` prefix)
- PostgreSQL (`lkfl_platform`) — audit log (был `lkfl_llm`, merged)
- Ollama API, OpenAI API — LLM provider endpoints
- **M12:** direct call из `internal/cel/` (был NATS `llm.generate`/`llm.result`)

---

## Пакет `internal/eligibility/` — M07 + ADR-021

**Назначение.** Eligibility engine: CEL evaluation (ADR-021). Заменяет AND/OR/Groups → CEL expression.

**Почему отдельный пакет:** Eligibility engine вызывается из 3 мест: engagement-flow (свой пакет), recommendations (другой пакет), billing rule engine (третий пакет). Значит eligibility НЕ должна быть подпакетом engagement.

### Структура

| Файл | Назначение |
|------|-----------|
| `engine.go` | `EligibilityEngine` — Check, EvaluateCEL |
| `types.go` | Типы: EligibilityResult, CELContext для eligibility |

### Публичный API

```go
type EligibilityEngine struct {
    db        *pgxpool.Pool
    userRepo  UserRepo
    celEval   *cel.CELEvaluator  // ← ADR-021: CEL evaluator
    logger    Logging
}

// Check eligibility для offer — загружает eligibility_cel из DB, evaluates через CEL
func (e *EligibilityEngine) Check(ctx context.Context, offerId uuid.UUID, userId uuid.UUID) (*EligibilityResult, error)

// EvaluateCEL — прямой вызов CEL evaluation (ADR-021)
func (e *EligibilityEngine) EvaluateCEL(ctx context.Context, celExpr string, context *cel.CELContext) (bool, error)

type EligibilityResult struct {
    Eligible bool
    Reasons  []string
}
```

### Зависимости

- `pkg/` — общие типы (Claims, Tenant)
- `user/` — для profile данных при evaluation
- `cel/` — **ADR-021:** CEL evaluation

---

## Модуль `internal/engagement/` — разделение на 4 подпакета

**Назначение.** Единая бизнес-логика энгейджментов, разделённая на 4 подпакета для предотвращения God Object.
Eligibility engine вынесен в `eligibility/` (T0701).

```
internal/engagement/
├── catalog/         # EngagementType, Offer, Category — CRUD, фильтры, поиск, кэш
├── flow/            # EngagementFlow, UserEngagement — Activate, Complete, Revert, ExecuteStep
├── collections/     # EngagementCollection — bundle создание, массовое подключение
└── survey/          # Survey Engine — бранчинг, TagMapper, analytics (M13/M14)
```

### Подпакет `internal/engagement/catalog/`

**Назначение.** Каталог энгейджментов: EngagementType, EngagementOffer, EngagementCategory. CRUD, фильтры, поиск, Redis кэш. A/B тесты, provider proposals.

### Структура

| Файл | Назначение |
|--------|----------|
| `types_engine.go` | EngagementType CRUD |
| `offer_engine.go` | EngagementOffer CRUD |
| `category_engine.go` | EngagementCategory CRUD |
| `search.go` | Поиск, фильтры, кэш (Redis `catalog:`) |
| `ab_test.go` | A/B тесты на карточках каталога |
| `proposals.go` | Provider proposals |

### Публичный API

```go
type CatalogService struct { /* DI: db, cache, logger */ }

func (s *CatalogService) List(ctx context.Context, q *CatalogQuery) ([]Engagement, Pagination, error)
func (s *CatalogService) Get(ctx context.Context, id uuid.UUID) (*Engagement, error)
func (s *CatalogService) FilterByType(ctx context.Context, eType string) ([]Engagement, error)
func (s *CatalogService) FilterByCategory(ctx context.Context, catId uuid.UUID) ([]Engagement, error)
func (s *CatalogService) Search(ctx context.Context, query string) ([]Engagement, error)
```

### Зависимости

- `db/` — PostgreSQL
- `pkg/` — общие типы
- Redis (`catalog:` prefix) — кэш каталога (TTL 6h)

---

### Подпакет `internal/engagement/flow/`

**Назначение.** Flow execution: EngagementFlow, UserEngagement, UserEngagementStep.
Активация, завершение, отмена, выполнение шагов. Billing events через direct call.

### Структура

| Файл | Назначение |
|--------|----------|
| `flow_engine.go` | FlowEngine: Activate, Complete, Revert, ExecuteStep |
| `user_engagement.go` | UserEngagement CRUD, status transitions |
| `billing_events.go` | **M12:** direct call billing.BillingService (debit/credit) |
| `approval.go` | Approval requests (flow step: approval) |
| `document.go` | Document generation metadata (flow step: document_generation) |

### Публичный API

```go
type FlowEngine struct { /* DI: db, billing, integrations, payments, asynq, logger, eligibility */ }

func (f *FlowEngine) Activate(ctx context.Context, offerId uuid.UUID, userId uuid.UUID) error
func (f *FlowEngine) Complete(ctx context.Context, engagementId uuid.UUID) error
func (f *FlowEngine) Revert(ctx context.Context, engagementId uuid.UUID) error
func (f *FlowEngine) ExecuteStep(ctx context.Context, engagementId uuid.UUID, stepId uuid.UUID, data map[string]string) error
func (f *FlowEngine) GetUserEngagements(ctx context.Context, userId uuid.UUID) ([]UserEngagement, error)
```

### Billing Events → Direct Call (M12: был NATS)

Каждый переход flow-статуса вызывает `billing.BillingService` напрямую:

| Transition | Billing Method | Direction | Payload |
|------------|-|-------|---------|
| `pending → in_progress` (benefit) | `DebitReserve(userId, amount, category, offerId)` | debit (заморозка) | `{ idempotency_key, userId, amount, category, offerId }` |
| `approved → active` (benefit) | `Debit(ctx, userId, amount, category, offerId)` | debit (подтверждение) | `{ idempotency_key, userId, amount, category, offerId }` |
| `in_progress → active` (benefit, instant) | `Debit(ctx, userId, amount, category, offerId)` | debit | `{ idempotency_key, userId, amount, category, offerId }` |
| `in_progress → completed` (activity) | `Credit(ctx, userId, amount, category, source, periodId)` | credit | `{ idempotency_key, userId, amount, category, source, periodId }` |
| `failed` (любой тип) | `Revert(ctx, txId)` | credit (возврат) | `{ idempotency_key, userId, amount, category, offerId }` |

> **P0:** все mutating события содержат `idempotency_key` (sha256 hash). Billing: `INSERT ... ON CONFLICT DO NOTHING`.

### Зависимости

- `db/` — PostgreSQL
- `eligibility/` — для eligibility check при activation
- `user/` — для profile данных в flow execution
- `pkg/` — общие типы
- **M12:** `billing/` — direct call `billing.BillingService.*()` (был NATS)
- **M12:** `integrations/` — direct call `integrations.ProviderGateway.Activate()` (был NATS)
- **M12:** `payments/` — direct call `payments.PaymentGateway.Authorize()` (был NATS)
- `gamification/` — Go-callback `OnEngagementCompleted()` (ADR-023)
- Redis — Asynq workers

---

### Подпакет `internal/engagement/collections/`

**Назначение.** Bundles: EngagementCollection. Создание наборов, массовое подключение офферов.

### Структура

| Файл | Назначение |
|--------|----------|
| `collection_engine.go` | CollectionsEngine: List, Get, Engage |

### Публичный API

```go
type CollectionsEngine struct { /* DI: db, flowEngine, logger */ }

func (c *CollectionsEngine) List(ctx context.Context) ([]Collection, error)
func (c *CollectionsEngine) Get(ctx context.Context, id uuid.UUID) (*Collection, error)
func (c *CollectionsEngine) Engage(ctx context.Context, collectionId uuid.UUID, userId uuid.UUID) error
```

### Зависимости

- `db/` — PostgreSQL
- `engagement/flow/` — для создания user_engagement записей при массовом подключении
- `pkg/` — общие типы
- Redis — catalog cache (TTL 6h, `catalog:` prefix)

### Подпакет `internal/engagement/survey/` — M13 T1301 (спроектирован, будет реализован в M14)

**Назначение.** Survey Engine: полноценный модуль опросов с бранчингом, tag-mapping и аналитикой.
Опросы остаются `EngagementType(type: "activity")`, но получают движок с ветвлением вопросов,
конвертацией ответов в теги и аналитикой результатов.

> **M13 T1301:** подпакет внутри engagement/ (не отдельный пакет). ADR-025.
> **Пока только архитектура.** Реальный Go-код будет создан в M14.
> Обратно совместимо: `ui_component="SurveyForm"` → survey_schema, `ui_component="EngagementForm"` → legacy form_schema.

### Структура

| Файл | Назначение |
|------|--|
| `doc.go` | Package documentation |
| `types.go` | SurveySchema, Question, QuestionType, TagMapping, SurveyTag |
| `resolver.go` | Resolver — рендер вопросов с бранчингом + state management (hydration из form_data JSONB) |
| `tag_mapper.go` | TagMapper — конвертация ответов в пользовательские теги (INSERT/UPDATE GREATEST weight, delete orphaned) |
| `analytics.go` | AnalyticsEngine — агрегация результатов (TotalResponses, CompletionRate, Distribution, TopTags) |

### Публичный API

```go
// --- Resolver ---
package survey

type Resolver struct { /* schema, answers map[string]any */ }

func NewResolverWithState(schema *SurveySchema, savedAnswers map[string]any) *Resolver
func (r *Resolver) GetNextQuestion() (*Question, bool)
func (r *Resolver) SubmitAnswer(questionId string, answer any) error
func (r *Resolver) IsComplete() bool
func (r *Resolver) GetFinalAnswers() map[string]any
func ValidateSchema(schema *SurveySchema) error

// --- TagMapper ---
type TagMapper struct { /* db, logger */ }

func NewTagMapper(db *sql.DB, logger Logger) *TagMapper
func (m *TagMapper) MapSurveyAnswers(
    ctx context.Context,
    tenantId uuid.UUID,
    userId uuid.UUID,
    surveyOfferId uuid.UUID,
    schema *SurveySchema,
    answers map[string]any,
) ([]SurveyTag, error)

// --- Analytics ---
type AnalyticsEngine struct { /* db, logger */ }

func NewAnalyticsEngine(db *sql.DB, logger Logger) *AnalyticsEngine
func (a *AnalyticsEngine) GetSurveyAnalytics(
    ctx context.Context,
    surveyOfferId uuid.UUID,
) (*SurveyAnalytics, error)
```

### Зависимости

- `db/` — PostgreSQL (user_survey_attributes + user_engagements + form_data JSONB)
- `cel/TagResolver` — InvalidateCache при MapSurveyAnswers
- **Минимальный diff в flow/flow_engine.go:** ExecuteStep при `ui_component="SurveyForm"` → hydrate resolver → SubmitAnswer → save form_data → TagMapper

---

## Пакет `internal/notification/` — НОВЫЙ

**Назначение.** Templates, каналы доставки (email/push/in-app), очередь, persistence.

**Логика вытаскивается из worker `notification-send` и api-handlers.**

### Структура

| Файл | Назначение |
|--------|----------|
| `notification.go` | `NotificationEngine` — основной API: Send, SendFromEvent |
| `channels.go` | `Channel` interface + registry |
| `email.go` | SMTP channel implementation |
| `push.go` | FCM/APNs channel implementation |
| `inapp.go` | In-app channel — запись в БД (визуализация через `/notifications`) |
| `templates.go` | Go template engine + шаблонизатор |
| `store.go` | Persistence: Create, ListUnread, MarkRead, MarkAllRead |

### Публичный API

```go
type NotificationEngine struct {
    templates  *TemplateEngine
    channels   ChannelRegistry
    store      *NotificationStore
    logger     Logger
}

// Channel interface — реализируется каждым каналом доставки
type Channel interface {
    Send(ctx context.Context, recipient string, payload *NotificationPayload) error
    Name() string
}

func (n *NotificationEngine) Send(ctx context.Context, userId uuid.UUID, templateName string, data map[string]string) error
func (n *NotificationEngine) SendFromEvent(ctx context.Context, event *EngagementEvent) error
func (n *NotificationEngine) ListUnread(ctx context.Context, userId uuid.UUID) ([]Notification, Pagination, error)
func (n *NotificationEngine) MarkRead(ctx context.Context, userId uuid.UUID, notificationId uuid.UUID) error
func (n *NotificationEngine) MarkAllRead(ctx context.Context, userId uuid.UUID) error
```

### Каналы

| Канал | Реализация | Данные |
|-----|------|----|
| `email` | SMTP + template | письмо с текстом уведомления |
| `push` | FCM/APNs | push-уведомление на мобильное |
| `inapp` | PostgreSQL (table `notifications`) | запись в БД → визуализация через API |

### Event → Template mapping

Каждое engagement-событие маппится на template:

| Event | Email | Push | In-App |
|-----|----|----|----|
| `pending → in_progress` | ✅ | — | ✅ |
| `approved → active` | ✅ | ✅ | ✅ |
| `completed` (activity) | ✅ | ✅ | ✅ |
| `expiring_soon` (30 дней) | ✅ | ✅ | ✅ |
| `expired` | ✅ | — | ✅ |
| `failed` | ✅ | ✅ | ✅ |

### Матрица предпочтений пользователя (P2)

**Проблема:** Email может улететь в spam, push может быть отключён. Пользователь должен выбирать, как его уведомлять.

**Решение:** таблица `user_notification_preferences` — настройки per user + per event type.

```sql
CREATE TABLE user_notification_preferences (
    id              SERIAL PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type      VARCHAR(50) NOT NULL,   -- 'engagement_approved', 'expiring_soon', 'marketing', ...
    email_enabled   BOOLEAN DEFAULT true,
    push_enabled    BOOLEAN DEFAULT true,
    inapp_enabled   BOOLEAN DEFAULT true,
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE (user_id, event_type)
);

-- Default preferences (копируются при регистрации пользователя)
CREATE TABLE default_notification_preferences (
    event_type      VARCHAR(50) PRIMARY KEY,
    email_enabled   BOOLEAN DEFAULT true,
    push_enabled    BOOLEAN DEFAULT true,
    inapp_enabled   BOOLEAN DEFAULT true
);

INSERT INTO default_notification_preferences VALUES
    ('engagement_approved', true, true, true),
    ('engagement_completed', true, true, true),
    ('expiring_soon', true, true, true),
    ('expired', true, false, true),
    ('failed', true, true, true),
    ('marketing', false, false, true);  -- маркетинг по умолчанию отключён
```

**Flow:** при `Send(userId, event)` → чтение preferences → фильтрация каналов → отправка только в выбранные.

```go
func (n *NotificationEngine) SendFromEvent(ctx context.Context, event *EngagementEvent) error {
    prefs := n.store.GetPreferences(ctx, event.UserId, event.Type)
    for channel := range prefs.EnabledChannels() {
        n.channels.Send(ctx, channel, event)
    }
}
```

### Шаблоны уведомлений (P3 — TODO)

| Тип шаблона | Где хранится | Кто меняет | Частота изменений |
|---|---|---|---|
| **Транзакционные** (заявка одобрена, баллы начислены) | Go templates в коде | Разработчик | Редко (связаны с API контрактом) |
| **Маркетинговые** (акция, новая льгота, reminder) | **TODO:** БД с versioning + approval workflow | Маркетолог / HR | Часто (каждая кампания) |

> **P3 TODO:** реализовать `notification_templates` таблицу для маркетинговых шаблонов:
> ```sql
> CREATE TABLE notification_templates (
>     id              SERIAL PRIMARY KEY,
>     name            VARCHAR(100) NOT NULL,
>     channel         VARCHAR(20) NOT NULL,  -- 'email', 'push'
>     subject         TEXT,                   -- для email
>     body            TEXT NOT NULL,          -- Go template string
>     variables       JSONB,                  -- список доступных переменных
>     version         INT DEFAULT 1,
>     status          VARCHAR(20) DEFAULT 'draft',  -- draft | approved | archived
>     approved_by     UUID REFERENCES users(id),
>     created_at      TIMESTAMPTZ DEFAULT now(),
>     updated_at      TIMESTAMPTZ
> );
> ```
> MVP — оставить Go templates в коде. Перенести в БД при появлении маркетинговой команды.

### Зависимости

- `db/` — notifications table + **P2:** `user_notification_preferences`
- `pkg/` — общие типы
- SMTP config, FCM/APNs config — через viper env

---

## Пакет `internal/recommendations/` — STUB (Phase 2, M10 T1001)

> **⚠️ M10 T1001:** Пакет заменён на stub. Full-реализация (5 файлов, Redis cache, PostgreSQL table) удалена — 0 user journeys зависят от endpoint'а, нет механизма обратной связи (feedback), `conversion_rate = 0` всегда. Phase 2 — после frontend integration.

**Назначение.** Stub — возвращает пустой slice. Полная рекомендательная система (context rules + segment rules + evaluation + hit counting) будет добавлена в Phase 2 после готовности frontend-интеграции.

### Структура (stub)

| Файл | Назначение |
|--------|----------|
| `engine.go` | `RecommendationsEngine` — Recommend → empty, Debug → nil |

### Публичный API (stub)

```go
type RecommendationsEngine struct {
    logger Logger
}

// Recommend — Phase 2 stub: возвращает пустой slice
func (r *RecommendationsEngine) Recommend(ctx context.Context, userId uuid.UUID, context *EngagementContext) ([]Recommendation, error) {
    return []Recommendation{}, nil
}

// Debug — Phase 2 stub: возвращает пустой результат
func (r *RecommendationsEngine) Debug(ctx context.Context, userId uuid.UUID) (*DebugResult, error) {
    return &DebugResult{
        ContextRules:    nil,
        SegmentRules:    nil,
        Recommendations: nil,
    }, nil
}
```

> **CRUD методы (CreateRule, UpdateRule, DeleteRule, ListRules)** — Phase 2 TODO, после frontend integration.

> **Phase 2 план:** двухслойная модель (context rules + segment rules), evaluation engine через CEL, hit counting + conversion rate, Redis cache (TTL 1h), PostgreSQL persistence.

### Зависимости (stub)

- `pkg/` — общие типы

> **Удалено:** `db/` (PostgreSQL), `user/` (profile), `cel/`, Redis cache — все восстановлены в Phase 2.

---

## Пакет `internal/gamification/` — M09 НОВЫЙ

**Назначение.** Система геймификации: ачивки (achievements), уровни лояльности (loyalty levels), триггеры присвоения, XLSX импорт. **ADR-023:** CEL — 5-й домен оценки.

**Почему отдельный пакет:** Геймификация — отдельный домен (не eligibility, не engagement). CEL оценивает условия присвоения, но факты присвоения — immutable records в БД. Trigger mechanism (Go-callback из engagement) требует обратного вызова без круговой зависимости.

### Структура

| Файл | Назначение |
|------|-|
| `models.go` | Типы: Achievement, AchievementGrant, LoyaltyLevelDefinition, UserLoyaltyLevel, ImportJob |
| `achievement.go` | AchievementEngine — CRUD achievement-шаблонов, GetUserAchievements, GetUserProgress |
| `grant_engine.go` | GrantEngine — CheckAndAward(ctx, userId): проверка всех CEL-условий, присвоение новых |
| `loyalty.go` | LoyaltyEngine — GetCurrentLevel, Upgrade, MonthlyCheck, CheckEligibility |
| `triggers.go` | TriggerHandler — OnEngagementCompleted, OnMonthlyCron, AwardManually, AwardByImport |
| `cel_integration.go` | BuildGamificationCELContext() — наполнение game.* полей CELContext |
| `xlsx_import.go` | XLSXImporter — валидация → preview → apply → import_jobs → error report |

### Публичный API

```go
type GamificationEngine struct {
    achievements   *AchievementEngine
    grants         *GrantEngine
    loyalty        *LoyaltyEngine
    triggers       *TriggerHandler
    xlsxImporter   *XLSXImporter
    celEvaluator   *cel.CELEvaluator  // ADR-023: CEL — 5-й домен
    db             *pgxpool.Pool
    asynqServer    *asynq.Server
    logger         Logger
}

// Achievement CRUD
func (e *GamificationEngine) CreateAchievement(ctx context.Context, in *AchievementInput) (*Achievement, error)
func (e *GamificationEngine) ListAchievements(ctx context.Context, q *AchievementListQuery) ([]Achievement, Pagination, error)
func (e *GamificationEngine) GetAchievement(ctx context.Context, id uuid.UUID) (*Achievement, error)
func (e *GamificationEngine) UpdateAchievement(ctx context.Context, id uuid.UUID, in *AchievementInput) error
func (e *GamificationEngine) DeleteAchievement(ctx context.Context, id uuid.UUID) error  // soft: grants остаются

// User-facing API
func (e *GamificationEngine) GetUserAchievements(ctx context.Context, userId uuid.UUID) ([]AchievementGrant, error)
func (e *GamificationEngine) GetUserProgress(ctx context.Context, userId uuid.UUID) ([]AchievementProgress, error)
func (e *GamificationEngine) GetCurrentLevel(ctx context.Context, userId uuid.UUID) (*UserLoyaltyLevel, error)
func (e *GamificationEngine) GetLevelHistory(ctx context.Context, userId uuid.UUID) ([]UserLoyaltyLevel, error)

// Grant engine — CEL evaluation
func (e *GamificationEngine) CheckAndAward(ctx context.Context, userId uuid.UUID) (*GrantResult, error)

// Loyalty
func (e *GamificationEngine) MonthlyCheck(ctx context.Context) error
func (e *GamificationEngine) UpgradeLevel(ctx context.Context, userId uuid.UUID, levelKey string) error

// Admin
func (e *GamificationEngine) AwardManually(ctx context.Context, userId uuid.UUID, achievementId uuid.UUID) error
func (e *GamificationEngine) ImportBatchXLSX(ctx context.Context, file io.Reader) (*ImportJob, error)
```

### Trigger contract (Go-callback)

```go
// TriggerHandler implements EngagementCompleter interface из engagement/
type TriggerHandler struct {
    grantEngine    *GrantEngine
    loyaltyEngine  *LoyaltyEngine
}

func (t *TriggerHandler) OnEngagementCompleted(ctx context.Context, engagementId uuid.UUID, userId uuid.UUID, category string) error {
    // 1. Загрузить achievement-шаблоны с trigger_on='engagement_completed'
    // 2. Для каждого: BuildGamificationCELContext(userId) → CEL evaluation → TRUE → INSERT grant
    // 3. Проверить loyalty upgrade
    // Best-effort: ошибка логируется, completion продолжается
}
```

### Зависимости

- `db/` — PostgreSQL (achievements, achievement_grants, loyalty_level_definitions, user_loyalty_levels, gamification_import_jobs)
- `user/` — профиль для CEL context построения
- `cel/` — **ADR-023:** CEL evaluation условий присвоения (5-й домен)
- `pkg/` — общие типы
- Redis (`asynq:` prefix) — Asynq workers (gamification-check-monthly, gamification-import-xlsx)

---

## Пакет `internal/billing/` — M12: бывший отдельный сервис

> **M12:** Billing-сервис (порт :8081, `billing:8081`, отдельный go.mod + Dockerfile) слит в Platform как `internal/billing/`. Контракты → Go-интерфейсы.

**Назначение.** Баланс пользователя, транзакции (credit/debit), периоды распределения, сгорание баллов, софинансирование.

**Почему internal пакет, а не отдельный сервис (M12):**
- One PG transaction: `engagement/flow/` активирует льготу и `billing/` deb'ит в рамках одной tx
- Нет NATS overhead: compile-time Go interface вместо runtime subject strings
- ACID гарантии сохраняются (одна DB, одна tx)
- ФСТЭК audit trail не требует отдельного бинарника

### Структура

| Файл | Назначение |
|--------|---------|
| `account.go` | `AccountEngine` — GetBalance, GetAccount |
| `transaction.go` | `TransactionEngine` — Credit, Debit, List, Revert |
| `period.go` | `PeriodEngine` — Activate, Expire, Burn, GetCurrent |
| `rule_engine.go` | `RuleEngine` — EvaluateRule, FilterRules (ADR-021 CEL) |
| `payroll.go` | `PayrollEngine` — SubmitPayroll, ListPayrollStatements (был `integrations/1c/`, M11 T1103) |

### Публичный API

```go
type BillingService struct {
    account    *AccountEngine
    tx         *TransactionEngine
    period     *PeriodEngine
    ruleEngine *RuleEngine
    payroll    *PayrollEngine
    db         *pgxpool.Pool
    celEval    *cel.CELEvaluator
    logger     Logger
}

// Account
func (s *BillingService) GetBalance(ctx context.Context, userId uuid.UUID) (*Balance, error)
func (s *BillingService) GetAccount(ctx context.Context, userId uuid.UUID) (*Account, error)

// Transaction
func (s *BillingService) Credit(ctx context.Context, userId uuid.UUID, amount float64, category string, source string, periodId uuid.UUID) (*Transaction, error)
func (s *BillingService) Debit(ctx context.Context, userId uuid.UUID, amount float64, category string, offerId uuid.UUID) (*Transaction, error)
func (s *BillingService) List(ctx context.Context, userId uuid.UUID, q *TxQuery) ([]Transaction, Pagination, error)
func (s *BillingService) Revert(ctx context.Context, txId uuid.UUID) error

// Period
func (s *BillingService) Activate(ctx context.Context, periodId uuid.UUID) error
func (s *BillingService) Expire(ctx context.Context, periodId uuid.UUID) error
func (s *BillingService) Burn(ctx context.Context) error
func (s *BillingService) GetCurrent(ctx context.Context) (*Period, error)

// Rule Engine (ADR-021: CEL-based)
func (s *BillingService) EvaluateRule(ctx context.Context, ruleId uuid.UUID, context *cel.CELContext) (bool, error)
func (s *BillingService) FilterRules(ctx context.Context, tenantId uuid.UUID, trigger string, context *cel.CELContext) ([]Rule, error)

// Payroll (M11 T1103: 1C integration → billing/payroll/)
func (s *BillingService) SubmitPayroll(ctx context.Context, stmt *PayrollStatement) error
func (s *BillingService) ListPayrollStatements(ctx context.Context, tenantId uuid.UUID) ([]PayrollStatement, error)
func (s *BillingService) GetPayrollStatus(ctx context.Context, id uuid.UUID) (*PayrollStatus, error)
```

### Зависимости

- `db/` — PostgreSQL (объединённая `lkfl_platform`)
- `cel/` — **ADR-021:** CEL evaluation billing rules
- `shared/pkg/celcontext` — CELContext type (common Go package, platform ↔ billing)

---

## Пакет `internal/integrationclient/` — M16: gRPC client к proxy

> **M16 (ADR-035):** `internal/integrations/` вынесен в отдельный бинарник `lkfl-integration-proxy`. В монолите остался `internal/integrationclient/` — typed gRPC client к proxy (localhost:8090). Interface `IntegrationService` для mock в тестах.

**Назначение.** Typed gRPC client к `lkfl-integration-proxy`. Обеспечивает compile-time safety (protoc генерация) и fault isolation (proxy упал → монолит работает с кэшем).

**Почему interface (не прямой вызов):** Test isolation — в тестах `engagement/flow/` подменяем `IntegrationService` на mock, не требуя запущенный proxy.

### Структура

| Файл | Назначение |
|------|-----------|
| `client.go` | `IntegrationClient` — typed gRPC client, wrapping `proto/integration/v1/` |
| `types.go` | `IntegrationService` interface (для mock), request/response типы |
| `options.go` | `DialOption` (timeout, retry, circuit breaker) |

### Публичный API

```go
// IntegrationService — interface для test isolation (mock в тестах)
type IntegrationService interface {
    // Activate — асинхронная активация льготы у провайдера
    // Возвращает job_id за < 10ms. Результат — через webhook callback.
    Activate(ctx context.Context, req *ActivateRequest) (*ActivateResponse, error)

    // Deactivate — асинхронная деактивация льготы у провайдера
    Deactivate(ctx context.Context, req *DeactivateRequest) (*DeactivateResponse, error)

    // GetProviderStatus — синхронная проверка статуса (< 5s)
    GetProviderStatus(ctx context.Context, req *ProviderStatusRequest) (*ProviderStatusResponse, error)

    // GetCatalog — синхронное чтение каталога из кэша proxy
    GetCatalog(ctx context.Context, req *CatalogRequest) (*CatalogResponse, error)

    // HealthCheck — проверка здоровья провайдера
    HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error)

    // Admin операции
    ListProviders(ctx context.Context, req *ListProvidersRequest) (*ListProvidersResponse, error)
    GetProvider(ctx context.Context, req *GetProviderRequest) (*GetProviderResponse, error)
    UpdateProvider(ctx context.Context, req *UpdateProviderRequest) (*UpdateProviderResponse, error)
    TriggerSync(ctx context.Context, req *TriggerSyncRequest) (*TriggerSyncResponse, error)
    GetSyncLogs(ctx context.Context, req *GetSyncLogsRequest) (*GetSyncLogsResponse, error)
}

// IntegrationClient — реализация IntegrationService через gRPC
type IntegrationClient struct {
    conn   *grpc.ClientConn
    client integrationv1.IntegrationServiceClient
    logger Logger
}

func NewIntegrationClient(addr string, opts ...DialOption) (*IntegrationClient, error)
func (c *IntegrationClient) Close() error

// Activate — асинхронная активация
func (c *IntegrationClient) Activate(ctx context.Context, req *ActivateRequest) (*ActivateResponse, error)

// Deactivate — асинхронная деактивация
func (c *IntegrationClient) Deactivate(ctx context.Context, req *DeactivateRequest) (*DeactivateResponse, error)

// GetProviderStatus — синхронный статус
func (c *IntegrationClient) GetProviderStatus(ctx context.Context, req *ProviderStatusRequest) (*ProviderStatusResponse, error)

// GetCatalog — синхронное чтение каталога из кэша
func (c *IntegrationClient) GetCatalog(ctx context.Context, req *CatalogRequest) (*CatalogResponse, error)

// HealthCheck — проверка здоровья
func (c *IntegrationClient) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error)

// ListProviders — список всех провайдеров
func (c *IntegrationClient) ListProviders(ctx context.Context, req *ListProvidersRequest) (*ListProvidersResponse, error)

// GetProvider — детали провайдера
func (c *IntegrationClient) GetProvider(ctx context.Context, req *GetProviderRequest) (*GetProviderResponse, error)

// UpdateProvider — обновление конфигурации провайдера
func (c *IntegrationClient) UpdateProvider(ctx context.Context, req *UpdateProviderRequest) (*UpdateProviderResponse, error)

// TriggerSync — запуск синхронизации каталога
func (c *IntegrationClient) TriggerSync(ctx context.Context, req *TriggerSyncRequest) (*TriggerSyncResponse, error)

// GetSyncLogs — лог синхронизации
func (c *IntegrationClient) GetSyncLogs(ctx context.Context, req *GetSyncLogsRequest) (*GetSyncLogsResponse, error)
```

### Зависимости

- `google.golang.org/grpc` — gRPC client
- `proto/integration/v1/` — generated protobuf types

---

## `integration-proxy/` — M16: отдельный бинарник lkfl-integration-proxy

> **M16 (ADR-035):** Gateway к внешним провайдерам льгот вынесен из монолита в отдельный бинарник. Fault isolation, credential isolation, goroutine safety, webhook isolation.

**Назначение.** Один бинарник, два listener'а:
- **gRPC на :8090** — для монолита (typed client `integrationclient/`)
- **HTTP на :8091** — для webhook'ов от внешних провайдеров

**Почему отдельный бинарник:**
- **Fault isolation** — panic в адаптере → restart proxy, не монолита
- **Credential isolation** — ключи провайдеров только в proxy (encrypted in `lkfl_integration.providers`)
- **Goroutine safety** — асинхронная активация, hot path монолита не блокируется
- **Webhook isolation** — публичные webhook endpoint'ы на proxy, не на монолите
- **Independent deploy** — новый адаптер → redeploy только proxy

### Структура

| Директория | Назначение |
|---|---|
| `adapters/` | ProviderAdapter реализации (11 провайдеров: 9 YAML config + 2 hard-coded) |
| `circuitbreaker/` | Circuit breaker per provider (failure threshold, recovery timeout) |
| `webhook/` | Webhook receiver + verifier (signature validation, payload processing) |
| `grpc/` | gRPC server + generated code (`proto/integration/v1/`) |
| `config/` | YAML config загрузка (`provider-configs/*.yaml`) |
| `cmd/integration-proxy/main.go` | Entry point: gRPC server + HTTP webhook server |

### Адаптеры провайдеров (`adapters/`)

| Адаптер | Провайдер | Тип | Формат |
|-----|-|-|-|
| `alpha/` | АльфаСтрахование | ДМС | YAML config |
| `worldclass/` | World Class | Фитнес | YAML config |
| `yandex/` | Яндекс Еда for Business | Питание | YAML config |
| `yandex-psych/` | Яндекс Психотерапия | Психолог | YAML config |
| `skillbox/` | Skillbox | Обучение | YAML config |
| `sber/` | СберСпорт | Спорт | YAML config |
| `sdek-store/` | СДЭК Store | Мерч | YAML config |
| `motherchild/` | Мать и дитя | Стоматология | YAML config |
| `mts/` | МТС | Связь | YAML config |
| `skyeng/` | Skyeng | Языки | YAML config |
| `giftcard/` | Подарок в квадрате | Подарочные карты | hard-coded |

### Circuit Breaker

| Параметр | Значение |
|----------|---------|
| Failure threshold | 10 ошибок за 60с |
| Recovery timeout | 30с (half-open probe) |
| State: closed | Нормальная работа |
| State: open | Все вызовы отклоняются, cached/error response |
| State: half-open | Один пробный запрос, success → closed, failure → open |

### Зависимости

- `google.golang.org/grpc` — gRPC server
- `proto/integration/v1/` — generated protobuf types
- PostgreSQL — schema `lkfl_integration` (providers, sync_log, webhook_events, dead_letters, activation_jobs, circuit_breaker_state)
- Redis — retry queues, job state
- HTTP client — для benefit-provider REST API

---

## Пакет `internal/payments/` — M12: бывший отдельный сервис

> **M12:** Payment-gateway-сервис (порт :8084, `payment-gateway:8084`, отдельный go.mod + Dockerfile) слит в Platform как `internal/payments/`. Контракты → Go-интерфейсы.

**Назначение.** Авторизация/подтверждение/отмена платежей (Visa/MC/МИР), передача заявлений на удержание из ЗП.

**PCI DSS isolation (M12):** Физическая изоляция заменена на separation of concerns на уровне кода:
- `internal/payments/` — thin wrapper, данные карты не сохраняются
- Separate credentials storage в HashiCorp Vault
- Token-only от пэйшлюза, не raw card data
- TLS 1.3 между payment gateway и банком

### Структура

| Файл | Назначение |
|--------|---------|
| `gateway.go` | `PaymentGateway` — Authorize, Capture, Void, CapturePayroll |
| `api.go` | Thin REST handlers: NewRouter(deps) |
| `auth.go` | Thin wrapper on `shared/pkg/auth` (JWT, tenant) |

### Публичный API

```go
type PaymentGateway struct {
    provider *PaymentProviderClient
    db       *pgxpool.Pool
    logger   Logger
}

func (g *PaymentGateway) Authorize(ctx context.Context, userId uuid.UUID, amount float64, method string) (*AuthResult, error)
func (g *PaymentGateway) Capture(ctx context.Context, authCode string) error
func (g *PaymentGateway) Void(ctx context.Context, authCode string) error
func (g *PaymentGateway) CapturePayroll(ctx context.Context, statementId uuid.UUID) error

type PaymentProvider interface {
    Authorize(ctx context.Context, cardToken string, amount float64) (*AuthResult, error)
    Capture(ctx context.Context, authCode string) error
    Void(ctx context.Context, authCode string) error
    Refund(ctx context.Context, authCode string, amount float64) error
}
```

### Зависимости

- `db/` — PostgreSQL (payment_transactions, payment_authorizations)
- `shared/pkg/auth` — JWT validation, middleware, RBAC
- Payment provider API (REST) — external, TLS 1.3

---

## Пакет `internal/content/` — FAQ, баннеры, описания

**Назначение.** Управление динамическим контентом: FAQ (поддержка), баннеры (главная), описания карточек (каталог), страницы (о платформе).

**Почему отдельный пакет:** Контент — независимый домен (не engagement, не notification). Admin CRUD через `admin_content.go` → `content.ContentEngine`. Public read-only через `support_handler.go` (FAQ), `catalog_handler.go` (banнеры).

### Структура

| Файл | Назначение |
|------|-----------|
| `content.go` | `ContentEngine` — Create, Get, Update, Delete, ListPublished, ListByType |
| `types.go` | Типы: Content (id, tenant_id, type, title, body, sort_order, is_published, ...) |

### Публичный API

```go
type ContentEngine struct { /* DI: db, logger */ }

func (e *ContentEngine) Create(ctx context.Context, in *ContentInput) (*Content, error)
func (e *ContentEngine) Get(ctx context.Context, id uuid.UUID) (*Content, error)
func (e *ContentEngine) Update(ctx context.Context, id uuid.UUID, in *ContentInput) error
func (e *ContentEngine) Delete(ctx context.Context, id uuid.UUID) error
func (e *ContentEngine) ListPublished(ctx context.Context, contentType string) ([]Content, error)
func (e *ContentEngine) ListByTenant(ctx context.Context, tenantId uuid.UUID, q *ContentQuery) ([]Content, Pagination, error)
```

### Зависимости

- `db/` — PostgreSQL (content table)
- `pkg/` — общие типы (Tenant, Pagination)

---

## Пакет `internal/api/` —thin handlers (2 router'а, M10 T1004)

**Назначение.** HTTP handlers без бизнес-логики. Каждый handler делегирует в соответствующий business-пакет.

> **M10 T1004:** Разделение на 2 router'а:
> - **Public API** (`/api/v1/...`) — endpoints для аутентифицированных пользователей, high rate limits, RBAC = любой authenticated
> - **Admin API** (`/admin/...`) — endpoints для HR/catalog_manager/admin, low rate limits, RBAC = admin-only, полный audit trail

### Структура

| Файл | Назначение | Router |
|--------  |---------|---|
| `router.go` | Chi router setup, wire public + admin routers | оба |
| `public_router.go` | Public router setup, PublicHandlerDeps, middleware chain | public |
| `admin_router.go` | Admin router setup, AdminHandlerDeps, middleware chain | admin |
| `auth_handler.go` | `/auth/*` → `auth/` (register only, остальное — Keycloak) | public |
| `user_handler.go` | `/user/*` → `user/` | public |
| `consent_handler.go` | `/user/consents/*` → `consent/` | public |
| `engagement_handler.go` | `/engagements/*`, `/engagement-offers/*` → `engagement/catalog/` | public |
| `user_engagement_handler.go` | `/user-engagements/*` → `engagement/flow/` | public |
| `collections_handler.go` | `/collections/*` → `engagement/collections/` | public |
| `notification_handler.go` | `/notifications/*` → `notification/` | public |
| `recommendation_handler.go` | `/recommendations/*` → `recommendations/` (stub, M10 T1001) | public |
 | `balance_handler.go` | `/balance/*` → direct call billing/BillingService | public |
| `document_handler.go` | `/documents/*` → Asynq `document-generate` | public |
| `support_handler.go` | `/support/*` → db/ | public |
| `gamification_handler.go` | **M09 T0901** — `/gamification/v1/*` → `gamification/` | public |
| `cel_handler.go` | **P1** — `/api/v1/cel/*` → generate + validate + **dry-run** + **dry-run/batch** | public |
| `admin_user.go` | **M07 T0703** — `/admin/users/*`, `/admin/periods/*` → `user/` + `consent/` | admin |
| `admin_catalog.go` | **M07 T0703** — `/admin/engagements/*`, `/admin/engagement-types/*` → `engagement/catalog/` | admin |
| `admin_flows.go` | **M07 T0703** — `/admin/engagement-flows/*` → `engagement/flow/` | admin |
| `admin_collections.go` | **M07 T0703** — `/admin/collections/*` → `engagement/collections/` | admin |
| `admin_recommendations.go` | **M07 T0703** — `/admin/recommendations/*` → `recommendations/` (stub, T1001) | admin |
| `admin_analytics.go` | **M07 T0703** — `/admin/analytics/*` → агрегация db/ | admin |
| `admin_content.go` | **M07 T0703** — `/admin/content/*`, `/admin/requests/*` → db/ + notification/ | admin |
| `middleware.go` | common middleware: logging, recovery, rate limiting | оба |
| `admin_middleware.go` | admin-specific middleware: audit trail, admin RBAC | admin |

### DI — 2 HandlerDeps struct (M10 T1004)

```go
// Public API: /api/v1/... — endpoints для authenticated пользователей
// M12: NATS убран → BillingService, PaymentGateway через Go interfaces
// M16: ProviderGateway → IntegrationClient (gRPC к proxy)
type PublicHandlerDeps struct {
    Auth              *auth.OIDCVerifier
    User              *user.UserRepository
    Consent           *consent.ConsentEngine
    Catalog           *catalog.CatalogService           // engagement/catalog/
    Flow              *flow.FlowEngine                  // engagement/flow/
    Collections       *collections.CollectionsEngine    // engagement/collections/
    Notification      *notification.NotificationEngine
    Gamification      *gamification.GamificationEngine  // M09
    Billing           *billing.BillingService           // M12: был NATS
    IntegrationClient integrationclient.IntegrationService // M16: gRPC к proxy
    Payments          *payments.PaymentGateway          // M12: был NATS
    DB                *sql.DB
    Redis             *redis.Client
    Asynq             *asynq.Server
    Logger            Logger
}

func NewPublicRouter(deps *PublicHandlerDeps) *chi.Mux

// Admin API: /admin/... — endpoints для HR / catalog_manager / admin
type AdminHandlerDeps struct {
    User          *user.UserRepository
    Consent       *consent.ConsentEngine
    Catalog       *catalog.CatalogService           // engagement/catalog/
    Flow          *flow.FlowEngine                  // engagement/flow/
    Collections   *collections.CollectionsEngine    // engagement/collections/
    Notification  *notification.NotificationEngine
    Gamification  *gamification.GamificationEngine  // M09
    DB            *sql.DB
    Logger        Logger
}

func NewAdminRouter(deps *AdminHandlerDeps) *chi.Mux
```

> **M10 T1004:** Old `HandlerDeps` (12 полей) разбит на 2 struct.
> - `PublicHandlerDeps` (10 полей originally): Auth, User, Consent, Engagement, Notification, Gamification, DB, Redis, Logger + **M12:** Billing, Integrations, Payments (был NATS, убран) + **M16:** Integrations → IntegrationClient (gRPC к proxy)
> - `AdminHandlerDeps` (7 полей): User, Consent, Engagement, Notification, Gamification, DB, Logger
> - `Eligibility` убран из Public (не нужен напрямую в handlers)
> - **M12:** `NATS` убран из Public. `Asynq` убран из Admin (не нужны в admin endpoints)
> - **M16:** `Integrations` → `IntegrationClient` (interface `integrationclient.IntegrationService`)
> - `Recommendations` убран (stub, M10 T1001)

### Middleware chains (M10 T1004)

```
Public:  Recovery → Logger → RateLimiter(high) → JWT → Tenant → RBAC(any) → Handler
Admin:   Recovery → Logger → RateLimiter(low)  → JWT → Tenant → RBAC(admin) → Audit → Handler
```

| Middleware | Public | Admin | Разница |
|---|---|---|---|
| Recovery | ✅ | ✅ | одинаковый |
| Logger | ✅ | ✅ | одинаковый |
| RateLimiter | HIGH (1000 req/min) | LOW (100 req/min) | admin — меньше пользователей |
| JWT | ✅ | ✅ | одинаковый |
| Tenant | ✅ | ✅ | одинаковый |
| RBAC | any authenticated | admin-only | admin: `hr`, `catalog_manager`, `admin` |
| Audit | — | ✅ | admin: полный audit trail каждого действия |

### Handler grouping

**Public handlers (10):** auth_handler, user_handler, consent_handler, engagement_handler, collections_handler, notification_handler, recommendation_handler, balance_handler, document_handler, support_handler, gamification_handler, cel_handler

**Admin handlers (5):** admin_user, admin_catalog, admin_recommendations, admin_analytics, admin_content

---

## DI граф — ASCII диаграмма (M10 T1004: 2 router'а + M16: proxy)

```
                     ┌─────────────────────────────┐
                     │         Nginx (:80)         │
                     │                             │
                     │  /api/v1/*     → :8080      │
                     │  /webhooks/*   → :8091      │
                     └──────┬──────────────┬───────┘
                            │              │
              ┌─────────────▼──┐    ┌──────▼──────────────┐
              │  lkfl-         │    │ lkfl-integration-   │
              │  server        │    │ proxy               │
              │  (:8080)       │    │ (:8090 gRPC)        │
              │                │    │ (:8091 HTTP webhook) │
              │  gRPC client   │───→│  Provider adapters  │
              │  (localhost)   │    │  Circuit breaker    │
              └───────┬────────┘    │  Retry/timeout      │
                      │             │  Credential store   │
           ┌──────────┴──────────┐  └────────┬───────────┘
           │                      │          │
   ┌───────▼──────────┐   ┌──────▼───────────▼──┐
   │  PublicRouter    │   │  AdminRouter         │
   │  (/api/v1/...)   │   │  (/admin/...)        │
   │  PublicHandlerDeps│   │  AdminHandlerDeps    │
   │  ────────────────│   │  ──────────────────  │
   │  Recovery→Logger │   │  Recovery→Logger     │
   │  JWT→Tenant→RBAC │   │  JWT→Tenant→RBAC→Audit│
   │  Public Handlers │   │  Admin Handlers      │
   └───┬──┬──┬──┬─────┘   └───┬──┬──┬──┬─────────┘
       │  │  │  │             │  │  │  │
 ┌─────▼──▼──▼──▼─────────────▼──▼──▼──▼──────────┐
 │              shared internal packages            │
 │  auth/ │ user/ │ consent/ │ eligibility/         │
 │  compliance/ │ engagement/ │ notification/        │
 │  gamification/ │ billing/ │ payments/            │
 │  content/ │ cel/ │ llm/ │ recommendations/(stub) │
 │  integrationclient/ → gRPC → proxy (M16)         │
 └───┬──────────┬──────────┬──────────┬─────────────┘
     │          │          │          │
     └──────────┴──────────┴──────────┴──────────┐
                                                 ▼
                ┌──────────────────────────────────────────┐
                │           external deps                  │
                │ PostgreSQL (lkfl_platform schema)        │
                │ Redis (key prefixes)                     │
                │ SMTP │ FCM/APNs │ Keycloak (OIDC)        │
                └──────────────────────────────────────────┘

  Proxy external deps:
                ┌──────────────────────────────────────────┐
                │ PostgreSQL (lkfl_integration schema)     │
                │ Внешние провайдеры (HTTP/REST)           │
                └──────────────────────────────────────────┘
```

### Зависимости между пакетами

| Пакет | Зависит от |
|-------|----------|
| `auth/` | `pkg/`, Redis (`jwt:` prefix), `shared/pkg/auth` |
| `user/` | `pkg/`, `consent/`, PostgreSQL |
| `consent/` | `pkg/`, PostgreSQL |
| `cel/` | `pkg/`, `llm/` (direct call), Redis (`cel:` prefix) |
| `llm/` | `pkg/`, PostgreSQL, Redis (`cel:` prefix), Ollama/OpenAI API |
| `eligibility/` | `pkg/`, `user/`, `cel/` |
| `compliance/` | `user/`, `engagement/flow/`, `notification/`, PostgreSQL |
| `engagement/catalog/` | `pkg/`, PostgreSQL, Redis (`catalog:` prefix) |
| `engagement/flow/` | `pkg/`, `user/`, `eligibility/`, `cel/`, `billing/` (direct call), `integrationclient/` (gRPC → proxy), `payments/` (direct call), Go-callback → `gamification/`, PostgreSQL, Redis |
| `engagement/collections/` | `pkg/`, `engagement/flow/`, PostgreSQL |
| `notification/` | `pkg/`, PostgreSQL |
| `recommendations/` | `pkg/` | **M10 T1001:** stub (Phase 2). |
| `gamification/` | `pkg/`, `user/`, `cel/`, PostgreSQL, Redis (`asynq:` prefix) |
| **`billing/`** | `pkg/`, `cel/`, PostgreSQL, `shared/pkg/celcontext` |
| **`integrationclient/`** | `google.golang.org/grpc`, `proto/integration/v1/` (gRPC → proxy) |
| **`payments/`** | `pkg/`, PostgreSQL, `shared/pkg/auth`, HTTP (payment provider API) |
| `api/` (Public) | auth, user, consent, engagement, notification, gamification, billing, integrationclient, payments, db, redis, logger |
| `api/` (Admin) | user, consent, engagement, notification, gamification, billing, db, logger |

---

## Asynq workers mapping

Монолит `lkfl-server` использует Asynq workers для фоновых задач бизнес-логики.

> **M16 (ADR-035):** catalog-sync вынесен в proxy. Proxy имеет собственный worker pool для HTTP вызовов к провайдерам (не Asynq монолита).

| Worker | Пакет | Метод | Описание |
|--|-|--------|-----|

| Worker | Пакет | Метод | Описание |
|--|-|--------|-----|
| `notification-send` | `notification/` | `NotificationEngine.SendFromEvent(ctx, event)` | Отправка по каналу (email/push/in-app) |
| `document-generate` | `engagement/flow/` | `FlowEngine.ExecuteStep(ctx, engagementId, stepId, "document_generation")` | Генерация PDF в flow step |
| `registry-import-xlsx` | `user/` | `RegistryImporter.ImportXLSX(ctx, file)` | Импорт XLSX реестра |
| `registry-import-api` | `user/` | `RegistryImporter.ImportAPI(ctx, hrClient)` | Импорт из HR-системы |
| `hr-sync-daily` | `user/` | `HRSync.PullRegistry(ctx)` | **M11 T1102:** ежедневный pull кадрового реестра из HR-системы (был через NATS → Integrations) |
| `consent-revoke` | `compliance/` | `ComplianceEngine.CascadeRevoke(ctx, userId)` | Каскадный отзыв ПДн (T0702: из consent/ → compliance/) |
| `engagement-activate` | `engagement/flow/` | `FlowEngine.Activate(ctx, offerId, userId)` | Запуск activation flow |
| `engagement-complete` | `engagement/flow/` | `FlowEngine.Complete(ctx, engagementId)` | Запуск completion flow |
| `engagement-revert` | `engagement/flow/` | `FlowEngine.Revert(ctx, engagementId)` | Возврат при отмене |
| `gamification-check-monthly` | `gamification/` | `LoyaltyEngine.MonthlyCheck(ctx)` | Ежемесячная проверка уровней лояльности + CEL-ачивок monthly_cron |
| `gamification-import-xlsx` | `gamification/` | `XLSXImporter.Import(ctx, file)` | Массовое присвоение ачивок и уровней из XLSX (2 шаблона: badges + levels) |

---

## Что НЕ меняется

| Компонент | До M06 | После M06 | M12 | M16 |
|-----|-|---|-|-|
| Бинарников | 2 (server + worker) | 2 (server + worker) | **M12:** 2 (lkfl-server + worker) | **M16:** 3 (lkfl-server + worker + integration-proxy) |
| NATS subjects | 9 (platform↔billing + platform↔integrations) | 9 | **M12:** Optional, 0 в mono-режиме. Go interfaces вместо NATS | ❌ нет |
| PostgreSQL schema | `lkfl_platform` | `lkfl_platform` (+ achievements, loyalty M09) | **M12:** Одна schema. `lkfl_billing`, `lkfl_payments`, `lkfl_integrations` → merged | **M16:** `lkfl_platform` (монолит) + `lkfl_integration` (proxy) |
| Redis DB 0 (JWT cache) | используется | Redis DB 0 | **M12:** key prefix `jwt:` | ❌ нет |
| Redis DB 1 (Asynq) | используется | Redis DB 1 | **M12:** key prefix `asynq:` | ❌ нет |
| Redis DB 2 (catalog cache) | используется | Redis DB 2 | **M12:** key prefix `catalog:` | ❌ нет |
| Redis DB 3 (rate limiting) | используется | Redis DB 3 | **M12:** key prefix `rate:` | ❌ нет |
| Redis DB 4 (CEL) | используется | Redis DB 4 | **M12:** key prefix `cel:` | ❌ нет |
| Nginx routes | `/api/` → platform:8080 | `/api/` → platform:8080 | **M12:** `/api/` → lkfl-server:8080, merged billing + payments | **M16:** `/api/` → server:8080, `/webhooks/` → proxy:8091 |
| go.mod | отдельно | отдельно | **M12:** один go.mod | **M16:** один go.mod (два бинарника) |
| Billing сервис | отдельно | отдельно | **M12:** `internal/billing/` | ❌ нет |
| Integrations сервис | отдельно | отдельно | **M12:** `internal/integrations/` | **M16:** `integration-proxy/` (отдельный бинарник) |
| Payment Gateway сервис | отдельно | отдельно | **M12:** `internal/payments/` | ❌ нет |
| Frontend | отдельно | отдельно | ❌ нет | ❌ нет |
| Keycloak IdP | центральный | центральный | ❌ нет | ❌ нет |

---

## Почему пакеты, а не gRPC микросервисы

### Сравнительный анализ

| Критерий | Пакеты (internal/*) | Отдельные микросервисы |
|--------|-|---|
| **Бинарников** | 2 без изменений | 5+ (engagement-svc, notification-svc, recommendations-svc, ...) |
| **Nginx routes** | без новых правил | 5+ новых upstream → сложность маршрутизации |
| **Общие зависимости (auth, user, consent)** | один import | дублирование или gRPC client → зависимость |
| **Тестируемость** | unit-тесты пакета изолированы (mock'и DI) | нужны integration-тесты на каждый сервис |
| **Масштаб (131 endpoint, 1 команда)** | один handler package → просто | overkill для ~40 business endpoints Platform |
| **Deploy** | один Docker image | оркестрация 5+ контейнеров |
| **Межсервисная latency** | 0 — function call | network call + serialization |
| **Transactional consistency** | одна DB → ACID | distributed transactions → 2PC / saga |
| **Масштаб когда применимо** | 300+ endpoints, 3+ команд, different SLA | — |

### Обоснование выбора

1. **Platform — это API-агрегатор.** Его задача — предоставлять HTTP API для Frontend. Бизнес-логика не требует отдельного deploy'а: engagement eligibility, notification send, recommendations — всё в рамках одного request-response цикла.

2. **Нет различий в SLA.** Все 9 пакетов обслуживают один API с одинаковыми требованиями к latency (<200ms p95). Нет потребности в отдельном scale-up.

3. **Биллинг** остаётся в монолите (ACID на финансы). **Integrations вынесен в proxy** (M16, ADR-035) — это исключение: внешние HTTP calls не являются бизнес-логикой, это I/O boundary, требующая fault isolation. Выносить notification и recommendations из Platform не имеет смысла — они зависят от тех же DB-таблиц.

4. **Unit-тестовая изоляция достижима** через DI injection. Каждый бизнес-пакет принимает интерфейс зависимостей, а не конкретные реализации.

### Когда иметь смысл вынести в отдельный сервис

- Пакет > 5000 строк кода
- SLA отличается от остальных (например, 99.99% vs 99.9%)
- Команда > 3 разработчиков работает на этом пакете параллельно
- Данные требуют отдельной storage технологии (Cassandra, Elasticsearch)
- Compliance требует физической изоляции

Ни один из 3 новых пакетов (`notification/`, `recommendations/`, вычленённый `engagement/`) не соответствует этим критериям на текущей стадии.

> **M16 исключение:** `integration-proxy/` вынесен как отдельный бинарник по причине fault isolation внешних HTTP вызовов (ADR-035). Это не нарушает принцип монолита для бизнес-логики — proxy является infra component, не business module.

---

## Миграционный план

### Фаза 1: Документация (M06, T0601)
- [x] Создать `пакеты-platform.md` (этот документ)
- [x] Создать ADR-013
- [x] Обновить `модули.md`
- [x] Обновить `спецификация/api.md`

### Фаза 1.5: Документация M07 (T0701–T0709)
- [x] Создать ADR-014 (eligibility extraction)
- [x] Создать ADR-015 (compliance package)
- [x] Создать ADR-016 (admin handler split)
- [x] Создать ADR-017 (generic REST adapter)
- [x] Создать ADR-018 (payment-gateway service)
- [x] Создать ADR-019 (wizard engine)
- [x] Создать ADR-020 (NATS subjects registry)
- [x] Обновить `модули.md` — 4 Go-сервиса, eligibility/compliance packages, admin handlers
- [x] Обновить `спецификация/api.md` — eligibility, wizard, integrations hub endpoints
- [x] Создать `nats-subjects.md` — master registry 19 subjects
- [x] Обновить `интеграции.md` — generic REST adapter, provider-configs
- [x] Обновить `безопасность.md` — PCI DSS payment-gateway, compliance audit trail
- [x] Обновить `контекст/настраиваемость.md` — REST-провайдер без кода
- [x] Cross-check (T0709) — все ссылки проверены

### Фаза 2: Код (следующая веха после M07)
- [ ] Создать subpackages `internal/engagement/catalog/`, `internal/engagement/flow/`, `internal/engagement/collections/` — eligibility вынесен
- [ ] Создать package `internal/eligibility/` (T0701) — Check, EvaluateRule, EvaluateGroup
- [ ] Создать package `internal/compliance/` (T0702) — CascadeRevoke, AuditTrail, EnforceRetention
- [ ] Создать package `internal/notification/` с Channel interface
- [ ] Создать package `internal/recommendations/` с two-layer engine
- [ ] Переписать `api/` на thin handlers (5 admin файлов, T0703)
- [ ] Обновить Asynq workers (consent-revoke → compliance/)
- [ ] Выделить payment-gateway/ как 4-й Go-сервис (T0705)
- [ ] Заменить 9 hard-coded adapters на Generic REST + YAML config (T0704)

### Фаза 3: Тесты (после код-миграции)
- [ ] Unit-тесты `eligibility.EligibilityEngine` (AND/OR/groups)
- [ ] Unit-тесты `compliance.ComplianceEngine` (cascade revoke, audit trail)
- [ ] Unit-тесты `notification.Channel` (mock channel)
- [ ] Unit-тесты `recommendations.Evaluator` (segment matching)
- [ ] Integration-тесты полного flow (activate → billing → notification)

### Фаза 4: M16 — Integration Proxy (ADR-035)
- [x] Создать ADR-035 (T1601) — Integration Proxy обоснование
- [x] Обновить `пакеты-platform.md` (T1604) — `integrations/` → `integrationclient/` + `integration-proxy/`
- [ ] Создать `internal/integrationclient/` — gRPC client к proxy
- [ ] Создать `proto/integration/v1/integration.proto` — gRPC service definition
- [ ] Создать `integration-proxy/` — adapters/, circuitbreaker/, webhook/, grpc/, config/
- [ ] Создать `cmd/integration-proxy/main.go` — entry point
- [ ] Обновить `engagement/flow/` — direct call → gRPC через integrationclient
- [ ] Обновить `api/` — PublicHandlerDeps: Integrations → IntegrationClient
- [ ] Unit-тесты `integrationclient.IntegrationClient` (mock IntegrationService)
- [ ] Unit-тесты `integration-proxy/adapters/` (mock provider HTTP)
- [ ] Integration-тесты полного flow (mono → gRPC → proxy → provider)

---

## Сводная таблица пакетов

### Монолит `lkfl-server` — 16 пакетов (14 business + tenant + api + integrationclient)

> **M16:** `internal/integrations/` удалён из монолита. На его месте `internal/integrationclient/` — gRPC client к proxy.

| Пакет | Файлов (ожидаемо) | Обязанностей (SRP) | Строк API | Тестовое покрытие (цель) |
|-------|---:|---:|-------:|---:|
| `auth/` | 3 | 1 (OIDC verification) | 4 | 80% |
| `user/` | 3 | 1 (CRUD) | 6 + 2 = 8 | 80% |
| `consent/` | 2 | 1 (PDn lifecycle only) | 4 | 80% |
| `eligibility/` | 2 | 1 (CEL evaluation) | 3 | 90% |
| `compliance/` | 3 | 3 (cascade, audit, retention) | 3 | 85% |
| `engagement/` | 5 | 3 (catalog, flow, collections) | 13 | 90% |
| `notification/` | 7 | 2 (delivery + persistence) | 4 + 4 = 8 | 85% |
| `recommendations/` | 5 | 2 (engine + rules CRUD) | 2 + 4 = 6 | 90% |
| `gamification/` | 5 | 3 (achievements, loyalty, triggers) | 6 | 85% |
| `cel/` | 4 | 1 (CEL generation + evaluation) | 4 | 90% |
| `llm/` | 3 | 1 (LLM agent routing + prompt mgmt) | 2 | 80% |
| `billing/` | 5 | 4 (balance, tx, periods, rules) | 8 | 90% |
| **`integrationclient/`** | **3** | **1 (gRPC client к proxy)** | **9** | **85%** |
| `payments/` | 3 | 1 (PCI DSS payments) | 4 | 85% |
| `content/` | 2 | 1 (dynamic content CRUD) | 6 | 80% |
| `api/` | 12+ | 1 (HTTP routing + middleware) | делегирует | 70% |

### Integration Proxy — отдельный бинарник

| Пакет | Файлов (ожидаемо) | Обязанностей (SRP) | Строк API | Тестовое покрытие (цель) |
|-------|---:|---:|-------:|---:|
| `integration-proxy/adapters/` | 11 | 1 (provider adapter per provider) | 0 (internal) | 80% |
| `integration-proxy/circuitbreaker/` | 2 | 1 (circuit breaker per provider) | 0 (internal) | 90% |
| `integration-proxy/webhook/` | 2 | 1 (webhook receiver + verifier) | 0 (internal) | 85% |
| `integration-proxy/grpc/` | 3 | 1 (gRPC server + generated code) | 9 (gRPC) | 80% |
| `integration-proxy/config/` | 1 | 1 (YAML config loading) | 0 (internal) | 80% |
