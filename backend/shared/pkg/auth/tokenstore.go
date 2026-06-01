// Package auth — TokenStore: хранение Keycloak токенов в Redis.
//
// Хранит access_token и refresh_token для server-side silent refresh
// и инвалидации Keycloak SSO сессии при logout.
//
// Redis keys:
//   - lkfl:kc:token:{user_sub} → access_token (TTL 15 мин)
//   - lkfl:kc:refresh:{user_sub} → refresh_token (TTL 7 дней)
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore — хранилище Keycloak токенов.
type TokenStore struct {
	redis redisClient
}

// NewTokenStore создаёт TokenStore.
func NewTokenStore(client redisClient) *TokenStore {
	return &TokenStore{redis: client}
}

// SaveTokens сохраняет access_token и refresh_token в Redis.
func (t *TokenStore) SaveTokens(ctx context.Context, userID string, accessToken, refreshToken string) error {
	tokenKey := tokenKey(userID)
	refreshKey := refreshKey(userID)

	// access_token — TTL 15 минут
	if err := t.redis.Set(ctx, tokenKey, accessToken, 15*time.Minute).Err(); err != nil {
		return fmt.Errorf("redis set access_token: %w", err)
	}

	// refresh_token — TTL 7 дней
	if err := t.redis.Set(ctx, refreshKey, refreshToken, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("redis set refresh_token: %w", err)
	}

	return nil
}

// GetAccessToken возвращает access_token для пользователя.
func (t *TokenStore) GetAccessToken(ctx context.Context, userID string) (string, error) {
	tokenKey := tokenKey(userID)
	token, err := t.redis.Get(ctx, tokenKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrSessionNotFound
		}
		return "", fmt.Errorf("redis get access_token: %w", err)
	}
	return token, nil
}

// GetRefreshToken возвращает refresh_token для пользователя.
func (t *TokenStore) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	refreshKey := refreshKey(userID)
	token, err := t.redis.Get(ctx, refreshKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrSessionNotFound
		}
		return "", fmt.Errorf("redis get refresh_token: %w", err)
	}
	return token, nil
}

// Delete удаляет оба токена пользователя.
func (t *TokenStore) Delete(ctx context.Context, userID string) error {
	tokenKey := tokenKey(userID)
	refreshKey := refreshKey(userID)

	_, err := t.redis.Del(ctx, tokenKey, refreshKey).Result()
	if err != nil {
		return fmt.Errorf("redis delete tokens: %w", err)
	}
	return nil
}

// UpdateAccessToken обновляет только access_token (при refresh).
func (t *TokenStore) UpdateAccessToken(ctx context.Context, userID string, accessToken string) error {
	tokenKey := tokenKey(userID)
	if err := t.redis.Set(ctx, tokenKey, accessToken, 15*time.Minute).Err(); err != nil {
		return fmt.Errorf("redis update access_token: %w", err)
	}
	return nil
}

// UpdateRefreshToken обновляет только refresh_token (при refresh).
func (t *TokenStore) UpdateRefreshToken(ctx context.Context, userID string, refreshToken string) error {
	refreshKey := refreshKey(userID)
	if err := t.redis.Set(ctx, refreshKey, refreshToken, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("redis update refresh_token: %w", err)
	}
	return nil
}

func tokenKey(userID string) string {
	return fmt.Sprintf("lkfl:kc:token:%s", userID)
}

func refreshKey(userID string) string {
	return fmt.Sprintf("lkfl:kc:refresh:%s", userID)
}
