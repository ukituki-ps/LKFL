# T1607 — Отчёт: `архитектура/модули.md` — +proxy binary

## Веха

M16-integration-proxy

## Что сделано

Файл `doc/архитектура/модули.md` обновлён для отражения архитектуры с двумя бинарниками (монолит + integration proxy).

### Изменения

| # | Секция | Изменение |
|---|--------|-----------|
| 1 | **TL;DR** | 1 бинарник → 2 бинарника (lkfl-server + lkfl-integration-proxy), 1 go.mod. Добавлен note: proxy — I/O boundary, не business module. Ссылка на ADR-035. |
| 2 | **Краткое описание** | Обновлено: два бинарника, 14 business-пакетов + tenant/ + api/ + integrationclient/ = 16 internal-пакетов монолита. Добавлен M16 note. |
| 3 | **ASCII diagram (структура проекта)** | Полностью переработана: добавлен `cmd/integration-proxy/`, `integration-proxy/` (adapters, circuitbreaker, webhook, grpc, config), `proto/integration/v1/`, `provider-configs/`. Убран `internal/integrations/`, добавлен `internal/integrationclient/`. |
| 4 | **Архитектурная диаграмма** | Новая секция: Nginx → mono (:8080) + proxy (:8090 gRPC, :8091 HTTP webhooks) → внешние провайдеры. |
| 5 | **lkfl-server** | Заголовок: «монолит (бизнес-логика)». Описание: убраны интеграции из назначения. M16 note в истории. |
| 6 | **lkfl-integration-proxy** | Новая секция: назначение, точка входа, структура (adapters, circuitbreaker, webhook, grpc, config). |
| 7 | **Таблица internal пакетов** | Убрана строка `integrations/`. Добавлена строка `integrationclient/` (gRPC client к proxy). Обновлены зависимости `engagement/flow/` (integrations/ → integrationclient/). |
| 8 | **Контракты между пакетами** | `engagement/flow/` → `integrations/` direct Go call → `integrationclient/` gRPC (localhost :8090) → proxy. Admin API аналогично. |
| 9 | **DI граф** | `integrations/` → `integrationclient/` → gRPC → proxy. Добавлен блок lkfl-integration-proxy (отдельный бинарник). |
| 10 | **Порты** | 8080 (mono HTTP), 8083 (asynq dashboard), 8090 (proxy gRPC), 8091 (proxy HTTP webhooks). |
| 11 | **Nginx routing** | `/webhooks/` → lkfl-integration-proxy:8091. Добавлен `/proxy-healthz` → proxy:8090. |
| 12 | **PostgreSQL** | Добавлена schema `lkfl_integration` (proxy): providers, provider_sync_log, webhook_events, dead_letters, activation_jobs, circuit_breaker_state. |
| 13 | **Redis** | Добавлен prefix `integration:` (proxy — provider status cache, circuit breaker state, webhook dedup). |
| 14 | **Зависимости между пакетами** | `integrations/` → `lkfl-integration-proxy`. Добавлен webhook flow: провайдеры → Nginx → proxy → монолит. gRPC note. |
| 15 | **Релизная политика** | Два бинарника → независимый deploy proxy. gRPC контракт → protoc генерация, один go.mod. |
| 16 | **Docker Compose** | Новая секция: 6 контейнеров (lkfl-server, lkfl-integration-proxy, postgres, redis, keycloak, nginx). |
| 17 | **Frontend** | Note: без изменений, proxy невидим для фронтенда. |

## Консистентность с ADR-035

Все изменения согласованы с [ADR-035](../архитектура/adr/035-integration-proxy.md):

- ✅ Два бинарника, один go.mod
- ✅ Proxy — I/O boundary, не business module
- ✅ gRPC localhost :8090, HTTP webhooks :8091
- ✅ `integration-proxy/adapters/` (не `internal/integrations/`)
- ✅ `integrationclient/` — gRPC client монолита
- ✅ PostgreSQL: `lkfl_platform` + `lkfl_integration`
- ✅ Redis: `integration:` prefix
- ✅ Nginx routing: `/webhooks/` → proxy
- ✅ Docker compose: 6 контейнеров
- ✅ Circuit breaker, credential isolation, fault isolation упомянуты

## Затраченное время

~25 минут

## Замечания

- Файл `пакеты-platform.md` требует отдельной задачи для обновления публичных API (включая removal `integrations/` и добавление `integrationclient/`).
- ADR-024 требует обновления секции «Exception: Integration Proxy» (упомянуто в ADR-035 как следствие).
