# T0801 — Архитектурная спецификация CEL + LLM — план

## План

### Wave 1: ADR и базовые документы

- [x] 1. Проанализировать 4 механизма условий в текущей архитектуре (billing, eligibility, flow, recommendations)
- [x] 2. Создать ADR-021 — CEL как единый движок бизнес-логики
- [x] 3. Создать ADR-022 — LLM Proxy как 5-й микросервис
- [x] 4. Определив 12 точек интеграции CEL (таблица по сервисам/пакетам)

### Wave 2: Детальные описания

- [x] 5. Создать cel-engine.md — schema, Go API, CELContext, 3 фазы миграции, DB schema changes, metrics, security
- [x] 6. Создать llm-proxy.md — API endpoints, agent router, prompt templates, DB schema, rate limiting, cost tracking, deployment, docker-compose

### Wave 3: Обновление документации (consistency)

- [x] 7. Обновить `архитектура/README.md` — диаграмма (LLM Proxy + cel/), ADR-список (22), содержимое (cel-engine.md + llm-proxy.md), ключевые решения
- [x] 8. Обновить `архитектура/модули.md` — LLM Proxy 6-й сервис, +cel/ в Platform, DI graph, Nginx routes, Keycloak, DB, Redis, release policy
- [x] 9. Обновить `архитектура/пакеты-platform.md` — cel/ package detail, eligibility/ → CEL-based
- [x] 10. Обновить `архитектура/стек.md` — cel-go, LLM Provider, 7 CEL/LLM метрик, CEL+LLM Dashboard
- [x] 11. Обновить `архитектура/биллинг-движок.md` — condition → condition_cel, LLM gen step, все примеры моделей с CEL
- [x] 12. Обновить `архитектура/engagement.md` — condition_expr → CEL + condition_source, примеры
- [x] 13. Обновить `архитектура/nats-subjects.md` — CEL NOT in NATS, Redis DB4
- [x] 14. Обновить `контекст/настраиваемость.md` — eligibility/billing/recommendations → CEL + LLM
- [x] 15. Обновить `спецификация/api.md` — 3 CEL endpoints: /generate, /validate, /preview с examples
- [x] 16. Финальная проверка согласованности всех файлов

## Зависимости

- После M07 (все задачи выполнены, архитектура стабильна)
- Нет внутренних зависимостей — архитектура определена в рамках T0801

## Правила

Только документация. Код не трогается. Физическая реализация → M09.
