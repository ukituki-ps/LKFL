# T2304 — CEL Domains (billing, eligibility, flow, game stub, recommendations stub)

## Веха

M23-cel-engine

## Тип

code

## Что сделать

### Domain type providers

```go
// billing.* — billing rules
// eligibility.* — eligibility conditions
// flow.* — flow step conditions
// game.* — gamification triggers (stub)
// recommendations.* — recommendation rules (stub)
```

### Примеры выражений

```
billing: balance.total >= offer.cost_cents
eligibility: tags.grade IN ['A', 'B', 'C'] && tags.tenure_months >= 6
flow: context.engagements.exists(e, e.type == 'dms')
```

## Критерии приёмки

- [ ] 5 доменов зарегистрированы
- [ ] billing, eligibility, flow — рабочие
- [ ] game, recommendations — stub (пустой slice)
- [ ] Примеры выражений работают
