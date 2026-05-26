# T2602 — internal/engagement/flow/ (Flow Engine)

## Веха

M26-engagement-flow

## Тип

code

## Контекст

Flow engine — step-by-step activation. JSON-driven steps definition.
Исходник: `doc/архитектура/engagement.md` (flow section).

## Что сделать

```go
package flow

type Service struct {
    repo         Repository
    eligibility  *eligibility.Service
    billing      *billing.Service
}

// StartFlow — начать активацию
func (s *Service) StartFlow(ctx context.Context, userID uuid.UUID, engagementID uuid.UUID, offerID uuid.UUID) (UserEngagement, error) {
    // 1. Eligibility check
    eligible, err := s.eligibility.Check(ctx, userID, engagementID)
    if !eligible {
        return UserEngagement{}, ErrNotEligible
    }
    // 2. Create user_engagement (status=in_progress)
    // 3. Create steps from flow definition
    // 4. Return first step
}

// CompleteStep — завершить шаг flow
func (s *Service) CompleteStep(ctx context.Context, engagementID uuid.UUID, stepKey string, data map[string]interface{}) (UserEngagement, error) {
    // 1. Mark step completed
    // 2. Check if all steps done
    // 3. If all done → trigger billing debit
}

// OnFlowComplete — callback при завершении flow
func (s *Service) OnFlowComplete(ctx context.Context, ue UserEngagement) error {
    // 1. Debit from balance (billing.CreateTransaction)
    // 2. Update user_engagement status → active
    // 3. Notify (F3)
}
```

## Критерии приёмки

- [ ] StartFlow — eligibility check + create
- [ ] CompleteStep — mark done + check all complete
- [ ] OnFlowComplete — billing debit + activate
- [ ] Unique constraint: one active per user+type
- [ ] Unit tests: eligible flow, not eligible, step completion, flow complete → debit
