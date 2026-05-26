# ADR-035: Integration Proxy — вынос внешних интеграций из монолита

**Статус:** Accepted
**Дата:** 2026-05-26
**Контекст:** M16-integration-proxy, T1601

---

## Контекст

ADR-024 (M12) принял modular monolith: один бинарник `lkfl-server`, 17 internal-пакетов, Go interfaces вместо NATS. Пакет `internal/integrations/` содержит ProviderGateway — gateway к внешним провайдерам льгот (11 адаптеров).

При этом `engagement/flow/` вызывает `integrations.ProviderGateway.Activate()` **синхронно** в hot path активации льготы. Прямой HTTP call из монолита к внешнему миру создаёт системные риски:

| Риск | Механизм | Влияние |
|------|----------|---------|
| **Блокировка горутин** | `Activate()` → HTTP call → 30s timeout | Горутина HTTP handler заблокирована 30с. При 100K+ сотрудников и пике — goroutine exhaustion, рост памяти, каскадный сбой |
| **Отсутствие fault isolation** | Один медленный провайдер в одном процессе | Memory leak в адаптере → OOM всего бинарника. Panic в адаптере → crash всего бинарника |
| **Credential blast radius** | Креденшиалы всех провайдеров в одном процессе | Компрометация → доступ ко всем интеграциям одновременно |
| **Webhook — публичная поверхность** | Webhook endpoint'ы на основном бинарнике (:8080) | Каждый провайдер может слать события на основной порт. Rate limit, DoS на главный бинарник |
| **Circuit breaker только в документации** | Описан в `интеграции.md`, не реализован в коде | Без кода — нет гарантии. При реализации легко пропустить |

**Проблема:** философия "1 бинарник" (ADR-024) конфликтует с requirement изоляции внешних вызовов. ADR-024 обосновывал монолит для agent-разработки (compile-time safety, один build), но внешние HTTP calls — это не бизнес-логика, это I/O boundary.

---

## Термины

| Термин | Определение |
|--------|-------------|
| **Монолит** | `lkfl-server` — основной бинарник, бизнес-логика, HTTP API для фронтенда |
| **Integration Proxy** | `lkfl-integration-proxy` — отдельный бинарник, gateway к внешним провайдерам |
| **Провайдер** | Внешний сервис льготы: ДМС, фитнес, питание, обучение и т.д. |
| **Адаптер** | Реализация `ProviderAdapter` interface для конкретного провайдера |
| **Hot path** | Путь запроса, напрямую влияющий на время ответа пользователю |

---

## Рассмотренные варианты

### Вариант 1: Оставить как есть (прямые вызовы из монолита)

```
lkfl-server (:8080)
  └── internal/integrations/
        └── HTTP call → внешний провайдер
```

| Плюсы | Минусы |
|-------|--------|
| Один бинарник, минимальная сложность | Блокировка горутин в hot path |
| Нет network overhead между модулями | Нет fault isolation |
| Простота разработки | Credential blast radius |

**Вердикт:** ❌ Отказ. Риски превышают benefits при 100K+ сотрудников и ФСТЭК-сертификации.

### Вариант 2: Asynq worker (без нового бинарника)

```
lkfl-server (:8080)
  ├── engagement/flow/ → enqueue Asynq job "activate-provider"
  └── cmd/worker/ → Asynq consumer → HTTP call → провайдер
```

Активация становится асинхронной. Пользователь получает "активация в процессе", результат — нотификацией.

| Плюсы | Минусы |
|-------|--------|
| Не нарушает "1 бинарник" | Асинхронный UX — пользователь ждёт нотификацию |
| Hot path разорван | Webhook всё ещё на основном бинарнике |
| Использует существующий Asynq | Credential blast radius не решён |
| | Нет fault isolation (тот же процесс) |

**Вердикт:** ❌ Отказ. Решает только goroutine blocking, но не fault isolation и credential blast radius.

### Вариант 3: Integration Proxy (ВЫБРАН)

```
                    ┌─────────────────────────────┐
                    │         Nginx (:80)         │
                    │                             │
                    │  /api/v1/*     → :8080      │
                    │  /webhooks/*   → :8090      │
                    └──────┬──────────────┬───────┘
                           │              │
               ┌───────────▼──┐    ┌──────▼──────────────┐
               │  lkfl-       │    │ lkfl-integration-   │
               │  server      │    │ proxy               │
               │  (:8080)     │    │ (:8090 gRPC)        │
               │              │    │                     │
               │  gRPC client │───→│  Provider adapters  │
               │  (localhost) │    │  Circuit breaker    │
               └──────────────┘    │  Retry/timeout      │
                                   │  Webhook receiver   │
                                   │  Credential store   │
                                   └────────┬────────────┘
                                            │
                                   ┌────────▼────────┐
                                   │  Внешние         │
                                   │  провайдеры      │
                                   │  (HTTP/REST)     │
                                   └─────────────────┘
```

| Плюсы | Минусы |
|-------|--------|
| **Полная fault isolation** — proxy упал, монолит работает с кэшем | Дополнительный бинарник (нарушает "1 бинарник" ADR-024) |
| **Credential isolation** — ключи провайдеров только в proxy | Network overhead (gRPC localhost ≈ 0.1ms — пренебрежимо) |
| **Webhook isolation** — webhook endpoint'ы на proxy, не на монолите | Нужно поддерживать gRPC contract между монолитом и proxy |
| **Independent deploy** — новый адаптер → redeploy только proxy | Дополнительный контейнер в docker-compose |
| **Circuit breaker на уровне процесса** — panic в адаптере → restart proxy, не монолита | |
| **Go interface → gRPC** — compile-time safety сохраняется (protoc генерация) | |

**Вердикт:** ✅ Выбран. Fault isolation и credential isolation критичны для production. Network overhead на localhost пренебрежим.

### Вариант 4: Sidecar pattern (Kubernetes)

Proxy как sidecar контейнер рядом с монолитом в одном pod.

| Плюсы | Минусы |
|-------|--------|
| Развёртывается вместе с монолитом | Требует Kubernetes |
| localhost communication | Overkill для текущего масштаба |

**Вердикт:** ❌ Отказ. Проект использует Docker Compose, не Kubernetes. Sidecar — premature optimization.

---

## Решение

### Архитектура

**Два бинарника, один репозиторий, один `go.mod`:**

```
lkfl/
├── cmd/server/main.go              # lkfl-server (:8080) — бизнес-логика
├── cmd/worker/main.go              # Asynq worker (тот же бинарник server)
├── cmd/integration-proxy/main.go   # lkfl-integration-proxy (:8090 gRPC)
├── internal/
│   ├── tenant/, auth/, user/, ...  # business-пакеты монолита (15 пакетов)
│   ├── api/                        # HTTP router монолита
│   └── integrationclient/          # gRPC client к proxy (НОВЫЙ)
├── integration-proxy/
│   ├── adapters/                   # ProviderAdapter реализации (11 провайдеров)
│   ├── circuitbreaker/             # Circuit breaker per provider
│   ├── webhook/                    # Webhook receiver + verifier
│   ├── grpc/                       # gRPC server + generated code
│   └── config/                     # YAML config загрузка
├── proto/
│   └── integration/v1/
│       └── integration.proto       # gRPC service definition
├── provider-configs/               # YAML конфиги провайдеров
├── go.mod                          # один go.mod на всё
└── Dockerfile                      # multi-stage: server + proxy
```

> **Важно:** один `go.mod` сохраняется. ADR-024 требование "один go.mod" не нарушается. Два бинарника — два `cmd/`, один модуль.

### gRPC Contract

```protobuf
syntax = "proto3";

package integration.v1;

service IntegrationService {
  // Сynchronous — быстрые/idempotent операции (< 5s)
  rpc GetProviderStatus(ProviderStatusRequest) returns (ProviderStatusResponse);
  rpc GetCatalog(CatalogRequest) returns (CatalogResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);

  // Asynchronous — медленные операции (activate/deactivate)
  rpc Activate(ActivateRequest) returns (ActivateResponse);
  rpc Deactivate(DeactivateRequest) returns (DeactivateResponse);

  // Admin — управление провайдерами
  rpc ListProviders(ListProvidersRequest) returns (ListProvidersResponse);
  rpc GetProvider(GetProviderRequest) returns (GetProviderResponse);
  rpc UpdateProvider(UpdateProviderRequest) returns (UpdateProviderResponse);
  rpc TriggerSync(TriggerSyncRequest) returns (TriggerSyncResponse);
  rpc GetSyncLogs(GetSyncLogsRequest) returns (GetSyncLogsResponse);
}

message ActivateRequest {
  string tenant_id = 1;
  string user_id = 2;
  string offer_id = 3;
  string provider_name = 4;
  map<string, string> metadata = 5;
}

message ActivateResponse {
  string job_id = 1;           // ID асинхронной задачи
  string status = 2;           // "queued", "processing"
  int64 estimated_seconds = 3; // оценка времени выполнения
}

message ProviderStatusRequest {
  string tenant_id = 1;
  string user_id = 2;
  string provider_name = 3;
}

message ProviderStatusResponse {
  string status = 1;           // "active", "inactive", "error"
  string policy_number = 2;
  map<string, string> details = 3;
}
```

### Синхронные vs Асинхронные операции

| Операция | Режим | Обоснование |
|----------|-------|-------------|
| `Activate` | **Асинхронный** | HTTP call к провайдеру 5-30с. Возвращает `job_id`. Монолит → "активация в процессе" → poll/webhook |
| `Deactivate` | **Асинхронный** | Аналогично activate |
| `GetProviderStatus` | **Синхронный** | Быстрый (< 5s), idempotent, используется в UI для отображения статуса |
| `GetCatalog` | **Синхронный** | Читает кэш из БД proxy, не вызывает провайдера |
| `HealthCheck` | **Синхронный** | Быстрый, используется для мониторинга |

**Почему Activate асинхронный:** Это решает goroutine blocking проблему. Монолит вызывает `Activate()` → получает `job_id` за < 10ms → сохраняет "in_progress" → пользователь видит прогресс. Proxy выполняет HTTP call в своём worker pool.

**Flow активации (асинхронный):**
```
1. Пользователь → POST /api/v1/engagements/:id/activate
2. Монолит → gRPC Activate() → proxy
3. Proxy → возвращает job_id за < 10ms
4. Монолит → сохраняет "in_progress" → отвечает пользователю "активация запущена"
5. Proxy → worker pool → HTTP call к провайдеру → результат
6. Proxy → HTTP POST /internal/webhook/callback (mono) → "активация завершена"
7. Монолит → обновляет статус → отправляет нотификацию пользователю
```

### Webhook handling

```
Внешний провайдер → Nginx /webhooks/:provider → Proxy (:8090)
                                                            │
                                                1. Verify signature
                                                2. Validate payload
                                                3. Store event in DB
                                                4. HTTP POST /internal/webhook/callback → Монолит
```

Webhook endpoint'ы живут **только** на proxy. Монолит не exposes публичных webhook endpoint'ов.

Proxy вызывает монолит через **внутренний HTTP endpoint** `/internal/webhook/callback` (не через gRPC, потому что это inbound call от proxy к монолиту, не request-response).

### Circuit Breaker

Каждый провайдер имеет собственный circuit breaker:

| Параметр | Значение |
|----------|---------|
| Failure threshold | 10 ошибок за 60с |
| Recovery timeout | 30с (half-open probe) |
| State: closed | Нормальная работа |
| State: open | Все вызовы отклоняются, возвращается cached/error response |
| State: half-open | Один пробный запрос, если success → closed, если failure → open |

### Worker Pool per Provider

Каждый провайдер имеет ограниченный пул горутин:

```go
type ProviderWorkerPool struct {
    maxConcurrent int    // default: 5
    queueSize     int    // default: 100
    semaphore     chan struct{}
    jobQueue      chan *ActivationJob
}
```

Это предотвращает scenario когда один провайдер потребляет все ресурсы.

### Credential Storage

| Где | Что хранится |
|-----|-------------|
| **Proxy** | Provider API credentials (encrypted in `providers` table, decrypted at runtime) |
| **Монолит** | Ничего — не знает о креденшиалах провайдеров |
| **HashiCorp Vault** | Master key для decryption (production) |
| **Environment variables** | Master key (development) |

### Database

**Один PostgreSQL, две schema:**

| Schema | Бинарник | Назначение |
|--------|----------|-----------|
| `lkfl_platform` | `lkfl-server` | Бизнес-данные (tenants, users, engagements, billing, ...) |
| `lkfl_integration` | `lkfl-integration-proxy` | Данные интеграций (providers, sync_log, webhook_events, dead_letters) |

Таблицы proxy (`lkfl_integration`):

| Таблица | Назначение |
|---------|-----------|
| `providers` | Конфигурация провайдеров (name, protocol, endpoints, auth_method, status) |
| `provider_sync_log` | Лог синхронизации каталога (provider, timestamp, count, errors) |
| `webhook_events` | Входящие webhook события (provider, payload, processed_at, status) |
| `dead_letters` | Отравшиеся сообщения (provider, operation, error, retry_count, created_at) |
| `activation_jobs` | Асинхронные задачи активации (job_id, status, result, created_at, completed_at) |
| `circuit_breaker_state` | Состояние circuit breaker (provider, state, last_failure, failure_count) |

> **Примечание:** таблица `providers` переезжает из `lkfl_platform` в `lkfl_integration`. Таблица `external_services` остаётся в `lkfl_platform` (используется монолитом для SSO redirect).

### Fault Tolerance

| Сценарий | Поведение |
|----------|-----------|
| Proxy недоступен | Монолит возвращает 503 с message "сервис интеграций временно недоступен". Каталог читается из локального кэша в PG. Активация откладывается |
| Провайдер недоступен | Circuit breaker open → cached response. Activation job → dead letter с retry schedule |
| Proxy упал (panic) | Docker restart policy → restart. Монолит продолжает работать (read-only mode для интеграций) |
| Network partition (localhost) | Не возможен в Docker network (same host). На bare metal — unlikely |

### Мониторинг

| Метрика | Тип | Описание |
|---------|-----|---------|
| `integration_provider_latency_seconds` | Histogram | Latency вызовов по провайдеру |
| `integration_provider_errors_total` | Counter | Ошибки по провайдеру |
| `integration_circuit_breaker_state` | Gauge | Состояние circuit breaker (0=closed, 1=open, 2=half-open) |
| `integration_sync_duration_seconds` | Histogram | Время синхронизации каталога |
| `integration_webhook_total` | Counter | Входящие webhooks по провайдеру |
| `integration_dead_letters_total` | Counter | Dead letters по провайдеру |
| `integration_active_jobs` | Gauge | Активные задачи в worker pool |

### Nginx configuration

```nginx
# API монолита
location /api/ {
    proxy_pass http://lkfl-server:8080;
}

# Webhooks от провайдеров
location /webhooks/ {
    proxy_pass http://lkfl-integration-proxy:8090;
    client_max_body_size 1M;
}

# Admin UI → монолит (он проксирует запросы к proxy для admin операций)
location /admin/ {
    proxy_pass http://lkfl-server:8080;
}

# Health checks
location /healthz {
    proxy_pass http://lkfl-server:8080;
}

location /proxy-healthz {
    proxy_pass http://lkfl-integration-proxy:8090;
}
```

### Ports

| Бинарник | Порт | Протокол | Назначение |
|----------|------|----------|-----------|
| `lkfl-server` | 8080 | HTTP | API для фронтенда |
| `lkfl-integration-proxy` | 8090 | gRPC | gRPC для монолита + HTTP для webhooks |
| Asynq dashboard | 8083 | HTTP | Мониторинг фоновых задач |

> Proxy использует gRPC и HTTP на одном порту (grpc-gateway или separate listener). Для простоты — два listener'а: gRPC на 8090, HTTP webhooks на 8091.

### Docker Compose

```yaml
services:
  lkfl-server:
    build: .
    command: server
    ports: ["8080:8080"]
    depends_on: [postgres, redis, keycloak]

  lkfl-integration-proxy:
    build: .
    command: integration-proxy
    ports: ["8090:8090", "8091:8091"]
    depends_on: [postgres, redis]
    environment:
      - INTEGRATION_DB_SCHEMA=lkfl_integration

  postgres:
    image: postgres:17
    # ...

  redis:
    image: redis:7
    # ...
```

---

## Последствия

### Положительные

1. **Fault isolation** — panic в адаптере → restart proxy, не монолита
2. **Credential isolation** — ключи провайдеров только в proxy
3. **Goroutine safety** — асинхронная активация, hot path не блокируется
4. **Webhook isolation** — публичные webhook endpoint'ы на proxy, не на монолите
5. **Independent scaling** — proxy масштабируется независимо при высокой нагрузке на интеграции
6. **Independent deploy** — новый адаптер → redeploy только proxy
7. **Circuit breaker на уровне процесса** — гарантирован, не зависит от дисциплины разработчика
8. **Один go.mod** — требование ADR-024 сохранено

### Отрицательные

1. **Два бинарника** — нарушает "1 бинарник" из ADR-024. **Обоснование:** ADR-024 обосновывал монолит для бизнес-логики. Внешние HTTP calls — это не бизнес-логика, это I/O boundary. Proxy — это infra component, не business module.
2. **gRPC contract** — нужно поддерживать `.proto` файл. **Митигация:** protoc генерация при build, один репозиторий.
3. **Дополнительный контейнер** — docker-compose: 5 → 6 контейнеров.
4. **Network latency** — gRPC localhost ≈ 0.1ms. Для синхронных операций (< 5s) — пренебрежимо. Для асинхронных — не применимо.

### Влияние на ADR-024

ADR-024 **НЕ отменяется**. Он остаётся valid для бизнес-логики (16 internal-пакетов монолита, один бинарник). Integration Proxy — это **исключение** для I/O boundary, аналогичное тому, как `cmd/worker/` уже является вторым entry point в том же бинарнике.

ADR-024 обновляется: добавляется секция "Exception: Integration Proxy" с обоснованием.

---

## Связь с другими ADR

| ADR | Статус | Изменение |
|-----|--------|-----------|
| ADR-024 (Modular Monolith) | ⚠️ Note | Добавлено исключение: Integration Proxy для I/O isolation |
| ADR-017 (Generic REST adapter) | ✅ Valid | Адаптеры переезжают в `integration-proxy/adapters/` |
| ADR-006 (Billing) | ✅ Valid | Не затрагивается |
| ADR-018 (Payment Gateway) | ✅ Valid | Payments остаётся в монолите (PCI DSS — иная изоляция) |

---

## Статус

✅ Accepted (M16, T1601–T1613)
