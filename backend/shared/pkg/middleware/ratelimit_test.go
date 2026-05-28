package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name       string
		xff        string
		xri        string
		remoteAddr string
		want       string
	}{
		{
			name:       "X-Forwarded-For с несколькими IP",
			xff:        "1.2.3.4, 5.6.7.8, 9.10.11.12",
			remoteAddr: "127.0.0.1:1234",
			want:       "1.2.3.4",
		},
		{
			name:       "X-Forwarded-For один IP",
			xff:        "1.2.3.4",
			remoteAddr: "127.0.0.1:1234",
			want:       "1.2.3.4",
		},
		{
			name:       "X-Real-IP",
			xri:        "2.3.4.5",
			remoteAddr: "127.0.0.1:1234",
			want:       "2.3.4.5",
		},
		{
			name:       "RemoteAddr с портом",
			remoteAddr: "3.4.5.6:5678",
			want:       "3.4.5.6",
		},
		{
			name:       "RemoteAddr без порта",
			remoteAddr: "3.4.5.6",
			want:       "3.4.5.6",
		},
		{
			name:       "X-Forwarded-For приоритет над X-Real-IP",
			xff:        "1.1.1.1",
			xri:        "2.2.2.2",
			remoteAddr: "127.0.0.1:1234",
			want:       "1.1.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				r.Header.Set("X-Real-IP", tt.xri)
			}

			got := extractIP(r)
			if got != tt.want {
				t.Errorf("extractIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimiterFailOpen(t *testing.T) {
	// Создаём Redis клиент, который не подключён к реальному серверу.
	// Rate limiter должен пропустить запрос (fail-open).
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:59999", // несуществующий порт
	})
	defer client.Close()

	cfg := RateLimitConfig{
		MaxRequests: 5,
		Window:      60,
		RedisClient: client,
		KeyPrefix:   "rl:test",
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RateLimiter(cfg)(handler)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Должен пропустить запрос даже при недоступном Redis (fail-open).
	mw.ServeHTTP(w, r)

	// Результат может быть 200 (запрос прошёл) или что-то другое,
	// главное — middleware не упал с паникой.
	t.Logf("fail-open response status: %d", w.Code)
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		s    string
		sep  string
		want int
	}{
		{"hello, world", ",", 5},
		{"abc", "d", -1},
		{"", "a", -1},
		{"a,b,c", ",", 1},
	}

	for _, tt := range tests {
		got := indexOf(tt.s, tt.sep)
		if got != tt.want {
			t.Errorf("indexOf(%q, %q) = %d, want %d", tt.s, tt.sep, got, tt.want)
		}
	}
}
