-- Migration: 001_add_user_survey_attributes.sql
-- ADR-025, M13 T1301: Survey Engine — теги из ответов опросов
--
-- user_survey_attributes хранит теги, вычисленные из ответов на опросы.
-- ADR-024 compliant: tenant_id — обязательное поле для multi-tenancy изоляции.
--
-- Lifecycle (обслуживается survey/TagMapper):
--   1. INSERT ON CONFLICT DO UPDATE SET weight = GREATEST(EXCLUDED.weight, weight)
--   2. DELETE orphaned тегов при изменении ответа внутри survey_offer_id
--   3. Cross-survey aggregation: MAX(weight) GROUP BY tag_key (TagResolver.AggregateSurveyTags)

-- Создание таблицы
CREATE TABLE IF NOT EXISTS user_survey_attributes (
    id              SERIAL PRIMARY KEY,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    survey_offer_id UUID NOT NULL REFERENCES engagement_offers(id) ON DELETE CASCADE,
    tag_key         VARCHAR(128) NOT NULL,
    tag_value       VARCHAR(256) NOT NULL DEFAULT '',
    weight          FLOAT NOT NULL DEFAULT 1.0 CHECK (weight >= 0 AND weight <= 1),
    question_id     VARCHAR(64) NOT NULL,
    answered_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, user_id, survey_offer_id, tag_key)
);

-- Индексы: tenant_id первым для partition pruning
CREATE INDEX IF NOT EXISTS idx_survey_attrs_tenant_user ON user_survey_attributes(tenant_id, user_id);
CREATE INDEX IF NOT EXISTS idx_survey_attrs_tenant_tag ON user_survey_attributes(tenant_id, tag_key);
CREATE INDEX IF NOT EXISTS idx_survey_attrs_tenant_offer ON user_survey_attributes(tenant_id, survey_offer_id);
