# T2506 — Asynq worker: Сгорание баллов

## Веха

M25-billing

## Тип

code

## Что сделать

```go
// Scheduled job: каждый день в 00:00
func RegisterBurnJob(scheduler *asynq.Scheduler) {
    scheduler.Register("* 0 0 * * *", asynq.Task("billing:burn"))
}

func ProcessBurn(ctx context.Context, task asynq.Task) error {
    // 1. Find all user_balances with expires_at < NOW()
    // 2. For each: CreateTransaction(debit, burn, amount=balance)
    // 3. Set balance = 0
    // 4. Log burned amount
}
```

## Критерии приёмки

- [ ] Scheduled job (cron: `* 0 0 * * *`)
- [ ] Find expiring balances
- [ ] Create burn transactions
- [ ] Set balance = 0
- [ ] Unit tests
