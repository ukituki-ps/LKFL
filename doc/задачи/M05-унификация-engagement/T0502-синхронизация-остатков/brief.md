# T0502 — Синхронизация остатков после M05/M06

## Контекст

M05 (T0501) унифицировала льготы и активности в единую абстракцию Engagement, обновив 14+ файлов. Аудит документации показал 8 несогласованностей, которые остались не исправленными: часть файлов была «забыта», в другой — старые термины остались в примерах и таблицах.

**Источник:** полный аудит всех `.md` файлов проекта.

## Проблема

8 несогласованностей делятся на 3 группы:

### Группа P0 (2 шт.) — старые triggers + старые ссылки в api.md

| # | Файл | Проблема | Стара́я строка | Новая строка |
|---|------|----------|---------------|-------------|
| 1 | `спецификация/api.md` стр. 511-512 | `generic BenefitPlan + ActivationFlow` | DMS/Matkapital → generic BenefitPlan + ActivationFlow | DMS/Matkapital → Unified EngagementType + EngagementOffer |
| 2 | `спецификация/journeys/tenant-onboarding.md` стр. 91-97 | Устаревшие event triggers | `trigger=event:activity_completed` → `trigger=event:engagement_credit`<br>`trigger=event:benefit_activate` → `trigger=event:engagement_debit` |

### Группа P1 (3 шт.) — устаревшие термины в примерах и кодовых сниппетах

| # | Файл | Проблема |
|---|------|----------|
| 3 | `архитектура/интеграции.md` стр. 271 | `SyncCatalog` возвращает `[]Benefit` — нужно `[]EngagementOffer` |
| 4 | `архитектура/безопасность.md` стр. 197 | `benefit_id` в audit trail — нужно `engagement_id` |
| 5 | `архитектура/стек.md` стр. 98-99 | Пример лога: `"benefit activated"` → `"engagement.debit.confirm"` |

### Группа P2 (3 шт.) — устаревшие счётчики

| # | Файл | Проблема |
|---|------|----------|
| 6 | `спецификация/README.md` стр. 26 | «27 артефактов» → 25 (after S09+S10a merge) |
| 7 | `спецификация/README.md` стр. 30 | «116 REST API endpoints» → 131 endpoints |
| 8 | `план/README.md` | «M00-создание-описания — текущая веха» → M06 |

## Решение

Пройтись по каждому файлу и заменить устаревшие строки на engagement-терминологию.

## Зависимости

- **T0501** (M05-унификация-энгейджмента) — предшествующая задача. Без M05 нет базовой модели для синхронизации.
- **Отсутствуют** зависимости от M06 — это задача «чистки следа», а не структурных изменений.

## Файлы-мишени

| Файл | Группа | Что менять |
|------|--------|-----------|
| `спецификация/api.md` | P0 | str 511-512: `generic BenefitPlan + ActivationFlow` → `Unified EngagementType + EngagementOffer` |
| `спецификация/journeys/tenant-onboarding.md` | P0 | str 91-97: event triggers → engagement_credit/debit |
| `архитектура/интеграции.md` | P1 | str 271: `[]Benefit` → `[]EngagementOffer` |
| `архитектура/безопасность.md` | P1 | str 197: `benefit_id` → `engagement_id` |
| `архитектура/стек.md` | P1 | str 98-99: пример лога на engagement-терминах |
| `спецификация/README.md` | P2 | str 26: 27→25 артефактов, str 30: 116→131 endpoints |
| `план/README.md` | P2 | «текущая веха» → M06 |

Всего: **7 файлов**, **8 правок**.

## Критерии приёмки

1. [ ] `спецификация/api.md`: `BenefitPlan + ActivationFlow` удалён, заменён на `EngagementType + EngagementOffer`
2. [ ] `спецификация/journeys/tenant-onboarding.md`: triggers `event:activity_completed` → `event:engagement_credit`, `event:benefit_activate` → `event:engagement_debit`
3. [ ] `архитектура/интеграции.md`: `[]Benefit` → `[]EngagementOffer` в ProviderAdapter interface
4. [ ] `архитектура/безопасность.md`: `benefit_id` → `engagement_id` в audit trail таблице
5. [ ] `архитектура/стек.md`: пример лога observability на engagement-терминах
6. [ ] `спецификация/README.md`: 27 → 25 артефактов, 116 → 131 endpoints
7. [ ] `план/README.md`: текущая веха → M06
8. [ ] Глобальный grep `BenefitPlan|ActivationFlow|UserBenefit` за пределами stub-redirect'ов и исторических report'ов M01-M04 → 0 лишних совпадений
9. [ ] Глобальный grep `event:benefit_activate|event:activity_completed` за пределами исторических report'ов → 0 лишних совпадений
