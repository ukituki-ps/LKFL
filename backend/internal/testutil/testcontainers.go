// Package testutil — утилиты для интеграционных тестов.
//
// Предоставляет функции для запуска testcontainers (PostgreSQL + Redis),
// применения миграций и создания тестового HTTP-сервера с реальными зависимостями.
//
// Все тесты, использующие этот пакет, должны иметь build tag `integration`:
//
//	//go:build integration
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	intauth "lkfl/internal/auth"
	"lkfl/internal/engagement/catalog"
	"lkfl/internal/tenant"
	"lkfl/internal/user"
	sharedauth "lkfl/shared/pkg/auth"
	"lkfl/shared/pkg/migrate"
)

// SetupTestDB запускает PostgreSQL testcontainer, применяет миграции,
// возвращает pool + cleanup-функцию.
func SetupTestDB(ctx context.Context) (*pgxpool.Pool, func(), error) {
	node, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("lkfl_test"),
		postgres.WithUsername("lkfl"),
		postgres.WithPassword("lkfl"),
		tc.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := node.ConnectionString(ctx)
	if err != nil {
		_ = node.Terminate(ctx)
		return nil, nil, fmt.Errorf("get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		_ = node.Terminate(ctx)
		return nil, nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = node.Terminate(ctx)
		return nil, nil, fmt.Errorf("ping: %w", err)
	}

	// Apply migrations (blind mode — no tracking table for test containers)
	if err := migrate.Apply(ctx, nil, pool, false); err != nil {
		pool.Close()
		_ = node.Terminate(ctx)
		return nil, nil, fmt.Errorf("apply migrations: %w", err)
	}

	cleanup := func() {
		pool.Close()
		_ = node.Terminate(ctx)
	}

	return pool, cleanup, nil
}

// SetupTestRedis запускает Redis testcontainer, возвращает client + cleanup.
func SetupTestRedis(ctx context.Context) (*redis.Client, func(), error) {
	node, err := tcredis.Run(ctx,
		"redis:7-alpine",
		tc.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start redis container: %w", err)
	}

	endpoint, err := node.ConnectionString(ctx)
	if err != nil {
		_ = node.Terminate(ctx)
		return nil, nil, fmt.Errorf("get connection string: %w", err)
	}

	// ConnectionString returns "redis://localhost:PORT" or just "localhost:PORT"
	url := endpoint
	if !strings.HasPrefix(url, "redis://") {
		url = "redis://" + url
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		_ = node.Terminate(ctx)
		return nil, nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		_ = node.Terminate(ctx)
		return nil, nil, fmt.Errorf("ping: %w", err)
	}

	cleanup := func() {
		_ = client.Close()
		_ = node.Terminate(ctx)
	}

	return client, cleanup, nil
}

// TestServer — обёртка над httptest.Server с helper-методами.
type TestServer struct {
	*httptest.Server
	DB    *pgxpool.Pool
	Redis *redis.Client
}

// SetupAllWithServer запускает PostgreSQL + Redis + HTTP-сервер для интеграционных тестов.
func SetupAllWithServer(ctx context.Context) (*TestServer, func(), error) {
	db, dbCleanup, err := SetupTestDB(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("setup db: %w", err)
	}

	redisClient, redisCleanup, err := SetupTestRedis(ctx)
	if err != nil {
		dbCleanup()
		return nil, nil, fmt.Errorf("setup redis: %w", err)
	}

	server, err := buildTestServer(db, redisClient)
	if err != nil {
		dbCleanup()
		redisCleanup()
		return nil, nil, fmt.Errorf("build server: %w", err)
	}

	ts := &TestServer{
		Server: server,
		DB:     db,
		Redis:  redisClient,
	}

	cleanup := func() {
		server.Close()
		redisCleanup()
		dbCleanup()
	}

	return ts, cleanup, nil
}

// SetupTestServer запускает httptest.Server с реальными зависимостями (DB + Redis).
// Возвращает сервер и cleanup-функцию (только закрывает сервер, не DB/Redis).
func SetupTestServer(db *pgxpool.Pool, redisClient *redis.Client) (*httptest.Server, func(), error) {
	server, err := buildTestServer(db, redisClient)
	if err != nil {
		return nil, nil, fmt.Errorf("build server: %w", err)
	}

	cleanup := func() {
		server.Close()
	}

	return server, cleanup, nil
}

func buildTestServer(db *pgxpool.Pool, redisClient *redis.Client) (*httptest.Server, error) {
	// Tenant service
	tenantRepo := tenant.NewRepository(db)
	tenantService := tenant.NewService(tenantRepo)

	// User service
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Auth service + handler (verifier не нужен для test middleware)
	authService := intauth.NewService(userRepo)
	authHandler := intauth.NewHandler(nil, redisClient, authService,
		"http://localhost:8080/auth/realms/lkfl", "", "lkfl-spa", "", nil)

	// Build router
	r := chi.NewRouter()

	// Middleware chain
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(testCORS())

	// Health
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Auth routes (public, no JWT)
	r.Get("/api/v1/auth/login", authHandler.LoginRedirect)
	r.Get("/api/v1/auth/callback", authHandler.LoginCallback)
	r.Post("/api/v1/auth/logout", authHandler.Logout)

	// Employee routes (test JWT + tenant middleware)
	r.Route("/api/v1/", func(r chi.Router) {
		r.Use(TestMiddleware())

		// Auth/me — doesn't need tenant resolution
		r.Get("/auth/me", authHandler.Me)

		// Routes requiring tenant resolution
		r.Group(func(r chi.Router) {
			r.Use(tenant.TenantMiddlewareWithService(tenantService, redisClient, nil))

			r.Get("/users/me", userHandler.Me)
			r.Put("/users/me", userHandler.UpdateMe)

			// Catalog
			catalogCache := catalog.NewCache(redisClient, nil)
			catalogRepo := catalog.NewRepository(db)
			catalogService := catalog.NewService(catalogRepo, catalogCache)
			catalogHandler := catalog.NewHandler(catalogService, nil)
			r.Get("/engagements/categories", catalogHandler.Categories)
			r.Get("/engagements", catalogHandler.List)
			r.Get("/engagements/{id}", catalogHandler.Get)
		})
	})

	// Admin routes (test JWT + RBAC + admin tenant middleware)
	r.Route("/admin/", func(r chi.Router) {
		r.Use(TestMiddleware())
		r.Use(tenant.AdminTenantMiddleware(tenantService, redisClient, nil))

		// Admin-only (tenants CRUD)
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

		// HR + Admin (users)
		r.Group(func(r chi.Router) {
			r.Use(sharedauth.RBACMiddleware([]string{"hr", "admin"}))
			r.Route("/users", func(r chi.Router) {
				r.Get("/", userHandler.AdminList)
				r.Get("/{id}", userHandler.AdminGet)
				r.Put("/{id}", userHandler.AdminUpdate)
				r.Post("/{id}/deactivate", userHandler.AdminDeactivate)
			})
		})

		// Catalog Manager + Admin
		r.Group(func(r chi.Router) {
			r.Use(sharedauth.RBACMiddleware([]string{"catalog_manager", "admin"}))

			adminCatalogCache := catalog.NewCache(redisClient, nil)
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

	return httptest.NewServer(r), nil
}

// TestMiddleware — middleware для тестов, который принимает тестовые JWT токены.
//
// Формат тестового токена: "test:{subject}:{comma-separated-roles}"
// Пример: Authorization: Bearer test:abc123:admin,employee
func TestMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				writeJSONError(w, http.StatusUnauthorized, "invalid token format")
				return
			}

			claims, roles := parseTestToken(tokenString)
			if claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), sharedauth.ClaimsKey, *claims)
			ctx = context.WithValue(ctx, sharedauth.RolesKey, roles)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TestToken создаёт тестовый токен для TestMiddleware.
// Формат: "test:{subject}:{role1,role2,...}"
func TestToken(subject string, roles ...string) string {
	rolesStr := strings.Join(roles, ",")
	return fmt.Sprintf("test:%s:%s", subject, rolesStr)
}

// TestTokenAdmin создаёт тестовый токен с ролью admin.
func TestTokenAdmin(subject string) string {
	return TestToken(subject, "admin")
}

// TestTokenEmployee создаёт тестовый токен с ролью employee.
func TestTokenEmployee(subject string) string {
	return TestToken(subject, "employee")
}

// TestTokenHR создаёт тестовый токен с ролями hr и employee.
func TestTokenHR(subject string) string {
	return TestToken(subject, "hr", "employee")
}

// TestTokenCatalogManager создаёт тестовый токен с ролями catalog_manager и employee.
func TestTokenCatalogManager(subject string) string {
	return TestToken(subject, "catalog_manager", "employee")
}

// parseTestToken парсит тестовый токен.
func parseTestToken(token string) (*sharedauth.Claims, []string) {
	if !strings.HasPrefix(token, "test:") {
		return nil, nil
	}

	parts := strings.SplitN(strings.TrimPrefix(token, "test:"), ":", 2)
	if len(parts) < 2 {
		return nil, nil
	}

	subject := parts[0]
	if subject == "" {
		return nil, nil
	}

	var roles []string
	if parts[1] != "" {
		for _, r := range strings.Split(parts[1], ",") {
			r = strings.TrimSpace(r)
			if r != "" {
				roles = append(roles, r)
			}
		}
	}

	return &sharedauth.Claims{
		Subject:    subject,
		Email:      subject + "@test.local",
		GivenName:  "Test",
		FamilyName: "User",
	}, roles
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"error":"%s"}`, message)
}

func testCORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Tenant-ID")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ─── HTTP Helper Methods ───

// GetWithToken выполняет GET запрос с авторизацией.
func (ts *TestServer) GetWithToken(path string, token string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", ts.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return ts.Do(req)
}

// GetWithTokenAndTenant выполняет GET запрос с авторизацией и X-Tenant-ID.
func (ts *TestServer) GetWithTokenAndTenant(path string, token string, tenantSlug string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", ts.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", tenantSlug)
	return ts.Do(req)
}

// PostWithToken выполняет POST запрос с авторизацией.
func (ts *TestServer) PostWithToken(path string, token string, body any) (*http.Response, error) {
	return ts.DoJSON("POST", path, token, body)
}

// PostWithTokenAndTenant выполняет POST запрос с авторизацией и X-Tenant-ID.
func (ts *TestServer) PostWithTokenAndTenant(path string, token string, tenantSlug string, body any) (*http.Response, error) {
	return ts.DoJSONWithTenant("POST", path, token, tenantSlug, body)
}

// PutWithToken выполняет PUT запрос с авторизацией.
func (ts *TestServer) PutWithToken(path string, token string, body any) (*http.Response, error) {
	return ts.DoJSON("PUT", path, token, body)
}

// PutWithTokenAndTenant выполняет PUT запрос с авторизацией и X-Tenant-ID.
func (ts *TestServer) PutWithTokenAndTenant(path string, token string, tenantSlug string, body any) (*http.Response, error) {
	return ts.DoJSONWithTenant("PUT", path, token, tenantSlug, body)
}

// PatchWithToken выполняет PATCH запрос с авторизацией.
func (ts *TestServer) PatchWithToken(path string, token string, body any) (*http.Response, error) {
	return ts.DoJSON("PATCH", path, token, body)
}

// PatchWithTokenAndTenant выполняет PATCH запрос с авторизацией и X-Tenant-ID.
func (ts *TestServer) PatchWithTokenAndTenant(path string, token string, tenantSlug string, body any) (*http.Response, error) {
	return ts.DoJSONWithTenant("PATCH", path, token, tenantSlug, body)
}

// DeleteWithToken выполняет DELETE запрос с авторизацией.
func (ts *TestServer) DeleteWithToken(path string, token string) (*http.Response, error) {
	req, _ := http.NewRequest("DELETE", ts.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return ts.Do(req)
}

// DeleteWithTokenAndTenant выполняет DELETE запрос с авторизацией и X-Tenant-ID.
func (ts *TestServer) DeleteWithTokenAndTenant(path string, token string, tenantSlug string) (*http.Response, error) {
	req, _ := http.NewRequest("DELETE", ts.URL+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", tenantSlug)
	return ts.Do(req)
}

// DoJSON выполняет запрос с JSON телом.
func (ts *TestServer) DoJSON(method, path string, token string, body any) (*http.Response, error) {
	return ts.DoJSONWithTenant(method, path, token, "", body)
}

// DoJSONWithTenant выполняет запрос с JSON телом и X-Tenant-ID.
func (ts *TestServer) DoJSONWithTenant(method, path string, token string, tenantSlug string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	req, _ := http.NewRequest(method, ts.URL+path, strings.NewReader(string(data)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if tenantSlug != "" {
		req.Header.Set("X-Tenant-ID", tenantSlug)
	}
	return ts.Do(req)
}

func (ts *TestServer) Do(req *http.Request) (*http.Response, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// ─── DB Seed Helpers ───

// CreateTenant создаёт tenant через DB напрямую.
func (ts *TestServer) CreateTenant(ctx context.Context, slug, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := ts.DB.QueryRow(ctx,
		`INSERT INTO lkfl_platform.tenants (slug, name, status, settings) VALUES ($1, $2, $3, $4) RETURNING id`,
		slug, name, "active", "{}",
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create tenant: %w", err)
	}
	return id, nil
}

// CreateTenantWithStatus создаёт tenant с указанным статусом.
func (ts *TestServer) CreateTenantWithStatus(ctx context.Context, slug, name, status string) (uuid.UUID, error) {
	var id uuid.UUID
	err := ts.DB.QueryRow(ctx,
		`INSERT INTO lkfl_platform.tenants (slug, name, status, settings) VALUES ($1, $2, $3, $4) RETURNING id`,
		slug, name, status, "{}",
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create tenant: %w", err)
	}
	return id, nil
}

// CreateBrandConfig создаёт brand config.
func (ts *TestServer) CreateBrandConfig(ctx context.Context, tenantID uuid.UUID, primaryColor, secondaryColor string) error {
	_, err := ts.DB.Exec(ctx,
		`INSERT INTO lkfl_platform.tenant_brand_config (tenant_id, primary_color, secondary_color, css_variables)
		 VALUES ($1, $2, $3, $4)`,
		tenantID, primaryColor, secondaryColor, "{}",
	)
	return err
}

// CreateUser создаёт пользователя.
func (ts *TestServer) CreateUser(ctx context.Context, tenantID uuid.UUID, email, firstName, lastName, keycloakSub string) (uuid.UUID, error) {
	var id uuid.UUID
	err := ts.DB.QueryRow(ctx,
		`INSERT INTO lkfl_platform.users (tenant_id, email, first_name, last_name, status, keycloak_sub, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		tenantID, email, firstName, lastName, "active", keycloakSub, "{}",
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}

// AddUserRole добавляет роль пользователю.
func (ts *TestServer) AddUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	_, err := ts.DB.Exec(ctx,
		`INSERT INTO lkfl_platform.user_roles (user_id, role) VALUES ($1, $2)`,
		userID, role,
	)
	return err
}

// CreateCategory создаёт категорию.
func (ts *TestServer) CreateCategory(ctx context.Context, tenantID uuid.UUID, slug, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := ts.DB.QueryRow(ctx,
		`INSERT INTO lkfl_platform.engagement_categories (tenant_id, slug, name, sort_order)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		tenantID, slug, name, 0,
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create category: %w", err)
	}
	return id, nil
}

// CreateEngagementType создаёт тип энгейджмента.
func (ts *TestServer) CreateEngagementType(ctx context.Context, tenantID, categoryID uuid.UUID, slug, name, engType, status string) (uuid.UUID, error) {
	var id uuid.UUID
	err := ts.DB.QueryRow(ctx,
		`INSERT INTO lkfl_platform.engagement_types (tenant_id, category_id, slug, name, type, status, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		tenantID, categoryID, slug, name, engType, status, "{}",
	).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create engagement type: %w", err)
	}
	return id, nil
}

// ReadBody читает тело ответа и закрывает его.
func ReadBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

// ReadJSONBody читает тело ответа как JSON и десериализует в v.
func ReadJSONBody[T any](resp *http.Response, v *T) error {
	if resp == nil || resp.Body == nil {
		return fmt.Errorf("no response")
	}
	defer func() { _ = resp.Body.Close() }()
	return json.NewDecoder(resp.Body).Decode(v)
}

// ─── unused imports guard ───
var (
	_ = os.Stdout
)
