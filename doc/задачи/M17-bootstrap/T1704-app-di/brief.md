# T1704 — DI Wiring (app/)

## Веха

M17-bootstrap

## Тип

code

## Контекст

`app/` — пакет dependency injection. Единая точка инициализации всех зависимостей.
Паттерн: config → infrastructure (DB, Redis) → business packages → handlers → router → server.

Описан в `doc/архитектура/модули.md` строка 62 (`app/ — DI wiring`).

## Что сделать

Создать структуру `app/`:

```
app/
├── config.go        # Config struct + viper loading
├── wire.go          # DI wiring function
├── server.go        # HTTP server setup (chi router + middleware)
└── wire_gen.go      # (auto-generated, если используем google/wire)
```

### `app/config.go`

```go
package app

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Keycloak KeycloakConfig
    Sentry   SentryConfig
    Log      LogConfig
}

type ServerConfig struct {
    Port        int    `mapstructure:"SERVER_PORT"`
    ReadTimeout int    `mapstructure:"SERVER_READ_TIMEOUT"`    // seconds
    WriteTimeout int   `mapstructure:"SERVER_WRITE_TIMEOUT"`   // seconds
}

type DatabaseConfig struct {
    DSN string `mapstructure:"DB_DSN"`
    MaxConns int `mapstructure:"DB_MAX_CONNS"` // default: 25
    MinConns int `mapstructure:"DB_MIN_CONNS"` // default: 5
    MaxLifetime int `mapstructure:"DB_MAX_LIFETIME"` // minutes
}

type RedisConfig struct {
    URL      string `mapstructure:"REDIS_URL"`
    MaxRetries int  `mapstructure:"REDIS_MAX_RETRIES"`
}

type KeycloakConfig struct {
    Issuer        string `mapstructure:"KEYCLOAK_ISSUER"`
    ClientID      string `mapstructure:"KEYCLOAK_CLIENT_ID"`
    ClientSecret  string `mapstructure:"KEYCLOAK_CLIENT_SECRET"`
}

type SentryConfig struct {
    DSN string `mapstructure:"SENTRY_DSN"`
}

type LogConfig struct {
    Level string `mapstructure:"LOG_LEVEL"` // info, debug, warn, error
    Format string `mapstructure:"LOG_FORMAT"` // json (production), text (dev)
}
```

Config loading:
1. `viper.AddConfigPath(".")` — `.env` файл
2. `viper.AutomaticEnv()` — ENV variables override
3. `viper.SetDefault()` — default values
4. `viper.Unmarshal(&cfg)` → validation (required fields)

### `app/wire.go`

```go
package app

// Provide — инициализация всех зависимостей
func Provide(cfg Config) (*Server, func(), error) {
    // 1. Logger (structured JSON)
    logger := newLogger(cfg.Log)

    // 2. DB pool (pgx)
    dbPool, err := newDBPool(cfg.Database, logger)
    if err != nil {
        return nil, nil, fmt.Errorf("db pool: %w", err)
    }

    // 3. Redis client
    redisClient, err := newRedisClient(cfg.Redis, logger)
    if err != nil {
        dbPool.Close()
        return nil, nil, fmt.Errorf("redis: %w", err)
    }

    // 4. Sentry (optional)
    if cfg.Sentry.DSN != "" {
        initSentry(cfg.Sentry.DSN)
    }

    // 5. Prometheus registry
    reg := prometheus.NewRegistry()
    registerMetrics(reg)

    // 6. OIDC verifier
    oidcVerifier, err := newOIDCVerifier(cfg.Keycloak)
    if err != nil {
        dbPool.Close()
        redisClient.Close()
        return nil, nil, fmt.Errorf("oidc: %w", err)
    }

    // 7. Server (router + middleware + handlers)
    srv := NewServer(cfg.Server, dbPool, redisClient, oidcVerifier, logger, reg)

    // Cleanup function
    cleanup := func() {
        dbPool.Close()
        redisClient.Close()
    }

    return srv, cleanup, nil
}
```

### `app/server.go`

```go
package app

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
    "github.com/redis/go-redis/v9"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/coreos/go-oidc/v3/oidc"
    "github.com/prometheus/client_golang/prometheus"
)

type Server struct {
    httpServer *http.Server
    router     *chi.Mux
    config     ServerConfig
    logger     Logger
}

func NewServer(cfg ServerConfig, db *pgxpool.Pool, redis *redis.Client,
    verifier *oidc.IDTokenVerifier, logger Logger, reg *prometheus.Registry) *Server {

    r := chi.NewRouter()

    // Middleware chain (order matters!)
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.Timeout(30 * time.Second))
    r.Use(PrometheusMiddleware(reg))

    // CORS
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"*"}, // Production: конкретные домены из config
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
        ExposedHeaders:   []string{"Link"},
        MaxAge:           300,
    }))

    // Health
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    // Metrics
    r.Mount("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

    s := &Server{
        router: r,
        config: cfg,
        logger: logger,
    }

    s.httpServer = &http.Server{
        Addr:         fmt.Sprintf(":%d", cfg.Port),
        Handler:      r,
        ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
        WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    return s
}

func (s *Server) Start() error {
    s.logger.Info("server starting", "port", s.config.Port)
    return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("server shutting down")
    return s.httpServer.Shutdown(ctx)
}
```

## Требования

- DI без google/wire (ручная инициализация, проще для 17 пакетов)
- Config validation — required fields проверяются при запуске
- DB pool — `pgxpool` с connection limits
- Redis — `go-redis/v9` с retry
- Logger — structured JSON (production), text (dev)
- Prometheus — custom registry (не default, чтобы избежать conflicts)
- CORS — configurable (dev: `*`, prod: конкретные домены)
- Middleware chain — order matters (RequestID → RealIP → Logger → Recoverer → Timeout → Prometheus)

## Критерии приёмки

- [ ] `app/config.go` — Config struct с viper loading + validation
- [ ] `app/wire.go` — Provide() функция, инициализация всех зависимостей
- [ ] `app/server.go` — chi router + middleware chain + healthz + metrics
- [ ] Config loading: `.env` → ENV override → defaults
- [ ] DB pool подключается к PostgreSQL
- [ ] Redis client подключается к Redis
- [ ] OIDC verifier инициализируется из Keycloak issuer
- [ ] `/healthz` → 200 OK
- [ ] `/metrics` → Prometheus metrics endpoint
