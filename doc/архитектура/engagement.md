# Архитектура Энгейджмент (Engagement) — детальное описание

## TL;DR (для агентов)

> Этот файл описывает **Engagement** — единую абстракцию любого взаимодействия сотрудника с платформой. 4 подпакета.
> - **EngagementType (тип)** → строка 43 | **EngagementOffer (оффер)** → строка 87
> - **EngagementFlow (поток выполнения)** → строка 158 | **UserEngagement (экземпляр у сотрудника)** → строка 286
> - **Интеграция с биллингом/compliance/eligibility** → строка 381
> - **Eligibility через CEL** → строка 436 | **EngagementCollection (наборы)** → строка 474
> - **Диаграмма связей** → строка 509 | **Примеры заполнения** → строка 545
> - **Публичный API `engagement/`** → `пакеты-platform.md` строка 553
> - **Геймификация через engagement** → строка 919

> **Связь:** `архитектура/модули.md` → Platform `internal/engagement/` модуль (4 подпакета: catalog/, flow/, collections/, survey/). `архитектура/пакеты-platform.md` → детализация 16 пакетов монолита. `архитектура/пакеты-platform.md` → `internal/eligibility/` (M07 T0701). `контекст/настраиваемость.md` → Каталог энгейджментов.
> `спецификация/journeys/сотрудник.md` → J02–J08, J12, J13a. `спецификация/journeys/hr.md` → J18, J20a.

---

## Содержание

| Раздел | Строка |
|--------|--------|
| Концепция (Engagement — единая абстракция) | 19 |
| EngagementType — тип энгейджмента | 54 |
| EngagementOffer — конкретный оффер | 98 |
| EngagementFlow — поток выполнения | 169 |
| UserEngagement — экземпляр у сотрудника | 297 |
| Интеграция с остальной архитектурой | 392 |
| Eligibility — условия доступа (CEL) | 447 |
| EngagementCollection — набор энгейджментов | 485 |
| Диаграмма связей | 520 |
| Примеры заполнения | 556 |
| Миграция от старых сущностей | 830 |
| Интеграция с Геймификацией | 930 |
| Индексация БД | 983 |

---

## Концепция

Engagement — единая абстракция любого взаимодействия сотрудника с платформой, которое связано с биллингом.

```
EngagementType (тип) → EngagementOffer (оффер) → EngagementFlow (поток)
                                                       ↓
                                                  UserEngagement (юзер)
                                                       ↓
                                               Billing (debit / credit)
```

Дискриминатор `type` разделяет семантику:

| Поле | value | Биллинг | Пример |
|------|-------|---------|--------|
| `type: "benefit"` | Льгота | debit (списание баллов) | ДМС, фитнес, обучение, подарочная карта |
| `type: "activity"` | Активность | credit (начисление баллов) | опрос, событие, реферал, ENPS, health check |

### Маппинг из старых сущностей

```
ЛЬГОТЫ (M03)            АКТИВНОСТИ           → ENGAGEMENT
───────────             ──────────            ────────────
Benefit                 Activity Type   →     EngagementType
BenefitPlan             Activity        →     EngagementOffer
ActivationFlow          —               →     EngagementFlow
Completion              —               →     (встроено в UserEngagement)
UserBenefit             Completion      →     UserEngagement
Eligibility             Target Audience →     Eligibility
BenefitCollection       —               →     EngagementCollection
```

---

## EngagementType — тип энгейджмента

Переиспользуемый шаблон. Создаётся админом (catalog_manager для benefit, hr для activity).

```yaml
engagement_type:
  id: uuid
  tenant_id: uuid

  # Дискриминатор
  type: "benefit" | "activity"

  # Идентификация
  slug: "dms-alpha" | "survey-q2" | "referral" | "event-sport-day"
  name: "ДМС АльфаСтрахование" | "Опрос Q2 2026"
  description: "Добровольное медицинское страхование" | "Опрос удовлетворённости"
  icon_url: "/assets/engagement/dms.svg"              # опционально

  # Классификация
  category_id: uuid                                     # benefit_categories или activity_categories
  provider_adapter: "alpha"                              # только для benefit (slug адаптера)

  # Промо (только для benefit)
  is_promoted: false
  promo_banner_url: ""

  # Статус
  catalog_status: available | seasonal | archived

  # Мета
  created_by: user_id
  created_at: timestamp
```

### Различия по типу

| Поле | benefit | activity |
|------|---------|----------|
| `provider_adapter` | обязателен (slug) | пустой (нет внешн. провайдера) |
| `is_promoted`, `promo_banner_url` | используется | игнорируется |
| `catalog_status` | управляет видимостью в каталоге | управляет доступностью |

---

## EngagementOffer — конкретный оффер

Один EngagementType → N EngagementOffers.

Для benefit — это тарифы (базовый/расширенный/премиум ДМС). Для activity — конкретный экземпляр с датой, бонусом и аудиторией.

```yaml
engagement_offer:
  id: uuid
  tenant_id: uuid
  type_id: uuid                                     # родительский EngagementType

  # Идентификация
  name: "Базовая программа" | "Опрос Q2 — отдел IT"
  description: "Терапевт, стоматолог, лаборатория" | "5 вопросов, бонус 500 баллов"
  tier_level: 1                                     # порядок сортировки

  # Биллинг — явное направление
  billing_direction: debit | credit                 # debit=списание, credit=начисление
  billing_rule_id: uuid                             # ссылка на Billing Rule

  # Стоимость / Награда
  cost:
    amount: 3500 | 500                              # сумма (debit) или бонус (credit)
    currency: "points"                              # open string — из tenant.currency_config
    billing_model: subscription | one_time | per_use | per_session
    subscription_period: monthly | quarterly | yearly   # только для subscription

  # Поток выполнения
  flow_id: uuid                                     # ссылка на EngagementFlow

  # Условия доступа (ADR-021: CEL expression)
  eligibility_cel: "user.grade in ['A', 'B', 'C'] && user.years_of_service >= 0"
  eligibility_source: "грейд A, B или C и любой стаж"

  # Период действия
  start_date: date                                  # для activity — начало, для benefit — effective_from
  end_date: date                                    # опционально

  # Доплаты (только для benefit, debit)
  co_payment:
    enabled: false
    methods: []                                      # card | payroll_deduction

  # Методы ограничения (только для activity, credit)
  cooldown: "7d"                                    # мин. интервал между выполнениями
  max_completions: 3 | "unlimited"
  max_completions_scope: "global" | "per_period" | "per_user"

  # Метаданные (произвольные данные провайдера / UI)
  offer_metadata: JSONB

  # Статус
  status: active | paused | archived
```

### Унификация benefit vs activity

| Поле | benefit (debit) | activity (credit) |
|------|:-:|:-:|
| `billing_direction` | debit | credit |
| `billing_model` | subscription / per_use / one_time | one_time (всегда) |
| `cost.amount` | цена льготы | бонус за выполнение |
| `start_date` / `end_date` | effective_from / effective_to | период активности |
| `co_payment` | используется | игнорируется |
| `cooldown` / `max_completions` | игнорируется | используется |
| `tier_level` | тариф внутри продукта | — (обычно 1) |
| `offer_metadata` | `plan_metadata` (клиники, покрытия) | form config, UI hints |

---

## EngagementFlow — поток выполнения

Унифицированный формат шагов для любого типа. Для benefit — цепочка активации. Для activity — цепочка выполнения.

```yaml
engagement_flow:
  id: uuid
  tenant_id: uuid
  name: "Подключение ДМС с аппрувом" | "Опрос + начисление"
  description: "Анкета → Аппрув HR → Provider API" | "N вопросов → completion_criteria → credit"

  # Шаги (упорядочены)
  steps:
    - order: 1
      type: "form" | "approval" | "provider_api" | "provider_redirect"
            | "instant" | "document_generation" | "condition_check"
      # Данные шага (зависят от type)
      form_schema: JSONB                         # для form
      approval_role: "hr"                        # для approval
      auto_approve_after: "48h"                  # для approval (опционально)
      adapter_method: "Activate"                 # для provider_api
      redirect_url: "https://..."                # для provider_redirect
      sso: true                                  # для provider_redirect
      template: "dms_policy"                     # для document_generation
      condition_expr: 'answers.size() >= 5'       # для condition_check (ADR-021: CEL expression)
      condition_source: 'заполнено не менее 5 ответов'     # для condition_check (русский текст → LLM → CEL)
      ui_component: "SurveyForm"                 # React-компонент

  # Временные ограничения
  max_duration: "7d"                              # макс. время прохождения потока

  # Статус
  status: active | archived
```

> **ADR-021:** `condition_expr` — CEL expression. HR вводит `condition_source` на русском → LLM Proxy генерирует CEL → system validates → saved.
> Evaluation: `cel.Evaluator.Evaluate(condition_expr, CELContext)` — в процессе Platform.
> **M13 T1301:** При `ui_component: "SurveyForm"` → будет парситься `survey_schema` из `engagement/survey/` (поддержка бранчинга, tag-mapping) — реализация в M14.
> При `ui_component: "EngagementForm"` → legacy `form_schema` (обратно совместимо).

### Типы шагов

| Тип | Описание | benefit | activity |
|-----|----------|:-------:|:--------:|
| `instant` | Автоматическая активация через ProviderAdapter.Activate() | ✅ | — |
| `form` | Заполнение формы (JSON-схема → React) | ✅ | ✅ |
| `approval` | Ожидание аппрува роли (HR) | ✅ | ✅ |
| `provider_redirect` | Перенаправление на сайт провайдера (SSO) | ✅ | — |
| `provider_api` | Автоматическая активация через API провайдера | ✅ | — |
| `document_generation` | Генерация PDF-документа | ✅ | — |
| `condition_check` | Проверка условия завершения (**ADR-021:** CEL expression) | — | ✅ |

### Примеры потоков

**ДМС (benefit):**
```yaml
steps:
  - order: 1
    type: "form"
    form_schema:
      fields:
        - name: "surname"
          type: "text"
          required: true
        - name: "dob"
          type: "date"
          required: true
    ui_component: "EngagementForm"
  - order: 2
    type: "approval"
    approval_role: "hr"
    auto_approve_after: "72h"
  - order: 3
    type: "provider_api"
    adapter_method: "Activate"
  - order: 4
    type: "document_generation"
    template: "dms_policy"
```

**Фитнес (benefit, instant):**
```yaml
steps:
  - order: 1
    type: "instant"
    adapter_method: "Activate"
```

**Опрос (activity):**
```yaml
steps:
  - order: 1
    type: "form"
    form_schema:
      questions:
        - id: "q1"
          text: "Оцените уровень стресса (1-10)"
          type: "scale"
        - id: "q2"
          text: "Что можно улучшить?"
          type: "textarea"
    ui_component: "SurveyForm"
  - order: 2
    type: "condition_check"
    condition_source: "заполнено не менее 2 ответов и все обязательные поля"
    condition_expr: 'answers.size() >= 2 && context.answers_all_required_filled == true'
```

**Событие (activity):**
```yaml
steps:
  - order: 1
    type: "condition_check"
    condition_source: "сотрудник пришёл на мероприятие"
    condition_expr: 'context.check_in == true'
```

**Реферал (activity):**
```yaml
steps:
  - order: 1
    type: "condition_check"
    condition_source: "приглашённый пользователь зарегистрировался"
    condition_expr: 'context.referred_user.status == "registered"'
```

---

## UserEngagement — экземпляр у сотрудника

Заменяет UserBenefit + Completion. Одна запись на одно взаимодействие сотрудника с оффером.

```yaml
user_engagement:
  id: uuid
  tenant_id: uuid
  user_id: uuid
  offer_id: uuid                                    # ссылка на EngagementOffer

  # Идентификация типа
  engagement_type: "benefit" | "activity"           # дублируется из parent (для быстрых запросов)

  # Поток выполнения
  flow_status: pending | in_progress | approved | active | completed | failed | expired
  current_step: 1                                    # текущий шаг в EngagementFlow
  started_at: timestamp
  completed_at: timestamp                            # когда стал active/completed

  # Данные формы (из form-шагов)
  form_data: JSONB

  # Финансы
  billing_direction: debit | credit                 # дублируется из offer
  billing_amount: 3500 | 500
  currency: "points"
  last_billing_at: timestamp                        # для subscription/per_use
  billing_count: 0                                  # кол-во списаний/начислений
  billing_transaction_id: uuid                      # ссылка на последнюю транзакцию

  # Период действия у сотрудника
  valid_from: date
  valid_to: date                                    # опционально

  # Связь с набором (если подключено из набора)
  collection_id: uuid                               # опционально — EngagementCollection

  # Уведомления
  notifications_sent:
    start: boolean
    reminder: boolean
    completed: boolean

  # Мета
  created_at: timestamp
  updated_at: timestamp
```

### Статус-машина

**Для benefit (debit):**
```
pending ──→ in_progress ──→ approved ──→ active
   │              │              │           │
   │              │              │           ▼
   │              │              │       expired
   │              │              │
   │              ▼              │
   └─── failed ←─────────────────┘
```

**Для activity (credit):**
```
in_progress ──→ completed
       │              │
       ▼              │
    failed ───────────┘
```

| Переход | Триггер | benefit | activity |
|---------|---------|---------|----------|
| `pending → in_progress` | Нажал «Подключить» / «Начать» | ✅ | ✅ |
| `in_progress → approved` | HR одобрил / автоаппрув | ✅ | ✅ |
| `in_progress → active` | Instant / provider_api OK | ✅ | — |
| `approved → active` | Provider API OK / redirect OK | ✅ | — |
| `in_progress → completed` | condition_check passed | — | ✅ |
| `active/completed → expired` | valid_to < now | ✅ | ✅ |
| `* → failed` | Ошибка / отклонение | ✅ | ✅ |

### Биллинг-события

> ⚠️ **Historical (pre-M12):** NATS subjects registry (`nats-subjects.md`) удалён. Все subjects заменены на direct Go calls. См. [ADR-024](./adr/024-modular-monolith.md) и [ADR-020 ❌ Superseded](./adr/020-nats-subjects-registry.md).
> **P0:** все mutating события содержат `idempotency_key` в payload. Billing проверяет уникальность перед применением.

| transition | billing event (subject) | direction | payload |
|------------|---------|-----------|---------|
| `pending → in_progress` (benefit) | `billing.debit.reserve` | debit (заморозка) | `{ idempotency_key, userId, amount, category, offerId }` |
| `approved → active` (benefit) | `billing.debit.confirm` | debit (подтверждение) | `{ idempotency_key, userId, amount, category, offerId }` |
| `in_progress → active` (benefit, instant) | `billing.debit.confirm` | debit | `{ idempotency_key, userId, amount, category, offerId }` |
| `in_progress → completed` (activity) | `billing.credit` | credit | `{ idempotency_key, userId, amount, category, source, periodId }` |
| `failed` (любой) | `billing.debit.reverse` | credit (возврат) | `{ idempotency_key, userId, amount, category, offerId }` |

---

## Интеграция с остальной архитектурой

### EngagementOffer ↔ Billing

| Контракт | Описание |
|---|----|
| `billing_rule_id` → rule | billing engine ищет rule по ID |
| `billing_direction = debit` | Platform публикует `billing.debit.reserve` / `billing.debit.confirm` |
| `billing_direction = credit` | Platform публикует `billing.credit` |
| `billing_model = subscription` | Billing Rule trigger = `cron`, frequency = monthly/quarterly/yearly |
| `billing_model = per_use` | Billing Rule trigger = `event`, Platform публикует при каждом обращении |
| `billing_model = one_time` | Billing Rule trigger = `event`, frequency = one-time |
| `billing_model = per_session` | Аналогично per_use, сгруппированно по сессиям |

### EngagementType ↔ Integrations (только benefit)

| Контракт | Описание |
|---|----|
| `provider_adapter` → slug | Ссылка на адаптер в Integration Proxy. Adapter: Activate(), Deactivate(), Status(), SyncCatalog() |
| EngagementFlow.step.provider_api | Вызывает Adapter.Activate(ctx, req) |
| EngagementFlow.step.provider_redirect | Redirect на URL провайдера, SSO через Keycloak Identity Broker |

### UserEngagement ↔ Notification

| Событие | Уведомление |
|---|-|
| `pending → in_progress` | «Заявка на [название] отправлена» |
| `in_progress → approved` | «Заявка на [название] одобрена» |
| `approved → active` (benefit) | «[Название] подключено» + push/email + лента событий |
| `in_progress → completed` (activity) | «+N баллов за [название]» + push/email + лента событий |
| `active/completed → expired` | «[Название] истекает через 30 дней» → «истекло» |
| `* → failed` | «Не удалось: [причина]» |

### EngagementCollection ↔ Frontend

| Компонент | Описание |
|---|-|
| `CollectionBanner` | Баннер набора на Главной (S03) и в Каталоге (S04) |
| `CollectionCard` | Карточка набора: состав, цена, скидка, кнопка «Подключить» |
| `BundleCheckout` | Модалка подтверждения → проверка баланса → подключение всех offers |

### RBAC — доступ по type

| Роль | type=benefit | type=activity |
|-----|-|-----|
| `employee` | Может подключать (debit) | Может выполнять (credit) |
| `hr` | Нет доступа CRUD | Admin: создание, метрики, апрув |
| `catalog_manager` | Admin: CRUD type, offer, flow | Нет доступа |
| `admin` | Полный доступ | Полный доступ |

Фильтрация на уровне API middleware:
- HR requests → `?type=activity` auto-applied + middleware guard
- Catalog Manager requests → `?type=benefit` auto-applied + middleware guard
- Admin requests → оба типа

## Eligibility — условия доступа (ADR-021: CEL)

Унифицированный формат для eligibility (льготы) и target_audience (активности).

> **ADR-021:** eligibility выражается через CEL expression. HR вводит текст на русском → LLM генерирует CEL → system validates → saved как `eligibility_cel`.
> Старый формат AND/OR/Groups/segments — **legacy, удалён**.

```yaml
# Canonical формат (ADR-021):
eligibility_cel: "user.grade in ['A', 'B'] && user.years_of_service >= 3"
eligibility_source: "грейд A или B и стаж от 3 лет"

# Сложные условия:
eligibility_cel: "user.grade == 'A' || (user.grade == 'B' && user.years_of_service >= 5)"
eligibility_source: "грейд A ИЛИ (грейд B и стаж от 5 лет)"
```

### Доступные поля в eligibility_cel

Из профиля пользователя (Keycloak + hr-sync):

| CEL-поле | Описание | Пример |
|----------|----------|--------|
| `user.grade` | грейд | `user.grade in ['A', 'B']` |
| `user.years_of_service` | стаж (лет) | `user.years_of_service >= 3` |
| `user.has_children` | есть дети | `user.has_children == true` |
| `user.department` | отдел | `str_contains(user.department, 'IT')` |
| `user.status` | активен / уволен | `user.status == 'active'` |
| `user.tenant_id` | tenant | — |
| `tags.*` | динамические теги | `tags.is_remote == true` |

### Legacy формат (удалён)

> **Удалено после ADR-021.** Старый YAML-формат AND/OR/Groups больше не используется.
> Все существующие правила миграированы в `eligibility_cel` через миграционный скрипт.

---

## EngagementCollection — набор энгэйджментов

Заменяет BenefitCollection. Содержит только benefit-офферы.

```yaml
engagement_collection:
  id: uuid
  tenant_id: uuid
  name: "Зима 2026 — Зимние виды спорта"
  description: "Горные лыжи, сноуборд, хоккей — специальные цены"
  banner_url: "/assets/collections/winter2026.jpg"

  # Состав набора (только benefit-офферы)
  offers:
    - offer_id: uuid
      quantity: 1
    - offer_id: uuid
      quantity: 1

  # Общая стоимость (опционально — скидка за набор)
  bundle_price:
    amount: 1200
    currency: "points"
    discount_percent: 10

  # Период действия
  effective_from: date
  effective_to: date

  # Статус
  status: active | archived
```

---

## Диаграмма связей

```
┌─┬──────────────┬───────────┬──────┐     ┌──────────────────┐
│  EngagementType│  type:    │      │
│  (шпаблон)   │  benefit/activity  │
│                  slug, name,│ icon │
│                  category, adapter│
└──┬─────────────┴───────────┴──────┘
   │ 1:N
   ▼
┌──────────────┬──────────────────┐
│  EngagementOffer   billing_dir: │
│  (оффер)      │  debit/credit     │
│                  cost, flow_id   │
│                  eligibility     │
│                  billing_rule_id │
└──┬─────────────┴────────────────┘
   │ N:1              │ 1:N
   ▼                  ▼
┌──────────────┐  ┌──────────────────┐
│ EngagementFlow│  │ UserEngagement    │
│ (поток)      │  │ (экземпляр юзера) │
│ steps[]      │  │ flow_status       │
│              │  │ billing_event     │
└──────────────┘  │ form_data         │
                   └──────────────────┘

┌──────────────────┐     ┌──────────────┐
│ EngagementCollection│◄──│ EngagementOffer│
│ (набор)            │     │ (в наборе N)  │
└──────────────────┘     └──────────────┘
```

---

## Примеры заполнения

### ДМС (type=benefit, 3 тарифа)

```yaml
engagement_type:
  id: "et-dms-alpha"
  tenant_id: "t-sdek"
  type: "benefit"
  slug: "dms-alpha"
  name: "ДМС АльфаСтрахование"
  description: "Добровольное медицинское страхование"
  icon_url: "/assets/engagement/alpha.svg"
  category_id: "cat-health"
  provider_adapter: "alpha"
  catalog_status: available

engagement_offers:
  # Тариф 1 — Базовый
  - id: "eo-dms-basic"
    type_id: "et-dms-alpha"
    name: "Базовая программа"
    description: "Терапевт, стоматолог, лаборатория"
    tier_level: 1
    billing_direction: debit
    cost:
      amount: 3500
      currency: "points"
      billing_model: subscription
      subscription_period: monthly
    flow_id: "ef-dms-full"
    eligibility_cel: "user.grade in ['A', 'B', 'C'] && user.years_of_service >= 0"
    eligibility_source: "грейд A, B или C и любой стаж"
    billing_rule_id: "rule-debit-dms-basic"
    status: active

  # Тариф 2 — Расширенный
  - id: "eo-dms-ext"
    type_id: "et-dms-alpha"
    name: "Расширенная программа"
    description: "Базовая + офтальмология, физиотерапия, педиатр"
    tier_level: 2
    billing_direction: debit
    cost:
      amount: 5000
      currency: "points"
      billing_model: subscription
      subscription_period: monthly
    flow_id: "ef-dms-full"
    eligibility_cel: "user.grade in ['A', 'B'] && user.years_of_service >= 1"
    eligibility_source: "грейд A или B и стаж от 1 года"
    billing_rule_id: "rule-debit-dms-extended"
    status: active

  # Тариф 3 — Премиум
  - id: "eo-dms-prem"
    type_id: "et-dms-alpha"
    name: "Премиум"
    description: "Расширенная + VIP-клиники, роддом"
    tier_level: 3
    billing_direction: debit
    cost:
      amount: 7500
      currency: "points"
      billing_model: subscription
      subscription_period: monthly
    flow_id: "ef-dms-full"
    eligibility_cel: "user.grade == 'A' && user.years_of_service >= 3"
    eligibility_source: "грейд A и стаж от 3 лет"
    billing_rule_id: "rule-debit-dms-premium"
    status: active

engagement_flow:
  id: "ef-dms-full"
  name: "Подключение ДМС с аппрувом"
  steps:
    - order: 1
      type: "form"
      form_schema:
        fields:
          - name: "surname"
            type: "text"
            required: true
          - name: "dob"
            type: "date"
            required: true
          - name: "gender"
            type: "select"
            options: ["M", "F"]
      ui_component: "EngagementForm"
    - order: 2
      type: "approval"
      approval_role: "hr"
      auto_approve_after: "72h"
    - order: 3
      type: "provider_api"
      adapter_method: "Activate"
    - order: 4
      type: "document_generation"
      template: "dms_policy"
  max_duration: "14d"
  status: active

# Экземпляр у сотрудника
user_engagement:
  id: "ue-ivanov-dms"
  user_id: "user-ivanov"
  offer_id: "eo-dms-ext"                     # Расширенная программа
  engagement_type: "benefit"
  flow_status: active
  started_at: "2026-03-01T10:00:00Z"
  completed_at: "2026-03-05T14:30:00Z"
  billing_direction: debit
  billing_amount: 5000
  last_billing_at: "2026-04-01T00:00:00Z"
  billing_count: 2
  valid_from: "2026-03-01"
  valid_to: "2027-03-01"
```

### Фитнес World Class (type=benefit, 1 тариф)

```yaml
engagement_type:
  id: "et-fitness-wc"
  type: "benefit"
  slug: "fitness-world-class"
  name: "Фитнес World Class"
  description: "Абонемент на групповые занятия"
  icon_url: "/assets/engagement/worldclass.svg"
  category_id: "cat-fitness"
  provider_adapter: "worldclass"
  catalog_status: available

engagement_offer:
  id: "eo-fitness-group"
  type_id: "et-fitness-wc"
  name: "Групповые занятия"
  billing_direction: debit
  cost:
    amount: 500
    currency: "points"
    billing_model: subscription
    subscription_period: monthly
  flow_id: "ef-instant"
  eligibility_cel: "user.grade in ['A', 'B'] || (user.grade == 'C' && user.years_of_service >= 2)"
  eligibility_source: "грейд A или B ИЛИ (грейд C и стаж от 2 лет)"
  billing_rule_id: "rule-debit-fitness"
  status: active

engagement_flow:
  id: "ef-instant"
  name: "Мгновенная активация"
  steps:
    - order: 1
      type: "instant"
      adapter_method: "Activate"
  status: active
```

### Опрос Q2 (type=activity)

```yaml
engagement_type:
  id: "et-survey"
  type: "activity"
  slug: "survey"
  name: "Корпоративный опрос"
  description: "Опрос удовлетворённости"
  icon_url: "/assets/engagement/survey.svg"
  category_id: "cat-gamification"
  catalog_status: available

engagement_offer:
  id: "eo-survey-q2"
  type_id: "et-survey"
  name: "Опрос Q2 2026"
  billing_direction: credit
  cost:
    amount: 500
    currency: "points"
    billing_model: one_time
  flow_id: "ef-survey"
  eligibility_cel: "user.status == 'active'"
  eligibility_source: "активный сотрудник"
  billing_rule_id: "rule-credit-survey"
  cooldown: "7d"
  max_completions: 1
  max_completions_scope: "per_user"
  start_date: "2026-06-01"
  end_date: "2026-06-30"
  status: active

engagement_flow:
  id: "ef-survey"
  name: "Опрос → начисление"
  steps:
    - order: 1
      type: "form"
      form_schema:
        questions:
          - id: "stress"
            text: "Оцените стресс (1-10)"
            type: "scale"
            required: true
          - id: "feedback"
            text: "Что можно улучшить?"
            type: "textarea"
            required: false
      ui_component: "SurveyForm"
    - order: 2
      type: "condition_check"
      condition_expr: 'answers["stress"] != null'
  status: active

# Экземпляр у сотрудника
user_engagement:
  id: "ue-ivanov-survey"
  user_id: "user-ivanov"
  offer_id: "eo-survey-q2"
  engagement_type: "activity"
  flow_status: completed
  started_at: "2026-06-05T09:00:00Z"
  completed_at: "2026-06-05T09:15:00Z"
  billing_direction: credit
  billing_amount: 500
  billing_transaction_id: "tx-credit-500-ivanov"
  form_data:
    stress: 4
    feedback: "Больше гибких льгот"
```

### Реферал (type=activity)

```yaml
engagement_type:
  id: "et-referral"
  type: "activity"
  slug: "referral"
  name: "Реферальная программа"
  description: "Пригласите коллегу → получите бонус"
  icon_url: "/assets/engagement/referral.svg"
  category_id: "cat-gamification"
  catalog_status: available

engagement_offer:
  id: "eo-referral-default"
  type_id: "et-referral"
  name: "Реферал — стандарт"
  billing_direction: credit
  cost:
    amount: 300
    currency: "points"
    billing_model: one_time
  flow_id: "ef-referral"
  eligibility_cel: "user.status == 'active' && user.years_of_service >= 0.5"
  eligibility_source: "активный сотрудник со стажем от 6 месяцев"
  billing_rule_id: "rule-credit-referral"
  max_completions: "unlimited"
  max_completions_scope: "global"
  status: active

engagement_flow:
  id: "ef-referral"
  name: "Referral → check → credit"
  steps:
    - order: 1
      type: "condition_check"
      condition_expr: 'context.referred_user_status == "registered"'
  status: active
```

---

## Миграция от старых сущностей

### Benefit → EngagementType

| Поле (Benefit) | Поле (EngagementType) | Изменение |
|----------------|----------------------|-----------|
| `id` | `id` | без изменений |
| `tenant_id` | `tenant_id` | без изменений |
| `name`, `description`, `icon_url` | те же | без изменений |
| `category_id` | `category_id` | без изменений |
| `provider_adapter` | `provider_adapter` | без изменений |
| `catalog_status` | `catalog_status` | без изменений |
| `is_promoted`, `promo_banner_url` | те же | без изменений |
| `created_by`, `created_at` | те же | без изменений |
| — | `type: "benefit"` | **новое** — автоматически `benefit` для всех legacy |
| — | `slug` | **новое** — генерируется из name |

### BenefitPlan → EngagementOffer

| Поле (BenefitPlan) | Поле (EngagementOffer) | Изменение |
|-------------------|------------------------|-----------|
| `type_id` | `type_id` | benefit_id → type_id |
| `name`, `description`, `tier_level` | те же | без изменений |
| `cost.amount`, `cost.currency` | те же | без изменений |
| `billing_model`, `subscription_period` | те же | без изменений |
| `activation_flow_id` | `flow_id` | переименовано |
| `eligibility` | `eligibility_cel` / `eligibility_source` | **ADR-021:** миграция AND/OR/Groups → CEL expression + русский источник |
| `billing_rule_id` | `billing_rule_id` | без изменений |
| `co_payment` | `co_payment` | без изменений |
| `plan_metadata` | `offer_metadata` | переименовано |
| `effective_from/to` | `start_date`/`end_date` | переименовано |
| `status` | `status` | без изменений |
| — | `billing_direction: "debit"` | **новое** — автоматически `debit` для всех legacy |
| — | `cooldown`, `max_completions` | **новое** — по умолчанию пусто |

### ActivationFlow → EngagementFlow

| Поле (ActivationFlow) | Поле (EngagementFlow) | Изменение |
|----------------------|-----------------------|-----------|
| `id` | `id` | без изменений |
| `tenant_id` | `tenant_id` | без изменений |
| `name`, `description` | те же | без изменений |
| `steps[].order`, `type` | те же | без изменений |
| `steps[].form_schema`, `approval_role` | те же | без изменений |
| `steps[].adapter_method`, `redirect_url` | те же | без изменений |
| `max_duration` | `max_duration` | без изменений |
| `status` | `status` | без изменений |
| — | `steps[].type: "condition_check"` | **новое** — недоступно в ActivationFlow, только activity |

### Activity Type → EngagementType

| Поле (Activity Type) | Поле (EngagementType) | Изменение |
|---------------------|----------------------|-----------|
| `id` | `id` | без изменений |
| `tenant_id` | `tenant_id` | без изменений |
| `slug` | `slug` | без изменений |
| `name`, `description` | те же | без изменений |
| — | `type: "activity"` | **новое** — автоматически `activity` для всех legacy |
| — | `icon_url` | **новое** — пусто по умолчанию |
| — | `category_id` | **новое** |
| — | `provider_adapter: ""` | **новое** — пусто для activity |
| — | `catalog_status` | **новое** |

### Activity → EngagementOffer

| Поле (Activity) | Поле (EngagementOffer) | Изменение |
|----------------|-----------------------|-----------|
| `type_id` | `type_id` | без изменений |
| `name`, `description` | те же | без изменений |
| `reward.billing_rule_id` | `billing_rule_id` | переименовано |
| `start_date`, `end_date` | те же | без изменений |
| — | `billing_direction: "credit"` | **новое** |
| — | `cost.amount`, `cost.currency` | **новое** — из reward |
| — | `cost.billing_model: "one_time"` | **новое** |
| — | `flow_id` | **новое** — auto-generated flow с condition_check |
| `target_audience` | `eligibility` | **перемещено**, формат унифицирован |
| `cooldown`, `max_completions` | те же | без изменений |
| `notifications` | — | перенесено в UserEngagement.notifications_sent |

### Completion + UserBenefit → UserEngagement

| Поле | Source | Изменение |
|------|--------|-----------|
| `id` | новый UUID | генерируется |
| `user_id` |Completion.user_id + UserBenefit.user_id| объединено |
| `offer_id` | Completion.activity_id → offer_id; UserBenefit.plan_id → offer_id | переименовано |
| `engagement_type` | определяется по parent offer | **новое** |
| `flow_status` | UserBenefit.activation_status + Completion.status | объединено |
| `current_step` | UserBenefit.current_step | без изменений |
| `started_at`, `completed_at` | те же из обоих | объединено |
| `form_data` | UserBenefit.form_data + Completion.data | объединено |
| `billing_direction` | benefit → debit, activity → credit | **новое** |
| `billing_amount` | UserBenefit.total_cost / Completion.reward | унифицировано |
| `billing_transaction_id` | Completion.reward_transaction_id | переименовано |
| `last_billing_at`, `billing_count` | UserBenefit.те же | без изменений |
| `valid_from`, `valid_to` | UserBenefit.те же | без изменений |
| `collection_id` | UserBenefit.collection_id | без изменений |

---

## Интеграция с Геймификацией (M09, ADR-023)

FlowEngine.Complete() запускает `gamification/TriggerHandler.OnEngagementCompleted()` как **Go-callback** — не NATS, не event bus.

### Почему Go-callback, а не NATS?

- Gamification — in-process пакет внутри Platform, не отдельный сервис
- NATS-subscriber вводит race-condition: "billing credit начислен, но ачивка не присвоена"
- Callback гарантирует atomicity: после completion → award проверка → сохранение grants
- Нет отдельного NATS subscriber, нет дублирования обработки
- Проще тестировать: mock EngagementCompleter в unit-тестах FlowEngine

### Contract

```go
// в gamification/triggers.go
type EngagementCompleter interface {
    OnEngagementCompleted(ctx context.Context, engagementId uuid.UUID, userId uuid.UUID, category string) error
}

// FlowEngine получает через DI
type FlowEngine struct {
    // ... существующие поля ...
    gamificationTrigger EngagementCompleter // opционально, nil = без геймификации
}

func (f *FlowEngine) Complete(ctx context.Context, engagementId uuid.UUID) error {
    // ... существующая логика completion ...

    // M09: триггер геймификации (Go-callback, после billing events publish)
    if f.gamificationTrigger != nil {
        category := getCategory(engagementId) // из engagement type/offer
        err := f.gamificationTrigger.OnEngagementCompleted(ctx, engagementId, userId, category)
        if err != nil {
            f.logger.Warn("gamification trigger failed (best-effort)", "err", err)
            // НЕ return error — completion продолжается
        }
    }

    return nil
}
```

### Что делает OnEngagementCompleted

1. Загрузка всех achievement-шаблонов tenant'а с `trigger_on='engagement_completed'`
2. Для каждого шаблона: построение CELContext (`game.*` поля из DB) → CEL evaluation → TRUE → INSERT achievement_grant
3. Проверка loyalty upgrade: если engagement_count перешел порог следующего уровня → INSERT user_loyalty_level

**Best-effort:** ошибка геймификации логируется, но completion продолжается (не блокирует основной flow).

---

## Индексация БД

Рекомендуемые индексы (PostgreSQL):

```sql
-- Быстрые запросы каталога
CREATE INDEX idx_engagement_type_type ON engagement_types(tenant_id, catalog_status);
CREATE INDEX idx_engagement_type_category ON engagement_types(category_id);

-- Офферы типа
CREATE INDEX idx_engagement_offer_type ON engagement_offers(type_id, status);
CREATE INDEX idx_engagement_offer_billing ON engagement_offers(billing_direction);

-- Пользовательские экземпляры
CREATE INDEX idx_user_engagement_user ON user_engagements(user_id, engagement_type);
CREATE INDEX idx_user_engagement_status ON user_engagements(user_id, flow_status);
CREATE INDEX idx_user_engagement_offer ON user_engagements(offer_id);

-- Eligibility
CREATE INDEX idx_eligibility_segments ON eligibility USING gin(segments);

-- Для HR
CREATE INDEX idx_user_engagement_created ON user_engagements(created_at);
```
