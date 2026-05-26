# M14 — Survey Implementation

> **⛔ ОТМЕНЕНА (2026-05-26).** Решение команды — отложить реализацию кода до лучших времён.
> Архитектура Survey Engine (M13 T1301, ADR-025) сохранена и актуальна.

## Контекст

M13 T1301 завершил проектирование Survey Engine (ADR-025):
- Схема survey: question → option → branch_on → scoring
- SurveyEngine API: Create/GetResults/Analytics
- TagMapper integration: survey results → user tags
- FlowEngine integration: survey как step type
- Analytics API: per-question, per-branch, response rate
- DB: surveys, survey_questions, survey_responses, survey_branches, user_survey_attributes (5 таблиц)

> **Важно:** это первая задача по коду в проекте. Go-код отсутствует — создаётся с нуля.

## Phase 0 — Project Bootstrap

Перед реализацией Survey Engine необходимо создать инфраструктуру:

| Артефакт | Описание |
|----------|----------|
| `go.mod` | Зависимости: pgx, chi, go-oidc, go-redis, asynq, cel-go, excelize, gofpdf |
| `cmd/server/main.go` | HTTP entry point |
| `cmd/worker/main.go` | Asynq worker |
| `docker-compose.yml` | PostgreSQL 17, Redis 7, Keycloak 25 |
| `Dockerfile` | Multi-stage build |
| `Makefile` | build, test, lint, migrate, compose |
| `migrations/` | Atlas SQL-first |
| `internal/` | Пустые пакеты для DI wiring |
| `shared/pkg/auth/` | verifier, middleware, rbac |
| `.github/workflows/ci.yml` | build + test + lint |

## Что делать

1. Создать структуру Go-проекта (Phase 0)
2. Реализовать `internal/engagement/survey/` пакет с SurveyEngine API
3. Интегрировать survey steps в FlowEngine.ExecuteStep
4. Реализовать TagMapper для survey → tag mapping
5. Добавить TagResolver.AggregateSurveyTags() в CELContext builder

## Ожидаемая структура файлов

```
lkfl/
├── go.mod
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── .github/workflows/
│   └── ci.yml
├── cmd/
│   ├── server/main.go
│   └── worker/main.go
├── internal/
│   ├── tenant/
│   ├── auth/
│   ├── user/
│   ├── consent/
│   ├── cel/
│   ├── llm/
│   ├── eligibility/
│   ├── compliance/
│   ├── engagement/
│   │   ├── catalog.go
│   │   ├── flow.go
│   │   ├── collections.go
│   │   └── survey/
│   │       ├── engine.go
│   │       ├── validator.go
│   │       ├── tag_mapper.go
│   │       ├── scheduler.go
│   │       └── analytics.go
│   ├── notification/
│   ├── gamification/
│   ├── billing/
│   ├── integrations/
│   ├── payments/
│   ├── content/
│   ├── recommendations/
│   └── api/
├── shared/pkg/
│   ├── auth/
│   └── celcontext/
└── migrations/
```

## Зависимости

- ADR-025 (Survey Engine spec)
- M13 T1301 (architecture design)
- internal/cel/ (для CEL context builder extension)
- internal/engagement/flow/ (FlowEngine.StepType.Survey dispatch)
