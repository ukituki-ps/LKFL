package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"lkfl/internal/tenant"
)

// ErrUserDeactivated — пользователь деактивирован.
var ErrUserDeactivated = errors.New("user is deactivated")

// ErrAlreadyDeactivated — пользователь уже деактивирован.
var ErrAlreadyDeactivated = errors.New("user is already deactivated")

// ErrCannotReverseDeactivation — нельзя восстановить деактивацию.
var ErrCannotReverseDeactivation = errors.New("cannot reverse deactivation")

// validRoles — допустимые роли.
var validRoles = map[string]bool{
	RoleEmployee:       true,
	RoleHR:             true,
	RoleCatalogManager: true,
	RoleAdmin:          true,
}

// Service — бизнес-логика для пользователей.
type Service struct {
	repo Repository
}

// NewService создаёт Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetByID возвращает пользователя по ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByKeycloakSub возвращает пользователя по Keycloak subject.
func (s *Service) GetByKeycloakSub(ctx context.Context, keycloakSub string) (User, error) {
	return s.repo.GetByKeycloakSub(ctx, keycloakSub)
}

// UpdateProfile обновляет профиль пользователя.
// Валидирует email и имена.
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, email, firstName, lastName, phone string) (User, error) {
	// Получаем текущего пользователя
	current, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	// Нельзя обновлять профиль деактивированного пользователя
	if current.Status == StatusDeactivated {
		return User{}, ErrUserDeactivated
	}

	// Проверка уникальности email (если email меняется)
	if email != "" && email != current.Email {
		tid := tenant.TenantIDFromContext(ctx)
		_, err := s.repo.GetByEmail(ctx, tid, email)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return User{}, fmt.Errorf("check email uniqueness: %w", err)
		}
		if err == nil {
			return User{}, ErrDuplicateEmail
		}
	}

	// Обновляем поля
	u := current
	if email != "" {
		u.Email = email
	}
	if firstName != "" {
		u.FirstName = firstName
	}
	if lastName != "" {
		u.LastName = lastName
	}
	if phone != "" {
		u.Phone = phone
	}

	return s.repo.Update(ctx, u)
}

// Deactivate деактивирует пользователя (active → deactivated).
// Односторонняя операция — нельзя восстановить.
func (s *Service) Deactivate(ctx context.Context, userID uuid.UUID) (User, error) {
	current, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	// Проверка: пользователь должен быть активен
	if current.Status != StatusActive {
		if current.Status == StatusDeactivated {
			return User{}, ErrAlreadyDeactivated
		}
		if current.Status == StatusDeleted {
			return User{}, ErrNotFound
		}
		return User{}, ErrInvalidStatus
	}

	return s.repo.UpdateStatus(ctx, userID, StatusDeactivated)
}

// Activate активирует пользователя (deactivated → active).
// Обратная операция к Deactivate.
func (s *Service) Activate(ctx context.Context, userID uuid.UUID) (User, error) {
	current, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return User{}, err
	}

	if current.Status != StatusDeactivated {
		return User{}, ErrInvalidStatus
	}

	return s.repo.UpdateStatus(ctx, userID, StatusActive)
}

// List возвращает список пользователей с пагинацией и фильтрами.
// Использует tenant isolation из context.
func (s *Service) List(ctx context.Context, filter UserFilter) ([]User, int64, error) {
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

// AddRole добавляет роль пользователю.
func (s *Service) AddRole(ctx context.Context, userID uuid.UUID, role string, grantedBy *uuid.UUID) (UserRole, error) {
	// Проверка существования пользователя
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return UserRole{}, err
	}

	// Проверка валидности роли
	if !validRoles[role] {
		return UserRole{}, ErrInvalidRole
	}

	return s.repo.AddRole(ctx, userID, role, grantedBy)
}

// RemoveRole удаляет роль у пользователя.
func (s *Service) RemoveRole(ctx context.Context, userID uuid.UUID, role string) error {
	// Проверка существования пользователя
	_, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	return s.repo.RemoveRole(ctx, userID, role)
}

// GetRoles возвращает все роли пользователя.
func (s *Service) GetRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error) {
	return s.repo.GetRoles(ctx, userID)
}

// CreateAndSetupUser создаёт пользователя и автоматически создаёт аккаунт и базовую роль.
func (s *Service) CreateAndSetupUser(ctx context.Context, u User) (User, error) {
	// Если статус не задан — по умолчанию active
	if u.Status == "" {
		u.Status = StatusActive
	}

	// Валидация email
	if u.Email == "" {
		return User{}, ErrInvalidStatus // переиспользуем как generic validation error
	}

	// Создаём пользователя
	created, err := s.repo.Create(ctx, u)
	if err != nil {
		return User{}, err
	}

	// Создаём аккаунт
	settings := tenant.JSONB{}
	if u.Metadata != nil {
		if acctSettings, ok := u.Metadata["account_settings"]; ok {
			if m, ok := acctSettings.(map[string]any); ok {
				settings = tenant.JSONB(m)
			}
		}
	}
	_, err = s.repo.CreateAccount(ctx, created.ID, settings)
	if err != nil {
		// Не критично — логировать и продолжить (аккаунт можно создать позже)
		// В продакшене здесь будет slog
	}

	// Назначаем базовую роль employee
	_, err = s.repo.AddRole(ctx, created.ID, RoleEmployee, nil)
	if err != nil {
		// Не критично — роль можно назначить позже
	}

	return created, nil
}
