# M13 — Survey Engine: полноценный модуль опросов и викторин

## Описание

Текущая архитектура обрабатывает опросы как `EngagementType(type: "activity")` с flow-шагами `form` + `condition_check`. Это рабочий минимум для «опрос → баллы», но не покрывает три критических пробела:

1. **Tag Mapping от ответов** — TagResolver в `cel/` вычисляет теги только из профиля (grade, стаж, отдел). Ответы на опросы не конвертируются в теги → рекомендации, eligibility и billing rules не учитывают реальные интересы сотрудника.
2. **Бранчинг вопросов** — `form_schema` — плоский список вопросов без условий видимости. Нет `show_if`, нет ветвления по ответу.
3. **Агрегация результатов** — нет endpoint'а для HR/админа: «Покажи мне сводку по опросу T-XXX». Менеджер каталога не может валидировать спрос перед закупкой.

**Решение:** интегрировать Survey Engine как подпакет `internal/engagement/survey/` (не отдельный пакет) — опросы останутся `type: "activity"`, но получат полноценный движок с бранчингом, tag-mapping и аналитикой.

### Почему подпакет `engagement/survey/`, а не `internal/survey/`

| Критерий | Подпакет `engagement/survey/` | Отдельный `internal/survey/` |
|---------|--|----------|
| Ownership lifecycle | Flow lifecycle обрабатывает `engagement/flow/` напрямую | Нужен NATS/gRPC между пакетами |
| Биллинг | `engagement/flow/` → `billing.Credit()` уже есть | Нужна транзитная зависимость |
| Геймификация | `engagement/flow/` → `gamification/TriggerHandler` уже есть | Нужна транзитная зависимость |
| Интеграции | survey-ответы влияют на eligibility → engagement/ | Кросс-пакетный вызов |
| БД | Таблицы в `lkfl_platform` (уже там) | Собственные таблицы |
| Разделитель ответственности | Survey — подтип activity, не отдельный домен | Survey = новый домен (overkill) |

**Зависимости Survey Engine от существующей архитектуры:**

```
engagement/survey/
   └── TagMapper ─→ cel/TagResolver (расширение тегов)

engagement/flow/
   └── FlowStep(type="form") ─→ survey/Resolver (рендер с бранчингом)
   └── FlowEngine.Complete() ─→ billing/Credit (награда — существующий flow)

admin/api/
   └── /admin/engagements/:offerId/survey-analytics ─→ survey/Analytics
```

### Что будет спроектировано в T1301 (только документация)

1. ADR-025: Архитектура Survey Engine
2. Спецификация `survey_schema` (бранчинг, tag-mapping) — без Go-кода
3. API-дизайн `survey/TagMapper` в ADR — без реализации
4. Расширение `cel/TagResolver` — описание в ADR, без изменения tag_resolver.go
5. DB schema `user_survey_attributes` — SQL-заготовка для M14
6. Analytics API спецификация — endpoint, response shape, RBAC
7. State management Resolver — алгоритм в ADR, гидратация из form_data
8. Обновление `архитектура/модули.md`, `пакеты-platform.md`, `теги.md`

## Будущая структура (план для M14-implementation)

```
backend/internal/
├── engagement/
│   ├── survey/              # ← БУДЕТ создан в M14
│   │   ├── types.go          # SurveySchema, Question, QuestionType, TagMapping, SurveyState
│   │   ├── resolver.go       # SurveyResolver — рендер с бранчингом + hydration из form_data
│   │   ├── tag_mapper.go     # TagMapper — ответ → тег с весом (lifecycle: insert/update/delete orphaned)
│   │   └── analytics.go      # Analytics — агрегация (user_engagements + form_data + user_survey_attributes)
│   ├── catalog.go
│   ├── flow.go
│   ├── collections.go
│   ├── billing_events.go
│   └── types.go
```

**M13 создаёт только спекацию.** Реальный Go-код в `survey/` будет написан в M14.

## Ключевые архитектурные решения

| Решение | Обоснование |
|---|-|-|
| State в `user_engagement.form_data` (JSONB) | Resolver stateful, HTTP stateless. Persist между вызовами. Минимальный diff в flow.go. |
| `tenant_id` в DB schema | ADR-024: каждая бизнес-таблица содержит tenant_id. Partition pruning. |
| `GREATEST(new, existing)` при UPDATE | Сохраняет максимальный вес при перепрохождении опроса. |
| Orphaned tag deletion при изменении ответа | User ответил "sport" → перешёл на "cinema". Старый тег `interest:sport` удаляется для этого survey_offer_id. |
| multiple_choice + branch_on = OR-union | Выбор ["sport", "cinema"] → объединение всех веток. |
| text question + branch_on = error | Нет дискретных options → branch_on не применим. |

## Задачи вехи

| Задача | Описание | Статус |
|--------|--|--------|
| T1301 | Архитектура Survey Engine: ADR-025, survey_schema spec, TagMapper API design, TagResolver extension spec, DB schema (tenant_id), Analytics API spec, state management, doc update — **только документация, без кода** | ✅ выполнено |

## M14-survey-implementation (следующая веха)

Резервируется для фактической реализации Go-кода по спецификации T1301.
