# T2001 — Отчёт

## Статус

✅ Выполнено

## Созданные файлы

| Файл | Описание |
|------|----------|
| `backend/migrations/20260526130000_engagement.sql` | Up-миграция: 3 таблицы + индексы |
| `backend/migrations/20260526130000_engagement.sql.down` | Down-миграция: DROP в обратном порядке |

## Подтверждение критериев приёмки

- [x] `engagement_categories` — slug, name, icon, sort_order, tenant FK (CASCADE)
- [x] `engagement_types` — CHECK на type (benefit/activity) и status (draft/active/promo/hidden/completed)
- [x] `engagement_offers` — cost_cents BIGINT NOT NULL DEFAULT 0, tenant FK (CASCADE)
- [x] Unique indexes: `idx_cat_tenant_slug`, `idx_eng_type_tenant_slug`
- [x] Filter indexes: `idx_eng_type_category`, `idx_eng_type_status`, `idx_eng_type_type`
- [x] Tenant isolation indexes: `idx_cat_tenant`, `idx_eng_type_tenant`, `idx_offer_tenant`
- [x] Down миграция: DROP в обратном порядке (offers → types → categories)

## Паттерны проекта

- Формат: раздельные файлы `.sql` + `.sql.down` (без goose аннотаций, как T1901)
- Schema: `lkfl_platform`
- UUID PRIMARY KEY DEFAULT gen_random_uuid()
- TIMESTAMPTZ NOT NULL DEFAULT NOW()
- JSONB DEFAULT '{}'
- BIGINT для cost_cents
- ON DELETE CASCADE для всех FK

## Замечания

Миграция следует формату T1901 (раздельные up/down файлы без goose аннотаций), что отличается от формата в brief.md (goose аннотации в одном файле). Референс T1901 выбран как актуальный паттерн проекта.
