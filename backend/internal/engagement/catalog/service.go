package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrInvalidFilter — невалидный фильтр.
var ErrInvalidFilter = errors.New("invalid filter parameters")

// ErrInvalidStatus — недопустимый статус.
var ErrInvalidStatus = errors.New("invalid status")

// ErrInvalidType — недопустимый тип энгейджмента.
var ErrInvalidType = errors.New("invalid engagement type")

// ErrTenantMismatch — тип энгейджмента принадлежит другому tenant'у.
var ErrTenantMismatch = errors.New("engagement type belongs to another tenant")

// validStatuses — допустимые статусы.
var validStatuses = map[string]bool{
	StatusDraft:     true,
	StatusActive:    true,
	StatusPromo:     true,
	StatusHidden:    true,
	StatusCompleted: true,
}

// validTypes — допустимые типы.
var validTypes = map[string]bool{
	TypeBenefit:  true,
	TypeActivity: true,
}

// cacheListResult — структура для сериализации результата ListTypes в кэш.
type cacheListResult struct {
	Data  []EngagementType `json:"data"`
	Total int64            `json:"total"`
}

// cacheTypeResult — структура для сериализации результата GetTypeByID в кэш.
type cacheTypeResult struct {
	Data EngagementType `json:"data"`
}

// cacheCategoriesResult — структура для сериализации результата GetCategories в кэш.
type cacheCategoriesResult struct {
	Data []EngagementCategory `json:"data"`
}

// Service — бизнес-логика для каталога энгейджментов.
type Service struct {
	repo  Repository
	cache *Cache
}

// NewService создаёт Service.
// Если cache == nil, кэширование отключено.
func NewService(repo Repository, cache *Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

// ListTypes возвращает список типов энгейджментов с фильтрами и пагинацией.
// Публичный метод: по умолчанию показывает active и promo.
// Использует Redis кэш если доступен.
func (s *Service) ListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error) {
	// Валидация пагинации
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	// Валидация фильтров
	if filter.Type != "" && !validTypes[filter.Type] {
		return nil, 0, ErrInvalidType
	}

	// Попытка получить из кэша
	if s.cache != nil {
		if cached, ok := s.cache.GetList(ctx, filter); ok {
			var result cacheListResult
			if err := json.Unmarshal(cached, &result); err == nil {
				return result.Data, result.Total, nil
			}
			// Deserialization error — fallback к DB
		}
	}

	// Статус по умолчанию: если не задан, показываем active+promo
	// (репозиторий обрабатывает это по умолчанию)

	types, total, err := s.repo.ListTypes(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Сохранить в кэш
	if s.cache != nil {
		data, _ := json.Marshal(cacheListResult{Data: types, Total: total})
		_ = s.cache.SetList(ctx, filter, data)
	}

	return types, total, nil
}

// GetTypeByID возвращает тип энгейджмента по ID.
// Использует Redis кэш если доступен.
func (s *Service) GetTypeByID(ctx context.Context, id uuid.UUID) (EngagementType, error) {
	t, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		return EngagementType{}, err
	}

	// Загружаем офферы
	offers, err := s.repo.GetOffersByType(ctx, id)
	if err != nil {
		// Не критично — возвращаем тип без офферов
		offers = []EngagementOffer{}
	}
	t.Offers = offers

	// Сохранить в кэш (используем tenant_id типа для tenant isolation в key)
	if s.cache != nil {
		data, _ := json.Marshal(cacheTypeResult{Data: t})
		_ = s.cache.SetType(ctx, t.TenantID.String(), id.String(), data)
	}

	return t, nil
}

// GetTypeByTenantID возвращает тип энгейджмента по ID с проверкой tenant isolation.
// Использует Redis кэш если доступен.
func (s *Service) GetTypeByTenantID(ctx context.Context, tenantID, id uuid.UUID) (EngagementType, error) {
	// Попытка получить из кэша (key включает tenant_id для isolation)
	if s.cache != nil {
		if cached, ok := s.cache.GetType(ctx, tenantID.String(), id.String()); ok {
			var result cacheTypeResult
			if err := json.Unmarshal(cached, &result); err == nil {
				// Проверка tenant isolation
				if result.Data.TenantID == tenantID {
					return result.Data, nil
				}
			}
			// Tenant mismatch или deserialization error — идём в DB
		}
	}

	t, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		return EngagementType{}, err
	}

	// Tenant isolation: тип должен принадлежать запрашивающему tenant'у
	if t.TenantID != tenantID {
		return EngagementType{}, ErrTenantMismatch
	}

	// Загружаем офферы
	offers, err := s.repo.GetOffersByType(ctx, id)
	if err != nil {
		// Не критично — возвращаем тип без офферов
		offers = []EngagementOffer{}
	}
	t.Offers = offers

	// Сохранить в кэш
	if s.cache != nil {
		data, _ := json.Marshal(cacheTypeResult{Data: t})
		_ = s.cache.SetType(ctx, tenantID.String(), id.String(), data)
	}

	return t, nil
}

// GetCategories возвращает список категорий tenant'а.
// Использует Redis кэш если доступен.
func (s *Service) GetCategories(ctx context.Context, tenantID uuid.UUID) ([]EngagementCategory, error) {
	// Попытка получить из кэша
	if s.cache != nil {
		if cached, ok := s.cache.GetCategories(ctx, tenantID.String()); ok {
			var result cacheCategoriesResult
			if err := json.Unmarshal(cached, &result); err == nil {
				return result.Data, nil
			}
			// Deserialization error — fallback к DB
		}
	}

	categories, err := s.repo.GetCategories(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Сохранить в кэш
	if s.cache != nil {
		data, _ := json.Marshal(cacheCategoriesResult{Data: categories})
		_ = s.cache.SetCategories(ctx, tenantID.String(), data)
	}

	return categories, nil
}

// GetOffersByType возвращает офферы для типа энгейджмента.
func (s *Service) GetOffersByType(ctx context.Context, typeID uuid.UUID) ([]EngagementOffer, error) {
	return s.repo.GetOffersByType(ctx, typeID)
}

// AdminCreateCategory создаёт новую категорию.
// Проверяет уникальность slug внутри tenant'а.
func (s *Service) AdminCreateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error) {
	if c.Slug == "" || c.Name == "" {
		return EngagementCategory{}, ErrInvalidFilter
	}

	// Проверка уникальности slug
	existing, err := s.repo.GetCategories(ctx, c.TenantID)
	if err != nil {
		return EngagementCategory{}, fmt.Errorf("check category slug: %w", err)
	}
	for _, cat := range existing {
		if cat.Slug == c.Slug {
			return EngagementCategory{}, ErrDuplicateSlug
		}
	}

	return s.repo.AdminCreateCategory(ctx, c)
}

// AdminUpdateCategory обновляет категорию.
// Проверяет уникальность slug (если slug меняется).
func (s *Service) AdminUpdateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error) {
	if c.Slug == "" || c.Name == "" {
		return EngagementCategory{}, ErrInvalidFilter
	}

	// Проверка уникальности slug (если slug меняется)
	existing, err := s.repo.GetCategories(ctx, c.TenantID)
	if err != nil {
		return EngagementCategory{}, fmt.Errorf("check category slug: %w", err)
	}
	for _, cat := range existing {
		if cat.Slug == c.Slug && cat.ID != c.ID {
			return EngagementCategory{}, ErrDuplicateSlug
		}
	}

	return s.repo.AdminUpdateCategory(ctx, c)
}

// AdminCreateType создаёт новый тип энгейджмента.
// Проверяет существование категории и уникальность slug.
func (s *Service) AdminCreateType(ctx context.Context, t EngagementType) (EngagementType, error) {
	if t.Slug == "" || t.Name == "" {
		return EngagementType{}, ErrInvalidFilter
	}

	// Валидация типа
	if t.Type != "" && !validTypes[t.Type] {
		return EngagementType{}, ErrInvalidType
	}
	if t.Type == "" {
		t.Type = TypeBenefit
	}

	// Валидация статуса
	if t.Status != "" && !validStatuses[t.Status] {
		return EngagementType{}, ErrInvalidStatus
	}
	if t.Status == "" {
		t.Status = StatusDraft
	}

	// Проверка существования категории
	if t.CategoryID != uuid.Nil {
		categories, err := s.repo.GetCategories(ctx, t.TenantID)
		if err != nil {
			return EngagementType{}, fmt.Errorf("check category: %w", err)
		}
		categoryExists := false
		for _, cat := range categories {
			if cat.ID == t.CategoryID {
				categoryExists = true
				break
			}
		}
		if !categoryExists {
			return EngagementType{}, ErrCategoryNotFound
		}
	}

	// Проверка уникальности slug внутри tenant'а
	if _, err := s.repo.GetTypeBySlug(ctx, t.TenantID, t.Slug); err == nil {
		// GetBySlug вернул тип — slug уже занят
		return EngagementType{}, ErrDuplicateSlug
	} else if !errors.Is(err, ErrNotFound) {
		// Реальная ошибка БД
		return EngagementType{}, fmt.Errorf("check type slug: %w", err)
	}

	return s.repo.AdminCreateType(ctx, t)
}

// AdminUpdateType обновляет тип энгейджмента.
func (s *Service) AdminUpdateType(ctx context.Context, t EngagementType) (EngagementType, error) {
	if t.Slug == "" || t.Name == "" {
		return EngagementType{}, ErrInvalidFilter
	}

	// Валидация типа
	if t.Type != "" && !validTypes[t.Type] {
		return EngagementType{}, ErrInvalidType
	}

	// Валидация статуса
	if t.Status != "" && !validStatuses[t.Status] {
		return EngagementType{}, ErrInvalidStatus
	}

	// Проверка существования типа
	existing, err := s.repo.GetTypeByID(ctx, t.ID)
	if err != nil {
		return EngagementType{}, err
	}

	// Проверка уникальности slug (если slug меняется)
	if existing.Slug != t.Slug {
		if _, err := s.repo.GetTypeBySlug(ctx, t.TenantID, t.Slug); err == nil {
			return EngagementType{}, ErrDuplicateSlug
		} else if !errors.Is(err, ErrNotFound) {
			return EngagementType{}, fmt.Errorf("check type slug: %w", err)
		}
	}

	// Проверка существования категории (если указана)
	if t.CategoryID != uuid.Nil {
		categories, err := s.repo.GetCategories(ctx, t.TenantID)
		if err != nil {
			return EngagementType{}, fmt.Errorf("check category: %w", err)
		}
		categoryExists := false
		for _, cat := range categories {
			if cat.ID == t.CategoryID {
				categoryExists = true
				break
			}
		}
		if !categoryExists {
			return EngagementType{}, ErrCategoryNotFound
		}
	}

	return s.repo.AdminUpdateType(ctx, t)
}

// AdminDeleteType удаляет тип энгейджмента (soft delete через status=hidden).
// TODO: в F2 добавить проверку 0 активаций перед удалением.
func (s *Service) AdminDeleteType(ctx context.Context, id uuid.UUID) error {
	// Проверка существования
	_, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: проверить что нет активных активаций (F2)

	return s.repo.AdminDeleteType(ctx, id)
}

// AdminCreateOffer создаёт новый оффер.
// Проверяет существование типа энгейджмента.
func (s *Service) AdminCreateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error) {
	if o.Name == "" {
		return EngagementOffer{}, ErrInvalidFilter
	}

	// Проверка существования типа
	_, err := s.repo.GetTypeByID(ctx, o.EngagementTypeID)
	if err != nil {
		return EngagementOffer{}, fmt.Errorf("engagement type not found: %w", err)
	}

	return s.repo.AdminCreateOffer(ctx, o)
}

// AdminUpdateOffer обновляет оффер.
func (s *Service) AdminUpdateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error) {
	if o.Name == "" {
		return EngagementOffer{}, ErrInvalidFilter
	}

	return s.repo.AdminUpdateOffer(ctx, o)
}

// AdminDeleteOffer удаляет оффер.
func (s *Service) AdminDeleteOffer(ctx context.Context, id uuid.UUID) error {
	return s.repo.AdminDeleteOffer(ctx, id)
}

// AdminListTypes возвращает список всех типов энгейджментов (все статусы).
// Admin метод: показывает все статусы, не только active+promo.
func (s *Service) AdminListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error) {
	// Валидация пагинации
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	// Валидация фильтров
	if filter.Type != "" && !validTypes[filter.Type] {
		return nil, 0, ErrInvalidType
	}

	return s.repo.AdminListTypes(ctx, filter)
}

// AdminDeleteCategory удаляет категорию.
func (s *Service) AdminDeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.repo.AdminDeleteCategory(ctx, id)
}
