# T2306 — Prometheus metrics: CEL

## Веха

M23-cel-engine

## Тип

code

## Что сделать

```go
celEvaluationTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "cel_evaluation_total"},
    []string{"domain", "status"},
)

celEvaluationDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "cel_evaluation_duration_seconds",
        Help:    "CEL evaluation latency",
        Buckets: prometheus.DefBuckets,
    },
    []string{"domain"},
)
```

## Критерии приёмки

- [ ] `cel_evaluation_total{domain, status}`
- [ ] `cel_evaluation_duration_seconds{domain}`
- [ ] Registration в app/wire.go
- [ ] Available in /metrics
