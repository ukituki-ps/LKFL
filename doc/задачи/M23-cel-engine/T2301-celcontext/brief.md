# T2301 — shared/pkg/celcontext/ (CELContext type + builder)

## Веха

M23-cel-engine

## Тип

code

## Контекст

`shared/pkg/celcontext/` — тип CELContext и builder для сборки контекста оценки CEL выражений.
Исходник: `doc/архитектура/cel-engine.md` (строка 7), `doc/архитектура/пакеты-platform.md` (строка 384).

## Что сделать

```go
package celcontext

// CELContext — контекст для CEL evaluation
type CELContext struct {
    UserID    uuid.UUID            `json:"user_id"`
    TenantID  uuid.UUID            `json:"tenant_id"`
    Tags      map[string]string    `json:"tags"`       // user tags: grade, department, tenure, family
    Period    *PeriodInfo          `json:"period"`     // current billing period
    Balance   *BalanceInfo         `json:"balance"`    // user balance
    Engagements []EngagementInfo   `json:"engagements"` // active engagements
}

// Builder — сборщик CELContext из различных источников
type Builder struct {
    tagResolver   TagResolver
    billingRepo   BillingRepository
    engagementRepo EngagementRepository
}

func (b *Builder) Build(ctx context.Context, userID uuid.UUID) (CELContext, error) {
    // 1. Get user tags
    tags, err := b.tagResolver.Resolve(ctx, userID)
    // 2. Get current period
    period, _ := b.billingRepo.GetCurrentPeriod(ctx)
    // 3. Get balance
    balance, _ := b.billingRepo.GetBalance(ctx, userID)
    // 4. Get active engagements
    engagements, _ := b.engagementRepo.GetActive(ctx, userID)
    // Return CELContext
}
```

## Критерии приёмки

- [ ] `CELContext` struct
- [ ] `Builder` с Resolve()
- [ ] Tags, Period, Balance, Engagements
- [ ] Unit tests
