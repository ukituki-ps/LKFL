// Package app — dependency injection wiring.
//
// Единая точка инициализации всех зависимостей приложения:
// config → infrastructure (DB, Redis) → business → handlers → router → server.
package app

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config — корневая конфигурация приложения.
// Плоская структура: Viper + AutomaticEnv + mapstructure работают
// корректно только с плоскими ключами (не вложенными).
type Config struct {
	// Server
	ServerPort         int    `mapstructure:"SERVER_PORT"`
	ServerReadTimeout  int    `mapstructure:"SERVER_READ_TIMEOUT"`
	ServerWriteTimeout int    `mapstructure:"SERVER_WRITE_TIMEOUT"`

	// Database
	DBDSN         string `mapstructure:"DB_DSN"`
	DBMaxConns    int    `mapstructure:"DB_MAX_CONNS"`
	DBMinConns    int    `mapstructure:"DB_MIN_CONNS"`
	DBMaxLifetime int    `mapstructure:"DB_MAX_LIFETIME"` // minutes

	// Redis
	RedisURL        string `mapstructure:"REDIS_URL"`
	RedisMaxRetries int    `mapstructure:"REDIS_MAX_RETRIES"`

	// Keycloak
	KeycloakIssuer       string `mapstructure:"KEYCLOAK_ISSUER"`
	KeycloakClientID     string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	KeycloakClientSecret string `mapstructure:"KEYCLOAK_CLIENT_SECRET"`
	// KeycloakPublicURL — URL для browser-редиректов (видимый из браузера).
	// Если пуст — используется KeycloakIssuer.
	KeycloakPublicURL string `mapstructure:"KEYCLOAK_PUBLIC_URL"`

	// Sentry
	SentryDSN string `mapstructure:"SENTRY_DSN"`

	// Log
	LogLevel  string `mapstructure:"LOG_LEVEL"`  // info, debug, warn, error
	LogFormat string `mapstructure:"LOG_FORMAT"` // json (production), text (dev)

	// CORS
	CORSAllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"` // comma-separated origins
	CORSMaxAge         int    `mapstructure:"CORS_MAX_AGE"`         // seconds

	// Security
	RateLimitAuth    int `mapstructure:"RATE_LIMIT_AUTH"`
	RateLimitCatalog int `mapstructure:"RATE_LIMIT_CATALOG"`
	RateLimitAdmin   int `mapstructure:"RATE_LIMIT_ADMIN"`
}

// Convenience methods for backward compatibility

// DatabaseDSN returns the database DSN.
func (c Config) DatabaseDSN() string { return c.DBDSN }

// RedisURL returns the Redis URL.
func (c Config) RedisURLGet() string { return c.RedisURL }

// KeycloakIssuer returns the Keycloak issuer URL.
func (c Config) KeycloakIssuerGet() string { return c.KeycloakIssuer }

// LoadConfig загружает конфигурацию из файлов и переменных окружения.
//
// Приоритет (от низшего к высшему):
//  1. Значения по умолчанию (SetDefault)
//  2. .env файл в рабочей директории
//  3. Переменные окружения (ENV)
func LoadConfig() (Config, error) {
	v := viper.New()

	// 1. Значения по умолчанию
	setDefaults(v)

	// 2. .env файл
	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")

	// Игнорируем ошибку отсутствия .env — он необязателен
	_ = v.ReadInConfig()

	// 3. Переменные окружения — читаем напрямую из os.Getenv
	// Viper AutomaticEnv не работает с mapstructure Unmarshal для вложенных структур.
	// Читаем каждый ключ напрямую и устанавливаем в Viper.
	envKeys := []string{
		"DB_DSN", "DB_MAX_CONNS", "DB_MIN_CONNS", "DB_MAX_LIFETIME",
		"REDIS_URL", "REDIS_MAX_RETRIES",
		"KEYCLOAK_ISSUER", "KEYCLOAK_CLIENT_ID", "KEYCLOAK_CLIENT_SECRET", "KEYCLOAK_PUBLIC_URL",
		"SERVER_PORT", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
		"SENTRY_DSN",
		"LOG_LEVEL", "LOG_FORMAT",
		"CORS_ALLOWED_ORIGINS", "CORS_MAX_AGE",
		"RATE_LIMIT_AUTH", "RATE_LIMIT_CATALOG", "RATE_LIMIT_ADMIN",
	}
	for _, key := range envKeys {
		if val := os.Getenv(key); val != "" {
			v.Set(key, val)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// setDefaults задаёт значения по умолчанию для всех полей конфигурации.
func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVER_READ_TIMEOUT", 30)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 30)

	// Database
	v.SetDefault("DB_MAX_CONNS", 25)
	v.SetDefault("DB_MIN_CONNS", 5)
	v.SetDefault("DB_MAX_LIFETIME", 60)

	// Redis
	v.SetDefault("REDIS_MAX_RETRIES", 3)

	// Log
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")

	// CORS
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	v.SetDefault("CORS_MAX_AGE", 3600)

	// Security — rate limiting
	v.SetDefault("RATE_LIMIT_AUTH", 10)
	v.SetDefault("RATE_LIMIT_CATALOG", 100)
	v.SetDefault("RATE_LIMIT_ADMIN", 60)
}

// validate проверяет наличие обязательных полей конфигурации.
func validate(cfg Config) error {
	var missing []string

	if cfg.DBDSN == "" {
		missing = append(missing, "DB_DSN")
	}
	if cfg.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if cfg.KeycloakIssuer == "" {
		missing = append(missing, "KEYCLOAK_ISSUER")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %v", missing)
	}

	return nil
}

// IsDevelopment возвращает true, если приложение запущено в dev-режиме.
func (c Config) IsDevelopment() bool {
	return c.LogLevel == "debug"
}
