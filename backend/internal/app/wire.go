package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// Logger — интерфейс логгера, используемый во всём приложении.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Provide — инициализация всех зависимостей приложения.
//
// Порядок инициализации:
//  1. Logger
//  2. DB pool (pgx)
//  3. Redis client
//  4. Sentry (опционально)
//  5. Prometheus registry
//  6. OIDC verifier
//  7. Server (router + middleware + handlers)
//
// Возвращает Server и cleanup-функцию для graceful shutdown.
// При ошибке на любом шаге все инициализированные ресурсы освобождаются.
func Provide(cfg Config) (*Server, func(), error) {
	// 1. Logger
	logger := newLogger(cfg.Log)

	// 2. DB pool
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

	// 4. Sentry (опционально)
	if cfg.Sentry.DSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:           cfg.Sentry.DSN,
			EnableTracing: true,
		}); err != nil {
			logger.Warn("sentry init failed", "error", err)
		}
	}

	// 5. Prometheus registry
	reg := prometheus.NewRegistry()

	// 6. OIDC verifier
	verifier, err := newOIDCVerifier(cfg.Keycloak)
	if err != nil {
		dbPool.Close()
		_ = redisClient.Close()
		return nil, nil, fmt.Errorf("oidc: %w", err)
	}

	// 7. Server
	srv := NewServer(cfg.Server, dbPool, redisClient, verifier, logger, reg)

	// Cleanup
	cleanup := func() {
		dbPool.Close()
		_ = redisClient.Close()
		if cfg.Sentry.DSN != "" {
			sentry.Flush(2 * time.Second)
		}
	}

	return srv, cleanup, nil
}

// newLogger создаёт структурированный логгер.
//
// Формат зависит от конфигурации:
//   - json (по умолчанию) — для production
//   - text — для development
func newLogger(cfg LogConfig) *slog.Logger {
	level := parseLogLevel(cfg.Level)

	if cfg.Format == "text" {
		return slog.New(slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: level},
		))
	}

	return slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: level},
	))
}

// parseLogLevel преобразует строковый уровень в slog.Level.
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// newDBPool создаёт пул подключений к PostgreSQL.
func newDBPool(cfg DatabaseConfig, logger Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = rint32(cfg.MaxConns)
	poolCfg.MinConns = rint32(cfg.MinConns)
	poolCfg.MaxConnLifetime = time.Duration(cfg.MaxLifetime) * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	logger.Info("db pool connected",
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
		"max_lifetime_min", cfg.MaxLifetime,
	)

	return pool, nil
}

// newRedisClient создаёт клиент Redis с настройками retry.
func newRedisClient(cfg RedisConfig, logger Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	opt.MaxRetries = cfg.MaxRetries
	opt.MaxRetryBackoff = 500 * time.Millisecond
	opt.MinRetryBackoff = 8 * time.Millisecond
	opt.PoolSize = 10
	opt.PoolTimeout = 30 * time.Second
	opt.ReadTimeout = 5 * time.Second
	opt.WriteTimeout = 5 * time.Second
	opt.DisableIndentity = false

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	logger.Info("redis connected", "url", cfg.URL)

	return client, nil
}

// newOIDCVerifier создаёт верификатор OIDC токенов из Keycloak.
func newOIDCVerifier(cfg KeycloakConfig) (*oidc.IDTokenVerifier, error) {
	_, verifier, err := newOIDCProvider(context.Background(), cfg.Issuer, cfg.ClientID)
	return verifier, err
}

// newOIDCProvider создаёт OIDC provider и verifier.
func newOIDCProvider(ctx context.Context, issuerURL, clientID string) (*oidc.Provider, *oidc.IDTokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, nil, fmt.Errorf("new oidc provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
	return provider, verifier, nil
}

// rint32 преобразует int в int32 для pgxpool.
func rint32(v int) int32 {
	return int32(v)
}
