// Package metrics — кастомные Prometheus метрики LKFL.
//
// Метрики разделены по доменам:
//
//   - Catalog — запросы каталога (счётчики + гистограммы)
//   - Tenant — разрешение tenant (счётчики + гистограммы)
//   - Auth — аутентификация (счётчики)
//   - Redis — кэш (счётчики hits/misses/evictions)
//
// Все метрики регистрируются в переданном prometheus.Registry при создании.
// Метрики nil-safe: если Metrics == nil, методы no-op.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics — кастомные метрики LKFL.
type Metrics struct {
	// Catalog
	CatalogQueryTotal    *prometheus.CounterVec
	CatalogQueryDuration *prometheus.HistogramVec

	// Tenant
	TenantResolveTotal    *prometheus.CounterVec
	TenantResolveDuration *prometheus.HistogramVec

	// Auth
	AuthLoginTotal    *prometheus.CounterVec
	AuthCallbackTotal *prometheus.CounterVec

	// Redis cache
	RedisCacheHits      prometheus.Counter
	RedisCacheMisses    prometheus.Counter
	RedisCacheEvictions prometheus.Counter
}

// New создаёт и регистрирует метрики в реестре.
func New(reg *prometheus.Registry) *Metrics {
	m := &Metrics{
		// ─── Catalog ───
		CatalogQueryTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "catalog_query_total",
				Help: "Total number of catalog queries by type.",
			},
			[]string{"type", "status"},
		),
		CatalogQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "catalog_query_duration_seconds",
				Help:    "Catalog query duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"type"},
		),

		// ─── Tenant ───
		TenantResolveTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tenant_resolve_total",
				Help: "Total number of tenant resolution attempts.",
			},
			[]string{"method", "status"},
		),
		TenantResolveDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tenant_resolve_duration_seconds",
				Help:    "Tenant resolution duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),

		// ─── Auth ───
		AuthLoginTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_login_total",
				Help: "Total number of login attempts.",
			},
			[]string{"status"},
		),
		AuthCallbackTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "auth_callback_total",
				Help: "Total number of auth callback attempts.",
			},
			[]string{"status"},
		),

		// ─── Redis cache ───
		RedisCacheHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_cache_hits_total",
				Help: "Total number of Redis cache hits.",
			},
		),
		RedisCacheMisses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_cache_misses_total",
				Help: "Total number of Redis cache misses.",
			},
		),
		RedisCacheEvictions: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "redis_cache_evictions_total",
				Help: "Total number of Redis cache evictions.",
			},
		),
	}

	reg.MustRegister(
		m.CatalogQueryTotal,
		m.CatalogQueryDuration,
		m.TenantResolveTotal,
		m.TenantResolveDuration,
		m.AuthLoginTotal,
		m.AuthCallbackTotal,
		m.RedisCacheHits,
		m.RedisCacheMisses,
		m.RedisCacheEvictions,
	)

	return m
}

// ObserveCatalogQueryDuration — таймер для наблюдения длительности запроса каталога.
// Верните вызов возвращённой функции для записи метрики.
func (m *Metrics) ObserveCatalogQueryDuration(queryType string) func() {
	if m == nil {
		return func() {}
	}
	start := time.Now()
	return func() {
		m.CatalogQueryDuration.WithLabelValues(queryType).Observe(time.Since(start).Seconds())
	}
}

// ObserveTenantResolveDuration — таймер для наблюдения длительности разрешения tenant.
// Верните вызов возвращённой функции для записи метрики.
func (m *Metrics) ObserveTenantResolveDuration(method string) func() {
	if m == nil {
		return func() {}
	}
	start := time.Now()
	return func() {
		m.TenantResolveDuration.WithLabelValues(method).Observe(time.Since(start).Seconds())
	}
}
