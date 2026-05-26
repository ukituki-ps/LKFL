# T3101-T3106 — Survey Engine (ADR-025)

## Веха

M31-survey

## T3101 — Migrations: Survey
```sql
CREATE TABLE lkfl_platform.surveys (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255), description TEXT,
    status VARCHAR(20) CHECK (status IN ('draft','active','completed','archived')),
    start_date DATE, end_date DATE, settings JSONB, created_at, updated_at
);
CREATE TABLE lkfl_platform.survey_questions (
    id UUID PK, survey_id UUID FK, type VARCHAR(20) CHECK (type IN ('single_choice','multiple_choice','text','rating','boolean')),
    text TEXT, options JSONB, required BOOLEAN DEFAULT true, sort_order INT, created_at
);
CREATE TABLE lkfl_platform.survey_branches (
    id UUID PK, question_id UUID FK, target_question_id UUID FK,
    condition TEXT,  -- CEL expression for branching
    created_at
);
CREATE TABLE lkfl_platform.survey_responses (
    id UUID PK, survey_id UUID FK, user_id UUID FK, question_id UUID FK,
    answer JSONB, submitted_at
);
CREATE TABLE lkfl_platform.user_survey_attributes (
    id UUID PK, user_id UUID FK, survey_id UUID FK,
    attribute_key VARCHAR(100), attribute_value TEXT, created_at
);
```

## T3102 — internal/engagement/survey/ (Survey Engine)
- Survey CRUD
- Branching logic (CEL evaluation)
- Response collection
- Analytics aggregation

## T3103 — Survey + Activity integration
- Опрос как activity type
- Survey completion → activity completion → credit
- Unit tests

## T3104 — TagMapper
- Survey response → user tags → CEL context update
- Redis: `cel:tags:{user_id}` invalidation
- Unit tests

## T3105 — API: Survey
```
GET  /api/v1/surveys/:id              — опрос для пользователя
POST /api/v1/surveys/:id/submit       — ответ
GET  /api/v1/surveys/:id/progress     — прогресс заполнения
```

## T3106 — Analytics API
```
GET /admin/surveys/:id/analytics — результаты, статистика по вопросам
```

## Критерии приёмки
- [ ] Все 6 задач
- [ ] Branching logic работает
- [ ] Survey → activity → credit
- [ ] TagMapper → user tags
