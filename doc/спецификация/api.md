# API-контракты

## TL;DR (для агентов)

> Этот файл — **118 REST API endpoints** платформы. 936 строк.
> - **User/Consent** → строка 81 | **Dashboard** → строка 95
> - **Engagements (каталог, офферы, user engagements)** → строка 105
> - **Wizard Config** → строка 162 | **CEL Generation (admin)** → строка 195
> - **Balance/Billing/Payments** → строка 306–339
> - **Documents/Support/Notifications** → строка 354–415
> - **Admin endpoints (Users, Periods, Engagements, Analytics)** → строка 427–514
> - **Auth: Keycloak OIDC, JWT Bearer** → строка 69
> - **Пагинация: cursor-based** → строка 9

REST API для платформы гибких льгот. Все endpoints возвращают JSON. Аутентификация — Keycloak (OIDC). JWT Bearer Keycloak → api-gateway → backend.

> Roles в Keycloak: `employee`, `hr`, `catalog_manager`, `admin`, `integration_admin`.

---

## Пагинация

Все списочные endpoints используют **cursor-based pagination**.

**Запрос:**
```
GET /engagements?cursor=eyJpZCI6MTAwfQ&limit=20
```

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|-------------|----------|
| `cursor` | string | — | Base64-кодированный курсор (из `nextCursor` предыдущего ответа) |
| `limit` | int | 20 | Максимум 100 |

**Ответ:**
```json
{
  "data": [...],
  "pagination": {
    "nextCursor": "eyJpZCI6MTIwfQ",
    "hasMore": true,
    "total": 127
  }
}
```

При отсутствии `nextCursor` или `nextCursor = null` — это последняя страница.

---

## Формат ошибки

Все ошибки возвращают единый формат:

```json
{
  "error": {
    "code": "BENEFIT_NOT_FOUND",
    "message": "Льгота с ID не найдена",
    "details": {
      "benefitId": "abc-123"
    }
  }
}
```

| HTTP код | error.code | Описание |
|----------|-----------|----------|
| 400 | `VALIDATION_ERROR` | Неверные данные запроса |
| 401 | `UNAUTHORIZED` | Нет JWT / токен истёк |
| 403 | `FORBIDDEN` | Нет роли / tenant mismatch |
| 404 | `NOT_FOUND` | Ресурс не найден |
| 409 | `CONFLICT` | Дубликат / состояние несовместимо |
| 422 | `BUSINESS_ERROR` | Бизнес-правило нарушено (нехватка баллов, период закрыт) |
| 429 | `RATE_LIMITED` | Превышен лимит запросов |
| 500 | `INTERNAL_ERROR` | Внутренняя ошибка (не раскрывать детали клиенту) |
| 503 | `SERVICE_UNAVAILABLE` | Внешний сервис недоступен (provider, Keycloak) |

---

## Auth + Keycloak

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/auth/register` | POST | S01: кастомная регистрация — ФИО,email,ДР,телефон,пароль,согласия→валидация по реестру→создание user в Keycloak |

Всё остальное (login, refresh, logout, password reset) — стандартный Keycloak OIDC.

**1 endpoint.**

---

## User / Consent (S01, J13c)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/user/me` | GET | S01: профиль вне Keycloak: грейд, стаж, отдел, есть_дети |
| `/user/me` | PUT | S01: обновление контактов |
| `/user/consents` | GET | S01: список подписанных согласий |
| `/user/consents` | POST | S01: подписание |
| `/user/consents/revoke` | POST | J13c: отзыв ПДн → каскад → disable в Keycloak |

**5 endpoints.**

---

## Dashboard (S03)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/dashboard` | GET | S03: greeting, stat_cards, active_items, event_feed, quick_actions, recommendations, external_links |

**1 endpoint.**

---

## Engagements (S04, S09, S10, M01–M05, M06, J02–J08, J12, J13a)

> **Backend → `internal/engagement/catalog/`** (`M06`)

Единый каталог — льготы (benefit, debit) и активности (activity, credit). Один endpoint, фильтрация по `?type=`.

> **RBAC guard:** `employee` видит оба типа. `hr` видит только `type=activity`. `catalog_manager` видит только `type=benefit`.

| Endpoint | Method | Обоснование |
|---------|--------|-----|
| `/engagements` | GET | S04, S09: каталог энгэйджментов — пагинация, `?type=benefit\|activity`, `?category=`, `?status=`, `?search=` |
| `/engagements/:id` | GET | S10: детали engagement type — описание, имя, иконка, категория, адаптер (для benefit), список офферов (кратко) |
| `/engagements/types` | GET | S04, S09: фильтры — типы (`benefit` / `activity`) |
| `/engagements/categories` | GET | S04: фильтры — категории (конфигурируемый enum, объединяет benefit_categories + activity_categories) |

**4 endpoints.** RBAC middleware автоматически добавляет `?type=` на основе роли.

---

## Engagement Offers (S10, S09, M06, J02, J05, J12)

> **Backend → `internal/eligibility/engine.go`** (M07: eligibility вынесен в отдельный пакет, T0701, ADR-014)

Конкретные офферы (тарифы льготы / экземпляры активностей). Каждый offer со своей ценой, условиями, потоком.

| Endpoint | Method | Обоснование |
|---------|--------|-----|
| `/engagement-offers/:id` | GET | S10: детали оффера — cost, billing_direction (debit/credit), billing_model, eligibility, flow |
| `/engagement-offers/:id/check-eligibility` | POST | J02, J12: «могу ли я подключить?» → проверяет eligibility → `{eligible: bool, reasons: []}` |

**2 endpoints.**

---

## User Engagements (S03, S04, S09, M06, J02, J05, J07, J12)

> **Backend → `internal/engagement/flow/`** (M06: flow execution engine)
> **M07 T0706:** wizard engine — frontend получает `?include=wizard-config` → сервер возвращает flow steps как JSON wizard config для generic `Wizard.tsx` (ADR-019)

Экземпляр engagement конкретного сотрудника (UserEngagement). Заменяет UserBenefit + Completion.
Создание benefit = подключение (debit). Создание activity = начало выполнения (→ credit).

| Endpoint | Method |Обоснование |
|---------|--------|-----|
| `/user-engagements` | GET | S03: список подключённых и выполненных — `?type=benefit\|activity`, `?status=PENDING\|ACTIVE\|COMPLETED\|...`, plan/offer, status, activated_at, valid_to, billing_count |
| `/user-engagements` | POST | J02, J07, J12: подключение benefit / начало activity → создаёт user_engagement → запускает flow → Billing Rule (debit или credit) |
| `/user-engagements/:id` | GET | S09: детали — оффер, статус, шаги flow, form_data, документы |
| `/user-engagements/:id/steps/:stepId` | GET | S09: детали текущего шага engagement flow |
| `/user-engagements/:id/steps/:stepId/submit` | POST | S09: отправить данные шага (form → данные, approval → запрос аппрува, condition_check → выполнение) |
| `/user-engagements/:id/completes` | POST | J12: завершение activity → condition_check passed → Billing Rule credit |
| `/user-engagements/:id/upgrade` | POST | J05: апгрейд тарифа → переход между offers одного benefit → пересчёт биллинга |
| `/user-engagements/:id/revoke` | POST | отключение benefit → Billing Rule credit (возврат) → деактивация у провайдера |

**8 endpoints.** Заменяет старые `/user-benefits/*` + `/activities/my` + `/activities/:id/start` + `/activities/:id/submit`.

---

## Wizard Config (S09, M07 T0706, ADR-019)

> **Backend → `internal/engagement/flow/`** (flow steps → wizard JSON config)
> **Frontend → `modals/Wizard.tsx` + `store/wizards.ts`** (generic renderer)

Generic wizard engine заменяет DmsWizard + MatCapitalWizard. Frontend загружает wizard config с сервера → `Wizard.tsx` рендерит шаги из JSON.

| Endpoint | Method | Обоснование |
|----------|--------|-------------|--|
| `/wizards` | GET | S09, T0706: список доступных wizard-конфигов для tenant'а — `[{id, name, category}]` |
| `/wizards/:id` | GET | S09, T0706: детали wizard-конфига — `steps[]`, `onComplete`, `validation` → fed into `Wizard.tsx` |
| `/wizards/:id/validate` | POST | T0706 server-side validation: check step data before submit → `{valid, errors: []}` |

**3 endpoints.** Могут быть кэшированы на frontend (wizard configs нечасто меняются).

### Response пример `/wizards/dms-upgrade`:

```json
{
  "id": "dms-upgrade",
  "name": "Подключение ДМС",
  "steps": [
    { "id": "option", "component": "OptionSelector", "validate": "required" },
    { "id": "payment-method", "component": "PaymentMethod", "validate": "card_or_payroll" },
    { "id": "confirmation", "component": "ReviewSummary", "validate": "all_fields" },
    { "id": "done", "component": "SuccessScreen" }
  ],
  "onComplete": { "type": "api", "method": "POST", "url": "/user-engagements" }
}
```

---

## CEL Generation (ADR-021 + ADR-022) — Admin API

> **Backend → `internal/cel/generator.go`** (ADR-021: CEL engine)
> **LLM → `internal/llm/`** (M10 T1002: был LLM Proxy :8085, теперь in-process. ADR-022 исторический)
> **Admin UI:** HR вводит условие на русском → backend вызывает `llm/` → возвращает CEL + валидацию

**Важно:** эти endpoints вызываются ТОЛЬКО при CRUD правил (создание/редактирование billing rules, eligibility conditions, recommendation segments, flow conditions). НЕ в hot path транзакций. CEL evaluation — локально через `cel/` package.

| Endpoint | Method | Обоснование |
|---------|-|------|
| `/cel/generate` | POST | **ADR-021/022:** генерация CEL expression из русского текста. Request: `{source text, agent, tenant_id, context_schema}`. Response: `{cel_expression, model, model_version, token_usage, latency_ms}` |
| `/cel/validate` | POST | **ADR-021:** валидация CEL expression. Request: `{cel_expression, context_schema}`. Response: `{valid, errors: []}` |
| `/cel/preview` | POST | **ADR-021:** preview evaluation — проверить CEL на тестовом контексте. Request: `{cel_expression, test_context}`. Response: `{result: bool, trace: [{step, value}]}` |

**3 endpoints.**

### Request пример `/cel/generate`:

```json
{
  "source_text": "только Senior и выше со стажем больше 2 лет и без детей",
  "agent": "cel-generator",
  "tenant_id": "sdek-uuid-123",
  "context_schema": {
    "available_fields": ["user.grade", "user.years_of_service", "user.has_children", "user.department"]
  }
}
```

### Response пример `/cel/generate`:

```json
{
  "cel_expression": "user.grade in ['Senior', 'Lead', 'Director', 'VP'] && user.years_of_service > 2 && !user.has_children",
  "model": "ollama-qwen3.6:27b",
  "model_version": "v1.0.0",
  "token_usage": { "prompt": 85, "completion": 32 },
  "latency_ms": 280,
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "validated": true
}
```

### Request пример `/cel/validate`:

```json
{
  "cel_expression": "user.grade in ['A', 'B'] && user.years_of_service >= 3",
  "context_schema": { ... }
}
```

### Response пример `/cel/validate` — OK:

```json
{
  "valid": true,
  "errors": []
}
```

### Response пример `/cel/validate` — ошибка:

```json
{
  "valid": false,
  "errors": ["ERROR:(1:42) undefined identifier: user.nonexistent_field"]
}
```

### Request пример `/cel/preview` — тест на контексте:

```json
{
  "cel_expression": "user.grade == 'Senior' && user.years_of_service > 2",
  "test_context": {
    "user": { "grade": "Senior", "years_of_service": 5 }
  }
}
```

### Response пример `/cel/preview`:

```json
{
  "result": true,
  "trace": [
    { "step": "user.grade == 'Senior'", "value": true },
    { "step": "user.years_of_service > 2", "value": true }
  ]
}
```

---

## Engagement Collections (S04, M04, M06, J13a)

> **Backend → `internal/engagement/collections/`** (M06: collections engine)

Наборы энгеэйджментов (только benefit-офферы).

| Endpoint | Method | Обоснование |
|---------|--------|-----|
| `/collections` | GET | S04: активные наборы — пагинация, `?status=active` |
| `/collections/:id` | GET | J13a(2): состав набора — offers (с cost), bundle_price, скидка, период |
| `/collections/:id/engage` | POST | J13a(3-5): одна кнопка → создаёт user_engagement для каждого offer в наборе → Billing Rule debit |

**3 endpoints.**

---

## Balance (S05, B01)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/balance` | GET | S05: total + by_category + expiration_date + days_until_expiration |
| `/balance/transactions` | GET | S05: история — `?type=credit\|debit`, пагинация |
| `/balance/period` | GET | S05/J13: текущий период — даты, дни до конца |

**3 endpoints.** `/balance/categories` убран — дублирует `/balance`.

---

## Billing (B01, J20a, J21-J22)

> **M12:** endpoints перенесены в `/api/v1/` (было `/billing/v1/`). Billing — `internal/billing/` пакет, не отдельный сервис. Операционные вызовы → direct Go call из `engagement/flow/`.

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/api/v1/billing/rules` | GET | список правил |
| `/api/v1/billing/rules` | POST | создание (HR: грейд→баллы, менеджер: debit for benefit) |
| `/api/v1/billing/rules/:id` | GET | детали |
| `/api/v1/billing/rules/:id` | PUT | редактирование (J20a) |
| `/api/v1/billing/rules/:id` | DELETE | удаление |
| `/api/v1/billing/periods/:id/activate` | POST | H02: открыть период распределения |
| `/api/v1/billing/periods/:id/expire` | POST | H02: закрыть период → остановка распределения |
| `/api/v1/billing/periods/current` | GET | S05/J13: текущий активный период |

**8 endpoints.** Все billing endpoints — `/api/v1/billing/`.

> **M12:** операционные вызовы (credit, debit, balance, transactions) — direct Go call `billing.BillingService.*()`, не NATS. Здесь описаны только управленческие endpoints (rules, periods), доступные HR и менеджеру через Nginx.

---

## Payments (S08, J08a)

> **M12:** endpoints перенесены в `/api/v1/payments/` (был отдельный сервис `/payments/v1/`). `internal/payments/` пакет.

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/api/v1/payments/:userBenefitId` | POST | S08: создание доплаты к льготе — `?method=card\|salary` |
| `/api/v1/payments/transactions/:id` | GET | статус — frozen/confirmed/cancelled |
| `/api/v1/payments/transactions/:id/confirm` | POST | подтверждение карты → Billing Rule credit |
| `/api/v1/payments/transactions/:id/cancel` | POST | отмена → разморозка |

**4 endpoints.** Все payment endpoints — `/api/v1/payments/`.

---

## Documents (S06, J10)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/documents` | GET | список — название, тип, дата, статус. `?type=policy\|application\|consent` |
| `/documents/:id` | GET | метаданные + предпросмотр |
| `/documents/:id/download` | GET | скачивание PDF |

**3 endpoints.**

---

## Support (S07, J09)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/support/faq` | GET | **Public** — вопросы/ответы (контент из M07) |
| `/support/tickets` | POST | создание обращения — тема, сообщение |
| `/support/tickets` | GET | мои обращения со статусами |
| `/support/tickets/:id` | GET | детали + ответ |

**4 endpoints.** FAQ — без auth.

---

## Recommendations (S03, M06, J13b)

> **Backend → `internal/recommendations/engine.go`** (M06: recommendations engine)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/recommendations` | GET | J13b: персональные — сегмент + контекст (правила из M06) |

**1 endpoint.**

---

## External (S03, E01, J11)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/external/services` | GET | S03: список доступных внешних сервисов (для отображения на Главной) |

**1 endpoint.** Сквозная авторизация — исключительно через Keycloak Identity Broker (OIDC/SAML redirect). Платформа не генерирует токены для внешних сервисов.

---

## Notifications (S03, T02, M06)

> **Backend → `internal/notification/store.go`** (M06: notification persistence)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/notifications` | GET | S03: колокольчик — непрочитанные, пагинация |
| `/notifications/:id/read` | POST | отметка прочитанное |
| `/notifications/read-all` | POST | все прочитанные |

**3 endpoints.**

---

## Relative Consent (R01, J14a)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/relative/consent/:token` | GET | R01: **Public** — токен из SMS → текст соглашения |
| `/relative/consent/:token/accept` | POST | **Public** — «Принять» → SMS-код |
| `/relative/consent/:token/verify` | POST | **Public** — ввод SMS-кода → подтверждение → БД |

**3 endpoints.** Без Keycloak — публичный флоу.

---

## Admin — Users (H01, J16, J17, T04)

> **M07 T0703:** backend → `internal/api/admin_user.go` (выделен из единственного admin_handler.go)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/admin/users` | GET | H01: таблица — ФИО,email,статус, `?status=`, пагинация |
| `/admin/users` | POST | H01: добавление одного пользователя (→ Keycloak) |
| `/admin/users/import` | POST | J16: загрузка JSON-реестра из HR-системы → валидация |
| `/admin/users/import/:jobId` | GET | статус импорта, error_report |
| `/admin/users/:id` | GET | H01: поиск по ФИО+ДР → детали |
| `/admin/users/:id` | PUT | изменение данных (→ Keycloak) |
| `/admin/users/:id/deactivate` | POST | J17/T04: каскад — блокировка+открепление |
| `/admin/users/export` | GET | H01: выгрузка Excel |

**8 endpoints.** CRUD через наш API, Keycloak синхронизируется.

---

## Admin — Periods (H02, J15)

> **M07 T0703:** backend → `internal/api/admin_user.go` (периоды сгруппированы с user/admin_handler)

| Endpoint | Method | Обоснование |
|---------|--------|-------------|
| `/admin/periods` | GET | H02: список периодов |
| `/admin/periods` | POST | J15: создание — даты открытия/закрытия, mass-notify |
| `/admin/periods/:id` | GET | details + метрики `registered%`, `distributed%` |
| `/admin/periods/:id/close` | POST | J15: закрытие → блокировка распределения |

**4 endpoints.**

---

## Admin — Engagements (H03, J18, J20a, M01–M05, J21, J22)

> **M07 T0703:** backend → `internal/api/admin_catalog.go` (выделен из admin_handler.go)

Управление энгейджментами: создание офферов (tariffs for benefit / instances for activity), статусы, метрики.

> **RBAC guard:** `hr` видит только `type=activity`. `catalog_manager` видит только `type=benefit`. `admin` видит оба типа.

| Endpoint | Method | Обоснование |
|---------|--------|-----|
| `/admin/engagements` | GET | H03, M01: список — `?type=benefit\|activity`, пагинация |
| `/admin/engagements` | POST | J18, J21: создание оффера (benefit: тариф; activity: бонус, период, аудитория) |
| `/admin/engagements/:id` | GET | H03, J22: детали + метрики — для activity: %участие, начисления, конверсия; для benefit: просмотры, активации, GMV |
| `/admin/engagements/:id` | PUT | J22, M05: редактирование оффера |
| `/admin/engagements/:id/status` | PATCH | M05: active → paused → archived |
| `/admin/engagements/:id/close` | POST | закрытие (activity) / архивация (benefit) |
| `/admin/engagements/analytics` | GET | H03, M02: агрегаты — участие, начисления, конверсия "получил→потратил" (activity); просмотры, GMV, покрытие (benefit) |

**7 endpoints.** Заменяет Admin Activities (6) + Admin Benefits (7) + Admin Benefit Plans (6) → один унифицированный модуль.

---

## Admin — Requests / Approval (H04, J20)

> **M07 T0703:** backend → `internal/api/admin_content.go` (requests + content сгруппированы)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/admin/requests` | GET | J20(1): заявки — маткапитал, психолог, апгрейд ЗП, `?status=waiting\|approved\|rejected` |
| `/admin/requests/:id` | GET | J20(2): детали — сотрудник, документы, тип |
| `/admin/requests/:id/approve` | POST | J20(3): одобрение → уведомление → активация льготы |
| `/admin/requests/:id/reject` | POST | J20(4): отклонение → уведомление + причина → возврат баллов |

**4 endpoints.** Единый модуль — не размазан по benefits/dms/matkapital.

---

## Admin — Analytics (H05, J19)

> **M07 T0703:** backend → `internal/api/admin_analytics.go` (выделен из admin_handler.go)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/admin/analytics/dashboard` | GET | J19: заявленные/зарегистрированные/распределившие, расход по категориям, конверсия, топ льгот |
| `/admin/analytics/reports/monthly` | GET | список ежемесячных отчётов (хранение N месяцев) |
| `/admin/analytics/reports/monthly/:month` | GET | Excel отчёт за месяц |
| `/admin/analytics/reports/by-user` | GET | J19(4): по ФИО+ДР → детали сотрудника |
| `/admin/analytics/reports/by-period` | GET | H05: по выбранному периоду распределения |

**5 endpoints.**

---

## Admin — Engagement Types (M01, M05, J21, J22, J26, H03, J18)

> **M07 T0703:** backend → `internal/api/admin_catalog.go`

Управление типами энгейджментов: шаблоны benefit + activity.

> **RBAC guard:** `hr` видит/создаёт только `type=activity`. `catalog_manager` видит/создаёт только `type=benefit`.

| Endpoint | Method | Обоснование |
|---------|--------|-----|
| `/admin/engagement-types` | GET | все типы, `?type=benefit\|activity`, пагинация |
| `/admin/engagement-types` | POST | J21, J18: создание типа (name, description, provider_adapter, category_id) |
| `/admin/engagement-types/:id` | GET | J22: детали типа + список офферов |
| `/admin/engagement-types/:id` | PUT | J22: редактирование типа (name, description, icon_url) |
| `/admin/engagement-types/:id` | DELETE | J26: удаление (только если 0 офферов и 0 подключений) |
| `/admin/engagement-types/:id/catalog-status` | PATCH | M05: смена статуса в каталоге (available → seasonal → archived) |
| `/admin/engagement-types/analytics` | GET | M02: просмотры, активации, GMV, конверсия, покрытие по категориям |

**7 endpoints.** Заменяет Admin Benefits + Admin Activity Types.

---

## Admin — Engagement Flows (M01, M05, J21, J18)

> **M07 T0703:** backend → `internal/api/admin_catalog.go`

Управление потоками выполнения: шаги, формы, аппрувы, редиректы, condition_check.

| Endpoint | Method | Обоснование |
|---------|--------|-----|
| `/admin/engagement-flows` | GET | все потоки, `?status=` |
| `/admin/engagement-flows` | POST | J21, J18: создание потока (name, steps[]) |
| `/admin/engagement-flows/:id` | GET | детали потока + шаги |
| `/admin/engagement-flows/:id` | PUT | редактирование потока |
| `/admin/engagement-flows/:id` | DELETE | удаление потока (только если 0 офферов не используют) |

**5 endpoints.** Заменяет Admin Activation Flows + completion_criteria config.

---

## Admin — Engagement Collections (M04, J27)

> **M07 T0703:** backend → `internal/api/admin_catalog.go`

Управление наборами энгейджментов (только benefit-офферы).

| Endpoint | Method | Обоснование |
|---------|--------|-----|
| `/admin/collections` | GET | M04: созданные наборы |
| `/admin/collections` | POST | J27: создание набора из offers (offers: [{offer_id, quantity}], bundle_price, период) |
| `/admin/collections/:id` | GET | детали набора |
| `/admin/collections/:id` | PUT | M04: редактирование состава/периода/баннера |
| `/admin/collections/:id` | DELETE | архивация |

**5 endpoints.**

---

## Admin — Recommendations (M06, J29)

> **M07 T0703:** backend → `internal/api/admin_recommendations.go` (выделен из admin_handler.go)

> **Backend → `internal/recommendations/rules.go`** (M06: rule CRUD engine)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/admin/recommendations/rules` | GET | J29: список правил контекст + сегмент |
| `/admin/recommendations/rules` | POST | создание правила |
| `/admin/recommendations/rules/:id` | PUT | редактирование |
| `/admin/recommendations/rules/:id` | DELETE | удаление |
| `/admin/recommendations/debug/:userId` | GET | J29(4): «какое правило сработало для сотрудника X» |

**5 endpoints.**

---

## Admin — Content (M07, J28)

> **M07 T0703:** backend → `internal/api/admin_content.go` (content + requests сгруппированы)

| Endpoint | Method | Обоснование |
|----------|--------|-------------|
| `/admin/content/faq` | GET | список вопросов |
| `/admin/content/faq` | POST | добавление |
| `/admin/content/faq/:id` | PUT | редактирование ответа |
| `/admin/content/faq/:id` | DELETE | удаление |
| `/admin/content/banners` | GET | список баннеров → отображение S03 |
| `/admin/content/banners` | POST | добавление |
| `/admin/content/banners/:id` | PUT | редактирование |
| `/admin/content/banners/:id` | DELETE | удаление |

**8 endpoints.**

---

## Сводная таблица

> **M06:** колонка `Backend → пакет` показывает mapping endpoints на внутренние пакеты Platform (7 пакетов после M06).
> **M07  T0701→T0702:** eligibility → собственный пакет, compliance выделен из consent (9 пакетов).
> **M07 T0703:** admin_handler.go → 5 файлов (admin_user, admin_catalog, admin_recommendations, admin_analytics, admin_content).
> **M12:** lkfl-server modular monolith. Billing → `internal/billing/`, Payments → `internal/payments/`, Integrations → `internal/integrations/`. NATS → direct Go calls.
> Детали: [`пакеты-platform.md`](../архитектура/пакеты-platform.md).

| Модуль | Endpoints | Backend → пакет | Roles | Auth |
|------|---:|--|--|---|
| Auth + Keycloak | 1 | `api/` → `auth/` | all | keycloak (register only) |
| User / Consent | 5 | `api/` → `user/` + `consent/` | employee | keycloak |
| Dashboard | 1 | `api/` → агрегация | employee | keycloak |
| Engagements | 4 | `api/` → `engagement/catalog` | employee, hr, catalog_manager | keycloak (+ RBAC type guard) |
| Engagement Offers | 2 | `api/` → `engagement/catalog/` + `eligibility/` | employee | keycloak |
| User Engagements | 8 | `api/` → `engagement/flow` | employee | keycloak |
| Engagement Collections | 3 | `api/` → `engagement/collections` | employee | keycloak |
| Balance | 3 | `api/` → billing.BillingService (M12: был NATS) | employee | keycloak |
| Billing | 8 | **M12:** internal/billing (был Billing сервис) | hr, catalog_manager | keycloak |
| Payments | 4 | **M12:** internal/payments (был NATS integrations) | employee | keycloak |
| Documents | 3 | `api/` → Asynq document-generate | employee | keycloak |
| Support | 4 | `api/` → db | employee, **public(faq)** | keycloak |
| Recommendations | 1 | `api/` → `recommendations/engine` | employee | keycloak |
| External | 1 | **M16:** `api/` → `integrationclient/` → gRPC → proxy (был `integrations.ProviderGateway`, M12) | employee | keycloak |
| Notifications | 3 | `api/` → `notification/store` | employee | keycloak |
| Relative Consent | 3 | `api/` → db | **none** — public | token only |
| Wizard Config | 3 | `api/` → `engagement/flow` | employee | keycloak (M07 T0706) |
| Admin: Users | 8 | `api/` → `admin_user.go` (T0703) | hr, admin | keycloak |
| Admin: Periods | 4 | `api/` → `admin_user.go` (T0703) | hr | keycloak |
| Admin: Engagements | 7 | `api/` → `admin_catalog.go` (T0703) | hr(activity), catalog_manager(benefit) | keycloak (+ RBAC type guard) |
| Admin: Requests | 4 | `api/` → `admin_content.go` (T0703) | hr | keycloak |
| Admin: Analytics | 5 | `api/` → `admin_analytics.go` (T0703) | hr | keycloak |
| Admin: Engagement Types | 7 | `api/` → `admin_catalog.go` (T0703) | hr(activity), catalog_manager(benefit) | keycloak (+ RBAC type guard) |
| Admin: Engagement Flows | 5 | `api/` → `admin_catalog.go` (T0703) | hr(activity), catalog_manager(benefit) | keycloak (+ RBAC type guard) |
| Admin: Collections | 5 | `api/` → `admin_catalog.go` (T0703) | catalog_manager | keycloak |
| Admin: Recommendations | 5 | `api/` → `admin_recommendations.go` (T0703) | catalog_manager | keycloak |
| Admin: Content | 8 | `api/` → `admin_content.go` (T0703) | catalog_manager | keycloak |
| **Итого lkfl-server** | **134** | | | |

---

## Integrations Hub Admin API (M07 T0708, M12, M16)

> **M12:** Backend → `internal/integrations/` (был отдельный сервис). Platform обращалась через direct Go call `integrations.ProviderGateway.*()`, не NATS.
> **M16:** Монолит делегирует запросы к `lkfl-integration-proxy` через gRPC. HTTP endpoint'ы — facade pattern. Монолит → gRPC → proxy → ответ → HTTP response. Backend: `integrationclient/` → gRPC → proxy (было: `integrations.ProviderGateway.*()`).
> **HTTP API:** управленческий доступ для Integration Admin через Nginx (`/admin/integrations/` → lkfl-server:8080 → gRPC → lkfl-integration-proxy:8090).
> **RBAC guard:** `integration_admin` только. Остальные роли получают 403 Forbidden.
> **Архитектура:** [ADR-035](../архитектура/adr/035-integration-proxy.md) — вынос внешних интеграций в отдельный бинарник для fault isolation, credential isolation, goroutine safety.

### Providers CRUD

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/admin/providers` | GET | `integration_admin` | Список всех провайдеров + системных интеграций |
| `/admin/providers` | POST | `integration_admin` | Создание новой интеграции |
| `/admin/providers/:id` | GET | `integration_admin` | Детали провайдера (config, endpoints, status) |
| `/admin/providers/:id` | PUT | `integration_admin` | Редактирование config |
| `/admin/providers/:id` | DELETE | `integration_admin` | Деактивация (soft delete) |
| `/admin/providers/:id/health` | POST | `integration_admin` | Ручной health check → trigger |

### Health & Monitoring

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/admin/health/dashboard` | GET | `integration_admin` | Агрегированные health metrics всех интеграций |

### Sync Control

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/admin/sync/trigger` | POST | `integration_admin` | Ручной trigger синхронизации (provider/hr/catalog) |
| `/admin/sync/schedule` | GET | `integration_admin` | Текущие расписания синхронизации |
| `/admin/sync/schedule` | PUT | `integration_admin` | Изменение расписания |
| `/admin/sync/errors` | GET | `integration_admin` | Error log (последние N ошибок) |
| `/admin/sync/sla` | GET | `integration_admin` | SLA dashboard (error_rate, latency, uptime) |

**Итого: 12 endpoints.**

### Response Schema — Provider

```json
{
  "id": "prov-alpha-001",
  "name": "Alpha Insurance",
  "category": "dms",
  "protocol": "rest",
  "endpoints": {
    "activate": "https://alpha.api/v1/activate",
    "deactivate": "https://alpha.api/v1/deactivate",
    "status": "https://alpha.api/v1/status",
    "sync_catalog": "https://alpha.api/v1/catalog"
  },
  "auth_method": "bearer_token",
  "status": "active",
  "health": {
    "last_check": "2026-05-24T10:00:00Z",
    "status": "up",
    "error_rate_30d": 0.02,
    "latency_p95_ms": 125
  },
  "created_at": "2025-01-15T08:00:00Z",
  "updated_at": "2026-04-20T09:30:00Z",
  "circuit_breaker_state": "closed",
  "worker_pool_status": {
    "active": 2,
    "queued": 5,
    "max_concurrent": 5
  }
}
```

> **M16:** Поля `circuit_breaker_state` и `worker_pool_status` добавлены через proxy delegation.
> - `circuit_breaker_state`: enum — `closed` (нормальная работа), `open` (все вызовы отклоняются, cached response), `half_open` (пробный запрос).
> - `worker_pool_status`: статус worker pool провайдера в proxy. `active` — текущие активные задачи, `queued` — задачи в очереди, `max_concurrent` — лимит одновременных задач (default: 5).

### Response Schema — Health Dashboard

```json
{
  "total_providers": 14,
  "up": 12,
  "degraded": 1,
  "down": 1,
  "providers": [
    {
      "id": "prov-ready4-001",
      "name": "Ready4",
      "status": "degraded",
      "error_rate_30d": 0.15,
      "latency_p95_ms": 450,
      "last_sync": "2026-05-24T09:00:00Z",
      "last_error": "timeout after 30s"
    }
  ],
  "sla": {
    "target_uptime": 99.5,
    "actual_uptime_30d": 98.7,
    "target_latency_p95_ms": 200,
    "actual_latency_p95_ms": 178
  }
}
```

### Response Schema — Sync Trigger

```json
// POST /admin/sync/trigger
{
  "triggered": [
    {
      "provider_id": "prov-alpha-001",
      "type": "catalog_sync",
      "status": "queued",
      "estimated_completion": "2026-05-24T10:05:00Z"
    }
  ],
  "failed": []
}
```

### Error Format (единый со всеми сервисами)

```json
{
  "error": {
    "code": "PROVIDER_NOT_FOUND",
    "message": "Интеграция с ID prov-unknown не найдена",
    "details": {
      "id": "prov-unknown"
    }
  }
}
```

#### gRPC Error Codes (M16 — proxy delegation)

Монолит получает gRPC error от proxy и мапит на HTTP status:

| gRPC Code | HTTP Status | Описание | Поведение монолита |
|-----------|------------|----------|-------------------|
| `UNAVAILABLE` | 503 | Proxy недоступен (down, network error) | Возвращает 503 с message "сервис интеграций временно недоступен". Каталог читается из локального кэша в PG |
| `DEADLINE_EXCEEDED` | 504 | Timeout gRPC вызова (default: 5s) | Возвращает 504 с message "превышено время ожидания ответа от сервиса интеграций" |
| `INTERNAL` | 500 | Ошибка внутри proxy (panic, unexpected) | Возвращает 500 с message "внутренняя ошибка сервиса интеграций" |
| `NOT_FOUND` | 404 | Провайдер не найден в proxy | Возвращает 404 с message "интеграция не найдена" |

```json
// Пример gRPC error → HTTP response
{
  "error": {
    "code": "PROXY_UNAVAILABLE",
    "message": "Сервис интеграций временно недоступен",
    "grpc_code": "UNAVAILABLE",
    "http_status": 503,
    "details": {
      "retry_after_seconds": 30,
      "fallback": "catalog_read_from_cache"
    }
  }
}
```

### Pagination (cursor-based)

```
GET /admin/providers?cursor=prov-skillbox-003&limit=20
```

Response:
```json
{
  "data": [...],
  "pagination": {
    "next_cursor": "prov-mts-005",
    "has_more": true,
    "limit": 20
  }
}
```

---

## Сводная таблица (обновлено M07)

> **M07:** 134 endpoints Platform + Billing, + 12 endpoints Integrations Hub (T0708). Итого: 146 endpoints.
> **M12:** объединение сервисов → единый `/api/v1/`. **M16:** вынос proxy → admin endpoints перенесены. Итого: **118 endpoints**.

| Модуль | Endpoints | Backend → пакет | Roles | Auth |
|--------|---:|--|--|--|
| Auth + Keycloak | 1 | `api/` → `auth/` | all | keycloak (register only) |
| User / Consent | 5 | `api/` → `user/` + `consent/` | employee | keycloak |

> **Примечание:**
> - `Activities` (5) убраны → объединены в `User Engagements` (8).
> - `Catalog / Benefits` (4) + `Benefit Plans` (2) → объединены в `Engagements` (4) + `Engagement Offers` (2).
> - `Admin Activities` (6) + `Admin Benefits` (7) + `Admin Benefit Plans` (6) → объединены в `Admin Engagements` (7).
> - DMS (11) и Matkapital (3) убраны — теперь generic EngagementOffer + EngagementFlow + UserEngagement.
> - **M07 T0706:** Wizard Config (3 endpoints) — generic wizard engine заменяет DmsWizard + MatCapitalWizard (ADR-019).
> - **M12:** lkfl-server modular monolith. Billing, Payments, Integrations → internal пакеты. `/billing/v1/` → `/api/v1/billing/`. NATS → Go interfaces.
> - **M16:** Integrations Hub → proxy delegation. `internal/integrations/` → `internal/integrationclient/` → gRPC → `lkfl-integration-proxy` (ADR-035). 12 admin endpoints — facade pattern.
> - Итого: 134 endpoints. Backend по пакетам (M06+M07+M12+M16): auth, user, consent, eligibility, compliance, engagement, notification, recommendations, gamification, billing, integrationclient (→ proxy), payments, api.

---

## Backend → пакет mapping (M06+M07+M12)

Каждый раздел API выше делегирует в соответствующий internal-пакет. Кратно:

| API-раздел | Backend пакет | Ключевые методы |
|------|--|--|
| Engagements (catalog) | `internal/engagement/catalog/` | `Catalog.List()`, `Catalog.Get()` |
| Engagement Offers (eligibility) | `internal/eligibility/` (M07 T0701) | `Eligibility.Check()`, `Eligibility.EvaluateCEL()` |
| User Engagements (flow) | `internal/engagement/flow/` | `Flow.Activate()`, `Flow.Complete()`, `Flow.Revert()` |
| Engagement Collections | `internal/engagement/collections/` | `Collections.Engage()` |
| Balance | **M12:** `internal/billing/` (был NATS) | `BillingService.GetBalance()`, `BillingService.GetTransactions()` |
| Billing (admin) | **M12:** `internal/billing/` (был отдельный сервис) | `RuleEngine.EvaluateRule()`, `PeriodEngine.Activate()` |
| Payments | **M12:** `internal/payments/` (был отдельный сервис) | `PaymentGateway.Authorize()`, `PaymentGateway.Capture()`, `PaymentGateway.Void()` |
| External providers | **M16:** `internal/integrationclient/` → gRPC → `lkfl-integration-proxy` (был `internal/integrations/`, M12) | `IntegrationClient.Activate()`, `IntegrationClient.SyncCatalog()` → gRPC → proxy |
| Notifications | `internal/notification/` (store) | `Store.ListUnread()`, `Store.MarkRead()` |
| Recommendations | `internal/recommendations/` (engine) | `Engine.Recommend()`, `Engine.Debug()` |
| Admin: Recommendations | `internal/recommendations/` (rules CRUD) | `Rules.Create()`, `Rules.Update()`, `Rules.Delete()` |
| Admin: Engagements, Types, Flows | `internal/engagement/catalog/` + `internal/engagement/flow/` | CRUD через CatalogService + FlowEngine |
| Всё остальное | `api/` → db/ напрямую | thin handlers без business-логики |

---

## История изменений

### v2 → v3 (131 endpoints, M05 — Унификация Engagement)

| Изменение | Описание |
|----------|--|
| **Объединено** | Catalog/Benefits + Activities → Engagements; Benefit Plans → Engagement Offers; User Benefits + Activities/my → User Engagements |
| **Добавлено** | `/user-engagements/:id/completes` (activity completion), RBAC type guard на 4 admin-модуля, `/engagements/types` |
| **Удалено** | `Activities` (5 endpoints) → встроено в `User Engagements`; `DMS` (11), `Matkapital` (3) → generic EngagementOffer + EngagementFlow |
| **Переименовано** | `/collections/:id/activate` → `/collections/:id/engage`, `/admin/activation-flows` → `/admin/engagement-flows`, `/admin/benefits` → `/admin/engagement-types` |

### v1 → v2 (116 → 128, M03+M04 — Пересборка льгот)

| Что добавлено | Кол-во | Почему |
|---------------|------:|--------|
| Benefit Plans (сотрудник) | 2 | T0301 — тарифы: `/benefit-plans/:id` + check-eligibility |
| User Benefits | 7 | T0301 — экземпляр сотрудника: activate, deactivate, upgrade, steps |
| Admin: Benefit Plans | 6 | T0301 — CRUD тарифов отдельно от продуктов |
| Admin: Activation Flows | 5 | T0301 — CRUD потоков активации |
| `/activities/:id/start` | 1 | Т0301 — явный start для completion |
| **Всего добавлено** | **+21** | |

| Что убрано | Кол-во | Почему |
|-----------|------:|--------|
| `/benefits/:id/activate` | 1 | → `/user-engagements` POST (M05: engagement унификация) |
| `/benefits/:id/deactivate` | 1 | → `/user-engagements/:id/deactivate` (M05) |
| `/benefits/my` | 1 | → `/user-engagements` GET (M05) |
| DMS (11 endpoints) | 11 | → Unified EngagementType + EngagementOffer (M05) |
| Matkapital (3 endpoints) | 3 | → Unified EngagementType + EngagementOffer (M05) |
| `/activities/:id/survey/submit` | 1 | → `/activities/:id/submit` |
| `/activities/:id/attend` | 1 | → `/activities/:id/submit` |
| **Всего убрано** | **-19** | |

**Нетто v1→v2:** +12 endpoints.

### v4 → v5 (134 endpoints, M16 — Integration Proxy)

| Изменение | Описание |
|-----|--|
| **Изменено** | Backend package mapping: `internal/integrations/` → `internal/integrationclient/` → gRPC → `lkfl-integration-proxy` |
| **Изменено** | External providers: `ProviderGateway.Activate()` → `IntegrationClient.Activate()` → gRPC → proxy |
| **Добавлено** | Response schema: `circuit_breaker_state` (enum: closed, open, half_open), `worker_pool_status` (object) |
| **Добавлено** | Error responses: gRPC codes (UNAVAILABLE → 503, DEADLINE_EXCEEDED → 504, INTERNAL → 500, NOT_FOUND → 404) |
| **Добавлено** | Note: монолит делегирует запросы к proxy через gRPC (facade pattern) |

### v3 → v4 (134 endpoints, M12 — Модульный монолит)

| Изменение | Описание |
|-----|--|
| **Объединено** | `/billing/v1/*` → `/api/v1/billing/*`; `/payments/*` → `/api/v1/payments/*`; 4 health endpoint'а → один `/api/healthz` |
| **Изменено** | Backend package mapping: NATS billing → `internal/billing/`; separate payment-gateway → `internal/payments/`; NATS integrations → `internal/integrations/` |
| **Удалено** | `SERVICE_UNAVAILABLE` для internal deps (billing, integrations, payments); integrations:8082 upstream; billing:8081 upstream; payment-gateway:8084 upstream |
