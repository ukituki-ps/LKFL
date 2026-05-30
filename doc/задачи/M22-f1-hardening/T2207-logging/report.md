# T2207 — Logging (Loki) — Отчёт

## Что сделано

### 1. Пакет `shared/pkg/logger/` — кастомный slog handler

Создан пакет `lkfl/shared/pkg/logger/` с:

- **`SvcHandler`** — обёртка над `slog.Handler`, добавляющая атрибут `svc` (имя сервиса) в каждую запись лога. Реализует интерфейс `slog.Handler` (Go 1.21+ с `context.Context` в `Enabled` и `Handle`).
- **`New(Options)`** — фабрика для создания логгера с настройками:
  - `Level` — уровень логирования (debug/info/warn/error)
  - `Format` — формат вывода (json по умолчанию, text для dev)
  - `Service` — имя сервиса для атрибута `svc`
  - `Writer` — вывод логов (по умолчанию `os.Stdout`)
- **`parseLevel`** — парсинг строкового уровня в `slog.Level`

Формат JSON-лога (Loki-совместимый):
```json
{"time":"2025-01-01T12:00:00Z","level":"INFO","svc":"lkfl-server","msg":"catalog query executed","tenant_id":"sdek","duration_ms":42}
```

Дополнительные атрибуты (`tenant_id`, `user_id`, `trace_id`) добавляются через `logger.With()` в middleware/handlers.

### 2. Обновлён `internal/app/wire.go`

- `newLogger()` теперь использует `logger.New()` из пакета `shared/pkg/logger/`
- Удалён дублирующий код (`parseLogLevel`, ручное создание handler'ов)
- Сервисное имя: `"lkfl-server"`

### 3. `infra/loki/loki.yml` — конфигурация Loki

- Порт: 3100
- Хранение: boltdb-shipper + filesystem
- Ретеншн: 30 дней (`reject_old_samples_max_age: 30d`)
- Макс. записей на запрос: 5000

### 4. `infra/loki/promtail.yml` — конфигурация Promtail

- Сбор логов через Docker API (`docker_sd_configs`)
- Auto-discovery контейнеров по label `com.docker.compose.service`
- Целевые сервисы: `lkfl-server`, `lkfl-integration-proxy`
- Relabel: `service` label для фильтрации в Grafana Explore

### 5. Обновлён `docker-compose.yml`

Добавлены сервисы:

- **`loki`** (grafana/loki:3.0.0)
  - Порт: 3100
  - Volume: `lkfl_loki_data` для персистентности
  - Healthcheck: `/ready`
  - Сеть: `lkfl_backend`

- **`promtail`** (grafana/promtail:3.0.0)
  - Volume: docker.sock (ro) для Docker API
  - Зависимость: loki (service_healthy)
  - Сеть: `lkfl_backend`

- **Volume `lkfl_loki_data`** добавлен в секцию volumes

### 6. Обновлён `infra/grafana/provisioning/datasources.yml`

Добавлен datasource Loki:
```yaml
- name: Loki
  type: loki
  access: proxy
  url: http://loki:3100
  uid: loki
```

### 7. Тесты

- `shared/pkg/logger/logger_test.go` — 6 тестов:
  - `TestNew_JSONFormat` — JSON вывод с обязательными полями
  - `TestNew_TextFormat` — text формат
  - `TestNew_DefaultService` — дефолтное имя сервиса
  - `TestNew_LogLevels` — все уровни (debug/info/warn/error)
  - `TestNew_FilteredByLevel` — фильтрация по уровню
  - `TestSvcHandler_WithAttrs` — With() для tenant_id, user_id

## Результаты

| Команда | Результат |
|---------|-----------|
| `go build ./...` | ✅ чистая компиляция |
| `go test ./... -short` | ✅ все тесты зелёные |

## Grafana Explore — примеры LogQL запросов

```logql
# Все логи lkfl-server
{service="lkfl-server"}

# Логи по уровню (парсинг JSON)
{service="lkfl-server"} |= `\"level\":\"ERROR\"`

# Логи по tenant
{service="lkfl-server"} |= `\"tenant_id\":\"sdek\"`

# Ошибки авторизации
{service="lkfl-server"} |= `\"level\":\"ERROR\"` |= `auth`

# Медленные запросы (> 1000ms)
{service="lkfl-server"} |= `\"duration_ms\"` | json | duration_ms > 1000
```

## Замечания

- Promtail использует Docker API mode (не file-based), т.к. Go-приложение пишет в stdout, а Docker json-file driver собирает логи контейнеров автоматически.
- `tenant_id`, `user_id`, `trace_id` добавляются вызывающим кодом через `logger.With()` — это соответствует паттерну middleware enrichment.
