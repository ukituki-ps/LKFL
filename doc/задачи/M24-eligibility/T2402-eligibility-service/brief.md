# T2402 — internal/eligibility/ (Eligibility Engine)

## Веха

M24-eligibility

## Тип

code

## Контекст

Eligibility engine на базе CEL. Проверяет, доступен ли engagement пользователю.
Исходник: `doc/архитектура/пакеты-platform.md` (строка 510).

## Что сделать

```go
package eligibility

type Service struct {
    repo     Repository
    celEngine *cel.Engine
    ctxBuilder *celcontext.Builder
}

// Check — проверить eligibility для user + engagement
func (s *Service) Check(ctx context.Context, userID uuid.UUID, engagementID uuid.UUID) (bool, error) {
    // 1. Get eligibility conditions for engagement
    conditions, err := s.repo.GetConditions(ctx, engagementID)
    // 2. Build CEL context for user
    celCtx, err := s.ctxBuilder.Build(ctx, userID)
    // 3. Evaluate each condition (AND logic — все должны быть true)
    for _, cond := range conditions {
        result, err := s.celEngine.Evaluate(ctx, cond.Expression, celCtx)
        if !result.(bool) {
            return false, nil
        }
    }
    return true, nil
}
```

## Критерии приёмки

- [ ] Service.Check() — user + engagement → bool
- [ ] CEL integration
- [ ] AND logic (all conditions must pass)
- [ ] Unit tests: eligible, not eligible, no conditions, CEL error
