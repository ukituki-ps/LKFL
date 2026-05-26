package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server — HTTP-сервер приложения.
type Server struct {
	httpServer *http.Server
	router     *chi.Mux
	config     ServerConfig
	logger     Logger
}

// NewServer создаёт HTTP-сервер с chi router и цепочкой middleware.
//
// Middleware chain (порядок важен!):
//  1. RequestID — уникальный ID запроса
//  2. RealIP — определение реального IP клиента
//  3. Logger — логирование запросов
//  4. Recoverer — обработка паник
//  5. Timeout — ограничение времени выполнения
//  6. Prometheus — сбор метрик
func NewServer(
	cfg ServerConfig,
	db *pgxpool.Pool,
	redis *redis.Client,
	verifier *oidc.IDTokenVerifier,
	logger Logger,
	reg *prometheus.Registry,
) *Server {
	r := chi.NewRouter()

	// Middleware chain (order matters!)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(PrometheusMiddleware(reg))

	// CORS — dev: *, prod: конкретные домены из config
	r.Use(corsMiddleware())

	// Health check
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Prometheus metrics
	r.Mount("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	s := &Server{
		router: r,
		config: cfg,
		logger: logger,
	}

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start запускает HTTP-сервер.
func (s *Server) Start() error {
	s.logger.Info("server starting", "port", s.config.Port)
	return s.httpServer.ListenAndServe()
}

// Shutdown выполняет graceful shutdown сервера.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}

// corsMiddleware создаёт CORS middleware.
//
// В development разрешены все origin, в production — конкретные домены.
func corsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Dev: разрешаем все origin.
			// TODO: в production — список доменов из config.
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Tenant-ID")
			w.Header().Set("Access-Control-Expose-Headers", "Link")
			w.Header().Set("Access-Control-Max-Age", "300")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// PrometheusMiddleware — middleware для сбора метрик Prometheus.
func PrometheusMiddleware(reg *prometheus.Registry) func(http.Handler) http.Handler {
	httpReqsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "status"},
	)

	httpReqDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	reg.MustRegister(httpReqsTotal, httpReqDuration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			httpReqsTotal.WithLabelValues(r.Method, fmt.Sprintf("%d", ww.Status())).Inc()
			// Route pattern — URL path без query params
		pattern := r.URL.Path
		httpReqDuration.WithLabelValues(r.Method, pattern).Observe(time.Since(start).Seconds())
		})
	}
}
