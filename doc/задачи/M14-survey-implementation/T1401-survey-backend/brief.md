# T1401 — Survey Engine Implementation

> **⛔ ОТМЕНЕНО.** Решение команды — отложить реализацию кода до лучших времён.

## Контекст

Эта задача была предназначена для создания Go-проекта с нуля и реализации Survey Engine на основе архитектуры M13 T1301 (ADR-025).

## Причина отмены

Решение команды — отложить написание кода. Архитектура Survey Engine (M13, ADR-025) сохранена и актуальна для будущей реализации.

## Что было спроектировано (M13)

- Survey schema: question → option → branch_on → scoring
- SurveyEngine API: Create/GetResults/Analytics
- TagMapper: survey results → user survey tags
- FlowEngine integration: survey как step type
- Analytics API: per-question, per-branch, response rate
- DB: surveys, survey_questions, survey_responses, survey_branches (4 таблицы)

## При будущем возобновлении

1. Bootstrap Go-проекта (go.mod, cmd/server, cmd/worker, internal/)
2. Реализация `internal/engagement/survey/` по ADR-025
3. TagResolver.AggregateSurveyTags
4. FlowEngine.ExecuteStep integration
5. Admin endpoint `/admin/engagements/:offerId/survey-analytics`
6. Unit-тесты для всех компонентов
