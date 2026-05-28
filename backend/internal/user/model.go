// Package user — CRUD пользователей, профиль, admin-операции.
//
// Реализация:
//
//	model.go       — User, UserProfile, Account, UserRole, UserFilter
//	repository.go  — Repository interface + pgx реализация
//	service.go     — бизнес-логика (валидация, status transitions)
//	handler.go     — HTTP handlers (employee + admin)
package user

import (
	"time"

	"github.com/google/uuid"
	"lkfl/internal/tenant"
)

// User — модель пользователя платформы.
type User struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenant_id"`
	Email       string       `json:"email"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	Phone       string       `json:"phone,omitempty"`
	Status      string       `json:"status"`
	KeycloakSub string       `json:"-"` // никогда не экспортировать в API
	Metadata    tenant.JSONB `json:"metadata,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// UserProfile — публичный профиль (без keycloak_sub и tenant_id).
type UserProfile struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ToProfile преобразует User в публичный профиль.
func (u User) ToProfile() UserProfile {
	return UserProfile{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

// Account — аккаунт пользователя (баланс, настройки).
type Account struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	TotalBalance int64        `json:"total_balance"`
	Settings     tenant.JSONB `json:"settings"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// UserRole — роль пользователя (RBAC).
type UserRole struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Role      string     `json:"role"`
	GrantedAt time.Time  `json:"granted_at"`
	GrantedBy *uuid.UUID `json:"granted_by,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UserFilter — фильтр для админ-списка пользователей.
type UserFilter struct {
	Status  string
	Search  string // поиск по email, first_name, last_name (ILIKE)
	Page    int
	PerPage int
}

// Статусы пользователя.
const (
	StatusActive      = "active"
	StatusDeactivated = "deactivated"
	StatusDeleted     = "deleted"
)

// Роли пользователя (RBAC).
const (
	RoleEmployee       = "employee"
	RoleHR             = "hr"
	RoleCatalogManager = "catalog_manager"
	RoleAdmin          = "admin"
)
