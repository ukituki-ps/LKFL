// Package app — dependency injection wiring.
//
// Единая точка инициализации всех зависимостей приложения:
// config → infrastructure (DB, Redis) → business → handlers → router → server.
package app

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config — корневая конфигурация приложения.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Keycloak KeycloakConfig
	Sentry   SentryConfig
	Log      LogConfig
}

// ServerConfig — настройки HTTP-сервера.
type ServerConfig struct {
	Port         int `mapstructure:"SERVER_PORT"`
	ReadTimeout  int `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout int `mapstructure:"SERVER_WRITE_TIMEOUT"`
}

// DatabaseConfig — настройки PostgreSQL.
type DatabaseConfig struct {
	DSN         string `mapstructure:"DB_DSN"`
	MaxConns    int    `mapstructure:"DB_MAX_CONNS"`
	MinConns    int    `mapstructure:"DB_MIN_CONNS"`
	MaxLifetime int    `mapstructure:"DB_MAX_LIFETIME"` // minutes
}

// RedisConfig — настройки Redis.
type RedisConfig struct {
	URL        string `mapstructure:"REDIS_URL"`
	MaxRetries int    `mapstructure:"REDIS_MAX_RETRIES"`
}

// KeycloakConfig — настройки OIDC/Keycloak.
type KeycloakConfig struct {
	Issuer       string `mapstructure:"KEYCLOAK_ISSUER"`
	ClientID     string `mapstructure:"KEYCLOAK_CLIENT_ID"`
	ClientSecret string `mapstructure:"KEYCLOAK_CLIENT_SECRET"`
}

// SentryConfig — настройки Sentry.
type SentryConfig struct {
	DSN string `mapstructure:"SENTRY_DSN"`
}

// LogConfig — настройки логирования.
type LogConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`  // info, debug, warn, error
	Format string `mapstructure:"LOG_FORMAT"` // json (production), text (dev)
}

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

	// 3. Переменные окружения (перезаписывают всё предыдущее)
	v.AutomaticEnv()

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
}

// validate проверяет наличие обязательных полей конфигурации.
func validate(cfg Config) error {
	var missing []string

	if cfg.Database.DSN == "" {
		missing = append(missing, "DB_DSN")
	}
	if cfg.Redis.URL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if cfg.Keycloak.Issuer == "" {
		missing = append(missing, "KEYCLOAK_ISSUER")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %v", missing)
	}

	return nil
}

// IsDevelopment возвращает true, если приложение запущено в dev-режиме.
func (c Config) IsDevelopment() bool {
	return c.Log.Level == "debug"
}


