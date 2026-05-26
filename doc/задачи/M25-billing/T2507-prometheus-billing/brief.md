# T2507 — Prometheus metrics: Billing

## Веха

M25-billing

## Тип

code

## Что сделать

```go
billingTransactionsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "billing_transactions_total"},
    []string{"type"}, // credit, debit, burn
)

billingBalanceQueryTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "billing_balance_query_total"},
    []string{},
)
```

## Критерии приёмки

- [ ] `billing_transactions_total{type}`
- [ ] `billing_balance_query_total`
- [ ] Registration в app/wire.go
