package cel

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// TagResolver — вычисление тегов пользователя из профиля + survey-ответов.
//
// Теги — вычисляемые (не хранимые): выводятся из первичных данных
// (user profile, survey responses). Кэшируются в Redis (cel: prefix, TTL 1h).
//
// ADR-021: CEL Engine — TagResolver как часть cel/ пакета.
// doc/архитектура/теги.md — полный каталог тегов и формул вычисления.
type TagResolver struct {
	db    *sql.DB
	cache *redis.Client
}

// NewTagResolver создаёт экземпляр TagResolver.
func NewTagResolver(db *sql.DB, cache *redis.Client) *TagResolver {
	return &TagResolver{db: db, cache: cache}
}

// User — минимальный профиль пользователя для вычисления тегов.
type User struct {
	ID             string  `db:"id"`
	Location       string  `db:"location"`
	YearsOfService float64 `db:"years_of_service"`
	Grade          string  `db:"grade"`
	Department     string  `db:"department"`
	Status         string  `db:"status"`
	TenantID       string  `db:"tenant_id"`
}

// Resolve вычисляет все теги для пользователя: profile-теги + survey-теги.
//
//  Профиль: is_remote, is_newbie, is_senior, tenure_1y, tenure_3y, tenure_5y
//  Survey: interest:sport, lifestyle:family, sport_intensity и др.
//
// Результат кэшируется в Redis: tag_cache:{tenant_id}:{user_id}, TTL 1h.
func (r *TagResolver) Resolve(ctx context.Context, user *User) map[string]interface{} {
	// Проверяем кэш
	cacheKey := "cel:tag_cache:" + user.TenantID + ":" + user.ID
	if cached, err := r.cache.Get(ctx, cacheKey).Result(); err == nil {
		var tags map[string]interface{}
		if err := json.Unmarshal([]byte(cached), &tags); err == nil {
			return tags
		}
	}

	// Вычисляем profile-теги
	tags := r.computeFromProfile(user)

	// Вычисляем survey-теги (M13 T1301)
	surveyTags := r.AggregateSurveyTags(ctx, user.TenantID, user.ID)
	for k, v := range surveyTags {
		tags[k] = v
	}

	// Сохраняем в кэш
	if data, err := json.Marshal(tags); err == nil {
		r.cache.Set(ctx, cacheKey, data, 3600) // TTL 1 час
	}

	return tags
}

// computeFromProfile вычисляет теги из HR-профиля пользователя.
//
// Каталог тегов (doc/архитектура/теги.md):
//   is_remote    — location == "remote"
//   is_newbie    — стаж < 3 месяцев (0.25 года)
//   is_senior    — стаж >= 5 лет
//   tenure_1y    — стаж >= 1 год
//   tenure_3y    — стаж >= 3 года
//   tenure_5y    — стаж >= 5 лет
func (r *TagResolver) computeFromProfile(user *User) map[string]interface{} {
	tags := make(map[string]interface{})

	if user.Location == "remote" {
		tags["is_remote"] = true
	}

	if user.YearsOfService < 0.25 {
		tags["is_newbie"] = true
	}

	if user.YearsOfService >= 5 {
		tags["is_senior"] = true
	}

	if user.YearsOfService >= 1 {
		tags["tenure_1y"] = true
	}

	if user.YearsOfService >= 3 {
		tags["tenure_3y"] = true
	}

	if user.YearsOfService >= 5 {
		tags["tenure_5y"] = true
	}

	return tags
}

// AggregateSurveyTags агрегирует теги из survey-ответов пользователя.
//
// SELECT tag_key, tag_value, MAX(weight)
// FROM user_survey_attributes
// WHERE tenant_id = ? AND user_id = ?
// GROUP BY tag_key, tag_value
//
// Возвращает map[string]interface{}, совместимый с CELContext.Tags.
// ADR-025: M13 T1301.
func (r *TagResolver) AggregateSurveyTags(
	ctx context.Context,
	tenantId string,
	userId string,
) map[string]interface{} {
	query := `
		SELECT tag_key, tag_value, MAX(weight) AS max_weight
		FROM user_survey_attributes
		WHERE tenant_id = $1 AND user_id = $2
		GROUP BY tag_key, tag_value
	`

	rows, err := r.db.QueryContext(ctx, query, tenantId, userId)
	if err != nil {
		return nil
	}
	defer rows.Close()

	tags := make(map[string]interface{})

	for rows.Next() {
		var tagKey, tagValue string
		var maxWeight float64

		if err := rows.Scan(&tagKey, &tagValue, &maxWeight); err != nil {
			continue
		}

		tags[tagKey] = maxTagValue{
			Value:  tagValue,
			Weight: maxWeight,
		}
	}

	return tags
}

// maxTagValue — значение тега с весом из survey (максимальный вес при перекрёстных опросах).
type maxTagValue struct {
	Value  any     `json:"value"`
	Weight float64 `json:"weight"`
}

// ResolveForBatch — оптимизированная версия для batch-операций.
// Вычисляет теги для списка пользователей.
func (r *TagResolver) ResolveForBatch(ctx context.Context, users []User) []map[string]interface{} {
	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = r.Resolve(ctx, &user)
	}
	return result
}

// InvalidateCache удаляет кэш тегов пользователя из Redis.
//
// Вызывать при:
//   - Обновление user profile (grade, location, department)
//   - MapSurveyAnswers (survey-ответы изменились)
//   - HR sync batch (изменился профиль)
//
// Ключ: tag_cache:{tenant_id}:{user_id}
func (r *TagResolver) InvalidateCache(ctx context.Context, tenantId string, userId string) error {
	cacheKey := "cel:tag_cache:" + tenantId + ":" + userId
	return r.cache.Del(ctx, cacheKey).Err()
}
