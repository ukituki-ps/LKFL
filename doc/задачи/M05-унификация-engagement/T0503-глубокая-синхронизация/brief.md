# T0503 — Глубокая синхронизация M05 (аудит 2)

## Контекст

Аудит 1 (T0502) нашёл и исправил 8 поверхностных несогласованностей. Втором более глубокий аудит обнаружил **ещё 12 проблем** в кодовых интерфейсах, YAML-примерах и ADR.

## Несоответствия

### Группа P0 — кодовые интерфейсы и конфигурация (4 шт.)

| № | Файл | Строка | Было | Стало |
|---|---|---|---|---|
| 1 | модули.md | 166 | `SyncCatalog(ctx) ([]Benefit, error)` | `[]EngagementOffer` |
| 2 | модули.md | 363 | БД: "users, **benefits**, consents..." | engagements |
| 3 | модули.md | 373 | Redis: "**Benefit** catalog cache" | Engagement |
| 4 | модули.md | 399-400 | NATS comment: "activity completion → credit", "benefit activate → debit" | engagement_credit/debit |

### Группа P1 — биллинг-движок YAML-примеры (4 шт.)

| № | Файл | Строка | Было | Стало |
|---|---|---|---|---|
| 5 | биллинг-движок.md | 41 | `benefit_cost * 0.7` | `engagement_offer_cost * 0.7` |
| 6 | биллинг-движок.md | 54 | `event: "benefit_activate"` | `event: "engagement_debit"` |
| 7 | биллинг-движок.md | 92 | context: `benefit_id, benefit_cost` | `engagement_offer_id, engagement_offer_cost` |
| 8 | биллинг-движок.md | 211-215 | `benefit_active`, `benefit_provider` | `engagement_active`, `engagement_provider` |

### Группа P2 — ADR и глоссарий (4 шт.)

| № | Файл | Строка | Было | Стало |
|---|---|---|---|---|
| 9 | hr.md journeys | 150 | `H03 → activity_types` | `engagement-types` |
| 10 | adr/011-monorepo.md | 27 | "User, Benefit, Transaction" | "User, Engagement, Transaction" |
| 11 | adr/004-redis.md | 23 | "Benefit catalog cache" | "Engagement catalog cache" |
| 12 | глоссарий.md | 26 | «Активность = действие для получения баллов» | «Энгейджмент (type=activity) = ...» |

## Критерии приёмки

1. [ ] модули.md — ProviderAdapter, БД-таблица, Redis, NATS-комментарии обновлены
2. [ ] биллинг-движок.md — все 4 YAML-примера на engagement-терминах
3. [ ] hr.md journeys — H03 activity_types → engagement-types
4. [ ] adr/011-monorepo.md, adr/004-redis.md — Benefit → Engagement
5. [ ] глоссарий.md — Активность → Энгейджмент (type=activity)
6. [ ] Глобальный grep `benefit_` (подстрока), `Benefit` (в контексте кода) → 0 лишних совпадений
7. [ ] Глобальный grep `activity_types` → 0 за пределами исторических M04
