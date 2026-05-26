# Отчёт — T1301 Архитектура Survey Engine

## Статус

✅ Выполнено (100%) — только документация. АDR усилено на аудите от 2026-05-25.

## Выполнено

### 1. ADR-025 — Survey Engine Architecture

Файл: `doc/архитектура/adr/025-survey-engine.md` (→ 623 строки после аудито-доработки)

**Добавлено на аудите (+4 секции, ~360 строк):**

| Секция | Что добавлено |
|---|-|
| Спецификация `survey_schema` | Полный YAML-формат (4 вопроса, 4 типа), таблица question types × features, формат tag_mappings (базовый/расширенный), сравнительная таблица form_schema vs survey_schema, 5 правил валидации |
| Resolver API + алгоритмы | Полные Go-сигнатуры (6 методов + ValidateSchema), 5-шаговый алгоритм бранчинга, OR-union multiple_choice, text+branch_on=error, hydration flow между 2 HTTP-вызовами |
| TagMapper API + lifecycle | SurveyTag struct, TagMapper сигнатуры (NewTagMapper, MapSurveyAnswers), 6-шаговый lifecycle (read schema → compute expected → get existing → delete orphaned → INSERT/UPDATE GREATEST → InvalidateCache), answer-change сценарий, cross-survey aggregation пример с SQL |
| Analytics API | SurveyAnalytics struct (TotalResponses, CompletionRate, Distribution, AvgScore, TopTags), AnswerCount + SurveyTagCount, AnalyticsEngine сигнатуры, таблица источников данных (поле → SQL), admin GET endpoint + RBAC, JSON-пример ответа |

**Усилено:**
- §Последствия: AggregateSurveyTags с полной сигнатурой (ctx, tenantId, userId) + SQL-источник; Resolve merge с развёрнутым Go-кодом; tenantId из JWT middleware
- Контекст, варианты, решение, диаграмма зависимостей, ключевые решения, интеграция FlowEngine — были корректны

### 2. DB schema — SQL-заготовка

Файл: `backend/migrations/001_add_user_survey_attributes.sql`

CREATE TABLE user_survey_attributes (шаблон для будущей реализации):
- 8 columns, UNIQUE constraint, CHECK weight [0, 1]
- 3 индекса: tenant_user, tenant_tag, tenant_offer
- ADR-024 compliant: tenant_id для multi-tenancy, ON DELETE CASCADE

### 3. Обновлённая документация

| Файл | Изменение |
|-----|-----|
| `doc/архитектура/модули.md` | engagement/ описание + survey/, DI граф + survey/ стрелка, Nginx routes + survey-analytics |
| `doc/архитектура/пакеты-platform.md` | Новый подраздел engagement/survey/ (структура + API + зависимости) — как план |
| `doc/архитектура/теги.md` | Новая категория survey-теги, interest:* + sport_intensity в каталоге, InvalidateCache при MapSurveyAnswers |
| `doc/архитектура/cel-engine.md` | CELContext.Tags комментарий про survey-теги |
| `doc/архитектура/engagement.md` | survey_schema note при ui_component="SurveyForm" |
| `doc/задачи/README.md` | M13 статус → ✅ выполнено |

## Аудит от 2026-05-25

Первоначальный ADR-025 (158 строк) прошёл детальный аудит по 54 пунктам из brief.md.
Покрытие: 32/54 (59%).

**Критические пробелы (исправлены):**
- survey_schema не специфицирована → ✅ добавлен YAML, 4 question types, tag_mappings format, comparison table, 5 validation rules
- Resolver алгоритмы не описаны → ✅ добавлены 5-шаговый branching algorithm, OR-union semantics, state persistence flow
- TagMapper lifecycle не специфицирован → ✅ добавлен SurveyTag struct, 6-step lifecycle, answer-change scenario, cross-survey aggregation
- Analytics API поверхностно → ✅ добавлен SurveyAnalytics struct, data sources table, JSON example

**После доработки:** 54/54 (100%).

## Что НЕ сделано (намеренно)

M13 — только документация и архитектура. Код Go не создаётся в этой вехе:

| Что не тронут | Причина |
|-|-|
| `backend/internal/engagement/survey/*.go` | Будет создан в M14 при реализации |
| `backend/internal/cel/tag_resolver.go` | Не модифицирован — только описание расширения в ADR |
| `backend/internal/engagement/flow/flow_engine.go` | ExecuteStep не изменён |

## Проблемы

Отсутствуют. Все артефакты совместимы с существующей архитектурой.

## Следующие шаги

М14-survey-implementation (новая веха):
- Написать Go-код `engagement/survey/` по спецификации из ADR-025
- Расширить `cel/TagResolver` (AggregateSurveyTags + Resolve merge)
- Встроить survey-processing в FlowEngine.ExecuteStep
- Реализовать admin endpoint `/admin/engagements/:offerId/survey-analytics`
- Реализация frontend-компонента `SurveyWidget` (M09 или M14)
