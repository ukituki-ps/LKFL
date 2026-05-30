# ADR — Архитектурные решения

<<<<<<< HEAD
> **36 файлов.** 27 Accepted, 4 Superseded, 5 Note.
=======
> **38 файлов.** 29 Accepted, 4 Superseded, 5 Note.
>>>>>>> origin/main

---

## Сводная таблица

| ADR | Название | Статус | Строк |
|-----|----------|--------|-------|
| [001](./001-go-react.md) | Выбор Go + React | ✅ | 50 |
| [002](./002-postgresql.md) | PostgreSQL как основная БД | ✅ | 28 |
| [003](./003-keycloak.md) | Keycloak (OIDC) — центральный IdP | ✅ | 40 |
| [004](./004-redis.md) | Redis для сессий и кэша | ✅ | 38 |
| [005](./005-nats.md) | NATS JetStream (был message broker) | ⚠️ Superseded | 60 |
| [006](./006-billing.md) | Биллинг как модуль (был отдельный сервис) | ⚠️ Note: M12 — merged | 59 |
| [007](./007-april-ui.md) | April UI + Mantine — frontend foundation | ✅ | 37 |
| [008](./008-white-label.md) | White-label через CSS переменные | ✅ | 47 |
| [009](./009-multi-tenancy.md) | Tenant-aware архитектура | ✅ | 47 |
| [010](./010-nginx.md) | Nginx — API Gateway + Reverse Proxy | ✅ | 52 |
| [011](./011-monorepo.md) | Monorepo (M12: один go.mod) | ⚠️ Note: M12 — single go.mod | 55 |
| [012](./012-zustand.md) | Zustand для state management | ✅ | 53 |
| [013](./013-пакеты-platform.md) | Internal пакеты вместо gRPC микросервисов | ⚠️ Note: M12 — 15 business + tenant + api | 104 |
| [014](./014-eligibility-extraction.md) | EligibilityEngine → собственный пакет | ✅ | 44 |
| [015](./015-compliance-package.md) | Compliance (cascade + audit + retention) | ✅ | 45 |
| [016](./016-admin-handler-split.md) | Разбиение admin_handler по доменам | ✅ | 53 |
| [017](./017-generic-rest-adapter.md) | Generic REST adapter (YAML, не Go-код) | ✅ | 71 |
| [018](./018-payment-gateway-service.md) | Payment Gateway → internal/payments/ | ⚠️ Note: M12 — merged | 66 |
| [019](./019-wizard-engine.md) | WizardEngine (JSON-driven generic wizard) | ✅ | 77 |
| [020](./020-nats-subjects-registry.md) | NATS subjects registry + billing namespace | ❌ Superseded | 41 |
| [021](./021-cel-unified-rule-engine.md) | **CEL — единый движок бизнес-логики** | ✅ | 182 |
| [022](./022-llm-proxy-service.md) | LLM Proxy как 5-й микросервис | ❌ Superseded (→ internal/llm/, M10) | 151 |
| [023](./023-gamification-system.md) | **Геймификация на базе CEL** | ✅ | 159 |
| [024](./024-modular-monolith.md) | **Modular Monolith — один бинарник** | ✅ | 209 |
| [025](./025-survey-engine.md) | **Survey Engine (engagement/survey/)** | ✅ | 633 |
| [026](./026-hr-sync-platform.md) | HR Sync → user/ (перенос из Integrations) | ✅ | 42 |
| [027](./027-payroll-in-billing.md) | 1C Payroll → billing/payroll/ | ✅ | 55 |
| [028](./028-nats-rename.md) | NATS namespace rename (integration → provider) | ⚠️ Superseded | 64 |
| [029](./029-ds-components-gap-tz.md) | ТЗ для April Design System (компоненты ЛК) | 📋 Note | 735 |
| [030](./030-ci-cd-pipeline.md) | CI/CD Pipeline | ✅ | 154 |
| [031](./031-api-data-fetching.md) | API Data Fetching — React Query | ✅ | 93 |
| [032](./032-api-types-codegen.md) | API Types — openapi-typescript | ✅ | 67 |
| [033](./033-frontend-testing.md) | Frontend Testing — Vitest + Playwright | ✅ | 118 |
| [034](./034-i18n-yagni.md) | i18n — YAGNI | ✅ | 84 |
| [035](./035-integration-proxy.md) | Integration Proxy — вынос внешних интеграций из монолита | ✅ Accepted | 438 |
<<<<<<< HEAD
| [036](./036-authorization-system.md) | **Авторизация — адаптация April → LKFL** | ✅ Accepted | ~600 |
=======
| [036](./036-ci-cd-deploy-worker.md) | CI/CD — serverAI self-hosted runners + Deploy Worker | ✅ Accepted | 122 |
| [037](./037-keycloak-reverse-proxy.md) | Keycloak behind reverse proxy — один nginx, чистый verifier.go | ✅ Accepted | 150 |
| [038](./038-staging-move-serverai.md) | Переезд staging с serverDev на serverAI | ✅ Accepted | 68 |
>>>>>>> origin/main

---

## Ключевые ADR (читать обязательно)

| ADR | Тема | Зачем |
|-----|------|-------|
| **024** | Modular Monolith | Главный архитектурный выбор: 1 бинарник, 16 пакетов монолита + proxy, без NATS |
| **021** | CEL Engine | Единый движок бизнес-правил (5 доменов) |
| **025** | Survey Engine | Опросы: бранчинг, tag-mapping, аналитика |
| **023** | Gamification | Ачивки, лояльность, триггеры |
| **013** | Internal пакеты | Почему пакеты, а не микросервисы |
| **009** | Multi-tenancy | Изоляция данных tenant_id |
| **008** | White-label | CSS-переменные для бренда |
| **031** | Data Fetching | React Query для 118 endpoints |
| **032** | API Types | openapi-typescodegen для types |

## Устаревшие (не использовать)

| ADR | Тема | Заменён |
|-----|------|---------|
| **005** | NATS JetStream | ADR-024 (Modular Monolith) |
| **020** | NATS subjects registry | ADR-024 |
| **022** | LLM Proxy (отдельный сервис) | `internal/llm/` (M10 T1002) |
| **028** | NATS namespace rename | ADR-024 |
