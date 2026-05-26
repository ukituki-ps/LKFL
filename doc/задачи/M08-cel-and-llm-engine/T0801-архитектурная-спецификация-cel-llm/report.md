# T0801 — Архитектурная спецификация CEL + LLM — отчёт

## Статус

✅ выполнено

## Что сделано

### Созданы 2 ADR

**ADR-021** (`архитектура/adr/021-cel-unified-rule-engine.md`) — CEL как единый движок бизнес-логики:
- Контекст: 4 независимых механизма условий → избыточная сложность
- Решение: Google CEL заменяет YAML/AND/OR/JSON conditions в billing, eligibility, flow, recommendations
- LLM integration: русский текст → LLM → CEL expression → cel-go validate
- 3 фазы миграции: A (billing+eligibility), B (recommendations+flow), C (compliance)

**ADR-022** (`архитектура/adr/022-llm-proxy-service.md`) — LLM Proxy как 5-й микросервис:
- Контекст: direct LLM client в каждом сервисе → 7+ дублируемых модулей (prompts, validation, model tracking)
- Решение: LLM Proxy (:8085) — agent router, prompt mgmt, cost tracking, audit trail
- Масштабируемо: future agents (moderation, analytics) = новый YAML entry, не код в N сервисах

### Созданы 2 файла детального описания

**cel-engine.md** (371 строка):
- Роль движка (4 домена → 1 CEL engine)
- Go API: CELGenerator, CELEvaluator, CELValidator, CELContext, LLMProvider interface
- CELContext schema — единая схема для всех доменов
- LLM Integration через LLM Proxy (ADR-022) — не direct client
- DB migrations: billing_rules, engagement_offers, engagement_flows, recommendation_rules
- 3 фазы внедрения с приоритетами
- Metrics: 3 Prometheus metrics (cel_generation, cel_evaluation, cel_validation)
- Security: sandbox, rate limit 4096 chars, Redis DB4 cache

**llm-proxy.md** (310 строк):
- API: POST /llm/v1/generate, /validate, GET /agents, GET /metrics
- Agent Router YAML config (cel-generator, content-moderation, analytics-summary)
- Prompt templates system
- DB Schema: llm_requests table with audit trail
- Rate limiting per agent
- Cost tracking per tenant
- Docker deployment + Nginx routing

### Обновлены 9 файлов документации

| Файл | Изменения |
|------|--|
| `архитектура/README.md` | +ADR-021/022 (22 ADR), +cel-engine + llm-proxy в содержимое, диаграмма: LLM Proxy + cel/, +CEL в ключевые решения |
| `архитектура/модули.md` | +LLM Proxy 6-й сервис, +cel/ в Platform (10 packages), DI graph update, Nginx /llm/v1/, Keycloak lkfl-llm-proxy, DB lkfl_llm, Redis DB4, release policy ADR-021/022 |
| `архитектура/пакеты-platform.md` | +cel/ detail (5 files, Go API, 35 lines), eligibility/ → CEL-based (EvaluateCEL) |
| `архитектура/стек.md` | +cel-go + LLM Provider в Backend, +LLM Infra в Infrastructure, +7 CEL/LLM metrics, +CEL+LLM Dashboard |
| `архитектура/биллинг-движок.md` | condition[] → condition_cel CEL, всех примеров моделей A-D с condition_source + condition_cel |
| `архитектура/engagement.md` | condition_expr → CEL + condition_source, examples: survey, event, referral |
| `архитектура/nats-subjects.md` | CEL NOT in NATS, Redis DB 0-4 |
| `контекст/настраиваемость.md` | eligibility → CEL+LLM, billing → CEL+LLM, models A-D с CEL, activity types с CEL |
| `спецификация/api.md` | +3 CEL endpoints: /generate, /validate, /preview с request/response examples |

### Численные показатели

| Метка | Было | Стало |
|--|--|--|
| ADR | 20 | 22 |
| Go-сервисов | 4 | 5 (+llm-proxy :8085) |
| Platform internal packages | 9 | 10 (+cel/) |
| PostgreSQL DB | 5 | 6 (+lkfl_llm) |
| Redis DB | 4 (0-3) | 5 (0-4, DB4 = CEL cache + LLM rate limit) |
| CEL integration points | 0 | 12 |
| CEL API endpoints | 0 | 3 |

## Проблемы

- Нет конфликтов при обновлении — все файлы редактировались последовательно
- CEL context schema определена на уровне архитектуры, будет уточнена при реализации M09

## Следующий шаг

**M09-реализация-cel-engine**: физическая реализация на Go — cel/ package в Platform, llm-proxy service, DB migrations, integration tests
