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

// NewVerifier создаёт OIDC verifier из Keycloak issuer URL.
//
// Параметр issuerURL — URL issuer'а Keycloak realm, например:
//
//	https://keycloak.example.com/realms/lkfl
//
// clientID — идентификатор OIDC-клиента (например, lkfl-spa).
//
// Возвращает *oidc.IDTokenVerifier, готовую к верификации ID токенов.
// При ошибке подключения к Keycloak пытается повторно 10 раз с интервалом 2 сек.
func NewVerifier(ctx context.Context, issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
	var provider *oidc.Provider
	var err error

	for i := 0; i < 30; i++ {
		provider, err = oidc.NewProvider(ctx, issuerURL)
		if err == nil {
			break
		}
		slog.Warn("oidc provider not ready, retrying", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("oidc provider (after 10 retries): %w", err)
	}

	return provider.Verifier(&oidc.Config{
		ClientID: clientID,
	}), nil
}
