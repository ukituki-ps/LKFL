# T2505 — Asynq worker: Массовое начисление

## Веха

M25-billing

## Тип

code

## Контекст

Asynq background job для массового начисления по периоду.
Обработка 10000+ пользователей без блокировки HTTP.

## Что сделать

```go
package billing

type DistributeJob struct {
    PeriodID uuid.UUID
}

func (j *DistributeJob) Process(ctx context.Context, task asynq.Task) error {
    var job DistributeJob
    json.Unmarshal(task.Payload(), &job)

    // 1. Get period
    // 2. Get all active users for tenant
    // 3. For each user: CreateTransaction(credit, rule.amount)
    // 4. Track progress in Redis: billing:distribute:{period_id}
    // 5. Update period status → closed
    // Batch: 100 users per DB transaction
}
```

### Progress tracking (Redis)

```
billing:distribute:{period_id} → {"total": 10000, "processed": 3500, "errors": 2, "status": "running"}
```

## Критерии приёмки

- [ ] Asynq job type: `billing:distribute`
- [ ] Batch processing (100 per tx)
- [ ] Progress tracking в Redis
- [ ] Error handling (partial failure → log + continue)
- [ ] Unit tests: 1000 users, partial failure
