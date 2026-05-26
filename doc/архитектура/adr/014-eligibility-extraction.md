# ADR-014 — Выношение EligibilityEngine в отдельный пакет `internal/eligibility/`

## Контекст

После M06 eligibility engine находился внутри `internal/engagement/` вместе с catalog, flow и collections. Три других пакета зависят от eligibility:

1. `engagement/flow/` — проверка eligibility перед Activate
2. `recommendations/engine.go` — segment matching использует eligibility evaluation
3. Billing Rule Engine — `evaluation` conditions совпадают с eligibility rules

Это создаёт тесную связанность: eligibility — не подответственность engagement, а кросс-функциональный сервис, который должен быть доступен всем трём доменам.

## Решение

Вынести eligibility evaluation в отдельный пакет `internal/eligibility/`:

```
internal/
├── eligibility/           # ← НОВЫЙ
│   ├── engine.go          # EligibilityEngine: Check, EvaluateRule, EvaluateGroup
│   └── types.go           # EligibilityRule, EligibilityGroup, EligibilityResult
├── engagement/
│   ├── catalog.go         # каталог + cache
│   ├── flow.go            # flow execution (uses eligibility via DI)
│   ├── collections.go     # collections management
│   └── billing_events.go  # publish billing events → NATS
└── ...
```

Новый пакет `eligibility/`:
- Зависит только от `pkg/` и `user/` (для получения профиля пользователя)
- Не зависит от `engagement/` или `db/` напрямую (evaluation — чистая логика)
- Public API: `Check(ctx, offerId, userId)`, `EvaluateRule(ctx, rule, profile)`, `EvaluateGroup(ctx, group, profile)`

## Последствия

- ✅ SRP восстановлен: engagement содержит только catalog + flow + collections
- ✅ Круговая зависимость устранена: recommendations может импортировать eligibility без импорта engagement
- ✅ Тестируемость: eligibility engine unit-тестируется изолированно (без mock DB)
- ⚠️ Count internal packages: 7 → 8

## Статус

✅ Accepted (M07, T0701)
