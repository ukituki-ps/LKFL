# ADR-021 — CEL (Common Expression Language) как единый движок бизнес-логики

## Контекст

В системе существует 4 независимых механизма оценки условий:

1. **Billing Rule Engine** — YAML-array `${field, operator, value}[]` для фильтрации правил начисления/списания
2. **Eligibility Engine** — struct-based AND/OR/groups evaluation для проверки доступа к офферам
3. **Engagement Flow `condition_expr`** — ad-hoc string expressions (`'answers.count >= 5'`) для condition_check шагов
4. **Recommendations Engine** — JSON segment conditions + scoring rules для персонализации

Каждый механизм имеет:
- Собственный синтаксис условий
- Собственный API создания/редактирования
- Собственную логику валидации
- Собственный набор поддерживаемых операторов

Это создаёт:
- Избыточную сложность поддержки 4 движков
- Невозможность переносить правила между доменами (billing rule → eligibility condition)
- Высокую кривую обучения для HR-менеджеров (каждый UI-constructor отличается)
- Несогласованность: `(A || B) && C` невозможно в billing YAML-array (только implicit AND)

## Решение

Ввести **CEL (Google Common Expression Language)** как единый формат выражений для всех 4 доменов.

**Технология:** `github.com/google/cel-go` — официальная Go-библиотека Google.

**Интеграция с LLM:** пользователь (HR-менеджер) формулирует условие на естественном русском языке → LLM генерирует валидный CEL expression → система автоматически проверяет синтаксис перед сохранением.

### Schema единого CEL Context

Все 4 домена используют один и тот же контекст-объект (nested layout):

```go
type CELContext struct {
    User struct {
        Grade          string `cel:"grade"`
        YearsOfService int    `cel:"years_of_service"`
        HasChildren    bool   `cel:"has_children"`
        Department     string `cel:"department"`
        Status         string `cel:"status"`
        TenantID       string `cel:"tenant_id"`
        UserID         string `cel:"user_id"`
    } `cel:"user"`

    Tags map[string]any `cel:"tags"`

    Benefit struct {
        Category string  `cel:"category"`
        Cost     float64 `cel:"cost"`
    } `cel:"benefit"`

    Date struct{ Today string } `cel:"date"`

    // Динамический контекст — domain-specific fields
    // context.last_engagement, context.engagement_active, context.engagement_provider
    Context map[string]any `cel:"context"`

    Answers map[string]any `cel:"answers"`
    Events  map[string]any `cel:"events"`

    Period struct {
        Start string `cel:"start"`
        End   string `cel:"end"`
    } `cel:"period"`

    Balance struct{ Total float64 } `cel:"balance"`
}
```

### Custom CEL functions

| Function | Описание | Пример |
|----------|---------|--------|
| `date_diff_days(a, b)` | Разница дат (ISO8601) в днях | `date_diff_days(date.today, context.last_engagement) >= 7` |
| `str_contains(str, substr)` | Вхождение подстроки | `str_contains(user.department, 'Sales')` |
| `now_iso()` | Текущая дата ISO8601 | `date.today == now_iso()` |

### Миграция (3 фазы)

| Фаза | Домен | Что меняется |
|------|------|-------------|
| **A** | Billing Rules + Eligibility | YAML-array → CEL + LLM generation endpoint (HIGH prio) |
| **B** | Recommendations + Flow conditions | JSON conditions → CEL + segment CEL (MEDIUM prio) |
| **C** | Compliance retention + Consent auto-rules | ad-hoc → CEL (LOW prio) |

### LLM Integration

```
Admin UI (textarea: "бесплатный фитнес для директоров и удалённых")
  → POST /api/v1/cel/generate
    → LLM Proxy (ADR-022)
      → System prompt: context_schema + strict output rules
        → CEL string: `user.grade == 'Director' || tags.is_remote == true`
          → cel-go validator (parse + type-check)
            → OK: сохранить условие + исходный текст + model version
            → FAIL: вернуть ошибку на UI
```

### Drift protection

CEL генерируется ОДИН РАЗ при изменении текста условия. При повторном сохранении без изменений CEL сохраняется без регенерации (hash comparison исходного текста).

### Migration: Legacy Conditions Format

Каждый из 4 доменов использует свой legacy-формат условий:

| Домен | Legacy формат | Поле БД | Пример → CEL |
|------|-----------|-----|-|
| Billing Rules | YAML-array `${field, operator, value}[]` | `billing_rules.condition` | `[{"field":"user.grade","operator":"eq","value":"Senior"}]` → `user.grade == 'Senior'` |
| Eligibility | struct-based AND/OR/groups | `engagement_offers.eligibility_json` | `{"and":[{"field":"years_of_service","op":">=","val":3}]}` → `user.years_of_service >= 3` |
| Flow conditions | ad-hoc string `'answers.count >= 5'` | `engagement_flows.condition_expr` | `'answers.count >= 5'` → `answers['count'] >= 5` |
| Recommendations | JSON segment conditions | `recommendation_rules.segment_json` | `{"segment":"senior","criteria":{"grade":["Senior","Lead"]}}` → `user.grade in ['Senior', 'Lead']` |

**Migration script концепция:**

1. `SELECT * FROM billing_rules WHERE condition IS NOT NULL AND condition_cel IS NULL`
2. Для каждой записи: парсить legacy формат → генерировать CEL expression (детерминированно, без LLM)
3. Валидировать через cel-go parser
4. Сохранить в `condition_cel`
5. Shadow run: 1 месяц параллельного evaluation (legacy + CEL), сравнение результатов
6. Drop legacy полей после shadow period

### Rollback Strategy

При CEL evaluation error (panic, timeout, type mismatch):

```
IF CEL evaluation fails:
  1. Log error to Sentry + audit_logs
  2. Fallback to allow (eligibility) / deny (billing debit)
  3. Alert admin
  4. Manual: disable CEL per tenant → tenant.cel_enabled = false
```

**Per-tenant feature flag:** `tenants.cel_enabled BOOLEAN NOT NULL DEFAULT TRUE`
Отключение CEL на уровне tenant'а мгновенно переходит on hardcoded fallback:
- Eligibility: allow all (без ограничений)
- Billing debit: deny (не списывать баллы при ошибке)

### Audit trail

Каждое сохраняется условие хранит:
- `condition_source TEXT` — исходный текст на русском
- `condition_cel TEXT` — сгенерированный CEL
- `condition_llm_model TEXT` — модель LLM
- `condition_llm_version TEXT` — версия модели

## Аргументы «за»

- **Единый синтаксис** — один движок вместо четырёх
- **Безопасность** — CEL sandbox, нет `eval()`, нет произвольного кода
- **Нативная Go-поддержка** — `cel-go` стабильна, хорошо протестирована
- **Никакой кривой обучения** — LLM генерирует CEL из русского текста
- **Составная логика** — CEL поддерживает `(A || B) && C`, вложенные выражения
- **ФСТЭК-комплаенс** — audit trail: текст + CEL + модель + версия

## Аргументы «против»

- Новая зависимость в go.mod (`google.golang.org/genproto/googleapis/api/expr/v1alpha1`)
- LLM dependency — нужен manual CEL input как fallback при недоступности LLM
- Hallucination protection — обязателен pre-validation через cel-go parser
- Migration effort — 4 домена, ~12 точек интеграции

## Вердикт

**За.** Преимущества единого синтаксиса + LLM-generation перевешивают стоимость миграции. CEL replaces YAML, ad-hoc и JSON conditions. LLM eliminates learning curve.

## Следствия

- Все 4 домена используют один `CELContext` schema
- Admin UI меняет form-constructors на textarea + "Генерировать" button
- Migration scripts: `billing_rules.condition → condition_cel`, `eligibility → eligibility_cel`
- Backwards-compatibility: поле `condition_cel` alongside old fields во время migration phase, затем drop old
- CEL validation обязателен при каждом PUT/POST с условием
- Feature flag `tenant_config.llm_enabled` для опционального LLM per tenant (Phase A+)

## Статус

✅ Accepted
