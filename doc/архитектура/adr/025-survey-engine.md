# ADR-025: Survey Engine — полноценный модуль опросов в подпакете engagement/survey/

**Статус:** Accepted
**Дата:** 2026-05-25
**Контекст:** M13-survey-engine, T1301

---

## Контекст

Платформа обрабатывает опросы как `EngagementType(type: "activity")` с flow-шагами `form` + `condition_check`. Этот минимум покрывает кейс «опрос → баллы», но не закрывает три системных пробела:

### Пробел 1 — Tag Mapping от ответов (КРИТИЧЕСКИЙ)

TagResolver в `cel/` вычисляет теги исключительно из HR-профиля:

```
Сейчас:  TagResolver → user.profile → { is_remote, is_senior, tenure_3y }
Нужно:  TagResolver → user.profile + survey.answers → { is_remote, interest:sport, lifestyle:family }
```

Без этого опросы собирают данные, которые **не влияют** на:
- eligibility-правила (кто допускается к какой льготе)
- billing rules (какие правила действуют)
- рекомендации (какие льготы показывать)
- segments (кем таргетируются активности)

CEL работает на `{user.*, tags, game.*}`. Survey-ответы не попадают ни в один из этих блоков.

### Пробел 2 — Бранчинг вопросов

`form_schema` — плоский список вопросов без условий видимости. Нет `show_if`, нет ветвления по ответу. Невозможно спросить «Какой жанр любишь?» только если пользователь ответил «Кино» на предыдущий вопрос.

### Пробел 3 — Агрегация результатов

Нет endpoint'а для HR/админа `GET /admin/engagements/:offerId/survey-analytics`. Менеджер каталога не может валидировать спрос перед закупкой.

---

## Рассмотренные варианты

### Вариант А: Подпакет `engagement/survey/` (ВЫБРАН)

Survey Engine как подпакет внутри `internal/engagement/survey/`. Опросы остаются `type: "activity"`, но получают полноценный движок с бранчингом, tag-mapping и аналитикой.

| Плюсы | Минусы |
|---|---|
| Flow lifecycle обрабатывает survey напрямую — нет NATS/gRPC | engagement/ получает внутреннюю зависимость |
| Биллинг `engagement/flow/` → `billing.Credit()` — уже есть, без изменений | Нужно разделить form processing (legacy → survey) |
| Геймификация `engagement/flow/` → `gamification/TriggerHandler` — уже есть | survey/schema должна быть backward-compat с form_schema |
| БД таблицы в `lkfl_platform` — одна БД, одна tx | |
| Survey = подтип activity, не отдельный домен | |

**Вердикт:** ✅ Выбран. Минимальный diff в архитектуре, максимальное переиспользование контрактов.

### Вариант Б: Отдельный `internal/survey/`

Самостоятельный пакет на уровне `internal/` с собственными таблицами и lifecycle.

| Плюсы | Минусы |
|---|---|
| Чистая граница ответственности | Нужен NATS/gRPC между survey/ → billing/, survey/ → gamification/ |
| Независимое тестирование | Транзитные зависимости: survey/ → billing.Credit() невозможно без NATS |
| Подготовленность к future split | Оверкилл: survey — подтип activity, не самостоятельный домен |
| | Собственные таблицы → проблема cross-join с user_engagements |

**Вердикт:** ❌ Отклонён. Survey — подтип activity, не отдельный домен. Overhead отдельного пакета неоправдан.

---

## Решение

Подпакет `internal/engagement/survey/` с четырьмя файлами:

```
internal/engagement/
├── catalog/
│   ├── types_engine.go
│   ├── offer_engine.go
│   ├── category_engine.go
│   ├── search.go
│   ├── ab_test.go
│   └── proposals.go
├── flow/
│   ├── flow_engine.go
│   ├── user_engagement.go
│   ├── billing_events.go
│   ├── approval.go
│   └── document.go
├── collections/
│   └── collection_engine.go
└── survey/
    ├── types.go        # SurveySchema, Question, QuestionType, TagMapping, SurveyTag
    ├── resolver.go     # SurveyResolver — рендер с бранчингом + hydration из form_data
    ├── tag_mapper.go   # TagMapper — ответ → тег с весом (lifecycle: insert/update/delete orphaned)
    └── analytics.go    # Analytics — агрегация (user_engagements + form_data + user_survey_attributes)
```

### Диаграмма зависимостей

```
engagement/survey/
    ├── SurveySchema (types) ← EngagementFlow.step.form_schema при ui_component="SurveyForm"
    ├── Resolver         ← FlowEngine.ExecuteStep hydrates state из form_data (JSONB)
    ├── TagMapper ──────→ cel/TagResolver.InvalidateCache(userId)
    └── Analytics ──────→ admin API: GET /admin/engagements/:offerId/survey-analytics

cel/TagResolver.Resolve()
    ├── computeFromProfile(user)     ← существующий: is_remote, is_senior, ...
    └── AggregateSurveyTags(ctx, tenantId, userId)  ← НОВЫЙ: interest:sport, ...

cel/TagResolver.Resolve() → tags = merge(profile_tags, survey_tags)
```

### Ключевые решения

| Решение | Обоснование |
|---|-|
| State в `user_engagement.form_data` (JSONB) | Resolver stateful, HTTP stateless. Persist между вызовами. Минимальный diff в flow.go. |
| `tenant_id` в DB schema | ADR-024: каждая бизнес-таблица содержит tenant_id. Partition pruning. |
| `GREATEST(new, existing)` при UPDATE | Сохраняет максимальный вес при перепрохождении опроса. |
| Orphaned tag deletion | User ответил "sport" → перешёл на "cinema". Старый тег `interest:sport` удаляется. |
| multiple_choice + branch_on = OR-union | Выбор ["sport", "cinema"] → объединение всех веток. |
| text question + branch_on = error | Нет дискретных options → branch_on не применим. |

---

## Спецификация `survey_schema`

`survey_schema` — расширение `form_schema` для опросов. Сохраняется backward compatibility: `ui_component: "SurveyForm"` → парсится `survey_schema`; `ui_component: "EngagementForm"` → парсится `form_schema` (legacy).

### Полный YAML-формат

```yaml
survey_schema:
  version: 1
  questions:
    # === ВОПРОС 1: multiple_choice с branch_on + tag_mappings ===
    - id: "q1_interest"
      text: "Что тебе интересно? Выбери максимум 2"
      type: "multiple_choice"
        # Допустимые типы: single_choice | multiple_choice | scale | text
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
      # ВЕТВЛЕНИЕ: ответ → список ID показываемых вопросов
      branch_on:
        "sport": ["q2_sport_freq", "q3_sport_type"]
        "cinema": ["q2_cinema_genre"]
        "books": ["q2_books_genre"]
        "food": ["q2_food_pref"]
      # TAG MAPPING: ответ → пользовательский тег с весом
      tag_mappings:
        "sport": { tag: "interest:sport", weight: 1.0 }
        "cinema": { tag: "interest:cinema", weight: 0.8 }
        "books": { tag: "interest:books", weight: 0.7 }
        "food": { tag: "interest:food", weight: 0.9 }

    # === ВОПРОС 2: single_choice, зависимый от q1 ===
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

    # === ВОПРОС 3: scale ===
    - id: "q3_life_satisfaction"
      text: "Оцени своё здоровье от 1 до 5"
      type: "scale"
      required: true
      range: [1, 5]
      tag_mappings:
        "1": { tag: "health_status", value: "poor", weight: 0.3 }
        "2": { tag: "health_status", value: "poor", weight: 0.5 }
        "3": { tag: "health_status", value: "medium", weight: 0.6 }
        "4": { tag: "health_status", value: "good", weight: 0.8 }
        "5": { tag: "health_status", value: "excellent", weight: 1.0 }

    # === ВОПРОС 4: text (свободный ввод) ===
    - id: "q4_additional"
      text: "Что ещё предложишь для корпоративных льгот?"
      type: "text"
      required: false
```

### Типы вопросов

| type | Ответ | branch_on | tag_mappings | Особенности |
|------|-------|-----------|-------------|------|--
| `single_choice` | одно значение из `options[].value` | ✅ (ключи = option values) | ✅ (ключи = option values) | required: true/false |
| `multiple_choice` | массив option values | ✅ (OR-union всех выбранных веток) | ✅ (все выбранные option values) | required: true/false |
| `scale` | целое число в `range[0]..range[1]` | ✅ (ключи = string числа: "1", "2", ...) | ✅ (ключи = string числа) | range обязателен |
| `text` | произвольная строка | ❌ (error при наличии) | ❌ (нет дискретных options) | required: true/false |

### Формат `tag_mappings`

| Вариант | Когда | Пример |
|---------|--|-|
| Базовый: `{ tag, weight }` | single_choice/multiple_choice/scale — tag_key содержит всё | `"sport": { tag: "interest:sport", weight: 1.0 }` |
| Расширенный: `{ tag, value, weight }` | Нужно раздельно задавать key и value | `"daily": { tag: "sport_intensity", value: "high", weight: 0.9 }` |

При сохранении в `user_survey_attributes`:
- Если `value` указан → `tag_key = tag`, `tag_value = value`
- Если `value` не указан → `tag_key = tag`, `tag_value = ""` (пустая строка по умолчанию)

### Сравнение `form_schema` vs `survey_schema`

| Поле | form_schema | survey_schema |
|------|-|-|
| types | text, number, date, select, textarea | **single_choice, multiple_choice, scale, text** |
| branch_on | нет | **да** — ветвление по ответу (single_choice, multiple_choice, scale) |
| branchable_by | нет | **да** — указывает родительский вопрос |
| tag_mappings | нет | **да** — ответ → тег с весом |
| options.format | value (select) | **value/label** (унифицировано) |
| range | нет | **да** — для scale [min, max] |

### Правила валидации схемы

Схема проходит валидацию при создании оффера (`EngagementCatalog.Create()`). 5 обязательных проверок:

1. **branch_on ссылается на существующий question** — каждая ссылка в `branch_on[value][]` должна указывать на ID существующего вопроса
2. **branchable_by ссылается на существующий вопрос** — `branchable_by` должен совпадать с `id` другого вопроса в schema
3. **required вопросы достижимы** — если вопрос `required: true` и лежит за веткой (`branchable_by`), его родитель НЕ может быть `required: false` (иначе required вопрос может никогда не показаться)
4. **tag_mapping option.value существует** — каждый ключ в `tag_mappings` должен иметь соответствующий `options[].value`
5. **branch_on на text question → error** — вопросы типа `text` не имеют дискретных options → branch_on неприменим

---

## Resolver API + алгоритмы

### Сигнатуры

```go
package survey

// Resolver рендерит вопросы для конкретного пользователя.
// Stateful: хранит answers между вызовами GetNextQuestion/SubmitAnswer.
// Hydrate из form_data (JSONB) между HTTP-запросами через FlowEngine.
type Resolver struct {
    schema  *SurveySchema
    answers map[string]any  // заполненные ответы — hydrate из form_data
}

// NewResolverWithState создаёт Resolver с восстановленным state из БД.
// FlowEngine читает form_data (JSONB) → десериализует → передаёт здесь.
func NewResolverWithState(schema *SurveySchema, savedAnswers map[string]any) *Resolver

// GetNextQuestion возвращает следующий видимый вопрос.
// bool = true если есть вопросы, false если все заполнены.
func (r *Resolver) GetNextQuestion() (*Question, bool)

// SubmitAnswer обрабатывает ответ пользователя.
// Обновляет answers, пересчитывает видимые вопросы, убирает старые ветки.
func (r *Resolver) SubmitAnswer(questionId string, answer any) error

// IsComplete возвращает true когда все required вопросы заполнены.
// required определяется после применения бранчинга: скрытые вопросы не required.
func (r *Resolver) IsComplete() bool

// GetFinalAnswers возвращает финальные ответы для сохранения в form_data.
func (r *Resolver) GetFinalAnswers() map[string]any

// ValidateSchema проверяет корректность survey_schema.
func ValidateSchema(schema *SurveySchema) error
```

### Алгоритм бранчинга

Последовательность вычисления видимых вопросов:

```
ШАГ 1: Загрузить все вопросы из schema
ШАГ 2: Инициализировать visible = {q1} (первый вопрос всегда видим)
ШАГ 3: Для каждого ответа в answers:
  3a. Найти вопрос questionId в schema
  3b. Если question.branch_on определён:
      3b.i.  Для single_choice: ответ = одно значение → включить все Q из branch_on[ответ]
      3b.ii. Для multiple_choice: ответ = []string → объединить ВСЕ ветки (OR-union)
      3b.iii.Для scale: ответ = число → string(ответ) как ключ → включить Q из branch_on[string(ответ)]
  3c. Обновить visible = visible ∪ new_branch_questions
ШАГ 4: Исключить non-visible required вопросы из required-валидации
  4a. Если вопрос скрыт (не в visible) → он не считается required
ШАГ 5: GetNextQuestion() → первый вопрос из visible, у которого нет answers[question.id]
```

### Семантика multiple_choice + branch_on (OR-union)

```
Вопрос q1 (multiple_choice): "Выбери интересы"
  branch_on:
    "sport":  [q2_sport_freq, q3_sport_type]
    "cinema": [q2_cinema_genre]

Пользователь ответил: ["sport", "cinema"]

Результат:
  visible += [q2_sport_freq, q3_sport_type, q2_cinema_genre]
  → все включённые вопросы добавлены в видимые
  → если вопрос was required in original schema → остаётся required
```

### Семантика text + branch_on

```
text question НЕ поддерживает branch_on (нет дискретных options).
→ ValidateSchema возвращает error при попытке определить branch_on на text question.
```

### State Persistence (hydration между HTTP-вызовами)

Resolver — stateful объект, но HTTP без state. State persists в `user_engagement.form_data` (JSONB):

```
HTTP Call 1: POST /user-engagements/:id/execute-step
  → FlowEngine.ExecuteStep читает user_engagement.form_data (null/empty)
  → NewResolverWithState(schema, nil)
  → SubmitAnswer("q1", "sport")
  → answers = {"q1": "sport"}
  → FlowEngine сохраняет answers обратно в user_engagement.form_data (JSONB)
  → возвращает next_question = "q2_sport_freq"

HTTP Call 2: POST /user-engagements/:id/execute-step
  → FlowEngine.ExecuteStep читает user_engagement.form_data
  → form_data = {"q1": "sport"}
  → NewResolverWithState(schema, {"q1": "sport"})
  → Resolver «помнит» q1="sport" → branch_on активирует ["q2_sport_freq", q3_sport_type]
  → SubmitAnswer("q2_sport_freq", "daily")
  → answers = {"q1": "sport", "q2_sport_freq": "daily"}
  → FlowEngine сохраняет обновлённые answers обратно в form_data
  → возвращает next_question = "q3_sport_type"
```

Ключ: FlowEngine сохраняет intermediate form_data после каждого SubmitAnswer. Минимальный diff — один UPDATE в flow.go.

---

## TagMapper API + lifecycle

### SurveyTag структура

```go
package survey

// SurveyTag — результирующий тег из опроса.
type SurveyTag struct {
    TagKey   string   // "interest:sport", "sport_intensity"
    TagValue any      // "running", "high", "" (если не указан в mapping)
    Weight   float64  // 0.0 - 1.0
}
```

### Сигнатуры TagMapper

```go
package survey

type TagMapper struct {
    tagResolver *cel.TagResolver  // для InvalidateCache
    db          *pgxpool.Pool     // для INSERT/UPDATE/DELETE
    logger      Logger
}

func NewTagMapper(tagResolver *cel.TagResolver, db *pgxpool.Pool, logger Logger) *TagMapper

// MapSurveyAnswers конвертирует ответы опроса в пользовательские теги.
// Вызывается один раз при flow completion (resolver.IsComplete() == true).
func (m *TagMapper) MapSurveyAnswers(
    ctx context.Context,
    tenantId uuid.UUID,
    userId uuid.UUID,
    surveyOfferId uuid.UUID,
    schema *SurveySchema,
    answers map[string]any,
) ([]SurveyTag, error)
```

### Lifecycle MapSurveyAnswers (5 шагов)

```
ШАГ 1: Прочитать survey_schema для оффера
  → question[].tag_mappings — какие ответы дают какие теги

ШАГ 2: Вычислить expected-теги из ответов
  → для каждого questionId в answers:
    - найти question в schema
    - если question.tag_mappings определён:
      - map answer value → SurveyTag { tag_key, tag_value, weight }

ШАГ 3: Получить existing-теги из БД
  → SELECT * FROM user_survey_attributes
    WHERE tenant_id = ? AND user_id = ? AND survey_offer_id = ?

ШАГ 4: Удалить orphaned теги
  → orphaned = existing − expected (по tag_key)
  → DELETE FROM user_survey_attributes
    WHERE tenant_id = ? AND user_id = ? AND survey_offer_id = ?
    AND tag_key = ANY(orphaned_keys)

ШАГ 5: INSERT/UPDATE expected-теги
  → INSERT INTO user_survey_attributes (tenant_id, user_id, survey_offer_id, tag_key, tag_value, weight, question_id)
    VALUES (..., ...)
    ON CONFLICT (tenant_id, user_id, survey_offer_id, tag_key)
    DO UPDATE SET weight = GREATEST(EXCLUDED.weight, weight)
  → GREATEST сохраняет максимальный вес при перепрохождении опроса

ШАГ 6: Invalidate tag cache
  → tagResolver.InvalidateCache(userId)
  → DEL tag_cache:{tenant_id}:{user_id} из Redis
```

### Управление жизненным циклом тегов (answer change)

```
Сценарий: user ответил q1="sport" на опросе Α → тег interest:sport weight=0.9.
Потом прошёл опрос заново и ответил q1="cinema".

MapSurveyAnswers (вызов 2):
  expected = { interest:cinema@0.8 }     (из нового ответа)
  existing = { interest:sport@0.9 }      (из БД, старый ответ)
  orphaned = existing − expected = { interest:sport }
  → DELETE WHERE tag_key = 'interest:sport' AND survey_offer_id = 'survey:A'
  → INSERT interest:cinema@0.8
```

**Важно:** orphaned deletion работает только внутри одного survey_offer_id.
Перекрёстные теги из разных опросов не удаляются.

### Cross-survey aggregation (перекрёстные опросы)

```
Пользователь прошёл 3 опроса:
  Q1 опроса Α: interest:sport = 0.9
  Q2 опроса Β: interest:sport = 0.6
  Q3 опроса Γ: interest:cinema = 0.8

user_survey_attributes (хранит все строки):
  user:ivanov | tenant:T | survey:Α | interest:sport | ""  | 0.9
  user:ivanov | tenant:T | survey:Β | interest:sport | ""  | 0.6
  user:ivanov | tenant:T | survey:Γ | interest:cinema| ""  | 0.8

TagResolver.AggregateSurveyTags агрегирует через:
  SELECT tag_key, tag_value, MAX(weight)
  FROM user_survey_attributes
  WHERE tenant_id=T AND user_id=ivanov
  GROUP BY tag_key, tag_value

Результат:
  interest:sport = MAX(0.9, 0.6) = 0.9
  interest:cinema = 0.8
```

---

## Analytics API

### SurveyAnalytics структура

```go
package survey

// SurveyAnalytics — агрегированные результаты опроса.
type SurveyAnalytics struct {
    TotalResponses int                    `json:"totalResponses"`
    CompletionRate float64                `json:"completionRate"`
    Distribution   map[string][]AnswerCount `json:"distribution"`  // questionId → [{value, label, count}]
    AvgScore       map[string]float64     `json:"avgScore"`  // для scale-вопросов: среднее по ответам
    TopTags        []SurveyTagCount       `json:"topTags"`
}

// AnswerCount — подсчёт одного варианта ответа на вопрос.
type AnswerCount struct {
    Value string `json:"value"`
    Label string `json:"label"`
    Count int    `json:"count"`
}

// SurveyTagCount — частота тега и его средний вес.
type SurveyTagCount struct {
    TagKey    string   `json:"tagKey"`
    Count     int      `json:"count"`
    AvgWeight float64  `json:"avgWeight"`
}
```

### Сигнатуры AnalyticsEngine

```go
package survey

type AnalyticsEngine struct {
    db     *pgxpool.Pool
    logger Logger
}

func NewAnalyticsEngine(db *pgxpool.Pool, logger Logger) *AnalyticsEngine

// GetSurveyAnalytics агрегирует данные из трёх источников:
// - user_engagements (flow_status): TotalResponses, CompletionRate
// - user_engagements.form_data (JSONB): Distribution
// - user_survey_attributes: TopTags
func (a *AnalyticsEngine) GetSurveyAnalytics(
    ctx context.Context,
    surveyOfferId uuid.UUID,
) (*SurveyAnalytics, error)
```

### Источники данных

| Поле SurveyAnalytics | Источник | SQL логика |
|---|--|--|
| `TotalResponses` | `user_engagements` | `SELECT COUNT(*) FROM user_engagements WHERE offer_id = ? AND flow_status IN ('in_progress', 'completed')` |
| `CompletionRate` | `user_engagements` | `SUM(CASE WHEN flow_status = 'completed' THEN 1 ELSE 0 END) ::float / NULLIF(TotalResponses, 0)` |
| `Distribution[qId]` | `user_engagements.form_data->qId` (JSONB) | `SELECT fd->>'qId' as val, COUNT(*) FROM user_engagements WHERE offer_id=? AND flow_status='completed' GROUP BY val` |
| `AvgScore[qId]` | `user_engagements.form_data->qId` → cast to float | только для type=scale вопросов; `AVG(fd->>'qId'::float)` |
| `TopTags[]` | `user_survey_attributes` | `SELECT tag_key, COUNT(*) as cnt, AVG(weight) FROM user_survey_attributes WHERE survey_offer_id=? GROUP BY tag_key ORDER BY cnt DESC` |

### HTTP endpoint (admin)

```
GET  /admin/engagements/:offerId/survey-analytics
     → SurveyAnalytics (json)
     RBAC: hr, catalog_manager, admin
     tenant_id из JWT middleware (auto-join engagement_offers.tenant_id)
     Валидация: только для type=activity с survey_schema
```

### Пример ответа — кейс «Сбор спроса»

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
  "avgScore": {},
  "topTags": [
    {"tagKey": "interest:mens_health", "count": 156, "avgWeight": 0.95}
  ]
}
```

Решение менеджера каталога: `156 / 342 = 45.6% > 10%` → закупка обоснована.

---

## Последствия

### Изменение TagResolver

`cel/tag_resolver.go` получает метод `AggregateSurveyTags()`, который читает `user_survey_attributes` и возвращает `map[string]interface{}` (совместим с CELContext.Tags).

```go
package cel

// AggregateSurveyTags агрегирует теги из survey-ответов.
// tenantId из JWT middleware (ctx value), userId из User.
// Источник: user_survey_attributes — MAX(weight) GROUP BY tag_key, tag_value.
func (r *TagResolver) AggregateSurveyTags(
    ctx context.Context,
    tenantId uuid.UUID,  // из JWT middleware
    userId uuid.UUID,
) map[string]interface{} {
    // SELECT tag_key, tag_value, MAX(weight)
    // FROM user_survey_attributes
    // WHERE tenant_id = $1 AND user_id = $2
    // GROUP BY tag_key, tag_value
}

// Resolve — расширение: merge profile-тегов + survey-тегов
func (r *TagResolver) Resolve(ctx context.Context, user *User) map[string]interface{} {
    tags := r.computeFromProfile(user)               // is_remote, is_senior, ...
    surveyTags := r.AggregateSurveyTags(ctx, user.TenantID, user.ID)  // interest:sport, ...
    for k, v := range surveyTags {
        tags[k] = v  // survey-теги поверх profile-тегов
    }
    return tags  // → CELContext.Tags
}
```

**Важно:** `tenantId` берётся из JWT middleware `ctx` value — тот же tenant, который auto-join-ится при `engagement_offers.tenant_id = tenantId`.
InvalidateCache при MapSurveyAnswers — `DEL tag_cache:{tenant_id}:{user_id}` — использует существующий контракт.

### Новая БД-таблица

`user_survey_attributes` (tenant_id, user_id, survey_offer_id, tag_key, tag_value, weight, question_id, answered_at).

### Новый admin endpoint

`GET /admin/engagements/:offerId/survey-analytics` — RBAC: hr, catalog_manager, admin. Данные из трёх источников: `user_engagements` (TotalResponses, CompletionRate), `form_data JSONB` (Distribution), `user_survey_attributes` (TopTags).

### Redis cache invalidation

При MapSurveyAnswers — `DEL tag_cache:{tenant_id}:{user_id}`. Инвалидация через существующий контракт `cel/TagResolver.InvalidateCache(userId)`.

### Backward compatibility

- `form_schema` сохраняется для legacy-форм
- `ui_component: "SurveyForm"` → парсится `survey_schema`
- `ui_component: "EngagementForm"` → парсится `form_schema` (legacy)

---

## Интеграция с FlowEngine

ExecuteStep минимальный diff. При `step.Type == "form" && step.UIComponent == "SurveyForm"`:

1. Hydrate resolver из form_data (JSONB)
2. SubmitAnswer → сохранить intermediate answers обратно в form_data
3. При IsComplete() → TagMapper.MapSurveyAnswers()
4. Продолжить существующий flow (condition_check → credit)
