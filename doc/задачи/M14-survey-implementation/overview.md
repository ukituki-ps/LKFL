# M14 — Survey Engine Implementation

> **⛔ ОТМЕНЕНА (2026-05-26).** Решение команды — отложить реализацию кода.

## Описание

Первая задача по коду в проекте LKFL (отменена). Включала Phase 0 (project bootstrap) и реализацию Go-кода Survey Engine на основе архитектуры из M13 T1301.

## Цели

### Phase 0 — Project Bootstrap

0. Создать структуру Go-проекта: go.mod, cmd/server, cmd/worker, internal/, shared/pkg/, migrations/
1. Настроить docker-compose.yml (PostgreSQL 17, Redis 7, Keycloak 25)
2. Создать Makefile + GitHub Actions CI pipeline
3. Реализовать DI wiring в app/

### Phase 1 — Survey Engine

1. Реализовать `internal/engagement/survey/` пакет с SurveyEngine API
2. Интегрировать survey steps в FlowEngine.ExecuteStep
3. Поддержать branching logic (branch_on) в survey steps
4. Реализовать TagMapper для survey → tag mapping
5. Добавить TagResolver.AggregateSurveyTags() в CELContext builder

## Зависимости

- M13-survey-engine (T1301) — архитектура завершена
- ADR-025 — Survey Engine спецификация
- **Нет предыдущего Go-кода** — M14 создаёт проект с нуля

## M14 API Subset

Из 118 endpoints выделены survey-related endpoints для M14:

| Method | Path | Описание | Роль |
|--------|------|----------|------|
| `POST` | `/admin/engagements/surveys` | Create survey | hr, admin |
| `GET` | `/admin/engagements/surveys/:id` | Get survey details | hr, catalog_manager, admin |
| `GET` | `/admin/engagements/surveys/:id/results` | Survey analytics | hr, catalog_manager, admin |
| `POST` | `/user-engagements/:id/steps/:stepId/submit` | Survey step submit (answer) | employee |
| `GET` | `/user-engagements/:id/survey/:surveyId` | Get survey for user (next question) | employee |
| `GET` | `/admin/engagements/:offerId/survey-analytics` | Per-question, per-branch stats | hr, catalog_manager, admin |

> **Остальные 140 endpoints** — вне scope M14.

## Тестовая стратегия

### Mock strategy

| Компонент | Стратегия | Библиотека |
|-----------|-----------|------------|
| DB (PostgreSQL) | Interface mocking (`sql.DB` → `sqlmock`) | `github.com/DATA-DOG/go-sqlmock` |
| Redis | Interface mocking | `github.com/alicebob/miniredis` (in-memory) |
| CEL Evaluator | Interface mocking (`Evaluate(ctx, expr, ctx)` → stub) | ручной mock |
| External HTTP | WireMock / stub | `net/http/httptest` |

### Test fixtures

- **Survey JSON schemas** — валидные/невалидные schemas для in/out validation
- **CEL context fixtures** — готовые `CELContext` объекты для test evaluation
- **DB fixtures** — seed data для integration tests (tenants, users, engagements)

### Integration test

- **testcontainers** — PostgreSQL + Redis в Docker-контейнерах
- **Полный flow:** survey creation → answer submission → tag mapping → CEL evaluation → analytics

### Coverage requirement

- **Unit tests:** 80%+ coverage для `internal/engagement/survey/`
- **Integration tests:** coverage critical paths (create → submit → tag → analytics)

## Exit criteria

- `SurveyEngine.Create()` создаёт survey с validation schema
- `SurveyEngine.SubmitAnswer()` принимает ответ + валидирует
- `FlowEngine.ExecuteStep` dispatch-ит survey steps через `survey/scheduler.go`
- `TagMapper.Apply()` мапит результаты survey → теги пользователя
- TagResolver.AggregateSurveyTags() возвращает survey-теги для CEL evaluation
- Unit-test coverage: 80%+
- `go build ./...` — без ошибок
- `go test ./...` — все зелёные
