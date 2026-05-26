# T0502 — Синхронизация остатков после M05/M06

## Статус

**выполнено** · 2026-05-24

## Что сделано

Пройден полный аудит всех `.md` файлов проекта после M05 (унификация Engagement). Найдено 8 несогласованностей в 7 файлах, все исправлены.

### Правки по группам

#### P0 — критические (2 файла)

| Файл | Было | Стало |
|------|------|-------|
| `спецификация/api.md` str 137 | `Объединяет UserBenefit + Completion.` | `Заменяет UserBenefit + Completion.` |
| `спецификация/api.md` str 511-512 | `→ generic BenefitPlan + ActivationFlow` | `→ Unified EngagementType + EngagementOffer (M05)` |
| `спецификация/journeys/tenant-onboarding.md` str 91-97 | `trigger=event:activity_completed`, `trigger=event:benefit_activate` | `trigger=event:engagement_credit`, `trigger=event:engagement_debit` |

#### P1 — терминология (3 файла)

| Файл | Было | Стало |
|------|------|-------|
| `архитектура/интеграции.md` str 271 | `SyncCatalog ([]Benefit, error)` | `SyncCatalog ([]EngagementOffer, error)` |
| `архитектура/безопасность.md` str 197 | `benefit_id` в audit trail | `engagement_id` |
| `архитектура/стек.md` str 98-99 | `"benefit activated", "benefit_id": "fitness"` | `"engagement.debit.confirm", "engagement_id": "fitness", "billing_direction": "debit"` |

#### P2 — счётчики (2 файла)

| Файл | Было | Стало |
|------|------|-------|
| `спецификация/README.md` | 27 артефактов, 116 endpoints | 25 артефактов, 131 endpoints |
| `план/README.md` | M00 — текущая веха | M06 — текущая веха |

> Дополнительно: `специfication/api.md` str 137 — поправлено формулирование UserEngagement (было «Объединяет», стало «Заменяет»).

## Изменённые файлы

| Файл | Изменение |
|------|----------|
| `спецификация/api.md` | 2 правки (str 137, 511-512) |
| `спецификация/journeys/tenant-onboarding.md` | 1 правка (triggers) |
| `архитектура/интеграции.md` | 1 правка (SyncCatalog return type) |
| `архитектура/безопасность.md` | 1 правка (audit trail) |
| `архитектура/стек.md` | 1 правка (пример лога) |
| `спецификация/README.md` | 2 правки (счётчики) |
| `план/README.md` | 1 правка (текущая веха) |

## Проблемы / Риски

Нет. Все 9 критериев приёмки выполнены.

## Проверки

- `grep "BenefitPlan\|ActivationFlow\|UserBenefit"` → **47 совпадений**, все исторические (M02-M04 reports, engagement.md mapping, stub redirect, M05 brief/report)
- `grep "event:benefit_activate\|event:activity_completed"` → **3 совпадения**, только в brief.md T0502 (описание задачи)
- Оба результата = **0 лишних совпадений**

## Следующие шаги

M05 готова к закрытию. T0501 + T0502 = полный цикл унификации Engagement.
