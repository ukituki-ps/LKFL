# T0801 — Архитектурная спецификация: CEL Engine + LLM Proxy

## Веха

M08-cel-and-llm-engine

## Контекст

В платформе существует 4 независимых механизма оценки условий:
1. **Billing Rule Engine** — YAML-array `[{field, operator, value}]` для фильтрации правил начисления/списания
2. **Eligibility Engine** — struct-based AND/OR/groups evaluation для проверки доступа к офферам
3. **Engagement Flow condition_expr** — ad-hoc string expressions (`'answers.count >= 5'`) для condition_check шагов
4. **Recommendations Engine** — JSON segment conditions + scoring rules для персонализации

Проблемы: избыточная сложность, высокая кривая обучения для HR-менеджеров, отсутствие составных выражений `(A || B) && C` в billing YAML.

Предложены 2 архитектурных решения:
- **CEL (Common Expression Language)** — единый формат выражений для всех 4 доменов
- **LLM Proxy (5-й микросервис)** — централизованный шлюз для LLM-генерации CEL из русского текста

### Файлы-мишени

| Действие | Файл |
|---|-|
| Создать ADR | `архитектура/adr/021-cel-unified-rule-engine.md` |
| Создать ADR | `архитектура/adr/022-llm-proxy-service.md` |
| Создать детальное описание CEL | `архитектура/cel-engine.md` |
| Создать детальное описание LLM Proxy | `архитектура/llm-proxy.md` |
| Обновить (README) | `архитектура/README.md` — диаграмма + ADR-список + содержимое |
| Обновить (модули) | `архитектура/модули.md` — LLM Proxy серв, +cel/ пакет, DI graph, DB, Redis, Nginx, Keycloak, release policy |
| Обновить (пакеты) | `архитектура/пакеты-platform.md` — cel/ package, eligibility/ → CEL-based |
| Обновить (стек) | `архитектура/стек.md` — cel-go, LLM Provider, 7 CEL/LLM метрик, CEL+LLM Dashboard |
| Обновить (биллинг) | `архитектура/биллинг-движок.md` — condition → condition_cel, все примеры моделей |
| Обновить (engagement) | `архитектура/engagement.md` — condition_expr → CEL + condition_source |
| Обновить (nats) | `архитектура/nats-subjects.md` — CEL NOT in NATS |
| Обновить (матрица) | `контекст/настраиваемость.md` — eligibility/billing/recommendations → CEL + LLM |
| Обновить (API) | `спецификация/api.md` — 3 CEL endpoints: generate, validate, preview |

### Критерии приёмки

- [x] Созданы ADR-021 и ADR-022 в формате ХАДД
- [x] Созданы 2 файла детального описания (cel-engine.md, llm-proxy.md) с Go API, схемой DB, метриками, security
- [x] Обновлены все 9 файлов документации для консистентности
- [x] Определена миграционная стратегия (3 фазы: A/B/C)
- [x] Определены 12 точек интеграции CEL
- [x] Определён API LLM Proxy и 3 CEL endpoints
- [x] Диаграмма архитектуры обновлена (LLM Proxy + cel/ пакет)
- [x] Счётчики приведены в порядок: 22 ADR, 6 сервисов, 10 internal packages, 6 PG DB, 5 Redis DB
