# T2206 — Отчёт: Мониторинг (Prometheus + Grafana)

## Выполнено

### 1. Кастомные Prometheus метрики (`backend/internal/metrics/metrics.go`)

Создан новый пакет `metrics` с кастомными метриками:

- **Catalog:**
  - `catalog_query_total` — counter (type: list|get|categories, status: success|error)
  - `catalog_query_duration_seconds` — histogram (type)
- **Tenant:**
  - `tenant_resolve_total` — counter (method: host|header, status: success|error)
  - `tenant_resolve_duration_seconds` — histogram (method)
- **Auth:**
  - `auth_login_total` — counter (status: success)
  - `auth_callback_total` — counter (status: success|failure|error)
- **Redis cache:**
  - `redis_cache_hits_total` — counter
  - `redis_cache_misses_total` — counter
  - `redis_cache_evictions_total` — counter

Все метрики nil-safe (при Metrics == nil методы no-op).

### 2. Интеграция метрик в код

- **HTTP middleware** (`server.go`): добавлен `tenant_id` label в `http_requests_total` и `http_request_duration_seconds`
- **Auth handler** (`auth/handler.go`): метрики login/callback в LoginRedirect и LoginCallback
- **Tenant middleware** (`tenant/middleware.go`): метрики tenant resolution в HostResolver.Resolve
- **Catalog handler** (`catalog/handler.go`): метрики catalog queries в List/Get/Categories
- **Catalog cache** (`catalog/cache.go`): метрики Redis cache hits/misses/evictions
- **DI** (`wire.go`): создание `metrics.New(reg)` и передача через все компоненты

### 3. Prometheus конфиг

`infra/prometheus/prometheus.yml`:
- Scrape lkfl-server (:8080/metrics)
- Scrape lkfl-integration-proxy (:8091/metrics)
- Интервал: 15s

### 4. Grafana dashboard

`infra/grafana/dashboards/platform-f1.json` — JSON dashboard "LKFL Platform Overview F1":
- **Overview:** RPS, Latency P50/P95/P99, Error Rate
- **Catalog:** Queries RPS by Type, Query Latency P95
- **Tenant:** Resolution by Method, Resolution Latency P95
- **Auth:** Login Attempts, Callback Attempts
- **Redis Cache:** Hits/Misses, Cache Hit Ratio
- **Variables:** tenant_id, method
- **Time presets:** 1h, 6h, 24h, 7d

Provisioning config: `infra/grafana/dashboards/datasource.yml` + `infra/grafana/provisioning/datasources.yml`

### 5. Alerting rules

`infra/grafana/alerting/rules.yml`:
- HighErrorRate (>1%, page)
- HighLatencyP95 (>500ms, warning)
- HighLatencyP99 (>1000ms, page)
- CatalogQuerySlow (>1s, warning)
- TenantResolveFailures (warning)
- AuthCallbackFailures (>5%, page)
- RedisCacheMissRateHigh (>80%, warning)
- InstanceDown (page)

### 6. Docker Compose

Добавлены сервисы:
- **prometheus** (v2.53.0) — порт 9090, persistent volume, retention 15d
- **grafana** (11.1.0) — порт 3000, anonymous access, provisioning dashboards/alerting

## Изменённые файлы

| Файл | Действие |
|------|----------|
| `backend/internal/metrics/metrics.go` | Создан |
| `backend/internal/app/server.go` | Изменён (tenant_id label, metrics import) |
| `backend/internal/app/wire.go` | Изменён (metrics.New, DI) |
| `backend/internal/auth/handler.go` | Изменён (metrics field + tracking) |
| `backend/internal/tenant/middleware.go` | Изменён (metrics field + tracking) |
| `backend/internal/tenant/isolation.go` | Изменён (AdminTenantMiddleware + metrics) |
| `backend/internal/engagement/catalog/handler.go` | Изменён (metrics field + tracking) |
| `backend/internal/engagement/catalog/cache.go` | Изменён (metrics field + tracking) |
| `backend/internal/engagement/catalog/handler_test.go` | Изменён (nil metrics) |
| `backend/internal/engagement/catalog/cache_test.go` | Изменён (nil metrics) |
| `backend/internal/tenant/middleware_test.go` | Изменён (nil metrics) |
| `backend/internal/testutil/testcontainers.go` | Изменён (nil metrics) |
| `infra/prometheus/prometheus.yml` | Создан |
| `infra/grafana/dashboards/platform-f1.json` | Создан |
| `infra/grafana/dashboards/datasource.yml` | Создан |
| `infra/grafana/provisioning/datasources.yml` | Создан |
| `infra/grafana/alerting/rules.yml` | Создан |
| `docker-compose.yml` | Изменён (+prometheus, +grafana) |

## Проверка

- [x] `go build ./...` — чистая компиляция
- [x] `go test ./... -short` — все тесты зелёные
- [x] go.mod не изменён (prometheus/client_golang уже в зависимостях)

## Время

~2 часа

## Замечания

- Метрики nil-safe: все компоненты работают корректно при Metrics == nil
- Tenant resolution метрики собираются только в HostResolver (PathResolver не изменён — используется только в админ-контексте)
- Grafana dashboard provisioned через файлы — не требует ручного импорта
