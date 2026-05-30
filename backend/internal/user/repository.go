package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"lkfl/internal/tenant"
)

// ErrNotFound — пользователь не найден.
var ErrNotFound = errors.New("user not found")

// ErrAccountNotFound — аккаунт не найден.
var ErrAccountNotFound = errors.New("account not found")

// ErrRoleNotFound — роль не найдена.
var ErrRoleNotFound = errors.New("role not found")

// ErrDuplicateEmail — email уже занят в tenant'е.
var ErrDuplicateEmail = errors.New("email already exists")

// ErrInvalidStatus — недопустимый статус.
var ErrInvalidStatus = errors.New("invalid status")

// ErrInvalidRole — недопустимая роль.
var ErrInvalidRole = errors.New("invalid role")

// Repository — интерфейс для операций с пользователями.
type Repository interface {
	// Create создаёт нового пользователя.
	Create(ctx context.Context, u User) (User, error)

	// GetByID возвращает пользователя по ID.
	GetByID(ctx context.Context, id uuid.UUID) (User, error)

	// GetByKeycloakSub возвращает пользователя по Keycloak subject.
	GetByKeycloakSub(ctx context.Context, keycloakSub string) (User, error)

	// GetByEmail возвращает пользователя по email внутри tenant'а.
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error)

	// Update обновляет пользователя.
	Update(ctx context.Context, u User) (User, error)

	// UpdateStatus обновляет статус пользователя.
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (User, error)

	// List возвращает список пользователей с пагинацией и фильтрами.
	List(ctx context.Context, filter UserFilter) ([]User, int64, error)

	// CreateAccount создаёт аккаунт для пользователя.
	CreateAccount(ctx context.Context, userID uuid.UUID, settings tenant.JSONB) (Account, error)

	// GetAccountByUserID возвращает аккаунт по ID пользователя.
	GetAccountByUserID(ctx context.Context, userID uuid.UUID) (Account, error)

	// GetRoles возвращает все роли пользователя.
	GetRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error)

	// AddRole добавляет роль пользователю.
	AddRole(ctx context.Context, userID uuid.UUID, role string, grantedBy *uuid.UUID) (UserRole, error)

	// RemoveRole удаляет роль у пользователя.
	RemoveRole(ctx context.Context, userID uuid.UUID, role string) error
}

// pgRepository — pgx реализация Repository.
type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт pgx repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// Create создаёт нового пользователя.
func (r *pgRepository) Create(ctx context.Context, u User) (User, error) {
	query := `
		INSERT INTO lkfl_platform.users (tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata, created_at, updated_at
	`

	var user User
	var phonePtr *string
	err := r.pool.QueryRow(ctx, query,
		u.TenantID, u.Email, u.FirstName, u.LastName, u.Phone, u.Status, u.KeycloakSub, u.Metadata,
	).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName,
		&phonePtr, &user.Status, &user.KeycloakSub, &user.Metadata,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("user repository: create: %w", err)
	}
	if phonePtr != nil {
		user.Phone = *phonePtr
	}

	return user, nil
}

// GetByID возвращает пользователя по ID.
func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	query := `
		SELECT id, tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata, created_at, updated_at
		FROM lkfl_platform.users
		WHERE id = $1
	`

	var u User
	var phonePtr *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.FirstName, &u.LastName,
		&phonePtr, &u.Status, &u.KeycloakSub, &u.Metadata,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("user repository: get by id: %w", err)
	}
	if phonePtr != nil {
		u.Phone = *phonePtr
	}

	return u, nil
}

// GetByKeycloakSub возвращает пользователя по Keycloak subject.
func (r *pgRepository) GetByKeycloakSub(ctx context.Context, keycloakSub string) (User, error) {
	query := `
		SELECT id, tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata, created_at, updated_at
		FROM lkfl_platform.users
		WHERE keycloak_user_id = $1
	`

	var u User
	var phonePtr *string
	err := r.pool.QueryRow(ctx, query, keycloakSub).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.FirstName, &u.LastName,
		&phonePtr, &u.Status, &u.KeycloakSub, &u.Metadata,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("user repository: get by keycloak sub: %w", err)
	}
	if phonePtr != nil {
		u.Phone = *phonePtr
	}

	return u, nil
}

// GetByEmail возвращает пользователя по email внутри tenant'а.
func (r *pgRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error) {
	query := `
		SELECT id, tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata, created_at, updated_at
		FROM lkfl_platform.users
		WHERE tenant_id = $1 AND email = $2
	`

	var u User
	var phonePtr *string
	err := r.pool.QueryRow(ctx, query, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.FirstName, &u.LastName,
		&phonePtr, &u.Status, &u.KeycloakSub, &u.Metadata,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("user repository: get by email: %w", err)
	}
	if phonePtr != nil {
		u.Phone = *phonePtr
	}

	return u, nil
}

// Update обновляет пользователя.
func (r *pgRepository) Update(ctx context.Context, u User) (User, error) {
	query := `
		UPDATE lkfl_platform.users
		SET email = $1, first_name = $2, last_name = $3, phone = $4, metadata = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata, created_at, updated_at
	`

	var user User
	var phonePtr *string
	err := r.pool.QueryRow(ctx, query,
		u.Email, u.FirstName, u.LastName, u.Phone, u.Metadata, u.ID,
	).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName,
		&phonePtr, &user.Status, &user.KeycloakSub, &user.Metadata,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("user repository: update: %w", err)
	}
	if phonePtr != nil {
		user.Phone = *phonePtr
	}

	return user, nil
}

// UpdateStatus обновляет статус пользователя.
func (r *pgRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (User, error) {
	query := `
		UPDATE lkfl_platform.users
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata, created_at, updated_at
	`

	var user User
	var phonePtr *string
	err := r.pool.QueryRow(ctx, query, status, id).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.FirstName, &user.LastName,
		&phonePtr, &user.Status, &user.KeycloakSub, &user.Metadata,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("user repository: update status: %w", err)
	}

	return user, nil
}

// List возвращает список пользователей с пагинацией и фильтрами.
// Использует tenant.WithTenantID(ctx) для tenant isolation.
func (r *pgRepository) List(ctx context.Context, filter UserFilter) ([]User, int64, error) {
	// Базовый query
	baseQuery := `
		SELECT id, tenant_id, email, first_name, last_name, phone, status, keycloak_user_id, metadata, created_at, updated_at
		FROM lkfl_platform.users
	`

	// Tenant isolation — добавляет WHERE tenant_id = $1
	query, args := tenant.WithTenantID(ctx, baseQuery)
	argNum := 1
	if len(args) > 0 {
		argNum = len(args) + 1
	}

	// Фильтр по статусу
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}

	// Поиск по email, first_name, last_name (ILIKE)
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query += fmt.Sprintf(" AND (email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argNum, argNum, argNum)
		args = append(args, searchPattern)
		argNum++
	}

	// Count query — строим отдельно
	countBase := "SELECT COUNT(*) FROM lkfl_platform.users"
	countQuery, countArgs := tenant.WithTenantID(ctx, countBase)
	cArgNum := 1
	if len(countArgs) > 0 {
		cArgNum = len(countArgs) + 1
	}
	if filter.Status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", cArgNum)
		countArgs = append(countArgs, filter.Status)
		cArgNum++
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		countQuery += fmt.Sprintf(" AND (email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", cArgNum, cArgNum, cArgNum)
		countArgs = append(countArgs, searchPattern)
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("user repository: count: %w", err)
	}

	// Pagination
	perPage := filter.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	page := filter.Page
	if page < 1 {
		page = 1
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("user repository: list: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var phonePtr *string
		if err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.FirstName, &u.LastName,
			&phonePtr, &u.Status, &u.KeycloakSub, &u.Metadata,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("user repository: scan: %w", err)
		}
		if phonePtr != nil {
			u.Phone = *phonePtr
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("user repository: iterate: %w", err)
	}

	if users == nil {
		users = []User{}
	}

	return users, total, nil
}

// CreateAccount создаёт аккаунт для пользователя.
func (r *pgRepository) CreateAccount(ctx context.Context, userID uuid.UUID, settings tenant.JSONB) (Account, error) {
	query := `
		INSERT INTO lkfl_platform.accounts (user_id, total_balance, settings)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, total_balance, settings, created_at, updated_at
	`

	var account Account
	err := r.pool.QueryRow(ctx, query, userID, int64(0), settings).Scan(
		&account.ID, &account.UserID, &account.TotalBalance,
		&account.Settings, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		return Account{}, fmt.Errorf("user repository: create account: %w", err)
	}

	return account, nil
}

// GetAccountByUserID возвращает аккаунт по ID пользователя.
func (r *pgRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (Account, error) {
	query := `
		SELECT id, user_id, total_balance, settings, created_at, updated_at
		FROM lkfl_platform.accounts
		WHERE user_id = $1
	`

	var account Account
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&account.ID, &account.UserID, &account.TotalBalance,
		&account.Settings, &account.CreatedAt, &account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Account{}, ErrAccountNotFound
		}
		return Account{}, fmt.Errorf("user repository: get account: %w", err)
	}

	return account, nil
}

// GetRoles возвращает все роли пользователя.
func (r *pgRepository) GetRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error) {
	query := `
		SELECT id, user_id, role, granted_at, granted_by, expires_at
		FROM lkfl_platform.user_roles
		WHERE user_id = $1
		ORDER BY granted_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("user repository: get roles: %w", err)
	}
	defer rows.Close()

	var roles []UserRole
	for rows.Next() {
		var ur UserRole
		if err := rows.Scan(
			&ur.ID, &ur.UserID, &ur.Role,
			&ur.GrantedAt, &ur.GrantedBy, &ur.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("user repository: scan role: %w", err)
		}
		roles = append(roles, ur)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user repository: iterate roles: %w", err)
	}

	if roles == nil {
		roles = []UserRole{}
	}

	return roles, nil
}

// AddRole добавляет роль пользователю.
func (r *pgRepository) AddRole(ctx context.Context, userID uuid.UUID, role string, grantedBy *uuid.UUID) (UserRole, error) {
	query := `
		INSERT INTO lkfl_platform.user_roles (user_id, role, granted_by)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, role, granted_at, granted_by, expires_at
	`

	var ur UserRole
	err := r.pool.QueryRow(ctx, query, userID, role, grantedBy).Scan(
		&ur.ID, &ur.UserID, &ur.Role,
		&ur.GrantedAt, &ur.GrantedBy, &ur.ExpiresAt,
	)
	if err != nil {
		return UserRole{}, fmt.Errorf("user repository: add role: %w", err)
	}

	return ur, nil
}

// RemoveRole удаляет роль у пользователя.
func (r *pgRepository) RemoveRole(ctx context.Context, userID uuid.UUID, role string) error {
	query := `
		DELETE FROM lkfl_platform.user_roles
		WHERE user_id = $1 AND role = $2
	`

	res, err := r.pool.Exec(ctx, query, userID, role)
	if err != nil {
		return fmt.Errorf("user repository: remove role: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrRoleNotFound
	}

	return nil
}

// _ — проверка что pgRepository реализует Repository.
var _ Repository = (*pgRepository)(nil)
