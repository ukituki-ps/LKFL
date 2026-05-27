package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"lkfl/internal/metrics"
	shhttp "lkfl/shared/pkg/http"
)

// contextKey — типированный ключ для context.
type contextKey string

const (
	// TenantIDKey — ключ tenant ID в context.
	TenantIDKey contextKey = "tenant_id"

	// TenantSlugKey — ключ tenant slug в context.
	TenantSlugKey contextKey = "tenant_slug"
)

// TenantIDFromContext извлекает tenant ID из context.
func TenantIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(TenantIDKey).(uuid.UUID)
	return id
}

// TenantSlugFromContext извлекает tenant slug из context.
func TenantSlugFromContext(ctx context.Context) string {
	slug, _ := ctx.Value(TenantSlugKey).(string)
	return slug
}

// Resolver — интерфейс для разрешения tenant из request.
type Resolver interface {
	Resolve(r *http.Request) (Tenant, error)
}

// HostResolver — резолвит tenant из subdomain (sdek.example.com → slug=sdek).
// Fallback на X-Tenant-ID header для www/localhost.
type HostResolver struct {
	service *Service
	redis   *redis.Client
	metrics *metrics.Metrics
}

// NewHostResolver создаёт HostResolver.
// Если redisClient == nil, кэширование отключено.
// Если m == nil, метрики не собираются.
func NewHostResolver(service *Service, redisClient *redis.Client, m *metrics.Metrics) *HostResolver {
	return &HostResolver{
		service: service,
		redis:   redisClient,
		metrics: m,
	}
}

// Resolve реализует Resolver для HostResolver.
//
// Порядок разрешения:
//  1. Host header (subdomain → slug)
//  2. X-Tenant-ID header (localhost, www)
//
// Для JWT-based tenant resolution используйте JWTClaimsTenantMiddleware
// перед этим resolver'ом — он парсит issuer из JWT и ставит X-Tenant-ID.
func (r *HostResolver) Resolve(req *http.Request) (Tenant, error) {
	method := "host"
	if r.metrics != nil {
		done := r.metrics.ObserveTenantResolveDuration(method)
		defer done()
	}

	// X-Tenant-ID header имеет приоритет (ставится internal nginx или JWT middleware)
	if slug := req.Header.Get("X-Tenant-ID"); slug != "" {
		method = "header"
		tenant, err := r.resolveBySlug(req.Context(), slug)
		if r.metrics != nil {
			status := "success"
			if err != nil {
				status = "error"
			}
			r.metrics.TenantResolveTotal.WithLabelValues(method, status).Inc()
		}
		return tenant, err
	}

	// Убираем порт (localhost:8080 → localhost)
	host := req.Host
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	host = strings.Split(host, ".")[0]

	if host == "www" || host == "localhost" || host == "127" || host == "" {
		method = "header"
		if r.metrics != nil {
			r.metrics.TenantResolveTotal.WithLabelValues(method, "error").Inc()
		}
		return Tenant{}, ErrNoTenantHeader
	}

	tenant, err := r.resolveBySlug(req.Context(), host)
	if r.metrics != nil {
		status := "success"
		if err != nil {
			status = "error"
		}
		r.metrics.TenantResolveTotal.WithLabelValues(method, status).Inc()
	}
	return tenant, err
}

// ResolveBySlug — публичная версия resolveBySlug для использования из middleware.
func (r *HostResolver) ResolveBySlug(ctx context.Context, slug string) (Tenant, error) {
	return r.resolveBySlug(ctx, slug)
}

// resolveBySlug — общая логика разрешения по slug с Redis кэшированием.
func (r *HostResolver) resolveBySlug(ctx context.Context, slug string) (Tenant, error) {
	// Попытка получить из Redis кэша
	if r.redis != nil {
		if cached, err := r.getCachedTenant(ctx, slug); err == nil && cached != nil {
			return *cached, nil
		}
	}

	// Получаем из DB (без проверки status — middleware не фильтрует suspended)
	tenant, err := r.service.GetBySlugRaw(ctx, slug)
	if err != nil {
		return Tenant{}, err
	}

	// Кэшируем в Redis (без проверки status — middleware не фильтрует suspended)
	if r.redis != nil {
		_ = r.setCachedTenant(ctx, slug, tenant)
	}

	return tenant, nil
}

// getCachedTenant — получить tenant из Redis кэша.
func (r *HostResolver) getCachedTenant(ctx context.Context, slug string) (*Tenant, error) {
	key := "tenant:resolve:" + slug
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var tenant Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return nil, err
	}
	return &tenant, nil
}

// setCachedTenant — сохранить tenant в Redis кэш на 5 минут.
func (r *HostResolver) setCachedTenant(ctx context.Context, slug string, tenant Tenant) error {
	key := "tenant:resolve:" + slug
	data, err := json.Marshal(tenant)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, key, data, 5*time.Minute).Err()
}

// PathResolver — резолвит tenant из path segment (/t/{slug}/).
type PathResolver struct {
	service *Service
	redis   *redis.Client
}

// NewPathResolver создаёт PathResolver.
func NewPathResolver(service *Service, redisClient *redis.Client) *PathResolver {
	return &PathResolver{
		service: service,
		redis:   redisClient,
	}
}

// Resolve реализует Resolver для PathResolver.
func (r *PathResolver) Resolve(req *http.Request) (Tenant, error) {
	// /t/sdek/api/v1/... → sdek
	path := req.URL.Path
	if !strings.HasPrefix(path, "/t/") {
		return Tenant{}, ErrNoTenantInPath
	}

	parts := strings.Split(strings.TrimPrefix(path, "/t/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return Tenant{}, ErrNoTenantInPath
	}

	slug := parts[0]

	// Redis кэш
	if r.redis != nil {
		if cached, err := r.getPathCachedTenant(req.Context(), slug); err == nil && cached != nil {
			return *cached, nil
		}
	}

	tenant, err := r.service.GetBySlugRaw(req.Context(), slug)
	if err != nil {
		return Tenant{}, err
	}

	if r.redis != nil {
		_ = r.setPathCachedTenant(req.Context(), slug, tenant)
	}

	return tenant, nil
}

// getPathCachedTenant — получить tenant из Redis кэша для PathResolver.
func (r *PathResolver) getPathCachedTenant(ctx context.Context, slug string) (*Tenant, error) {
	key := "tenant:path:" + slug
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var tenant Tenant
	if err := json.Unmarshal(data, &tenant); err != nil {
		return nil, err
	}
	return &tenant, nil
}

// setPathCachedTenant — сохранить tenant в Redis кэш для PathResolver.
func (r *PathResolver) setPathCachedTenant(ctx context.Context, slug string, tenant Tenant) error {
	key := "tenant:path:" + slug
	data, err := json.Marshal(tenant)
	if err != nil {
		return err
	}
	return r.redis.Set(ctx, key, data, 5*time.Minute).Err()
}

// ErrNoTenantHeader — отсутствует X-Tenant-ID header.
var ErrNoTenantHeader = errors.New("no tenant header: X-Tenant-ID required for www/localhost")

// ErrNoTenantInPath — отсутствует tenant в path.
var ErrNoTenantInPath = errors.New("no tenant in path: use /t/{slug}/ prefix")

// Middleware создаёт chi middleware для tenant resolution.
//
// Resolver используется для определения tenant из request.
// Результат добавляется в context через TenantIDKey и TenantSlugKey.
// При ошибке возврата 401 Unauthorized.
func Middleware(resolver Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, err := resolver.Resolve(r)
			if err != nil {
				shhttp.WriteJSONError(w, http.StatusUnauthorized, "tenant not found")
				return
			}

			ctx := context.WithValue(r.Context(), TenantIDKey, tenant.ID)
			ctx = context.WithValue(ctx, TenantSlugKey, tenant.Slug)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SkipPaths возвращает middleware, пропускающий tenant resolution для
// указанных путей (healthz, metrics, admin).
func SkipPaths(paths []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range paths {
				if strings.HasPrefix(r.URL.Path, p) {
					next.ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TenantMiddlewareWithService создаёт middleware, используя Service и Redis напрямую.
// Это удобный конструктор, который создаёт HostResolver и оборачивает его в Middleware.
// Если m == nil, метрики не собираются.
func TenantMiddlewareWithService(service *Service, redisClient *redis.Client, m *metrics.Metrics) func(http.Handler) http.Handler {
	resolver := NewHostResolver(service, redisClient, m)
	return Middleware(resolver)
}
