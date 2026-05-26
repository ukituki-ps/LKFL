# T3201-T3207 — Gamification (ADR-023)

## Веха

M32-gamification

## T3201 — Migrations: Gamification
```sql
CREATE TABLE lkfl_platform.achievement_types (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255), description TEXT,
    icon VARCHAR(50), cel_rule_id UUID FK, reward_cents BIGINT,
    status VARCHAR(20), created_at, updated_at
);
CREATE TABLE lkfl_platform.achievement_grants (
    id UUID PK, user_id UUID FK, achievement_type_id UUID FK,
    granted_at TIMESTAMPTZ, metadata JSONB
);
CREATE UNIQUE INDEX idx_achievement_grant_user ON lkfl_platform.achievement_grants (user_id, achievement_type_id);
CREATE TABLE lkfl_platform.loyalty_levels (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255), sort_order INT,
    min_points BIGINT, max_points BIGINT, bonus_multiplier DECIMAL(3,2),
    created_at
);
CREATE TABLE lkfl_platform.user_loyalty_levels (
    id UUID PK, user_id UUID FK UNIQUE, loyalty_level_id UUID FK,
    current_points BIGINT DEFAULT 0, assigned_at TIMESTAMPTZ
);
CREATE TABLE lkfl_platform.trigger_rules (
    id UUID PK, tenant_id UUID FK, name VARCHAR(255),
    cel_rule_id UUID FK, reward_cents BIGINT, status VARCHAR(20),
    created_at
);
```

## T3202 — internal/gamification/ (Engine)
- AchievementGrantEngine: CEL domain `game.*`
- LoyaltyLevelManager: level calculation, transitions
- Trigger bonuses: event-based rewards

## T3203 — Gamification + Billing integration
- Achievement grant → credit
- Trigger bonus → credit
- Unit tests

## T3204 — XLSX импорт
- excelize для парсинга
- Массовое начисление ачивок
- Validation + error report

## T3205 — API: Gamification
```
GET /api/v1/achievements     — мои ачивки
GET /api/v1/loyalty          — уровень, прогресс
```

## T3206 — Admin API
```
CRUD achievement_types, loyalty_levels, trigger_rules
POST /admin/gamification/xlsx-import
```

## T3207 — Prometheus metrics
- `gamification_achievements_granted_total`
- `gamification_loyalty_level_changes_total`

## Критерии приёмки
- [ ] Все 7 задач
- [ ] Achievement → credit
- [ ] XLSX импорт
- [ ] Loyalty levels
