// Package auth — handlers и сервис для аутентификации LKFL.
//
// Реализует OIDC-поток через Keycloak: login redirect, callback, logout,
// создание/обновление пользователей при первом входе.
//
// Архитектура:
//
//	handler.go   — HTTP handlers (LoginRedirect, LoginCallback, Logout, Me)
//	service.go   — бизнес-логика (CreateOrUpdateUser, GetUserByKeycloakSub)
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"lkfl/internal/tenant"
	"lkfl/internal/user"
	sharedauth "lkfl/shared/pkg/auth"
)

// TenantResolver — интерфейс для разрешения tenant по slug.
// Используется в auth callback когда tenant middleware не установлен.
type TenantResolver interface {
	ResolveBySlug(ctx context.Context, slug string) (tenant.Tenant, error)
}

// Service — бизнес-логика аутентификации.
type Service struct {
	userRepo          user.Repository
	tenantResolver    TenantResolver // опционально — для callback без tenant middleware
	defaultTenantSlug string         // tenant slug из Keycloak issuer (fallback)
}

// NewService создаёт Service.
func NewService(userRepo user.Repository) *Service {
	return &Service{userRepo: userRepo}
}

// WithTenantResolver добавляет resolver для tenant по slug.
// Используется когда auth callback вызывается без tenant middleware.
func (s *Service) WithTenantResolver(resolver TenantResolver) *Service {
	s.tenantResolver = resolver
	return s
}

// SetDefaultTenantSlug задаёт tenant slug для fallback при callback.
func (s *Service) SetDefaultTenantSlug(slug string) {
	s.defaultTenantSlug = slug
}

// CreateOrUpdateUser — first login → create, subsequent → update.
//
// При первом входе пользователя создаёт запись в БД на основе OIDC claims.
// При повторных входах обновляет данные (email, имя, фамилия) из Keycloak.
//
// Tenant ID берётся из context (установлен tenant middleware ранее).
func (s *Service) CreateOrUpdateUser(ctx context.Context, claims *sharedauth.Claims, roles []string) (user.User, error) {
	// Ищем существующего пользователя по keycloak_sub
	existing, err := s.userRepo.GetByKeycloakSub(ctx, claims.Subject)
	if err == nil {
		// Пользователь существует — обновляем данные из Keycloak
		existing.Email = claims.Email
		existing.FirstName = claims.GivenName
		existing.LastName = claims.FamilyName
		updated, errUpdate := s.userRepo.Update(ctx, existing)

		// Синхронизируем роли из Keycloak (добавляем новые)
		for _, role := range roles {
			_, errAdd := s.userRepo.AddRole(ctx, existing.ID, role, nil)
			if errAdd != nil && !strings.Contains(errAdd.Error(), "duplicate") {
				slog.Warn("failed to sync role (may exist already)", "user_id", existing.ID, "role", role, "error", errAdd)
			}
		}

		if errUpdate != nil {
			return user.User{}, fmt.Errorf("update user: %w", errUpdate)
		}
		return updated, nil
	}

	// Пользователь не найден — создаём нового
	newUser := user.User{
		Email:       claims.Email,
		FirstName:   claims.GivenName,
		LastName:    claims.FamilyName,
		KeycloakSub: claims.Subject,
		Status:      user.StatusActive,
	}

	// Tenant ID из context (установлен tenant middleware)
	tid := tenant.TenantIDFromContext(ctx)
	if tid != uuid.Nil {
		newUser.TenantID = tid
	} else if s.tenantResolver != nil && s.defaultTenantSlug != "" {
		// Fallback: резолвим tenant по slug из Keycloak issuer
		t, err := s.tenantResolver.ResolveBySlug(ctx, s.defaultTenantSlug)
		if err == nil {
			newUser.TenantID = t.ID
		}
	}

	created, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return user.User{}, fmt.Errorf("create user: %w", err)
	}

	// Назначаем роли из Keycloak.
	// Для новых пользователей назначаем базовую роль "employee" + роли из Keycloak.
	// Для существующих — синхронизируем роли из Keycloak (добавляем новые).
	allRoles := roles
	if len(allRoles) == 0 {
		// Если Keycloak не вернул роли, назначаем роль сотрудника по умолчанию
		allRoles = []string{user.RoleEmployee}
	}
	for _, role := range allRoles {
		_, err := s.userRepo.AddRole(ctx, created.ID, role, nil)
		if err != nil && !strings.Contains(err.Error(), "duplicate") {
			// Неблокирующее логирование — если роль уже есть, это OK
			slog.Warn("failed to assign role (may exist already)", "user_id", created.ID, "role", role, "error", err)
		}
	}

	return created, nil
}

// GetUserByKeycloakSub возвращает пользователя по Keycloak subject ID.
func (s *Service) GetUserByKeycloakSub(ctx context.Context, keycloakSub string) (user.User, error) {
	return s.userRepo.GetByKeycloakSub(ctx, keycloakSub)
}
