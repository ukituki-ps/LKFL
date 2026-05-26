# T2403 — API: Eligibility Check

## Веха

M24-eligibility

## Тип

code

## Что сделать

```
POST /api/v1/eligibility/check
Body: { "engagement_id": "uuid" }
Response: { "eligible": true, "reason": null }
или
Response: { "eligible": false, "reason": "minimum tenure not met" }
```

## Критерии приёмки

- [ ] POST /api/v1/eligibility/check
- [ ] JWT + tenant middleware
- [ ] Response: eligible + reason
- [ ] Unit tests
