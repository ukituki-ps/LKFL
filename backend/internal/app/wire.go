package app

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"

	intauth "lkfl/internal/auth"
	"lkfl/internal/metrics"
	"lkfl/internal/tenant"
	"lkfl/internal/user"
	"lkfl/shared/pkg/auth"
	"lkfl/shared/pkg/logger"
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
	logger := newLogger(cfg.LogLevel, cfg.LogFormat)

	// 2. DB pool
	dbPool, err := newDBPool(cfg.DBDSN, cfg.DBMaxConns, cfg.DBMinConns, cfg.DBMaxLifetime, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("db pool: %w", err)
	}

	// 3. Redis client
	redisClient, err := newRedisClient(cfg.RedisURL, cfg.RedisMaxRetries, logger)
	if err != nil {
		dbPool.Close()
		return nil, nil, fmt.Errorf("redis: %w", err)
	}

	// 4. Sentry (опционально)
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:           cfg.SentryDSN,
			EnableTracing: true,
		}); err != nil {
			logger.Warn("sentry init failed", "error", err)
		}
	}

	// 5. Prometheus registry + custom metrics
	reg := prometheus.NewRegistry()
	appMetrics := metrics.New(reg)

	// 6. OIDC verifier
	verifier, err := auth.NewVerifier(context.Background(), cfg.KeycloakIssuer, cfg.KeycloakClientID)
	if err != nil {
		dbPool.Close()
		_ = redisClient.Close()
		return nil, nil, fmt.Errorf("oidc: %w", err)
	}

	// 6.5. Tenant service (system module — создается до бизнес-модулей)
	tenantRepo := tenant.NewRepository(dbPool)
	tenantService := tenant.NewService(tenantRepo)

	// 6.6. User repository + service
	userRepo := user.NewRepository(dbPool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// 6.7. Auth service + handler
	authService := intauth.NewService(userRepo)
	// Fallback tenant resolution для auth callback (без tenant middleware)
	tenantResolver := tenant.NewHostResolver(tenantService, redisClient, appMetrics)
	authService.WithTenantResolver(tenantResolver)
	// Извлекаем tenant slug из issuer URL: .../realms/lkfl-sdek → sdek
	parts := strings.Split(cfg.KeycloakIssuer, "/")
	for _, p := range parts {
		if len(p) > 5 && p[:5] == "lkfl-" {
			authService.SetDefaultTenantSlug(p[5:])
			break
		}
	}
	authHandler := intauth.NewHandler(verifier, redisClient, authService, cfg.KeycloakIssuer, cfg.KeycloakPublicURL, cfg.KeycloakClientID, cfg.KeycloakClientSecret, appMetrics)

	// 7. Server
	srv := NewServer(cfg, dbPool, redisClient, verifier, logger, reg, appMetrics, tenantService, authHandler, userHandler)

	// Cleanup
	cleanup := func() {
		dbPool.Close()
		_ = redisClient.Close()
		if cfg.SentryDSN != "" {
			sentry.Flush(2 * time.Second)
		}
	}

	return srv, cleanup, nil
}

// newLogger создаёт структурированный логгер.
//
// Формат зависит от конфигурации:
//   - json (по умолчанию) — для production, Loki-совместимый
//   - text — для development
//
// Каждый лог содержит атрибут svc (имя сервиса) для фильтрации
// в Grafana Explore / Loki.
func newLogger(level, format string) *slog.Logger {
	return logger.New(logger.Options{
		Level:   level,
		Format:  format,
		Service: "lkfl-server",
	})
}

// newDBPool создаёт пул подключений к PostgreSQL.
func newDBPool(dsn string, maxConns, minConns, maxLifetime int, logger Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	poolCfg.MaxConns = rint32(maxConns)
	poolCfg.MinConns = rint32(minConns)
	poolCfg.MaxConnLifetime = time.Duration(maxLifetime) * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	logger.Info("db pool connected",
		"max_conns", maxConns,
		"min_conns", minConns,
		"max_lifetime_min", maxLifetime,
	)

	return pool, nil
}

// newRedisClient создаёт клиент Redis с настройками retry.
func newRedisClient(url string, maxRetries int, logger Logger) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	opt.MaxRetries = maxRetries
	opt.MaxRetryBackoff = 500 * time.Millisecond
	opt.MinRetryBackoff = 8 * time.Millisecond
	opt.PoolSize = 10
	opt.PoolTimeout = 30 * time.Second
	opt.ReadTimeout = 5 * time.Second
	opt.WriteTimeout = 5 * time.Second
	opt.DisableIdentity = false

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	logger.Info("redis connected", "url", "****")

	return client, nil
}

// rint32 преобразует int в int32 для pgxpool.
func rint32(v int) int32 {
	return int32(v)
}
