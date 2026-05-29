package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	chi "github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"

	intauth "lkfl/internal/auth"
	"lkfl/internal/engagement/catalog"
	"lkfl/internal/metrics"
	"lkfl/internal/tenant"
	"lkfl/internal/user"
	sharedauth "lkfl/shared/pkg/auth"
	"lkfl/shared/pkg/middleware"
)

// Server — HTTP-сервер приложения.
type Server struct {
	httpServer *http.Server
	router     *chi.Mux
	config     Config
	logger     Logger
}

// httpMetrics — Prometheus метрики HTTP-запросов.
type httpMetrics struct {
	total    *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

// newHTTPMetrics создаёт и регистрирует метрики в реестре.
// Вызывается один раз в NewServer — безопасно для повторных вызовов
// при условии, что reg не используется повторно.
func newHTTPMetrics(reg *prometheus.Registry) *httpMetrics {
	m := &httpMetrics{
		total: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "status", "tenant_id"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "tenant_id"},
		),
	}
	reg.MustRegister(m.total, m.duration)
	return m
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
//  7. CORS — cross-origin policy
//  8. Rate Limiting — на уровне route групп
func NewServer(
	cfg Config,
	db *pgxpool.Pool,
	redis *redis.Client,
	verifier *oidc.IDTokenVerifier,
	logger Logger,
	reg *prometheus.Registry,
	appMetrics *metrics.Metrics,
	tenantService *tenant.Service,
	authHandler *intauth.Handler,
	userHandler *user.Handler,
) *Server {
	r := chi.NewRouter()

	// Middleware chain (order matters!)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))

	// Prometheus метрики — регистрация один раз, вне middleware
	httpM := newHTTPMetrics(reg)
	r.Use(PrometheusMiddleware(httpM))

	// CORS — rs/cors (production: конкретные домены из config)
	c := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(cfg.CORSAllowedOrigins, ","),
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           cfg.CORSMaxAge,
	})
	r.Use(c.Handler)

	// ─── Rate limiters ───
	authLimiter := middleware.RateLimiter(middleware.RateLimitConfig{
		MaxRequests: cfg.RateLimitAuth,
		Window:      60,
		RedisClient: redis,
		KeyPrefix:   "rl:auth",
	})

	catalogLimiter := middleware.RateLimiter(middleware.RateLimitConfig{
		MaxRequests: cfg.RateLimitCatalog,
		Window:      60,
		RedisClient: redis,
		KeyPrefix:   "rl:catalog",
	})

	adminLimiter := middleware.RateLimiter(middleware.RateLimitConfig{
		MaxRequests: cfg.RateLimitAdmin,
		Window:      60,
		RedisClient: redis,
		KeyPrefix:   "rl:admin",
	})

	// ─── Public routes (без auth) ───
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	r.Mount("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	// ─── Auth routes (публичные, без JWT) ───
	// Важно: define ДО /api/v1/ чтобы chi не применил JWT middleware к /auth/callback.
	// В chi: Route("/api/v1/", ...) matches ALL paths starting with /api/v1/ including
	// /api/v1/auth/* — middleware applies even if no sub-route matches.
	// Решение: explicit exclude /auth/* от JWT middleware через chi.Middleware.
	r.Route("/api/v1/auth/", func(r chi.Router) {
		r.Use(authLimiter)
		r.Get("/login", authHandler.LoginRedirect)
		r.Get("/callback", authHandler.LoginCallback)
		r.Post("/logout", authHandler.Logout)
	})

// ─── Employee routes (JWT + tenant middleware) ───
	// Excludes /api/v1/auth/* from JWT middleware — auth routes are public.
	r.Route("/api/v1/", func(r chi.Router) {
		// Wrap JWT middleware to skip /auth/* paths.
		// chi Route("/api/v1/", ...) prefix matches /api/v1/auth/ too,
		// so middleware applies even though sub-routes don't match.
		jwtMiddleware := sharedauth.JWTMiddleware(verifier)
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if strings.HasPrefix(req.URL.Path, "/api/v1/auth/") {
					next.ServeHTTP(w, req)
					return
				}
				jwtMiddleware(next).ServeHTTP(w, req)
			})
		})
		r.Use(tenant.TenantMiddlewareWithService(tenantService, redis, appMetrics))

		// User profile
		r.Get("/users/me", userHandler.Me)
		r.Put("/users/me", userHandler.UpdateMe)

		// Auth me (профиль через auth endpoint)
		r.Get("/auth/me", authHandler.Me)

		// Catalog (M20) — каталог льгот/активностей
		catalogCache := catalog.NewCache(redis, appMetrics)
		catalogRepo := catalog.NewRepository(db)
		catalogService := catalog.NewService(catalogRepo, catalogCache)
		catalogHandler := catalog.NewHandler(catalogService, appMetrics)
		r.Route("/engagements", func(r chi.Router) {
			r.Use(catalogLimiter)
			r.Get("/categories", catalogHandler.Categories)
			r.Get("/", catalogHandler.List)
			r.Get("/{id}", catalogHandler.Get)
		})
	})

	// ─── Admin routes (JWT + RBAC + admin tenant middleware) ───
	r.Route("/admin/", func(r chi.Router) {
		r.Use(sharedauth.JWTMiddleware(verifier))
		r.Use(tenant.AdminTenantMiddleware(tenantService, redis, appMetrics))
		r.Use(adminLimiter)

		// Admin-only
		r.Group(func(r chi.Router) {
			r.Use(sharedauth.RBACMiddleware([]string{"admin"}))
			r.Route("/tenants", func(r chi.Router) {
				th := tenant.NewHandler(tenantService)
				r.Post("/", th.Create)
				r.Get("/", th.List)
				r.Get("/{id}", th.GetByID)
				r.Put("/{id}", th.Update)
				r.Delete("/{id}", th.Delete)
				r.Get("/{id}/brand", th.GetBrandConfig)
				r.Put("/{id}/brand", th.UpsertBrandConfig)
			})
		})

		// HR + Admin
		r.Group(func(r chi.Router) {
			r.Use(sharedauth.RBACMiddleware([]string{"hr", "admin"}))
			r.Route("/users", func(r chi.Router) {
				r.Get("/", userHandler.AdminList)
				r.Get("/{id}", userHandler.AdminGet)
				r.Put("/{id}", userHandler.AdminUpdate)
				r.Post("/{id}/deactivate", userHandler.AdminDeactivate)
			})
		})

		// Catalog Manager + Admin (M20)
		r.Group(func(r chi.Router) {
			r.Use(sharedauth.RBACMiddleware([]string{"catalog_manager", "admin"}))

			// Catalog admin — создаём handlers для admin context
			adminCatalogCache := catalog.NewCache(redis, appMetrics)
			adminCatalogRepo := catalog.NewRepository(db)
			adminCatalogService := catalog.NewService(adminCatalogRepo, adminCatalogCache)
			adminCatalogHandler := catalog.NewAdminHandler(adminCatalogService, adminCatalogCache)

			r.Route("/engagements/categories", func(r chi.Router) {
				r.Post("/", adminCatalogHandler.CreateCategory)
				r.Put("/{id}", adminCatalogHandler.UpdateCategory)
				r.Delete("/{id}", adminCatalogHandler.DeleteCategory)
			})

			r.Route("/engagements/types", func(r chi.Router) {
				r.Post("/", adminCatalogHandler.CreateType)
				r.Get("/", adminCatalogHandler.ListTypes)
				r.Get("/{id}", adminCatalogHandler.GetType)
				r.Put("/{id}", adminCatalogHandler.UpdateType)
				r.Delete("/{id}", adminCatalogHandler.DeleteType)
				r.Patch("/{id}/status", adminCatalogHandler.UpdateStatus)
			})

			r.Route("/engagements/types/{typeId}/offers", func(r chi.Router) {
				r.Post("/", adminCatalogHandler.CreateOffer)
				r.Put("/{id}", adminCatalogHandler.UpdateOffer)
				r.Delete("/{id}", adminCatalogHandler.DeleteOffer)
			})
		})
	})

	s := &Server{
		router: r,
		config: cfg,
		logger: logger,
	}

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start запускает HTTP-сервер.
func (s *Server) Start() error {
	s.logger.Info("server starting", "port", s.config.ServerPort)
	return s.httpServer.ListenAndServe()
}

// Shutdown выполняет graceful shutdown сервера.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}

// PrometheusMiddleware — middleware для сбора метрик Prometheus.
//
// Использует chi.RouteContext для получения route pattern (не raw URL path),
// что предотвращает высокую кардинальность метрик.
//
// Извлекает tenant_id из context (устанавливается tenant middleware),
// использует "unknown" если tenant не определён (публичные маршруты).
func PrometheusMiddleware(m *httpMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			// Tenant ID из context (установлен tenant middleware) или "unknown"
			tenantID := "unknown"
			if tid := tenant.TenantIDFromContext(r.Context()); tid != uuid.Nil {
				tenantID = tid.String()
			}

			m.total.WithLabelValues(r.Method, fmt.Sprintf("%d", ww.Status()), tenantID).Inc()

			// Route pattern из chi контекста — URL pattern без query params,
			// напр. /api/v1/tenants/{tenant_id}/benefits вместо /api/v1/tenants/sdek/benefits.
			// Это предотвращает высокую кардинальность метрик.
			pattern := r.URL.Path
			if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
				pattern = routeCtx.RoutePattern()
			}
			m.duration.WithLabelValues(r.Method, pattern, tenantID).Observe(time.Since(start).Seconds())
		})
	}
}
