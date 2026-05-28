// Package middleware — HTTP middleware для сервера.
//
// Содержит rate limiting middleware на основе Redis sliding window.
package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimitConfig — настройка rate limiting.
type RateLimitConfig struct {
	// MaxRequests — максимальное количество запросов за окно.
	MaxRequests int
	// Window — длительность окна в секундах (по умолчанию 60).
	Window int
	// RedisClient — клиент Redis для хранения счётчиков.
	RedisClient *redis.Client
	// KeyPrefix — префикс ключа в Redis (напр. "rl:auth").
	KeyPrefix string
}

// RateLimiter — middleware для rate limiting через Redis sliding window.
//
// Алгоритм: Redis Sorted Set, где score = timestamp (ms), member = уникальный ID запроса.
// При каждом запросе удаляем старые записи вне окна и считаем текущие.
// Если количество превышает MaxRequests — возвращаем 429 Too Many Requests.
//
// Fail-open: если Redis недоступен, запрос пропускается.
func RateLimiter(cfg RateLimitConfig) func(http.Handler) http.Handler {
	if cfg.Window <= 0 {
		cfg.Window = 60
	}
	windowMillis := int64(cfg.Window * 1000)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			key := cfg.KeyPrefix + ":" + ip

			ctx := r.Context()
			now := time.Now().UnixMilli()
			windowStart := now - windowMillis

			// Sliding window: удалить старые записи и посчитать текущие.
			pipe := cfg.RedisClient.Pipeline()
			pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))
			member := strconv.FormatInt(now, 10) + ":" + ip
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: member})
			pipe.ZCard(ctx, key)
			pipe.Expire(ctx, key, time.Duration(cfg.Window+1)*time.Second)
			results, err := pipe.Exec(ctx)
			if err != nil {
				// Fail-open: если Redis недоступен, пропускаем запрос.
				next.ServeHTTP(w, r)
				return
			}

			countCmd, ok := results[2].(*redis.IntCmd)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}
			count := countCmd.Val()

			if int(count) > cfg.MaxRequests {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", strconv.Itoa(cfg.Window))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = fmt.Fprintf(w, `{"error":"rate limit exceeded"}`)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractIP извлекает IP-адрес клиента из запроса.
// Проверяет X-Forwarded-For (первый IP), затем X-Real-IP, затем RemoteAddr.
func extractIP(r *http.Request) string {
	// X-Forwarded-For: клиент, прокси1, прокси2
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Берём первый IP (клиент)
		idx := indexOf(xff, ",")
		if idx > 0 {
			return xff[:idx]
		}
		return xff
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// RemoteAddr может быть "IP:port"
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// indexOf — индекс первого вхождения sep в s (стандартная lib не экспортирует strconv).
func indexOf(s, sep string) int {
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			return i
		}
	}
	return -1
}
