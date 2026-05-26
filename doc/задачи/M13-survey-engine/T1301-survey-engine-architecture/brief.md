# T1301 — Архитектура Survey Engine

## Контекст

Платформа LKFL использует `EngagementType(type: "activity")` для представления опросов:
- Flow: `form` step → вопросы → `condition_check` → credit (баллы)
- Пример в `архитектура/engagement.md` §"Опрос Q2" (строки 769-849)

**Существующий минимальный кейс работает**, но предложение по Survey Module выявило три системных пробела:

### Пробел 1: Tag Mapping от ответов (КРИТИЧЕСКИЙ)

TagResolver в `cel/` вычисляет теги **только из профиля HR-данных**:
```
Сейчас:  TagResolver → user.profile → { is_remote, is_senior, tenure_3y }
Нужно:  TagResolver → user.profile + survey.answers → { is_remote, interest:sport, lifestyle:family }
```

Без этого опросы собирают данные, но они **не влияют** на:
- eligibility-правила (кто допускается к какой льготе)
- billing rules (какие правила действуют)
- рекомендации (какие льготы показывать)
- segments (кем таргетируются активности)

CEL работает на `{user.*, tags, game.*}`. Survey-ответы не попадают ни в один из этих блоков.

### Пробел 2: Бранчинг вопросов

`form_schema` — плоский список вопросов. Нет условий «покажи вопрос Б если ответ А»:

```yaml
# Неподдерживается:
questions:
  - id: "q1"
    text: "Что предпочитаешь? (Спорт/Кино/Книги)"
  - id: "q2"
    text: "Как часто занимаешься спортом?"
    show_if: 'q1 == "Спорт"'      # ← не работает
  - id: "q3"
    text: "Какой жанр любишь?"
    show_if: 'q1 == "Кино"'       # ← не работает
```

### Пробел 3: Агрегация результатов

Нет endpoint'а для HR/админа:
```
GET /admin/engagements/:offerId/survey-analytics
  → { totalResponses, byQuestion[], topTags[] }
```

Менеджер каталога не может валидировать спрос перед закупкой:
```
Опрос: «Интересна ли тема мужского здоровья?»
  → За > 10% → Закупка обоснована
  → За < 1% → Бюджет экономим
```

## Связь с существующей архитектурой

- `архитектура/engagement.md` — Engagement/activity-модель, flow-шаги для опросов
- `архитектура/теги.md` — TagResolver, каталог тегов (только profile-based)
- `архитектура/cel-engine.md` — CEL engine, 5 доменов (billing, eligibility, flow, game, recommendations)
- `архитектура/пакеты-platform.md` — 13 internal пакетов (после M12)
- `контекст/проблема.md` — «паралич выбора», «низкий engagement»
- `спецификация/journeys/hr.md` — HR-кейсы создания активностей
- `спецификация/journeys/сотрудник.md` — J12: сотрудник проходит опрос

---

## Задача

### 1. Написать ADR-025

Путь: `doc/архитектура/adr/025-survey-engine.md`

Содержание:
- **Контекст:** опросы как `activity` — рабочий минимум, но не полноценный инструмент
- **Варианты:**
  - Вариант А: подпакет `engagement/survey/` (выбран)
  - Вариант Б: отдельный пакет `internal/survey/` (отклонён — см. таблицу обоснования в overview.md)
- **Решение:** подпакет с TagMapper, Resolver, Analytics
- **Последствия:** расширение tag cache, новая БД-таблица, новый admin endpoint

### 2. Определить `survey_schema`

Расширить `form_schema` до `survey_schema`:

```yaml
survey_schema:
  version: 1
  questions:
    - id: "q1_interest"
      text: "Что тебе интересно? Выбери максимум 2"
      type: "multiple_choice"
        # single_choice | multiple_choice | scale | text
      required: true
      options:
        - value: "sport"
          label: "Спорт"
        - value: "cinema"
          label: "Кино"
        - value: "books"
          label: "Книги"
        - value: "food"
          label: "Еда"
      # ВЕТВЛЕНИЕ (опционально)
      branch_on:
        "sport": ["q2_sport_freq", "q3_sport_type"]
        "cinema": ["q2_cinema_genre"]
        "books": ["q2_books_genre"]
        "food": ["q2_food_type"]
      # TAG MAPPING (опционально)
      tag_mappings:
        "sport": { tag: "interest:sport", weight: 1.0 }
        "cinema": { tag: "interest:cinema", weight: 0.8 }
        "books": { tag: "interest:books", weight: 0.7 }
        "food": { tag: "interest:food", weight: 0.9 }

    - id: "q2_sport_freq"
      text: "Как часто занимаешься спортом?"
      type: "single_choice"
      required: false           # shown only when branch_on includes it
      branchable_by: "q1_interest"
      options:
        - value: "daily"
          label: "Ежедневно"
        - value: "weekly"
          label: "Раз в неделю"
        - value: "monthly"
          label: "Раз в месяц"
      tag_mappings:
        "daily": { tag: "sport_intensity", value: "high", weight: 0.9 }
        "weekly": { tag: "sport_intensity", value: "medium", weight: 0.7 }
        "monthly": { tag: "sport_intensity", value: "low", weight: 0.5 }
```

**Ключевые отличия от `form_schema`:**

| Поле | form_schema | survey_schema |
|------|-|-|
| types | text, number, date, select, textarea | **single_choice, multiple_choice, scale, text** |
| branch_on | нет | **да** — ветвление по ответу |
| tag_mappings | нет | **да** — ответ → тег с весом |
| options.format | value (select) | value/label (унифицировано) |
| range | нет | **да** — для scale [1, 5] |

**Обратная совместимость:** `form_schema` сохраняется. При `ui_component: "SurveyForm"` → парсится `survey_schema`. При `ui_component: "EngagementForm"` → парсится `form_schema` (legacy).

### 3. Спроектировать `survey/Resolver` + State Management

Путь: `internal/engagement/survey/resolver.go`

```go
package survey

// Resolver рендерит вопросы для конкретного пользователя
// Учёт бранчинга: показывает только видимые для текущего контекста вопросы
// State (answers) persists между HTTP-вызовами через user_engagement.form_data (JSONB)
type Resolver struct {
    schema  *SurveySchema
    answers map[string]any  // заполненные ответы — hydrate из form_data
}

// NewResolverWithState создаёт Resolver с восстановленным state из БД
// FlowEngine читает form_data (JSONB) → десериализует → передаёт в NewResolverWithState
func NewResolverWithState(schema *SurveySchema, savedAnswers map[string]any) *Resolver

// GetNextQuestion возвращает следующий видимый вопрос
// bool = true если есть вопросы, false если все заполнены
func (r *Resolver) GetNextQuestion() (*Question, bool)

// SubmitAnswer обрабатывает ответ пользователя
// Обновляет answers, пересчитывает видимые вопросы, убирает старые ветки
func (r *Resolver) SubmitAnswer(questionId string, answer any) error

// IsComplete возвращает true когда все required вопросы заполнены
// required определяется после применения бранчинга: скрытые вопросы не required
func (r *Resolver) IsComplete() bool

// GetFinalAnswers возвращает финальные ответы для сохранения в form_data
func (r *Resolver) GetFinalAnswers() map[string]any
```

**State Persistence (Решение критичного пробела):**

Resolver — stateful объект, но HTTP без state. Решение — persist answers в `user_engagement.form_data` (JSONB) между вызовами:

```
HTTP Call 1: POST /user-engagements/:id/execute-step
  → FlowEngine.ExecuteStep reads user_engagement.form_data (null/empty)
  → NewResolverWithState(schema, nil)
  → SubmitAnswer("q1", "sport")
  → answers = {"q1": "sport"}
  → FlowEngine saves answers back to user_engagement.form_data (JSONB)
  → returns: next_question = "q2_sport_freq"

HTTP Call 2: POST /user-engagements/:id/execute-step
  → FlowEngine.ExecuteStep reads user_engagement.form_data
  → form_data = {"q1": "sport"}
  → NewResolverWithState(schema, {"q1": "sport"})
  → Resolver knows q1="sport" → branch_on activates ["q2_sport_freq", ...]
  → SubmitAnswer("q2_sport_freq", "daily")
  → answers = {"q1": "sport", "q2_sport_freq": "daily"}
  → FlowEngine saves updated answers back to form_data
  → returns: next_question = "q3_sport_type"
```

**Ключевое:** FlowEngine сохраняется intermediate form_data после каждого SubmitAnswer. Это минимальное изменение в flow.go — добавить один UPDATE после resolver.SubmitAnswer.

**Алгоритм бранчинга:**
```
1. Загрузить question q1 → показать пользователю
2. q1 -> branch_on -> "Спорт": [q2_sport, q3_sport_type]
3. Пользователь ответил "Спорт" → включить q2_sport, q3_sport_type
4. q2_cinema, q2_books → скрыты (не в include, required == false)
5. Если скрытый вопрос required == true → validation error при создании схемы
```

**Семантика multiple_choice + branch_on (OR-union):**
```
Ответ ["sport", "cinema"]:
  → объединить ВСЕ ветки: ["q2_sport_freq", "q3_sport_type", "q2_cinema_genre"]
  → все включённые вопросы становятся required (если они were required in original branch)
  → user получает объединённый набор под-вопросов
```

**Семантика text (свободный ввод) + branch_on:**
```
text question НЕ поддерживает branch_on (нет дискретных options)
→ validation error при попытке определить branch_on на text question
```

**Валидация при создании схемы:**
- branch_on не ссылается на несуществующий question
- branchable_by ссылается на существующий question
- required вопросы достижимы (не скрыты branch_on другим required)
- tag_mapping option.value существует в options[]
- branch_on на text question → error (нестандартный тип ответа)

### 4. Спроектировать `survey/TagMapper`

Путь: `internal/engagement/survey/tag_mapper.go`

```go
package survey

type TagMapper struct {
    tagResolver *cel.TagResolver
    db          *pgxpool.Pool
    logger      Logger
}

// MapSurveyAnswers конвертирует ответы опроса в пользовательские теги
//
// Flow (вызывается один раз при flow completion — resolver.IsComplete() == true):
// 1. Прочитать survey_schema для оффера → question[].tag_mappings
// 2. Вычислить expected-теги из ответов (value → tag key + value + weight)
// 3. Получить existing-теги из user_survey_attributes WHERE user_id + survey_offer_id
// 4. Удалить orphaned теги: existing ∉ expected (ответ изменился → старый тег удалить)
// 5. INSERT/UPDATE expected-теги:
//    INSERT ON CONFLICT (tenant_id, user_id, survey_offer_id, tag_key)
//      DO UPDATE SET weight = GREATEST(EXCLUDED.weight, weight)
//    (если weight нового ответа > weight существующего → update, иначе оставить старый)
// 6. Invalidate tag cache в Redis (tag_cache:{tenant_id}:{user_id})
func (m *TagMapper) MapSurveyAnswers(
    ctx context.Context,
    tenantId uuid.UUID,  // из JWT middleware
    userId uuid.UUID,
    surveyOfferId uuid.UUID,
    answers map[string]any,
) ([]SurveyTag, error)

type SurveyTag struct {
    TagKey   string   // "interest:sport"
    TagValue any      // "running", 5, "high"
    Weight   float64  // 0.0 - 1.0
}
```

**Управление жизненным циклом тегов (answer change):**

```
Сценарий: user ответил q1="sport" на опросе А → тег interest:sport weight=0.9.
Потом перешёл опрос заново и ответил q1="cinema".

TagMapper.MapSurveyAnswers:
  expected = { interest:cinema@0.8 }     (из нового ответа)
  existing = { interest:sport@0.9 }      (из БД, старый ответ)
  orphaned = existing - expected = { interest:sport }
  → DELETE WHERE tag_key = 'interest:sport' AND survey_offer_id = 'survey:A'
  → INSERT interest:cinema@0.8
```

**Логика весов при дублировании (перекрестные опросы):**

```
Пользователь прошёл 3 опроса:
  Q1 опроса А: interest:sport = 0.9
  Q2 опроса Б: interest:sport = 0.6
  Q3 опроса В: interest:cinema = 0.8

user_survey_attributes (хранит все):
  user:ivanov | survey:A | interest:sport | 0.9
  user:ivanov | survey:B | interest:sport | 0.6
  user:ivanov | survey:C | interest:cinema | 0.8

TagResolver.Aggregate возвращает:
  interest:sport = MAX(0.9, 0.6) = 0.9
  interest:cinema = 0.8
```

**Важно:** deletion orphaned тегов работает только внутри одного survey_offer_id.
Перекрёстные теги (один tag_key из разных опросов) не удаляются — их aggregation делает
TagResolver.AggregateSurveyTags через GROUP BY tag_key.

### 5. Расширить `cel/TagResolver`

Путь: `internal/cel/tag_resolver.go` (добавить методы)

```go
package cel

// NEW: агрегация тегов из survey-ответов
// tenantId из JWT middleware (ctx value)
func (r *TagResolver) AggregateSurveyTags(
    ctx context.Context,
    tenantId uuid.UUID,
    userId uuid.UUID,
) map[string]interface{} {
    // 1. SELECT tag_key, tag_value, MAX(weight) FROM user_survey_attributes
    //    WHERE tenant_id = ? AND user_id = ? GROUP BY tag_key, tag_value
    // 2. Return map[string]interface{}  ← совместим с CELContext.Tags
}
```

**Изменение `Resolve()` (minimal diff):**
```go
func (r *TagResolver) Resolve(ctx context.Context, user *User) map[string]interface{} {
    tags := r.computeFromProfile(user)  // существующий: is_remote, is_senior, ...
    surveyTags := r.AggregateSurveyTags(ctx, user.TenantID, user.ID)  // NEW
    for k, v := range surveyTags {
        tags[k] = v  // survey-теги поверх profile-тегов
    }
    return tags
}
```

> **Примечание:** `InvalidateCache(userId)` — существующий метод `cel/TagResolver` (см. `теги.md` §Кэширование).
> При MapSurveyAnswers TagMapper вызывает InvalidateCache → Redis key `tag_cache:{tenant_id}:{user_id}` удаляется.
> Это не новая фича M13, а использование существующего контракта.

### 6. DB Schema: `user_survey_attributes`

> **ADR-024 compliant:** `tenant_id` — обязательное поле в каждой бизнес-таблице.

```sql
CREATE TABLE user_survey_attributes (
    id              SERIAL PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
      -- multi-tenancy: изоляция данных tenant'a
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    survey_offer_id UUID NOT NULL REFERENCES engagement_offers(id) ON DELETE CASCADE,
    tag_key         VARCHAR(128) NOT NULL,
      -- 'interest:sport', 'lifestyle:family', 'sport_intensity'
    tag_value       VARCHAR(256) NOT NULL,
      -- 'running', 'high', 'with_kids'
    weight          FLOAT NOT NULL DEFAULT 1.0 CHECK (weight >= 0 AND weight <= 1),
    question_id     VARCHAR(64) NOT NULL,
      -- ID вопроса из survey_schema (для tracing)
    answered_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, user_id, survey_offer_id, tag_key)
);

-- Индексы: tenant_id первым для partition pruning
CREATE INDEX idx_survey_attrs_tenant_user ON user_survey_attributes(tenant_id, user_id);
  -- для AggregateSurveyTags: WHERE tenant_id=? AND user_id=? GROUP BY
CREATE INDEX idx_survey_attrs_tenant_tag ON user_survey_attributes(tenant_id, tag_key);
  -- для analytics по тегам в рамках tenant'а
CREATE INDEX idx_survey_attrs_tenant_offer ON user_survey_attributes(tenant_id, survey_offer_id);
  -- для analytics по опросу + orphaned tag cleanup
```

### 7. Analytics API

```go
// survey/analytics.go

type AnalyticsEngine struct {
    db     *pgxpool.Pool
    logger Logger
}

type SurveyAnalytics struct {
    TotalResponses int                    `json:"totalResponses"`
    CompletionRate float64                `json:"completionRate"`
    AvgScore       map[string]float64     `json:"avgScore"`  // для scale-вопросов: среднее по ответам
    Distribution   map[string]any         `json:"distribution"`  // per question: value → count
    TopTags        []SurveyTagCount       `json:"topTags"`
}

type SurveyTagCount struct {
    TagKey   string  `json:"tagKey"`
    Count    int     `json:"count"`
    AvgWeight float64 `json:"avgWeight"`
}

// GetSurveyAnalytics агрегирует данные из трёх источников:
// - user_engagements (flow_status): TotalResponses, CompletionRate
// - user_engagements.form_data (JSONB): Distribution (распаковка ответов по вопросам)
// - user_survey_attributes: TopTags (агрегация тегов)
func (a *AnalyticsEngine) GetSurveyAnalytics(
    ctx context.Context,
    surveyOfferId uuid.UUID,
) (*SurveyAnalytics, error)
```

**Источники данных (явная спецификация):**

| Поле SurveyAnalytics | Источник | SQL |
|---|--|--|
| `TotalResponses` | `user_engagements` | `SELECT COUNT(*) FROM user_engagements WHERE offer_id = ? AND flow_status IN ('in_progress', 'completed')` |
| `CompletionRate` | `user_engagements` | `SUM(CASE flow_status = 'completed' THEN 1 ELSE 0 END) / NULLIF(TotalResponses, 0)` |
| `Distribution[qId]` | `user_engagements.form_data->qId` (JSONB) | `SELECT fd->'q1' as val, COUNT(*) FROM user_engagements WHERE offer_id=? AND flow_status='completed' GROUP BY val` |
| `AvgScore[qId]` | `user_engagements.form_data` → cast to float | только для type=scale вопросов |
| `TopTags[]` | `user_survey_attributes` | `SELECT tag_key, COUNT(*), AVG(weight) FROM user_survey_attributes WHERE survey_offer_id=? GROUP BY tag_key ORDER BY count DESC` |

HTTP endpoint (admin):
```
GET  /admin/engagements/:offerId/survey-analytics
     → SurveyAnalytics (json)
     RBAC: hr, catalog_manager, admin
     tenant_id из JWT middleware (auto-join engagement_offers.tenant_id)
     Только для type=activity с survey_schema (validation при запросе)
```

**Ответ для кейса «Сбор спроса»:**
```json
{
  "totalResponses": 342,
  "completionRate": 0.67,
  "distribution": {
    "q1_health": [
      {"value": "yes", "label": "Да", "count": 156},
      {"value": "no", "label": "Нет", "count": 42},
      {"value": "maybe", "label": "Не знаю", "count": 144}
    ]
  },
  "topTags": [
    {"tagKey": "interest:mens_health", "count": 156, "avgWeight": 0.95}
  ]
}
```

Решение менеджера: `156 / 342 = 45.6% > 10%` → закупка обоснована.

### 8. Интеграция в FlowEngine

```go
// engagement/flow/flow_engine.go — minimal diff в ExecuteStep()

func (f *FlowEngine) ExecuteStep(ctx context.Context, engagementId uuid.UUID, stepId uuid.UUID, data map[string]string) error {
    step := f.getStep(engagementId, stepId)
    ue := f.getUserEngagement(ctx, engagementId) // existing
    tenantId := ue.TenantID  // existing
    userId := ue.UserID      // existing
    offerId := ue.OfferID    // existing

    if step.Type == "form" && step.UIComponent == "SurveyForm" {
        // NEW: survey form processing
        schema := parseSurveySchema(step.FormSchema)

        // HYDRATE: восстановить state из form_data (JSONB) — между HTTP-вызовами
        savedAnswers := ue.FormData  // map[string]any из JSONB (deserialize)
        resolver := survey.NewResolverWithState(schema, savedAnswers)

        if err := resolver.SubmitAnswer(step.ID, data["answer"]); err != nil {
            return err
        }

        // PERSIST: сохранить intermediate answers обратно в form_data
        // (чтобы при следующем HTTP-вызове state восстановился)
        ue.FormData = resolver.GetFinalAnswers()
        if err := f.updateUserEngagementFormData(ctx, engagementId, ue.FormData); err != nil {
            return err
        }

        if !resolver.IsComplete() {
            return nil  // ещё есть вопросы → вернуть следующий
        }

        // Все вопросы заполнены → map tags
        return f.tagMapper.MapSurveyAnswers(ctx, tenantId, userId, offerId, resolver.GetFinalAnswers())
    }
    // ... existing logic для form, approval, provider_api ...
}
```

### 9. Обновить документацию

| Файл | Что добавить |
|------|-|
| `архитектура/модули.md` | `engagement/survey/` в подпакете, DI граф + analytics endpoint |
| `архитектура/пакеты-platform.md` | Подпакет survey/ в описании engagement/, types.go, resolver.go, tag_mapper.go, analytics.go |
| `архитектура/теги.md` | Новая категория: survey-теги (источник: user_survey_attributes, агрегация: MAX weight) |
| `архитектура/cel-engine.md` | CELContext.Tags — теперь включает survey-теги (источник: TagResolver.AggregateSurveyTags) |
| `архитектура/engagement.md` | survey_schema в flow step, пример "Опрос Q2" расширен |
| `задачи/README.md` | Добавить M13 в таблицу вех |

---

## Ожидаемый результат

**Важно: задача M13 — только документация и архитектурное проектирование. Код не создаётся.**

После выполнения T1301 в `doc/`:
1. ADR-025 записан в `doc/архитектура/adr/025-survey-engine.md`
2. `survey_schema` определён как спецификация в АDR (types, branch_on, tag_mappings, обратная совместимость с form_schema)
3. Go API подпакета `engagement/survey/` описан в ADR (Resolver, TagMapper, AnalyticsEngine — сигнатуры + алгоритмы)
4. Интеграция с TagResolver (`AggregateSurveyTags` + `Resolve` merge) описана в ADR
5. Интеграция с FlowEngine (ExecuteStep survey processing) описана в ADR
6. DB schema `user_survey_attributes` определена в миграции-шаблоне для будущего использования
7. Analytics API определён (SurveyAnalytics, endpoint, RBAC) в ADR
8. Документация обновлена (модули.md, пакеты-platform.md, теги.md, cel-engine.md, engagement.md, README.md)

**Результат — готовая спецификация для M14 (реализация кода).** Реализация Go-кода вынесена за пределы M13.
