# Архитектура

Этот раздел описывает **модульную структуру**, **технологический стек**, **интеграции**, **ADR** и **безопасность** платформы гибких льгот.

> **🗺️ Навигация:** [`doc/NAVIGATION.md`](../NAVIGATION.md) — карта «вопрос → файл:строка»

## Назначение

Архитектура переводит контекст в техническое проектирование. Она отвечает на вопросы:
- **Из каких сервисов состоит система?** — `lkfl-server` (монолит, 16 internal пакетов) + `lkfl-integration-proxy` (отдельный бинарник, gRPC), React SPA
- **Какой стек используем?** — Go 1.22, Keycloak OIDC, PostgreSQL, Redis, React + Mantine + @april/ui
- **Как интегрируемся с внешними системами?** — `lkfl-integration-proxy` (отдельный бинарник, gRPC :8090) — точка контакта с benefit-провайдерами. Монолит → proxy через `internal/integrationclient/` (gRPC client). HR → `internal/user/` напрямую (Asynq worker). 1C → `internal/billing/payroll/` напрямую (REST API). Платежи → `internal/payments/` (PCI DSS). [ADR-035]
- **Какие архитектурные решения приняты?** — 35 ADR в формате ХАДД (26 Accepted, 4 Superseded, 5 Note)
- **Как обеспечиваем безопасность?** — OWASP, 152-ФЗ, ФСТЭК, multi-tenancy, consent lifecycle, PCI DSS (Payment Gateway)

## Архитектура системы

```
                             ┌────────────────────────────────┐
                              │  Keycloak (Identity Provider)   │
                              │  OIDC · SSO · SAML broker       │
                              │  Realm per tenant · RBAC         │
                              │  Password policy · MFA · Audit   │
                              └──────┬────────────────┬──────────┘
                                     │ JWT validation │ Identity Broker
                            ┌────────┘                 └───────────────┐
                            ▼                                          ▼
                     ┌────────────┐                           ┌──────────────┐
                     │    Nginx   │                           │  External     │
                     │  (API GW)  │                           │  Providers    │
                     └──┬───┬───┬─┘                           │  (Ready4,    │
                        │   │   │                              │   PrimeZone, │
                        │   │   ├── /webhooks/* → :8091       │   Sber, etc.)│
                        │   │   │                                └──────┬─────┘
                        │   │   └── /api/* → :8080                    │
              ┌────────┘   │                                          │
              ▼           │                                          │
       ┌───────────┐      │                                          │
       │ Frontend  │      │                                          │
       │ (React SPA)│     │                                          │
       └───────────┘      │                                          │
                          │                                          │
       ┌──────────────────▼───────────────┐           ┌──────────────▼───────────────┐
       │   lkfl-server (:8080)            │           │   lkfl-integration-proxy      │
       │   бизнес-логика                  │           │   (:8090 gRPC + :8091 HTTP)   │
       │                                  │◄──gRPC──►│   (M16, ADR-035)              │
       │ ┌─────────────────────┐          │           │                               │
       │ │  Public Router      │          │           │  ┌─────────────────────────┐  │
       │ │  /api/v1/...        │          │           │  │  Provider adapters      │  │
       │ │  /api/v1/billing/   │          │           │  │  Circuit breaker        │  │
       │ │  /api/v1/payments/  │          │           │  │  Webhook receiver       │  │
       │ └─────────────────────┘          │           │  │  Worker pool            │  │
       │ ┌─────────────────────┐          │           │  └─────────────────────────┘  │
       │ │  Admin Router       │          │           │                               │
       │ │  /admin/...         │          │           └───────────────────────────────┘
       │ └─────────────────────┘          │
       │                                  │
       │ ┌─────────────────────┐          │
       │ │  auth/ user/        │          │
       │ │  consent/ cel/      │          │
       │ │  llm/ eligibility/  │          │
       │ │  compliance/        │          │
       │ │  engagement/        │          │
       │ │  ├─ catalog/ flow/  │          │
       │ │  ├─ collections/    │          │
       │ │  └─ survey/         │          │
       │ │  notification/      │          │
       │ │  gamification/      │          │
       │ │  billing/ ← M12     │          │
       │ │  integrationclient/ │          │  ← M16: gRPC client к proxy
       │ │  payments/ ← M12    │          │
       │ │  content/           │          │
       │ └─────────────────────┘          │
       └──┬──────────┬───────────┘
          │          │
      PostgreSQL   Redis
     (lkfl_platform  (key prefixes:
      +              jwt:, asynq:,
      lkfl_integration)  catalog:, cel:, rate:)
```

> **M12:** Один бинарник `lkfl-server` вместо 4 Go-сервисов. billing/, integrations/, payments/ — internal пакеты с Go-интерфейсами.
> **M16:** `integrations/` вынесен в отдельный бинарник `lkfl-integration-proxy` (ADR-035). В монолите `integrationclient/` — gRPC client к proxy.
> HR → `internal/user/HRSync` напрямую (Asynq worker). 1C → `internal/billing/payroll/` напрямую (REST API).

## Содержимое раздела

| Файл | Описание |
|--|---|
| `модули.md` | **M12:** `lkfl-server` — единый бинарник Go API, 17 internal пакетов (15 business + tenant + api), React SPA (кратко). DI через Go interfaces. **M16:** 16 пакетов монолита + proxy. |
| `фронтенд.md` | **M15:** Архитектура фронтенда — 10 разделов (A→J): обзор, routing, структура, API layer, state, компоненты, white-label, performance, mobile (кратко), формы (кратко). |
| `фронтенд-mobile-forms.md` | **M15:** Mobile + Forms детально — AprilMobileShellBar, модальности, breakpoints, touch-ориентиры, жесты; Zod + react-hook-form, wizard, survey. |
| `пакеты-platform.md` | **M12:** 17 внутренних пакетов (исторически). **M16:** 16 пакетов монолита: tenant/, auth, user, consent, cel, llm, eligibility, compliance, engagement, notification, gamification, **billing/**, **integrationclient/**, **payments/**, **content/**, **recommendations/** (stub), api — public API, DI graph, worker mapping. `integrations/` → `integrationclient/` + `integration-proxy/` (отдельный бинарник) |
| `стек.md` | Технологии, фреймворки, библиотеки с обоснованием и версиями. **ADR-021:** cel-go + **M10 T1002:** LLM in-platform + **M12:** NATS удалён |
| `интеграции.md` | 10 категорий (N провайдеров, **T1101:** только benefit-providers): абстрактные контракты, **M07 T0704:** Generic REST adapter + YAML config, ProviderAdapter, error handling. **M16:** Integration Proxy (gRPC contract, circuit breaker, webhook) — ADR-035 |
| `cel-engine.md` | **Детально:** CEL Rule Engine (ADR-021). Заменяет 4 механизма условий. LLM генерация через `internal/llm/` (M10 T1002). 3 фазы миграции. |
| `llm-proxy.md` | **M10 T1002:** Исторический ADR. LLM Proxy слит в Platform `internal/llm/`. Agent router, prompt mgmt, cost tracking, audit trail — теперь in-process. |
| `пакеты-platform.md § billing/` | **Детально:** Billing Rule Engine, формулы, условия, триггеры, периоды, модели начислений. **ADR-021:** условия → CEL (`condition_cel`). **M11 T1103:** +Payroll (1C) section. |
| `engagement.md` | **Детально:** Единая абстракция Engagement (льготы + активности), 4 подпакета (catalog/, flow/, collections/, survey/), YAML-схемы, примеры. **ADR-021:** `condition_expr` → CEL. |
| `теги.md` | TagResolver — вычисление тегов пользователя. Survey-теги (M13 T1301). |
| `активности.md` | ⚠️ Удалено M05 — содержимое объединено в [`engagement.md`](./engagement.md) |
| `льготы.md` | ⚠️ Удалено M05 — содержимое объединено в [`engagement.md`](./engagement.md) |
| `schema.md` | **P0:** Полная модель PostgreSQL 17 — 47 таблиц (41 `lkfl_platform` + 6 `lkfl_integration`), ER-диаграмма, индексы, constraints, Go mapping, Redis layout, миграционная стратегия (Atlas). **M16:** +lkfl_integration schema. |
| `adr/` | **35 архитектурных решений** в формате ADR (ХАДД). M12: ADR-024 (modular monolith). M15: ADR-031 (data fetching), ADR-032 (API types), ADR-033 (testing), ADR-034 (i18n). M16: ADR-035 (Integration Proxy). M13: ADR-025 (survey). ADR-023 (gamification), ADR-029 (DS gap), ADR-030 (CI/CD). **4 ❌ Superseded, 5 ⚠️ Note.** |
| `безопасность.md` | OWASP Top 10, 152-ФЗ, ФСТЭК, consent, audit trail, rate limiting, PCI DSS (`internal/payments/`) |
| `инфраструктура.md` | Dev стенд: архитектура деплоя, 11 багов (Б-001→Б-008), чек-лист деплоя, troubleshooting |

## ADR

| ADR | Тема | Статус |
|-----|------|--------|
| [ADR-001](adr/001-go-react.md) | Выбор Go + React для реализации | ✅ Accepted |
| [ADR-002](adr/002-postgresql.md) | PostgreSQL как основная БД | ✅ Accepted |
| [ADR-003](adr/003-keycloak.md) | Keycloak (OIDC) как центральный IdP | ✅ Accepted |
| [ADR-004](adr/004-redis.md) | Redis для сессий и кэша | ✅ Accepted |
| [ADR-005](adr/005-nats.md) | NATS JetStream как message broker | ⚠️ Superseded by ADR-024 (M12) |
| [ADR-006](adr/006-billing.md) | Биллинг как отдельный сервис | ⚠️ Note: M12 — merged |
| [ADR-007](adr/007-april-ui.md) | April UI + Mantine как frontend foundation | ✅ Accepted |
| [ADR-008](adr/008-white-label.md) | White-label через CSS переменные | ✅ Accepted |
| [ADR-009](adr/009-multi-tenancy.md) | Tenant-aware архитектура (multi-tenancy) | ✅ Accepted |
| [ADR-010](adr/010-nginx.md) | Nginx как API gateway и reverse proxy | ✅ Accepted |
| [ADR-011](adr/011-monorepo.md) | Monorepo с multi-module Go | ⚠️ Note: M12 — single go.mod |
| [ADR-012](adr/012-zustand.md) | Zustand для state management | ✅ Accepted |
| [ADR-013](adr/013-пакеты-platform.md) | Platform → internal пакеты (не gRPC микросервисы) | ⚠️ Note: M12 — 15 business + tenant + api |
| [ADR-014](adr/014-eligibility-extraction.md) | Eligibility → отдельный пакет (T0701) | ✅ Accepted (M07) |
| [ADR-015](adr/015-compliance-package.md) | Compliance (cascade + audit + retention) → отдельный пакет (T0702) | ✅ Accepted (M07) |
| [ADR-016](adr/016-admin-handler-split.md) | admin_handler.go → 5 файлов по доменам (T0703) | ✅ Accepted (M07) |
| [ADR-017](adr/017-generic-rest-adapter.md) | 11 hard-coded adapters → Generic REST + YAML config (T0704) | ✅ Accepted (M07) |
| [ADR-018](adr/018-payment-gateway-service.md) | Payment Gateway → 4-й Go-сервис (PCI DSS, T0705) | ⚠️ Note: M12 — merged |
| [ADR-019](adr/019-wizard-engine.md) | JSON-driven Wizard (DmsWizard + MatCapitalWizard → Wizard, T0706) | ✅ Accepted (M07) |
| [ADR-020](adr/020-nats-subjects-registry.md) | NATS master registry + billing namespace (T0707) | ❌ Superseded (M12) |
| [ADR-021](adr/021-cel-unified-rule-engine.md) | CEL — единый движок бизнес-логики (заменяет YAML/AND/OR/JSON conditions) | ✅ Accepted |
| [ADR-022](adr/022-llm-proxy-service.md) | LLM Proxy — 5-й микросервис (центральный шлюз ко всем LLM-моделям) | ⚠️ Note: M10 — merged |
| [ADR-024](adr/024-modular-monolith.md) | Modular Monolith — переход от микросервисов к одному бинарнику | ✅ Accepted (M12) |
| [ADR-026](adr/026-hr-sync-platform.md) | HR Sync → Platform user/ (перенос из Integrations, M11 T1102) | ✅ Accepted (M11) |
| [ADR-027](adr/027-payroll-in-billing.md) | 1C Payroll → Billing payroll/ (перенос из Integrations, M11 T1103) | ✅ Accepted (M11) |
| [ADR-028](adr/028-nats-rename.md) | NATS namespace rename `integration.*` → `provider.*` (M11 T1104) | ❌ Superseded by ADR-024 (M12 — NATS удалён) |
| [ADR-023](adr/023-gamification-system.md) | Gamification — 5-й CEL домен (ачивки, loyalty) | ✅ Accepted (M09) |
| [ADR-025](adr/025-survey-engine.md) | Survey Engine — бранчинг, TagMapper, analytics | ✅ Accepted (M13) |
| [ADR-029](adr/029-ds-components-gap-tz.md) | DS components gap — April UI vs Mantine | ✅ Accepted |
| [ADR-030](adr/030-ci-cd-pipeline.md) | CI/CD Pipeline — GitHub Actions + Atlas + Docker | ✅ Accepted |
| [ADR-031](adr/031-api-data-fetching.md) | API Data Fetching — React Query (@tanstack/react-query) | ✅ Accepted (M15) |
| [ADR-032](adr/032-api-types-codegen.md) | API Types — openapi-typescript (codegen из OpenAPI spec) | ✅ Accepted (M15) |
| [ADR-033](adr/033-frontend-testing.md) | Frontend Testing — Vitest + RTL + Playwright | ✅ Accepted (M15) |
| [ADR-034](adr/034-i18n-yagni.md) | i18n — YAGNI (lib/translations/ru.ts, не i18next) | ✅ Accepted (M15) |
| [ADR-035](adr/035-integration-proxy.md) | Integration Proxy — вынос внешних интеграций из монолита (gRPC, circuit breaker, webhook) | ✅ Accepted (M16) |

> **Примечание:** ADR-023 (Gamification) — M09. ADR-025 (Survey Engine) — M13. ADR-029 (DS Gap) — документирован в `adr/029-ds-components-gap-tz.md`.

## Ключевые решения

- **M12: Модульный монолит** — `lkfl-server`, один `go.mod`, 17 internal пакетов с Go-интерфейсами (исторически). [ADR-024](adr/024-modular-monolith.md). **M16:** 16 пакетов монолита + proxy. Future split через interface → gRPC/NATS.
- **Keycloak** вместо голых JWT — SSO, Identity Broker для внешних сервисов, ФСТЭК-комплаенс
- **NATS удалён** — не используется в mono-режиме. Межмодульная коммуникация — Go interfaces. Future split через interface → gRPC/NATS. ADR-005 и ADR-020 — Superseded.
- **`lkfl-integration-proxy`** — отдельный бинарник, единственная точка контакта с benefit-провайдерами (ADR-035). Монолит → proxy через `internal/integrationclient/` (gRPC). Fault isolation, credential isolation, goroutine safety. HR → `internal/user/` (Asynq worker), 1C → `internal/billing/payroll/` (REST) — исключения.
- **Локальный кэш** — платформа хранит копию каталога и статусов; обновления приходят через direct call
- **Multi-tenancy** — `tenant_id UUID` в каждой бизнес-таблице, brand CSS per tenant
- **Биллинг как internal пакет** — финансовые операции с ACID-гарантиями через общую PostgreSQL transaction. **M11 T1103:** +`payroll/` (1C бухгалтерия).
- **`internal/payments/`** — PCI DSS isolation на уровне кода: credentials separate, thin wrappers, token-only
- **HashiCorp Vault** — централизованное хранение ключей шифрования и credentials
- **CEL (Common Expression Language)** — единый движок бизнес-логики (ADR-021). Все условия: billing rules, eligibility, flow condition_check, recommendations segments — CEL expressions. LLM генерирует CEL из русского текста → cel-go валидирует → sandbox evaluation.
