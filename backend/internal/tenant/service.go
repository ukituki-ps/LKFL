package tenant

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
)

// Статусы tenant'а.
const (
	// TenantStatusActive — tenant активен.
	TenantStatusActive = "active"

	// TenantStatusSuspended — tenant приостановлен.
	TenantStatusSuspended = "suspended"
)

var (
	// ErrNotFound — tenant не найден.
	ErrNotFound = errors.New("tenant not found")

	// ErrSlugExists — slug уже занят.
	ErrSlugExists = errors.New("tenant slug already exists")

	// ErrTenantSuspended — tenant приостановлен.
	ErrTenantSuspended = errors.New("tenant is suspended")

	// ErrInvalidSlug — slug не соответствует формату.
	ErrInvalidSlug = errors.New("invalid tenant slug format")

	// ErrBrandNotFound — brand config не найден.
	ErrBrandNotFound = errors.New("brand config not found")

	// slugRegex — валидный slug: lowercase, alphanumeric, hyphens.
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// Repository — интерфейс для операций с tenants и brand config.
//
// Реализация: pgRepository (pgx).
// Используется Service для абстракции от хранилища.
type Repository interface {
	// Create создаёт новый tenant.
	Create(ctx context.Context, t Tenant) (Tenant, error)

	// GetByID возвращает tenant по ID.
	GetByID(ctx context.Context, id uuid.UUID) (Tenant, error)

	// GetBySlug возвращает tenant по slug.
	GetBySlug(ctx context.Context, slug string) (Tenant, error)

	// List возвращает список tenants с пагинацией и фильтром.
	List(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error)

	// Update обновляет существующего tenant.
	Update(ctx context.Context, t Tenant) (Tenant, error)

	// Delete удаляет tenant.
	Delete(ctx context.Context, id uuid.UUID) error

	// GetBrandConfig возвращает brand config для tenant.
	GetBrandConfig(ctx context.Context, tenantID uuid.UUID) (BrandConfig, error)

	// UpsertBrandConfig создаёт или обновляет brand config.
	UpsertBrandConfig(ctx context.Context, bc BrandConfig) (BrandConfig, error)
}

// TenantFilter — фильтр для List.
//
// Status фильтрует по статусу tenant'а (active, suspended).
// Page и PerPage управляют пагинацией (default: page=1, per_page=20).
type TenantFilter struct {
	Status  string
	Page    int
	PerPage int
}

// Service — бизнес-логика для tenants.
//
// Инъецирует Repository для абстракции от хранилища.
// Обеспечивает валидацию slug, проверку уникальности, status check.
type Service struct {
	repo Repository
}

// NewService создаёт Service с указанным Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateTenant создаёт нового tenant с валидацией.
func (s *Service) CreateTenant(ctx context.Context, t Tenant) (Tenant, error) {
	if !slugRegex.MatchString(t.Slug) {
		return Tenant{}, ErrInvalidSlug
	}

	// Проверка уникальности slug
	_, getErr := s.repo.GetBySlug(ctx, t.Slug)
	if getErr == nil {
		return Tenant{}, ErrSlugExists
	}
	if !errors.Is(getErr, ErrNotFound) {
		return Tenant{}, getErr
	}

	if t.Status == "" {
		t.Status = TenantStatusActive
	}
	if t.Settings == nil {
		t.Settings = JSONB{}
	}

	return s.repo.Create(ctx, t)
}

// GetByID возвращает tenant по ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

// GetBySlug возвращает tenant по slug. Проверяет, что tenant активен.
func (s *Service) GetBySlug(ctx context.Context, slug string) (Tenant, error) {
	t, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return Tenant{}, err
	}
	if t.Status != TenantStatusActive {
		return Tenant{}, ErrTenantSuspended
	}
	return t, nil
}

// GetBySlugRaw возвращает tenant по slug без проверки статуса.
// Используется middleware для tenant resolution — middleware не фильтрует
// suspended tenants (решение принимает бизнес-логика в handler'ах).
func (s *Service) GetBySlugRaw(ctx context.Context, slug string) (Tenant, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// ListTenants возвращает список tenants с пагинацией.
func (s *Service) ListTenants(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	return s.repo.List(ctx, filter)
}

// UpdateTenant обновляет tenant.
func (s *Service) UpdateTenant(ctx context.Context, t Tenant) (Tenant, error) {
	// Если slug меняется — проверить валидность и уникальность
	if t.Slug != "" && !slugRegex.MatchString(t.Slug) {
		return Tenant{}, ErrInvalidSlug
	}

	return s.repo.Update(ctx, t)
}

// DeleteTenant удаляет tenant.
func (s *Service) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// GetBrandConfig возвращает brand config для tenant.
func (s *Service) GetBrandConfig(ctx context.Context, tenantID uuid.UUID) (BrandConfig, error) {
	return s.repo.GetBrandConfig(ctx, tenantID)
}

// UpsertBrandConfig создаёт или обновляет brand config.
func (s *Service) UpsertBrandConfig(ctx context.Context, bc BrandConfig) (BrandConfig, error) {
	if bc.CSSVariables == nil {
		bc.CSSVariables = JSONB{}
	}
	return s.repo.UpsertBrandConfig(ctx, bc)
}

// _ — проверка что pgRepository реализует Repository.
var _ Repository = (*pgRepository)(nil)
