# T2404 — TagResolver (теги пользователя для CEL)

## Веха

M24-eligibility

## Тип

code

## Контекст

TagResolver собирает теги пользователя из metadata и других источников для CEL context.
Исходник: `doc/архитектура/теги.md`.

## Что сделать

```go
type TagResolver interface {
    Resolve(ctx context.Context, userID uuid.UUID) (map[string]string, error)
}

type UserTagResolver struct {
    userRepo Repository
}

func (r *UserTagResolver) Resolve(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
    user, err := r.userRepo.GetByID(ctx, userID)
    tags := make(map[string]string)
    // From metadata JSONB: grade, department, hire_date
    // Computed: tenure_months, has_family
    // Return tags
}
```

## Критерии приёмки

- [ ] TagResolver interface
- [ ] UserTagResolver implementation
- [ ] Tags: grade, department, tenure_months, has_family
- [ ] Redis cache: `cel:tags:{user_id}` TTL 10min
- [ ] Unit tests
