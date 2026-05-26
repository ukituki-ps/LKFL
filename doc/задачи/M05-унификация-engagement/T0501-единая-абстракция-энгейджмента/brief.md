# T0501 — Единая абстракция «Энгейджмент» (Engagement)

## Контекст

После M03 (пересборка льгот на 5 сущностей) выявлена **топологическая эквивалентность** модулей «Льготы» и «Активности»:

| Льготы (после M03) | Активности | Одинаковая роль |
|--|--|--|
| Benefit (продукт) | Activity Type (шаблон) | Конфигурируемый тип |
| BenefitPlan (тариф) | Activity (экземпляр) | Конкретный экземпляр с параметрами |
| ActivationFlow (поток шагов) | *(отсутствует — flat completion)* | Цепочка действий |
| UserBenefit (экземпляр юзера) | Completion (выполнение) | Пользователь + статус + данные |
| billing_rule_id → debit | billing_rule_id → credit | Связь с биллингом |
| Eligibility (AND/OR) | Target Audience (segments) | Фильтр доступа |

**Обе системы решают одну задачу:** сотрудник взаимодействует с платформой → получает/тратит баллы. Разница только в направлении биллинга (debit vs credit) и семантике.

## Проблема

Дублирование архитектурных паттернов:
- Две параллельные базы (5 таблиц льгот + 3 таблицы активностей)
- Два параллельных API-модуля (User Benefits + Activities)
- Два параллельных админских модуля (Admin Benefits + Admin Activities)
- Два разных wizard'а (Activation Wizard + Survey Form / Event Checkin)
- Фильтры каталога не объединены — сотрудник видит льготы и активности на разных страницах

## Решение

**Единая абстракция `Engagement`** с дискриминатором `type`:

```
EngagementType          (было: Benefit + Activity Type)
  type: "benefit" | "activity"      ← новое поле-дискриминатор
  ↓ 1:N
EngagementOffer         (было: BenefitPlan + Activity)
  billing_direction: debit | credit  ← replaces implicit direction by type
  activation_flow_id: uuid           ← для benefit; для activity — auto-generated
  ↓ N:1
EngagementFlow          (было: ActivationFlow + completion_criteria)
  steps: []Step                        ← унифицированный формат шагов
  ↓ N:1
UserEngagement          (было: UserBenefit + Completion)
  status: pending | in_progress | approved | active | completed | failed | expired
  billing_event: debit | credit        ← явное направление биллинга
Eligibility             (было: Eligibility + Target Audience)
  ← один формат AND/OR правил для обоих типов
```

### Что такое Engagement

Engagement = любое взаимодействие сотрудника, которое:
- Имеет конфигурируемый шаблон (тип)
- Имеет конкретный экземпляр (оффер)
- Имеет поток выполнения (flow)
- Привязано к биллингу (debit или credit)
- Имеет экземпляр у сотрудника (user_engagement)
- Имеет правила доступа (eligibility)

**`type: "benefit"`** — сотрудник тратит баллы (debit). Примеры: ДМС, фитнес, обучение, подарочная карта.
**`type: "activity"`** — сотрудник получает баллы (credit). Примеры: опрос, событие, реферал, ENPS.

### Ключевые изменения

1. **Новое поле `type: "benefit" | "activity"`** в EngagementType — дискриминатор режима
2. **Новое поле `billing_direction: debit | credit`** в EngagementOffer — явное направление (не подразумевается из type)
3. **EngagementFlow** заменяет ActivationFlow (льготы) И completion_criteria (активности) — один формат шагов
4. **UserEngagement** заменяет UserBenefit + Completion — одна статус-машина
5. **Eligibility** унифицирует `eligibility` (льготы) + `target_audience` (активности)
6. **Каталог объединён** — сотрудник видит льготы и активности в одном списке, фильтрует по `type`

### Роль новой абстракции в RBAC

| Роль | Льготы (type=benefit) | Активности (type=activity) |
|--|--|--|
| Сотрудник (`employee`) | Может подключать (debit) | Может выполнять (credit) |
| HR (`hr`) | Нет доступа к управлению | Управляет через H03 (создание, метрики) |
| Менеджер каталога (`catalog_manager`) | Управляет CRUD (M01) | Нет доступа |
| Admin (`admin`) | Полный доступ | Полный доступ |

Новые доступы:
- **HR** может управлять Activity-энгейджментами через `/admin/engagements?type=activity`
- **Catalog Manager** может управлять Benefit-энгейджментами через `/admin/engagements?type=benefit`
- Фильтрация по `type` на уровне API middleware (RBAC + type guard)

## Зависимости

- **T0301** (M03-архитектура-льгот) — предшествующая задача. Без M03 нет 5 сущностей льгот для унификации.
- **T0401** (M04-api-под-льготы) — предшествующая задача. API уже переведён на 5 сущностей.

## Файлы-мишени

Каждый файл требует обновления для согласованности с новой абстракцией.

### Архитектура (6 файлов)
- `архитектура/льготы.md` → переписать как часть Engagement
- `архитектура/активности.md` → переписать как часть Engagement
- `архитектура/engagement.md` — **НОВЫЙ** — полный документ унифицированной модели
- `архитектура/модули.md` → ссылки+схема сервисов под Engagement
- `архитектура/биллинг-движок.md` → event-контракты engagement, не benefit/activity
- `архитектура/README.md` → обновить nav-ссылки

### Спецификация (4 файла)
- `спецификация/api.md` → объединённые endpoints, type-фильтрация, RBAC guards
- `спецификация/артефакты.md` → S09+S10a объединены в один S09 «Engagement Wizard»
- `спецификация/критерии-приёмки.md` → обновитьjourneys-ссылки
- `спецификация/journeys/сотрудник.md` → J02, J07, J12 через engagement

### Контекст (1 файл)
- `контекст/настраиваемость.md` → одна модель вместо двух

### Задачи (2 файла)
- `задачи/README.md` → добавить веху M05
- `задачи/M05-унификация-энгейджмента/` → новая веха

## Критерии приёмки

1. ✅ Документ `архитектура/engagement.md` создан с полными YAML-схемами (5 сущностей)
2. ✅ Поле `type: "benefit" | "activity"` в EngagementType
3. ✅ Поле `billing_direction: debit | credit` в EngagementOffer
4. ✅ EngagementFlow заменяет ActivationFlow + completion_criteria
5. ✅ 4 примера заполнения: ДМС (benefit), фитнес (benefit), опрос (activity), реферал (activity)
6. ✅ ASCII-диаграмма связей между сущностями
7. ✅ `архитектура/льготы.md` + `архитектура/активности.md` заменены пересылкой (stub → engagement.md)
8. ✅ `архитектура/модули.md` обновлен (Platform модули, NATS subjects, Asynq workers)
9. ✅ `архитектура/биллинг-движок.md` обновлен (event-контракты engagement)
10. ✅ `архитектура/README.md` обновлен (nav-table)
11. ✅ `спецификация/api.md` переписан (объединённые endpoints + RBAC type guards)
12. ✅ `спецификация/артефакты.md` обновлен (S09+S10a → Engagement Wizard)
13. ✅ `спецификация/критерии-приёмки.md` обновлен
14. ✅ `спецификация/journeys/сотрудник.md`, `hr.md`, `менеджер-каталога.md` обновлены
15. ✅ `контекст/настраиваемость.md` — одна строка вместо двух
16. ✅ RBAC-таблица: access control по type для каждой роли
