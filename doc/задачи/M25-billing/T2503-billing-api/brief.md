# T2503 — Billing API (Public)

## Веха

M25-billing

## Тип

code

## Что сделать

```
GET  /api/v1/balance           — баланс (total + categories + expiration)
GET  /api/v1/transactions      — транзакции (фильтры: direction, category, date range, pagination)
GET  /api/v1/periods           — периоды (current + history)
```

### Response: Balance

```json
{
  "total": 15000,
  "categories": [
    {"name": "general", "balance": 10000, "expires_at": "2026-12-31", "days_until_expiration": 219},
    {"name": "fitness", "balance": 5000, "expires_at": "2026-06-30", "days_until_expiration": 35}
  ]
}
```

## Критерии приёмки

- [ ] GET /api/v1/balance
- [ ] GET /api/v1/transactions (фильтры + pagination)
- [ ] GET /api/v1/periods
- [ ] JWT + tenant middleware
- [ ] Unit tests
