# T2502 — internal/billing/ (Billing Engine)

## Веха

M25-billing

## Тип

code

## Контекст

Billing engine: правила, транзакции, баланс. Financial accuracy критична.
Исходник: `doc/архитектура/биллинг-движок.md`, `doc/архитектура/пакеты-platform.md` (строка 1021).

## Что сделать

```go
package billing

type Service struct {
    repo     Repository
    dbPool   *pgxpool.Pool
}

// CreateTransaction — атомарная транзакция с balance update
func (s *Service) CreateTransaction(ctx context.Context, t Transaction) (Transaction, error) {
    // DB transaction:
    // 1. INSERT transactions (status=frozen)
    // 2. UPDATE user_balances (balance += amount for credit, -= for debit)
    // 3. CHECK: balance >= 0 for debit
    // 4. UPDATE transactions (status=confirmed)
    // If any step fails — ROLLBACK
}

// GetBalance — баланс пользователя (total + by category)
func (s *Service) GetBalance(ctx context.Context, userID uuid.UUID) (Balance, error) {
    // SELECT category, balance_cents, expires_at FROM user_balances WHERE user_id = $1
    // Compute total, days_until_expiration
}
```

## Требования

- DB transaction (SERIALIZABLE или READ COMMITTED + SELECT FOR UPDATE)
- No double-spend: balance check перед debit
- BIGINT for amounts (копейки)
- Atomic: transaction + balance update в одном DB tx

## Критерии приёмки

- [ ] CreateTransaction — atomic DB tx
- [ ] GetBalance — total + categories + expiration
- [ ] No double-spend (balance check)
- [ ] BIGINT amounts
- [ ] Unit tests: credit, debit, insufficient balance, concurrent debit
