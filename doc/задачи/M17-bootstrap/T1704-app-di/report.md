# T1704 — Отчёт о выполнении

## Статус

выполнено

## Что сделано

Создан пакет `backend/internal/app/` — единая точка DI wiring приложения.

### `app/config.go` (142 строки)

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

### `app/wire.go` (188 строк)

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

### `app/server.go` (137 строк)

- **Server struct** — обёртка над http.Server + chi.Mux
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

## Проверка

- `go build ./...` — ✅ чистая компиляция
- Все зависимости из go.mod (без добавления новых)

## Время

~40 минут

## Замечания

- CORS middleware — заглушка с `*` origin; в production заменить на config-driven список доменов
- OIDC verifier не хранит provider (только verifier) — provider может понадобиться для introspection endpoint в будущем
- Prometheus middleware использует `r.Pattern` (chi-specific) — работает только с chi router
