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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc" // v2.3.0+incompatible
)

// newHTTPClient создаёт HTTP-клиент с поддержкой кастомных CA-сертификатов.
// Если SSL_CERT_FILE установлен — загружает cert и добавляет в root pool.
// Если TLS_INSECURE=true — отключает верификацию (только staging).
func newHTTPClient() *http.Client {
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}

	// Загрузить кастомный cert из SSL_CERT_FILE (для self-signed на staging)
	certFile := os.Getenv("SSL_CERT_FILE")
	if certFile != "" {
		data, err := os.ReadFile(certFile)
		if err == nil && pool.AppendCertsFromPEM(data) {
			slog.Info("loaded custom CA cert", "file", certFile)
		}
	}

	// Fallback: полное отключение верификации (только staging)
	if os.Getenv("TLS_INSECURE") == "true" {
		slog.Warn("TLS verification disabled (staging)")
		return &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}
}

// NewVerifier создаёт OIDC verifier из Keycloak issuer URL.
//
// Параметр issuerURL — URL issuer'а Keycloak realm, например:
//
//	https://keycloak.example.com/realms/lkfl
//
// clientID — идентификатор OIDC-клиента (например, lkfl-spa).
//
// Возвращает *oidc.IDTokenVerifier, готовую к верификации ID токенов.
// При ошибке подключения к Keycloak пытается повторно 30 раз с интервалом 2 сек.
func NewVerifier(ctx context.Context, issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
	var provider *oidc.Provider
	var err error

	// Установить кастомный HTTP клиент для discovery (go-oidc v2 использует http.DefaultClient)
	oldTransport := http.DefaultTransport
	http.DefaultTransport = newHTTPClient().Transport
	defer func() { http.DefaultTransport = oldTransport }()

	for i := 0; i < 30; i++ {
		provider, err = oidc.NewProvider(ctx, issuerURL)
		if err == nil {
			break
		}
		slog.Warn("oidc provider not ready, retrying", "attempt", i+1, "error", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("oidc provider (after 30 retries): %w", err)
	}

	return provider.Verifier(&oidc.Config{
		ClientID: clientID,
	}), nil
}
