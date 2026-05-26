# CEL Engine — единый движок бизнес-логики

## TL;DR (для агентов)

> Этот файл описывает **CEL (Google Common Expression Language)** — единый движок бизнес-правил платформы.
> - **Архитектура (5 доменов: billing, eligibility, flow, gamification, engagement)** → строка 21
> - **CELContext (единая схема контекста)** → строка 111
> - **Go API (Evaluate, Generate, Validate)** → строка 66
> - **LLM Integration (in-process генерация CEL из русского текста)** → строка 223
> - **14 точек интеграции CEL в системе** → строка 375
> - **Публичный API пакета `cel/`** → `пакеты-platform.md` строка 318
> - **TagResolver (теги пользователя для CEL)** → `теги.md`

> **ADR-021:** Google CEL заменяет 4 независимых механизма условия (billing YAML, eligibility AND/OR, flow condition_expr, recommendations JSON). Все условия — CEL expression. Генерация — через LLM из русского текста.

---

## Роль движка

CEL Engine — **единая точка evaluation** для всех бизнес-правил платформы:

| Домен | Что оценивает | До CEL (ADR-021) | После CEL |
|------|-------------|---------|-|
| **Биллинг** | Применяется ли правило к пользователю? | YAML-array `[{field, op, val}]` | `condition_cel: "user.grade in ['A','B']"` |
| **Eligibility** | Доступен ли оффер пользователю? | AND/OR/Groups struct | `eligibility_cel: "user.grade == 'Senior' && user.years_of_service >= 3"` |
| **Engagement Flow** | Пройден ли condition_check шаг? | ad-hoc: `'answers.count >= 5'` | `condition_expr: "answers.size() >= 5"` |
| **Recommendations** | Попадает ли пользователь в сегмент? | JSON segment conditions | `segment_cel: "user.department == 'Sales' && !tags.is_remote"` |
| **Gamification** | Присваивается ли ачивка / уровень лояльности? | — (нет слоя геймификации) | `condition_cel: "game.engagement_by_category['survey'] >= 5"` |

---

## Архитектура

```
┌────────── Admin UI ───────────┐
│ textarea: "бесплатный фитнес  │
│ для директоров и удалённых"   │
│                               │
│ [Генерировать CEL]            │
│                               │
│ CEL preview (read-only):      │
│ user.grade == 'Director'      │
│ || tags.is_remote == true     │
│ [✏️ Expert: ручной CEL]       │
│                               │
│ [Сохранить]                   │
└───┬──────────────────────┬────┘
    │ POST /api/v1/cel/    │ PUT /billing/v1/rules/:id
    │ generate              │
    ▼                       ▼
┌──────────────────────────────────────┐
│  Platform: cel/ package             │
│  ┌──────────────────────────────┐    │
│  │ CELGenerator                │    │
│  │ - Generate() → LLM call     │    │
│  │ - Validate() → cel-go parse │    │
│  │ - Evaluate() → cel-go       │    │
│  │   interpreter               │    │
│  └──────────────────────────────┘    │
│  ┌──────────────────────────────┐    │
│  │ CELValidator                │    │
│  │ - Syntax check              │    │
│  │ - Type check against schema │    │
│  │ - Drift detection           │    │
│  └──────────────────────────────┘    │
└───────────┬─────────────────┬────────┘
            │                 │
    ┌───────▼──────┐  ┌──────▼────────┐
    │ LLM (M10: in-process) │  │  CEL Context  │
    │ internal/llm/        │  │  Builder      │
    │ (был :8085, ADR-022) │  │  (CELContext) │
    └──────────────┘  └───────────────┘
```

---

## Go API

### Package: `platform/internal/cel/`

| Файл | Назначение |
|--------|------|
| `generator.go` | `CELGenerator — Generate(ctx, sourceText) → (cel, error)` (через `internal/llm/`, M10 T1002: был LLMProxyClient) |
| `evaluator.go` | `CELEvaluator — Evaluate(ctx, celExpr, context) → (bool, error)` |
| `validator.go` | `CELValidator — Validate(celExpr, schema) → (valid, errors[])` |
| `context.go` | `CELContext — типы контекста (nested: user.*, benefit.*, date.*, context.*)` |
| `schema.go` | `CELSchema — определение доступных полей и типов для CEL` |
| `functions.go` | Custom CEL functions: date_diff_days, str_contains, now_iso |
| `~proxy_client.go` | ~~`LLMProxyClient — HTTP-клиент для LLM Proxy :8085`~~ → M10 T1002: удалён. Теперь `internal/llm/LLMClient.GenerateCEL(ctx, sourceText)` in-process direct call |

### Public API

```go
type CELGenerator struct {
    llmProxy    *LLMProxyClient
    validator   *CELValidator
    logger      Logging
}

// Generate CEL expression из русского текста (через LLM Proxy — ADR-022)
func (g *CELGenerator) Generate(ctx context.Context, source string) (string, error)

type CELEvaluator struct {
    celEnv *cel.Env
}

// Evaluate CEL expression в контексте
func (e *CELEvaluator) Evaluate(ctx context.Context, celExpr string, c *CELContext) (bool, error)

type CELValidator struct {
    celEnv *cel.Env
}

// Validate CEL синтаксиса и типов
func (v *CELValidator) Validate(celExpr string) error

func (v *CELValidator) Compile(celExpr string) (*cel.Program, error)
```

---

## CELContext — единая схема контекста

```go
type CELContext struct {
    // Профиль пользователя — вложенный (user.*)
    User struct {
        Grade          string `cel:"grade"`
        YearsOfService int    `cel:"years_of_service"`
        HasChildren    bool   `cel:"has_children"`
        Department     string `cel:"department"`
        Status         string `cel:"status"`        // active | inactive | frozen
        TenantID       string `cel:"tenant_id"`
        UserID         string `cel:"user_id"`
    } `cel:"user"`

// Теги (динамические) — tags.*
Tags map[string]interface{} `cel:"tags"`
// tags.is_remote, tags.is_newbie, tags.pilot_program
// tags.interest:sport, tags.sport_intensity (M14 planned: survey-теги из TagResolver.AggregateSurveyTags, spec in M13 T1301)

    // Данные льготного оффера — вложенный (benefit.*)
    Benefit struct {
        Category string  `cel:"category"`   // fitness, dms, food, ...
        Cost     float64 `cel:"cost"`
    } `cel:"benefit"`

    // Дата/время — вложенный (date.*)
    Date struct{ Today string } `cel:"date"` // "2026-05-24" ISO8601 date

    // Динамический контекст (зависит от домена) — context.*
    Context map[string]interface{} `cel:"context"`
    // context.last_engagement, context.engagement_active, context.engagement_provider

    // Survey condition_check — answers.*
    Answers map[string]interface{} `cel:"answers"`

    // Billing event context — events.*
    Events map[string]interface{} `cel:"events"`
    // events.engagement_offer_cost, events.type

    // Период — period.*
    Period struct {
        Start string `cel:"start"`
        End   string `cel:"end"`
    } `cel:"period"`

    // Balance — вложенный (balance.*)
    Balance struct{ Total float64 } `cel:"balance"`

    // Геймификация — вложенный (game.*) — M09 ADR-023
    Game struct {
        Achievements         []string         `cel:"achievements"`           // ключи имеющихся ачивок
        AchievementCount     int              `cel:"achievement_count"`     // количество ачивок
        EngagementCount      int              `cel:"engagement_count"`      // всего завершённых энгейджментов
        EngagementByCategory map[string]int   `cel:"engagement_by_category"`// по категориям: {'survey': 3, 'referral': 2}
        BenefitCategories    int              `cel:"benefit_categories_count"` // кол-во категорий льгот, в которые юзер подключен
        LoyaltyLevel         string           `cel:"loyalty_level"`         // текущий уровень
        LoyaltyPoints        float64          `cel:"loyalty_points"`        // cumulative engagement points
        DaysSinceActive      int              `cel:"days_since_active"`    // дней с последней активности
        EnpsSubmitted        bool             `cel:"enps_submitted"`
        HasFamily            bool             `cel:"has_family"`      // есть родственники в системе ДМС
    } `cel:"game"`
}
```

> **Важно:** все поля доступны в CEL через префикс `user.*`, `benefit.*`, `date.*`, `context.*`, `events.*`, `balance.*`, `period.*`, `game.*`. Flat-поля: `tags.*`, `answers.*`.
>
> **M09 ADR-023:** `game.*` — геймификация (5-й домен CEL). Используется в `gamification/grant_engine.go` для evaluation условий присвоения ачивок.

### Custom CEL functions

Помимо стандартного CEL, добавлены 3 custom function для удобства:

| Function | Описание | Пример |
|----------|---------|--------|
| `date_diff_days(a, b)` | Разница между двумя датами (ISO8601) в днях | `date_diff_days(date.today, context.last_engagement) >= 7` |
| `str_contains(str, substr)` | Проверка вхождения подстроки | `str_contains(user.department, 'Sales')` |
| `now_iso()` | Текущая дата в ISO8601 | `date.today == now_iso()` |
```

### Available fields по доменам

| Поле (CEL) | Биллинг | Eligibility | Flow | Recommendations | Gamification |
|-----------|:-------:|:-----------:|:----:|:---------------:|:---:|
| `user.grade` | ✅ | ✅ | — | ✅ | — |
| `user.years_of_service` | ✅ | ✅ | — | ✅ | ✅ |
| `user.has_children` | ✅ | ✅ | — | ✅ | — |
| `user.department` | ✅ | ✅ | — | ✅ | — |
| `user.status` | ✅ | ✅ | — | — | — |
| `user.user_id` | ✅ | ✅ | — | ✅ | ✅ |
| `tags.*` | ✅ | ✅ | — | ✅ | — |
| `benefit.category` | ✅ | ✅ | — | ✅ | — |
| `benefit.cost` | ✅ | ✅ | — | — | — |
| `date.today` | — | ✅ | — | ✅ | ✅ |
| `answers.*` | — | — | ✅ | — | — |
| `events.*` | ✅ | — | — | — | — |
| `balance.total` | ✅ | — | — | — | ✅ |
| `period.start` / `period.end` | ✅ | ✅ | — | — | — |
| `context.*` | ✅ | — | — | — | — |
| `game.achievements` | — | — | — | — | ✅ |
| `game.achievement_count` | — | — | — | — | ✅ |
| `game.engagement_count` | — | — | — | — | ✅ |
| `game.engagement_by_category` | — | — | — | — | ✅ |
| `game.benefit_categories_count` | — | — | — | — | ✅ |
| `game.loyalty_level` | — | — | — | — | ✅ |
| `game.loyalty_points` | — | — | — | — | ✅ |
| `game.days_since_active` | — | — | — | — | ✅ |
| `game.enps_submitted` | — | — | — | — | ✅ |
| `game.has_family` | — | — | — | — | ✅ |

---

## LLM Integration — in-process (M10 T1002)

> **M10 T1002:** LLM Proxy слит в Platform как `internal/llm/`. `LLMProxyClient` удалён — заменён на in-process direct call `internal/llm/LLMClient.GenerateCEL(ctx, sourceText)`.

```go
// Было (pre-M10): LLMProxyClient — HTTP client для :8085
// Стало (M10 T1002):
type LLMClient struct {
    provider  LLMProvider   // OllamaClient, OpenAIClient
    validator *CELValidator
    logger    Logging
}

func (c *LLMClient) GenerateCEL(ctx context.Context, sourceText string) (string, error)
```

### LLM Integration — in-process

LLM-генерация идёт через **`internal/llm/`** — in-process пакет. Ни Platform, ни Billing не делают HTTP-вызовов к отдельному LLM Proxy :8085.

```
┌── Platform ─────────────────────┐
│ internal/cel/generator.go       │
│   → internal/llm/LLMClient     │
│     .GenerateCEL(source)       │
│ internal/cel/evaluator.go       │
│   → cel-go.Eval() (локально 5μs)│
│ internal/cel/validator.go       │
└──┬───────────────────────┘
    │ in-process Go call
┌───▼─────────────────────────┐   ┌── ollama ────┐  ┌── openai ────┐
│ internal/llm/               │   │ :11434        │  │ api.com      │
│ - LLMClient                 │───│               │  │              │
│ - OllamaClient              │   └───────────────┘  └──────────────┘
│ - OpenAIClient              │
│ - Agent router (in-memory)  │
│ - cost tracking             │
│ - audit trail (PostgreSQL)  │
└─────────────────────────────┘

> M10 T1002: был LLM Proxy :8085, теперь in-process direct call
```

**Почему in-process, а не отдельный сервис:**  M10 T1002: убран +1 HTTP hop (latency ~50ms на каждый CEL-gen вызов). Проще deployment: один бинарник, не 5 микросервисов. Cost tracking и audit trail → та же PG, тот же connection pool. Platform down → LLM down anyway (CEL gen нужен только при CRUD правил).

> **Детальное описание pre-M10 архитектуры LLM Proxy:** [`legacy/llm-proxy.md`](./legacy/llm-proxy.md) — Historical документ. Agent router, prompt templates, cost tracking, audit trail — принципы те же, но in-process.

### System Prompt

```
You are a CEL (Common Expression Language) expression generator.
Available context schema (nested):
{
  user: {
    grade: string,
    years_of_service: int,
    has_children: bool,
    department: string,
    status: string,
    tenant_id: string,
    user_id: string
  },
  tags: map<string, any>,              // tags.is_remote, tags.is_newbie, ...
  benefit: { category: string, cost: double },
  date: { today: string },             // ISO8601 date "2026-05-24"
  context: map<string, any>,           // context.last_engagement, context.engagement_active, ...
  answers: map<string, any>,           // answers.size(), answers.field_name
  events: map<string, any>,            // events.engagement_offer_cost, events.type
  period: { start: string, end: string },
  balance: { total: double },
  game: {                              // M09 ADR-023: Gamification domain
    achievements: string[],
    achievement_count: int,
    engagement_count: int,
    engagement_by_category: map<string, int>,
    benefit_categories_count: int,
    loyalty_level: string,
    loyalty_points: double,
    days_since_active: int,
    enps_submitted: bool,
    has_family: bool
  }
}

Custom functions:
- date_diff_days(a: string, b: string) → int  // разница дат в днях
- str_contains(str: string, substr: string) → bool

RULES:
- Output ONLY a valid CEL expression. No explanations, no quotes, no markdown.
- Use == for equality, != for inequality
- Use && for AND, || for OR, ! for NOT
- Use <, >, <=, >= for comparisons
- Use 'string_literal' for string values (single quotes)
- Use in for membership: user.grade in ['A', 'B']
- Use size() for map/slice length: answers.size() >= 5
- Use date_diff_days() for date arithmetic: date_diff_days(date.today, context.last_engagement) >= 7
- If condition is impossible to express, output: ERROR
```

---

## DB Schema миграции

### billing_rules

```sql
-- Old → New migration (Phase A)
ALTER TABLE billing_rules
  ADD COLUMN condition_cel TEXT,
  ADD COLUMN condition_source TEXT,
  ADD COLUMN condition_llm_model TEXT,
  ADD COLUMN condition_llm_version TEXT;

-- Populate from old condition JSONB
-- (migration script: translate [{field, op, val}] → CEL)
UPDATE billing_rules SET condition_cel = (SELECT translate_to_cel(condition));

-- После Phase A completion:
ALTER TABLE billing_rules DROP COLUMN condition;
```

### engagement_offers

```sql
ALTER TABLE engagement_offers
  ADD COLUMN eligibility_cel TEXT,
  ADD COLUMN eligibility_source TEXT;
```

### engagement_flows (steps)

```sql
-- steps JSONB: condition_expr уже TEXT, менять тип не нужно
-- Добавить source для audit trail
ALTER TABLE engagement_flows
  ALTER COLUMN steps TYPE jsonb
  USING jsonb_set(steps, '{condition_source}', '"auto-generated"');
```

### recommendation_rules

```sql
ALTER TABLE recommendation_rules
  ADD COLUMN segment_cel TEXT,
  ADD COLUMN segment_source TEXT,
  ADD COLUMN scoring_cel TEXT,
  ADD COLUMN scoring_source TEXT;
```

---

## 14 точек интеграции CEL

| # | Сервис | Пакет/Модуль | Таблица/Поле | Описание |
|---|--------|-------------|-------------|---------|
| 1 | Platform | `internal/cel/generator.go` | — | CELGenerator.Generate() — LLM → CEL |
| 2 | Platform | `internal/cel/evaluator.go` | — | CELEvaluator.Evaluate() — cel-go interpreter |
| 3 | Platform | `internal/cel/validator.go` | — | CELValidator.Validate() — syntax + type check |
| 4 | Platform | `internal/eligibility/` | `engagement_offers.eligibility_cel` | Eligibility: Check(offerId, userId) → CEL evaluation |
| 5 | Platform | `internal/engagement/flow/` | `engagement_flows.steps[].condition_expr` | Flow: condition_check step → CEL evaluation |
| 6 | Platform | `internal/recommendations/` | `recommendation_rules.segment_cel` | Recommendations: segment matching → CEL evaluation |
| 7 | Platform | `internal/recommendations/` | `recommendation_rules.scoring_cel` | Recommendations: scoring calculation → CEL evaluation |
| 8 | Platform | `internal/api/` | `/api/v1/cel/generate` | Admin API: LLM generation endpoint |
| 9 | Billing | `rule_engine/` | `billing_rules.condition_cel` | Billing: filter rules → CEL evaluation |
| 10 | Platform | `internal/api/` | `/api/v1/cel/validate` | Admin API: CEL syntax validation endpoint |
| 11 | Platform | `internal/api/` | `/api/v1/cel/preview` | Admin API: test CEL on sample context |
| 12 | Platform | `internal/compliance/` | `compliance_policies.retention_cel` | Compliance: data retention rules → CEL evaluation (Phase C) |
| 13 | Platform | `internal/gamification/` | `achievements.condition_cel` | Gamification: CEL-условия присвоения ачивок + уровней лояльности |
| 14 | Platform | `internal/api/` | `/gamification/v1/cel/generate-achievement` | Admin API: LLM генерация CEL для ачивок |
| 15 | Platform | `internal/api/` | `POST /api/v1/cel/dry-run` | **P1:** Dry-run на конкретном userId |
| 16 | Platform | `internal/api/` | `POST /api/v1/cel/dry-run/batch` | **P1:** Batch dry-run — покрытие по tenant'у |

---

## CEL Sandbox — Dry-Run (P1)

> **P1:** endpoint для тестирования CEL-правил на реальных данных пользователя **без применения**.

### Проблема

HR вводит условие «бесплатный фитнес для директоров». Без проверки правило может:
- Сработать для 0 пользователей (опечатка в CEL)
- Сработать для всех (слишком широкое условие → финансовые потери)
- Вызвать ошибку evaluation (неизвестное поле)

### `POST /api/v1/cel/dry-run` — тест на одном пользователе

```http
POST /api/v1/cel/dry-run
Authorization: Bearer <JWT>  (hr | catalog_manager | admin)

{
  "cel_expression": "user.grade == 'Director' && tags.is_remote == true",
  "user_id": "uuid-ivanov",
  "domain": "billing"
}
```

**Ответ:**

```json
{
  "result": true,
  "context": {
    "user": { "grade": "Director", "years_of_service": 7 },
    "tags": { "is_remote": true, "is_senior": true }
  },
  "evaluation_time_ms": 2,
  "warnings": []
}
```

### `POST /api/v1/cel/dry-run/batch` — проверка покрытия

```json
{
  "cel_expression": "user.grade in ['A', 'B'] && user.years_of_service >= 3",
  "tenant_id": "sdek"
}
```

**Ответ:**

```json
{
  "matched_count": 1247,
  "total_users": 10000,
  "coverage_percent": 12.47,
  "sample_matched": [
    { "user_id": "uuid-1", "grade": "A", "years_of_service": 5 },
    { "user_id": "uuid-2", "grade": "B", "years_of_service": 3 }
  ],
  "sample_unmatched": [
    { "user_id": "uuid-3", "grade": "C", "reason": "grade=C not in [A,B]" },
    { "user_id": "uuid-4", "grade": "A", "years_of_service": 1, "reason": "years_of_service=1 < 3" }
  ]
}
```

> **⚠️ Ограничение:** batch dry-run сканирует все users tenant'а. Rate limit: 1 req/min per admin. Для tenant'ов > 50K users — асинхронный режим (Asynq job).

### Audit trail правил

Каждое правило хранит историю изменений:

```sql
-- Версионирование правил (P1)
CREATE TABLE billing_rule_versions (
    id              SERIAL PRIMARY KEY,
    rule_id         UUID NOT NULL REFERENCES billing_rules(id),
    version         INT NOT NULL,
    condition_cel   TEXT,
    condition_source TEXT,
    changed_by      UUID NOT NULL REFERENCES users(id),
    changed_at      TIMESTAMPTZ DEFAULT now(),
    change_reason   TEXT
);

CREATE INDEX idx_billing_rule_versions_rule ON billing_rule_versions(rule_id, version DESC);

-- Расширение billing_rules
ALTER TABLE billing_rules
  ADD COLUMN version         INT DEFAULT 1,
  ADD COLUMN updated_by      UUID REFERENCES users(id),
  ADD COLUMN updated_at      TIMESTAMPTZ,
  ADD COLUMN audit_log       JSONB DEFAULT '[]'::jsonb;
```

---

## Фазы внедрения

### Фаза A — Core (HIGH prio)

| Таблицa | Поле новое | Точка применения |
|---------|---------|-------------|
| `billing_rules` | `condition_cel`, `condition_source`, **P1:** `version`, `updated_by`, `updated_at`, `audit_log` | Billing engine: фильтр правил перед применением |
| `billing_rules` | **P1:** `billing_rule_versions` (новая таблица) | Audit trail версий правил |
| `engagement_offers` | `eligibility_cel`, `eligibility_source` | Eligibility engine: Check(offerId, userId) |
| Admin UI | `/billing/v1/rules` textarea | HR: "писать на русском → CEL" |
| API | `POST /api/v1/cel/generate` | LLM generation endpoint |
| API | **P1:** `POST /api/v1/cel/dry-run` | Тест на конкретном userId |
| API | **P1:** `POST /api/v1/cel/dry-run/batch` | Проверка покрытия по tenant'у |

**Результат:** Billing + Eligibility работают на CEL. LLM генерирует CEL для Admin UI. HR тестирует правила перед сохранением.

### Фаза B — Recommendations + Flow (MEDIUM prio)

| Таблица | Поле новое | Точка применения |
|---------|---------|-------------|
| `recommendation_rules` | `segment_cel`, `scoring_cel` | Recommendations engine: segment matching + scoring |
| `engagement_flows` | `steps[].condition_source` | Flow engine: condition_check evaluation |
| Admin UI | `/admin/recommendations/rules` textarea | Менеджер: правила рекомендаций на русском |

**Результат:** Recommendations + Flow conditions мигрированы. 4 домена используют CEL. (5-й — Gamification, M09 ADR-023)

### Фаза C — Compliance (LOW prio)

| Таблица | Поле новое | Точка применения |
|---------|---------|-------------|
| `compliance_policies` | `retention_cel` | Compliance: data retention rules |
| `consents` | `auto_renewal_cel` | Consent: auto-renewal conditions |

**Результат:** Полная миграция. Старые форматы удалены.

---

## Metrics

| Метрика | Тип | Описание |
|--------|-----|------|
| `cel_generation_total{status="success"\|"error"\|"drift"}` | Counter | LLM CEL генерации |
| `cel_evaluation_total{domain="billing"\|"eligibility"\|"flow"\|"recommendations"\|"gamification"}` | Counter | CEL evaluation вызовы по доменам |
| `cel_evaluation_duration_seconds{domain}` | Histogram | Latency CEL evaluation по доменам |
| `cel_validation_errors_total` | Counter | Ошибки парсинга CEL |
| `cel_custom_functions_total{fn="date_diff_days"\|"str_contains"\|"now_iso"}` | Counter | Вызовы custom CEL functions |
| `gamification_achievement_awards_total{achievement_key}` | Counter | Кол-во присвоений ачивок по ключам |
| `gamification_loyalty_upgrades_total{from_level, to_level}` | Counter | Переходы между уровнями лояльности |
| `gamification_check_duration_seconds` | Histogram | Время проверки условий присвоения на пользователя |
| `cel_dryrun_total{mode="single"\|"batch"}` | Counter | P1: вызовы dry-run |
| `cel_dryrun_duration_seconds{mode}` | Histogram | P1: latency dry-run |
| `cel_dryrun_matched_users{domain}` | Gauge | P1: кол-во совпавших users в batch dry-run |

---

## Observability

JSON-логи:
```json
{
  "ts": "2026-05-24T12:00:00Z",
  "level": "info",
  "svc": "platform",
  "cel_domain": "billing",
  "cel_expr": "user.grade in ['A','B']",
  "cel_result": true,
  "user_id": "uuid",
  "rule_id": "grade_benefit"
}
```

---

## Security

- CEL sandbox: нет доступа к Go stdlib, системному IO, сети
- Rate limit on `/api/v1/cel/generate`: 10 req/min per user (LLM call cost)
- LLM tokens не логгируются (только request/response hash)
- CEL expression length limit: 4096 chars
- Pre-compile + cache: каждый уникальный CEL compile-ется один раз, результат кэшируется в Redis (TTL 24h)
