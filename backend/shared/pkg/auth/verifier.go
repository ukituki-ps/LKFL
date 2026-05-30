// Package auth — OIDC-верификация, JWT-мидлвэр и RBAC для LKFL.
//
// Пакет предоставляет общую функциональность аутентификации:
//   - OIDC verifier (go-oidc) для Keycloak
//   - Claims extraction из ID Token (включая Keycloak roles)
//   - JWT middleware для chi-роутера
//   - RBAC middleware (проверка ролей из context)
//
// Используется монолитом (lkfl-server) и потенциально другими компонентами.
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/coreos/go-oidc" // v2.3.0+incompatible
)

// VerifierOption — опция для конфигурации NewVerifier.
type VerifierOption func(*verifierConfig)

type verifierConfig struct {
	maxRetries int
	retryDelay time.Duration
}

func defaultConfig() verifierConfig {
	return verifierConfig{
		maxRetries: 30,
		retryDelay: 2 * time.Second,
	}
}

// WithMaxRetries задаёт максимальное количество попыток подключения к Keycloak.
func WithMaxRetries(n int) VerifierOption {
	return func(c *verifierConfig) {
		if n > 0 {
			c.maxRetries = n
		}
	}
}

// WithRetryDelay задаёт задержку между попытками подключения к Keycloak.
func WithRetryDelay(d time.Duration) VerifierOption {
	return func(c *verifierConfig) {
		if d > 0 {
			c.retryDelay = d
		}
	}
}

// NewVerifier создаёт OIDC verifier из Keycloak issuer URL.
//
// Параметр issuerURL — URL issuer'а Keycloak realm, например:
//
//	http://keycloak:8080/realms/lkfl-sdek  (staging, внутренний URL)
//	https://keycloak.example.com/realms/lkfl (production)
//
// clientID — идентификатор OIDC-клиента (например, lkfl-spa).
//
// Возвращает *oidc.IDTokenVerifier, готовую к верификации ID токенов.
// При ошибке подключения к Keycloak пытается повторно (по умолчанию 30 раз,
// каждые 2 сек). Конфигурируется через WithMaxRetries и WithRetryDelay.
//
// ADR-037: issuer всегда = внутренний URL Keycloak. Никаких хак-обёрток
// для TLS — внутри Docker-сети HTTP, TLS termination только на границе сети.
func NewVerifier(ctx context.Context, issuerURL, clientID string, opts ...VerifierOption) (*oidc.IDTokenVerifier, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	var provider *oidc.Provider
	var err error

	for i := 0; i < cfg.maxRetries; i++ {
		provider, err = oidc.NewProvider(ctx, issuerURL)
		if err == nil {
			break
		}
		slog.Warn("oidc provider not ready, retrying", "attempt", i+1, "max_retries", cfg.maxRetries, "error", err)
		time.Sleep(cfg.retryDelay)
	}

	if err != nil {
		return nil, fmt.Errorf("oidc provider (after %d retries): %w", cfg.maxRetries, err)
	}

	return provider.Verifier(&oidc.Config{
		ClientID: clientID,
	}), nil
}
