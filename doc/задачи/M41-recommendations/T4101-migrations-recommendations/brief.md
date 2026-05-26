# T4101-T4105 — Рекомендации (2-layer rule engine)

## Веха

M41-recommendations

## T4101 — Migrations
```sql
CREATE TABLE lkfl_platform.recommendation_rules (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255),
    layer VARCHAR(20) CHECK (layer IN ('context','segment')),
    cel_rule_id UUID FK, priority INT, enabled BOOLEAN DEFAULT true,
    created_at, updated_at
);
CREATE TABLE lkfl_platform.recommendation_contexts (
    id UUID PK, rule_id UUID FK, name VARCHAR(255),
    threshold_key VARCHAR(50), threshold_value DECIMAL(10,2),
    created_at
);
CREATE TABLE lkfl_platform.recommendation_segments (
    id UUID PK, rule_id UUID FK, name VARCHAR(255),
    cel_criteria TEXT, engagement_type_ids JSONB, priority INT,
    created_at
);
```

## T4102 — internal/recommendations/ (Engine)
- Context layer: thresholds (сгорание, новый период)
- Segment layer: criteria (грейд/стаж/семья)
- Replace stub `[]Recommendation{}`

## T4103 — Recommendations + CEL
- Segment criteria как CEL expression
- TagResolver integration

## T4104 — API
```
GET /api/v1/recommendations          — для Dashboard
GET /admin/recommendations/debug?userId= — debug endpoint
```

## T4105 — Admin API
```
CRUD rules, contexts, segments
Priority, enable/disable
```

## Критерии приёмки
- [ ] Все 5 задач
- [ ] Context + segment layers
- [ ] CEL criteria
- [ ] Dashboard recommendations
