// Package auth — SessionStore: хранение server-side сессий в Redis.
//
// Сессия — это случайный токен (32 байта hex), привязанный к Keycloak subject ID
// и tenant slug. Хранится в httpOnly cookie на стороне клиента, в Redis — на сервере.
//
// Redis key: lkfl:sess:{session_token} → JSON{user_id, tenant_id, created_at}
// TTL: 24 часа (настраиваемый).
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisClient — интерфейс Redis-операций, необходимых для SessionStore.
// Позволяет подменять Redis в unit-тестах.
type redisClient interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}

// ErrSessionNotFound — сессия не найдена или истёкла.
var ErrSessionNotFound = errors.New("session not found")

// SessionData — данные серверной сессии.
type SessionData struct {
	UserID    string    // Keycloak sub
	Email     string    // user email (needed for Keycloak Admin API lookup)
	TenantID  string    // resolved from issuer
	CreatedAt time.Time
}

// SessionStore — хранилище server-side сессий.
type SessionStore struct {
	redis redisClient
	ttl   time.Duration
}

// NewSessionStore создаёт SessionStore.
func NewSessionStore(client redisClient, ttl time.Duration) *SessionStore {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &SessionStore{redis: client, ttl: ttl}
}

// Create создаёт новую сессию и возвращает session token.
//
// Генерирует криптографически безопасный токен (32 байта → 64 hex).
// Сохраняет в Redis с указанным TTL.
func (s *SessionStore) Create(ctx context.Context, data SessionData) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal session data: %w", err)
	}

	key := sessionKey(token)
	if err := s.redis.Set(ctx, key, payload, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("redis set session: %w", err)
	}

	return token, nil
}

// Get возвращает данные сессии по token.
//
// Возвращает ErrSessionNotFound, если сессия не найдена или истекла.
func (s *SessionStore) Get(ctx context.Context, token string) (SessionData, error) {
	key := sessionKey(token)
	payload, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return SessionData{}, ErrSessionNotFound
		}
		return SessionData{}, fmt.Errorf("redis get session: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return SessionData{}, fmt.Errorf("unmarshal session data: %w", err)
	}

	return data, nil
}

// Delete удаляет сессию по token.
func (s *SessionStore) Delete(ctx context.Context, token string) error {
	key := sessionKey(token)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete session: %w", err)
	}
	return nil
}

// DeleteByUserID удаляет все сессии пользователя.
//
// Сканирует Redis по паттерну lkfl:sess:* и удаляет все сессии
// с совпадающим UserID. Используется при принудительном logout всех сессий.
func (s *SessionStore) DeleteByUserID(ctx context.Context, userID string) error {
	iter := 0
	cursor := uint64(0)
	for {
		keys, cursor, err := s.redis.Scan(ctx, cursor, "lkfl:sess:*", 100).Result()
		if err != nil {
			return fmt.Errorf("redis scan sessions: %w", err)
		}

		for _, key := range keys {
			payload, err := s.redis.Get(ctx, key).Result()
			if err != nil {
				continue
			}
			var data SessionData
			if err := json.Unmarshal([]byte(payload), &data); err != nil {
				continue
			}
			if data.UserID == userID {
				_ = s.redis.Del(ctx, key).Err()
			}
		}

		iter++
		if cursor == 0 || iter > 100 {
			break
		}
	}
	return nil
}

// generateSessionToken генерирует криптографически безопасный токен (32 байта → 64 hex).
func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// sessionKey формирует Redis key для сессии.
func sessionKey(token string) string {
	return fmt.Sprintf("lkfl:sess:%s", token)
}

// sessionCookieName — имя cookie для хранения session token.
const sessionCookieName = "lkfl_session"

// ExtractSessionCookie извлекает session token из httpOnly cookie запроса.
func ExtractSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// decodeAccessToken парсит access token Keycloak (JWT) и извлекает payload + roles.
//
// Используется только в тестах (не для продакшн-валидации — используйте ExtractClaims).
func decodeAccessToken(rawToken string) (map[string]interface{}, []string, error) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("decode JWT payload: %w", err)
	}

	var rawClaims map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &rawClaims); err != nil {
		return nil, nil, fmt.Errorf("parse JWT claims: %w", err)
	}

	roles := extractKeycloakRolesFromClaims(rawClaims)

	return rawClaims, roles, nil
}

// extractKeycloakRolesFromClaims извлекает Keycloak роли из raw claims.
// Поддерживает realm_access (access token) и resource_access (ID token).
func extractKeycloakRolesFromClaims(raw map[string]interface{}) []string {
	roles := []string{}
	
	// realm_access (access token): {"roles": ["admin", "employee"]}
	if ra, ok := raw["realm_access"].(map[string]interface{}); ok {
		if roleList, ok := ra["roles"].([]interface{}); ok {
			for _, r := range roleList {
				if role, ok := r.(string); ok {
					roles = append(roles, role)
				}
			}
		}
	}
	
	// resource_access (ID token): {"clientID": {"roles": ["admin"]}}
	if ra, ok := raw["resource_access"].(map[string]interface{}); ok {
		for _, v := range ra {
			if roleMap, ok := v.(map[string]interface{}); ok {
				if roleList, ok := roleMap["roles"].([]interface{}); ok {
					for _, r := range roleList {
						if role, ok := r.(string); ok {
							roles = append(roles, role)
						}
					}
				}
			}
		}
	}
	return roles
}

// isTokenExpired проверяет, истёк ли токен по exp claim.
func isTokenExpired(claims map[string]interface{}) bool {
	expRaw, ok := claims["exp"]
	if !ok {
		return false
	}
	exp, ok := expRaw.(float64)
	if !ok {
		return false
	}
	return time.Unix(int64(exp), 0).Before(time.Now())
}

// buildClaims строит Claims из raw claims и SessionData.
//
// Claims используется только в тестах (для проверки roundtrip).
func buildClaims(rawClaims map[string]interface{}, sd SessionData) Claims {
	claims := Claims{
		Subject: sd.UserID,
		TenantID: sd.TenantID,
	}
	if email, ok := rawClaims["email"].(string); ok {
		claims.Email = email
	}
	if username, ok := rawClaims["preferred_username"].(string); ok {
		claims.PreferredUsername = username
	}
	if name, ok := rawClaims["name"].(string); ok {
		claims.Name = name
	}
	if given, ok := rawClaims["given_name"].(string); ok {
		claims.GivenName = given
	}
	if family, ok := rawClaims["family_name"].(string); ok {
		claims.FamilyName = family
	}
	if iss, ok := rawClaims["iss"].(string); ok {
		claims.Issuer = iss
	}
	return claims
}

// decodeBase64URL декодирует base64url (без padding).
func decodeBase64URL(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	return base64.StdEncoding.DecodeString(s)
}

// clearSessionCookie удаляет session cookie.
func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}
