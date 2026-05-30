# T1704 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

Создан пакет `backend/internal/app/` — единая точка DI wiring приложения.

### `app/config.go` (143 строки)

- **Config struct** — корневая конфигурация с 7 подструктурами:
  - `ServerConfig` — PORT, ReadTimeout, WriteTimeout
  - `DatabaseConfig` — DSN, MaxConns, MinConns, MaxLifetime
  - `RedisConfig` — URL, MaxRetries
  - `KeycloakConfig` — Issuer, ClientID, ClientSecret
  - `SentryConfig` — DSN
  - `LogConfig` — Level, Format
- **LoadConfig()** — загрузка через viper: defaults → .env → ENV override
- **validate()** — проверка обязательных полей: DB_DSN, REDIS_URL, KEYCLOAK_ISSUER
- **IsDevelopment()** — хелпер для определения dev-режима

### `app/wire.go` (210 строк)

- **Logger interface** — Debug/Info/Warn/Error, реализуется через `log/slog`
- **Provide(cfg Config)** — функция DI wiring:
  1. Logger (JSON для production, text для dev)
  2. DB pool (pgxpool с ParseConfig + NewWithConfig)
  3. Redis client (go-redis/v9 с retry + backoff)
  4. Sentry (опционально, по наличию DSN)
  5. Prometheus registry (custom, не default)
  6. OIDC verifier (go-oidc v2.3.0+incompatible)
  7. Server (chi router + middleware)
- **Cleanup function** — graceful shutdown: DB.Close, Redis.Close, Sentry.Flush
- **Rollback при ошибках** — каждый шаг освобождает ресурсы предыдущих шагов

### `app/server.go` (179 строк)

- **Server struct** — обёртка над http.Server + chi.Mux
- **httpMetrics struct** — вынесенные метрики Prometheus (total, duration)
- **newHTTPMetrics(reg)** — создание и регистрация метрик один раз в NewServer()
- **NewServer()** — создание сервера с зависимостями
- **Middleware chain** (порядок):
  1. RequestID — уникальный ID запроса
  2. RealIP — определение реального IP
  3. Logger — логирование запросов
  4. Recoverer — обработка паник
  5. Timeout — 30s ограничение
  6. Prometheus — http_requests_total + http_request_duration_seconds
- **CORS** — dev: * (заглушка, TODO: config-driven в production)
- **/healthz** → 200 OK
- **/metrics** → Prometheus handler (custom registry)
- **Start()** / **Shutdown(ctx)** — lifecycle методы

## Технические решения

| Решение | Обоснование |
|---------|-------------|
| `log/slog` вместо zap/zerolog | Стандартная библиотека Go 1.21+, JSON/text форматы OOTB |
| `pgxpool.ParseConfig` + builder | pgx/v5 не имеет SetMaxConns на Pool, нужен Config |
| `go-oidc` v2.3.0+incompatible | Версия без /v3 суффикса (соответствует go.mod) |
| Custom Prometheus registry | Избежание конфликтов с default registry |
| Ручной CORS middleware | go-chi/cors не в go.mod, rs/cors требует другой API |
| Rollback в Provide() | Каждый шаг освобождает ресурсы предыдущих при ошибке |
| httpMetrics struct | Регистрация метрик вне middleware — безопасно для тестов и рестарта |
| chi.RouteContext().RoutePattern() | Кардинальность метрик по route pattern, не по raw URL path |

## Исправления (по итогам аудита)

1. **Prometheus кардинальность** — заменён `r.URL.Path` на `chi.RouteContext(r.Context()).RoutePattern()`.
    Ранее метрика `http_request_duration_seconds` имела высокую кардинальность:
    каждый уникальный URL (напр. `/api/v1/tenants/sdek/benefits/abc-123`) создавал
    отдельное значение label `path`. Теперь используется pattern
    `/api/v1/tenants/{tenant_id}/benefits/{benefit_id}`.

2. **Prometheus MustRegister** — регистрация метрик вынесена из `PrometheusMiddleware()`
    в `newHTTPMetrics()`, вызываемую один раз в `NewServer()`.
    Ранее повторный вызов `NewServer()` вызывал panic при `MustRegister` дубликата.

3. **Отступ в server.go:148** — исправлен сдвиг строки `pattern := r.URL.Path`.

4. **go.mod toolchain** — удалена директива `toolchain go1.24.4` (M17 audit).
    Minimum version остаётся `go 1.22`. `go mod tidy` на Go 1.24+ может
    восстановить toolchain директиву — это не блокирующая проблема.

5. **internal/bootstrap/deps.go → internal/deps/retained.go** — перемещён и обновлён.
    Пакет `bootstrap` переименован в `deps`. 8 blank imports заменены на 7 retained
    с привязкой к фазам (F2: validator, cel-go, bcrypt; F3: gofpdf, excelize;
    General: rs/cors, protobuf).

## Проверка

- `go build ./...` — ✅ чистая компиляция
- `go vet ./...` — ✅ без замечаний
- Все зависимости из go.mod (без добавления новых)

## Время

~40 минут (оригинал) + ~15 минут (исправления)

## Замечания

- CORS middleware — заглушка с `*` origin; в production заменить на config-driven список доменов
- OIDC verifier не хранит provider (только verifier) — provider может понадобиться для introspection endpoint в будущем
- Prometheus middleware использует chi.RouteContext — работает только с chi router
