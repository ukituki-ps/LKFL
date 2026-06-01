// Package auth — TokenRefresher: server-side refresh Keycloak токенов.
//
// При истечении access_token middleware вызывает TokenRefresher.Refresh(),
// который обменивает refresh_token на новые токены через Keycloak token endpoint.
//
// Поток:
//  1. POST /protocol/openid-connect/token с grant_type=refresh_token
//  2. Извлечь новый access_token + refresh_token из ответа
//  3. Сохранить в TokenStore
//  4. Вернуть decoded claims из нового access_token
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// TokenRefresher — обменивает refresh_token на новые токены.
type TokenRefresher struct {
	issuer       string
	clientID     string
	clientSecret string
	tokenStore   *TokenStore
}

// NewTokenRefresher создаёт TokenRefresher.
func NewTokenRefresher(issuer, clientID, clientSecret string, ts *TokenStore) *TokenRefresher {
	return &TokenRefresher{
		issuer:       issuer,
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenStore:   ts,
	}
}

// Refresh обменивает refresh_token на новые токены.
//
// Возвращает Claims из нового access_token и обновлённый refresh_token.
// Если refresh_token истёк или невалиден — возвращает ошибку.
func (r *TokenRefresher) Refresh(ctx context.Context, userID string) (accessToken string, refreshToken string, err error) {
	currentRefresh, err := r.tokenStore.GetRefreshToken(ctx, userID)
	if err != nil {
		return "", "", fmt.Errorf("get refresh_token: %w", err)
	}

	tokenEndpoint := r.issuer + "/protocol/openid-connect/token"
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", r.clientID)
	form.Set("refresh_token", currentRefresh)
	if r.clientSecret != "" {
		form.Set("client_secret", r.clientSecret)
	}

	resp, err := http.PostForm(tokenEndpoint, form)
	if err != nil {
		return "", "", fmt.Errorf("http POST to token endpoint: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("token refresh failed", "status", resp.StatusCode, "body", string(body))
		return "", "", fmt.Errorf("token refresh failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenSet struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &tokenSet); err != nil {
		return "", "", fmt.Errorf("parse token response: %w", err)
	}

	// Сохраняем новые токены
	if tokenSet.AccessToken != "" {
		_ = r.tokenStore.UpdateAccessToken(ctx, userID, tokenSet.AccessToken)
	}
	if tokenSet.RefreshToken != "" {
		_ = r.tokenStore.UpdateRefreshToken(ctx, userID, tokenSet.RefreshToken)
	}

	return tokenSet.AccessToken, tokenSet.RefreshToken, nil
}
