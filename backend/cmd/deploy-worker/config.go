package main

import (
	"os"
	"strconv"
)

// Config — конфигурация deploy-worker из переменных окружения.
type Config struct {
	Port          int
	WebhookSecret string
	GHCRToken     string
	GHCRUsername  string
	ComposeFile   string
	ComposeDir    string
}

// loadConfig загружает конфигурацию из переменных окружения.
func loadConfig() Config {
	port := 9092
	if p := os.Getenv("PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	return Config{
		Port:          port,
		WebhookSecret: os.Getenv("WEBHOOK_SECRET"),
		GHCRToken:     os.Getenv("GHCR_TOKEN"),
		GHCRUsername:  getEnvOrDefault("GHCR_USERNAME", "ukituki"),
		ComposeFile:   getEnvOrDefault("COMPOSE_FILE", "docker-compose.staging.yml"),
		ComposeDir:    getEnvOrDefault("COMPOSE_DIR", "."),
	}
}

// getEnvOrDefault возвращает значение переменной окружения или defaultVal, если пустая.
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
