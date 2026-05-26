# M08 — CEL и LLM Engine

## Описание

Проектирование и документирование двух архитектурных решений:
- **ADR-021: CEL (Common Expression Language)** как единый движок бизнес-логики — заменяет 4 независимых механизма условий (billing YAML, eligibility AND/OR, flow condition_expr, recommendations JSON)
- **ADR-022: LLM Proxy** как 5-й микросервис — централизованный шлюз ко всем LLM-моделям

M08 — чисто архитектурная веха. Физическая реализация в Go (cel/ package, llm-proxy service) запланирована на **M09-реализация-cel-engine** после стабилизации M00→M07.

### Что не так (проблема)

Проблема была описана в анализе решений от пользователя:

| Проблема | Текущее состояние | Критичность |
|--------|----|--------|
| 4 независимых механизма условий | billing YAML-array, eligibility AND/OR/Groups, flow condition_expr (ad-hoc), recommendations JSON | 🔴 Высокая |
| Высокая кривая обучения для HR-менеджеров | Каждый domain имеет свой UI-constructor / API format | 🔴 Высокая |
| Вложенные выражения `(A \|\| B) && C` | Невозможно в billing YAML (только implicit AND) | 🟡 Средняя |
| Дублирование prompt/validate/model config | Если каждый сервис имеет свой LLM client = 7+ дублируемых модулей | 🟡 Средняя |

### Что было сделано

M08 — архитектурная спецификация. Все изменения:
1. ✅ Созданы 2 ADR (ADR-021, ADR-022)
2. ✅ Создано 2 файла детального описания (cel-engine.md, llm-proxy.md)
3. ✅ Обновлено 9 файлов документации для консистентности
4. ✅ Определены 3 фазы миграции (A: billing+eligibility, B: recommendations+flow, C: compliance)
5. ✅ Определены 12 точек интеграции CEL
6. ✅ Определён API LLM Proxy + 3 CEL endpoints для Platform

### Миграционный план

| Фаза | Домен | Точки интеграции | Приоритет |
|------|------|-|--------|
| **A** | Billing Rules + Eligibility | billing_rules.condition_cel, engagement_offers.eligibility_cel, /cel/generate endpoint | 🔴 HIGH |
| **B** | Recommendations + Flow conditions | recommendation_rules.segment_cel, engagement_flows condition_expr CEL | 🟡 MEDIUM |
| **C** | Compliance retention + Consent auto-rules | compliance_policies.retention_cel, consents.auto_renewal_cel | 🟢 LOW |

### Веху можно закрывать когда

- [x] T0801 — архитектурная спецификация CEL Engine + LLM Proxy (2 ADR, 2 detail docs, 9 updated files)

## Задачи вехи

| Задача | Описание | Статус |
|\|--|\|--|
| T0801 | Архитектурная спецификация: CEL Engine (ADR-021) + LLM Proxy (ADR-022) | ✅ выполнено |
