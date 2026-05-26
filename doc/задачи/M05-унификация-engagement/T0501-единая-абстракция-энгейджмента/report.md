# T0501 — Единая абстракция «Энгейджмент»

## Статус

выполнено (100%)

## Что сделано

### Создано (1 файл новый)

| Файл | Размер | Описание |
|------|------:|-|--|
| `архитектура/engagement.md` | 1015 строк | Полное описание 5 сущностей (EngagementType, EngagementOffer, EngagementFlow, UserEngagement, Eligibility), YAML-схемы, ASCII-диаграмма, 4 примера заполнения (ДМС, фитнес, опрос, реферал), таблица миграции, RBAC, индексы БД |

### Заменены (2 файла → stub redirect)

| Файл | Новое содержимое |
|------|--|
| `архитектура/льготы.md` | Redirect stub → `engagement.md` |
| `архитектура/активности.md` | Redirect stub → `engagement.md` |

### Обновлены (15 файлов, 20+ секций)

| Файл | Что изменено |
|------|-|--|
| `архитектура/модули.md` | `benefit/` + `activity/` → `engagement/`, NATS subjects обновлены, 3 новых Asynq workers |
| `архитектура/биллинг-движок.md` | `benefit_activate` → `engagement_debit`, `activity_completed` → `engagement_credit`, софинансирование через engagement |
| `архитектура/README.md` | nav-таблица: engagement.md как основной документ, stub-метки |
| `спецификация/api.md` | Полный реврайт: Engagements, Offers, User Engagements, Admin Engagements, Types, Flows; удалена устаревшая секция Activities (5 endpoints); RBAC type guards; 131 endpoints |
| `спецификация/артефакты.md` | S09+S10a → S09 Engagement Wizard, S04 unified catalog, H03 activity engagements, M01 engagement types, сводная таблица 25 артефактов |
| `спецификация/критерии-приёмки.md` | 5 критериев обновлены на engagement терминологию, S10a→S09, S10b→S10 |
| `спецификация/journeys/сотрудник.md` | J02, J03, J04, J05, J06, J07, J10, J14b — все S10a/b/c → S09/S10 |
| `спецификация/journeys/hr.md` | J18 — геймификация → activity engagements |
| `спецификация/journeys/менеджер-каталога.md` | J21–J30 — type=benefit, engagement type, offer, flow терминология |
| `спецификация/journeys/README.md` | S10a→S09, S10b→S10, S10c→S09 во всех таблицах, S01-S10c→S01-S10 |
| `контекст/настраиваемость.md` | Удалены строки activity_types/activities (старая модель), удален дубликат "Сезонные наборы" |
| `глоссарий.md` | S01-S10c → S01-S10 |
| `задачи/README.md` | Добавлена веха M05 |
| `задачи/M05-унификация-engagement/overview.md` | Описание вехи + exit criteria |
| `задачи/M05-унификация-engagement/plan.yaml` | Все чекбоксы отмечены ✅ |

### Исправлено (2 опечатки)

| Файл | Что исправлено |
|------|---|
| `архитектура/engagement.md` | "EnagementType" → "EngagementType", "Maппинг" → "Маппинг" |

### 16 критериев приёмки — покрытие

| # | Критерий | Статус |
|--|--|--|
| 1 | engagement.md создан с YAML-схемами (5 сущностей) | ✅ |
| 2 | type: "benefit" \| "activity" в EngagementType | ✅ |
| 3 | billing_direction: debit \| credit в EngagementOffer | ✅ |
| 4 | EngagementFlow заменяет ActivationFlow + completion_criteria | ✅ |
| 5 | 4 примера: ДМС, фитнес, опрос, реферал | ✅ |
| 6 | ASCII-диаграмма связей | ✅ |
| 7 | льготы.md + активности.md → stub redirect | ✅ |
| 8 | модули.md обновлён | ✅ |
| 9 | биллинг-движок.md обновлён | ✅ |
| 10 | README.md обновлён | ✅ |
| 11 | api.md переписан (объединённые endpoints + RBAC guards) | ✅ |
| 12 | артефакты.md обновлён (S09+S10a → Engagement Wizard) | ✅ |
| 13 | критерии-приёмки.md обновлён | ✅ |
| 14 | journeys (сотрудник, hr, менеджер-каталога) обновлены | ✅ |
| 15 | настраиваемость.md — одна модель | ✅ |
| 16 | RBAC-таблица: access control по type | ✅ |

## Проблемы и риски

Нет. Все изменения — документационные. Код не трогают.

## Замечания

1. Предыдущая сессия пометила чекбоксы как выполненные, но не внесла часть изменений в файлы. Эта сессия выполнила фактическое применение:
   - Удалена устаревшая секция Activities из `спецификация/api.md` (5 endpoints → объединены в User Engagements)
   - Все ссылки S10a/S10b/S10c заменены на S09/S10
   - Удалены устаревшие строки из `настраиваемость.md`
   - Исправлены опечатки в `engagement.md`

## Следующие шаги

T0501 завершен. Следующие вехи (M06–M13) теперь могут использовать единую абстракцию Engagement.