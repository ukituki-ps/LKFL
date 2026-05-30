// Package tenant — управление tenants и white-label брендированием.
//
// Системный пакет для multi-tenancy: CRUD tenants, brand config,
// tenant resolver middleware, tenant isolation для query builder.
//
// Архитектура:
//
//	model.go       — Tenant, BrandConfig, JSONB типы
//	repository.go  — Repository interface + pgx реализация
//	service.go     — бизнес-логика (валидация slug, status check)
//	handler.go     — admin HTTP handlers
//	middleware.go  — chi middleware (tenant resolver)
//	isolation.go   — tenant isolation (query builder)
package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgRepository — pgx реализация Repository.
type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт pgx repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// Create создаёт новый tenant.
func (r *pgRepository) Create(ctx context.Context, t Tenant) (Tenant, error) {
	query := `
		INSERT INTO lkfl_platform.tenants (slug, name, status, settings)
		VALUES ($1, $2, $3, $4)
		RETURNING id, slug, name, status, settings, created_at, updated_at
	`

	var tenant Tenant
	err := r.pool.QueryRow(ctx, query,
		t.Slug, t.Name, t.Status, t.Settings,
	).Scan(
		&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status,
		&tenant.Settings, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	if err != nil {
		return Tenant{}, fmt.Errorf("tenant repository: create: %w", err)
	}

	return tenant, nil
}

// GetByID возвращает tenant по ID.
func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (Tenant, error) {
	query := `
		SELECT id, slug, name, status, settings, created_at, updated_at
		FROM lkfl_platform.tenants
		WHERE id = $1
	`

	var t Tenant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Slug, &t.Name, &t.Status,
		&t.Settings, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tenant{}, ErrNotFound
		}
		return Tenant{}, fmt.Errorf("tenant repository: get by id: %w", err)
	}

	return t, nil
}

// GetBySlug возвращает tenant по slug.
func (r *pgRepository) GetBySlug(ctx context.Context, slug string) (Tenant, error) {
	query := `
		SELECT id, slug, name, status, settings, created_at, updated_at
		FROM lkfl_platform.tenants
		WHERE slug = $1
	`

	var t Tenant
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&t.ID, &t.Slug, &t.Name, &t.Status,
		&t.Settings, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tenant{}, ErrNotFound
		}
		return Tenant{}, fmt.Errorf("tenant repository: get by slug: %w", err)
	}

	return t, nil
}

// List возвращает список tenants с пагинацией.
func (r *pgRepository) List(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
	// Count query
	var countQuery string
	var countArgs []any
	countQuery = "SELECT COUNT(*) FROM lkfl_platform.tenants"

	if filter.Status != "" {
		countQuery += " WHERE status = $1"
		countArgs = append(countArgs, filter.Status)
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("tenant repository: count: %w", err)
	}

	// Data query
	var dataQuery string
	var dataArgs []any
	argNum := 1

	dataQuery = "SELECT id, slug, name, status, settings, created_at, updated_at FROM lkfl_platform.tenants"

	if filter.Status != "" {
		dataQuery += fmt.Sprintf(" WHERE status = $%d", argNum)
		dataArgs = append(dataArgs, filter.Status)
		argNum++
	}

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

	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	dataArgs = append(dataArgs, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("tenant repository: list: %w", err)
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.Status,
			&t.Settings, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("tenant repository: scan: %w", err)
		}
		tenants = append(tenants, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("tenant repository: iterate: %w", err)
	}

	if tenants == nil {
		tenants = []Tenant{}
	}

	return tenants, total, nil
}

// Update обновляет существующего tenant.
func (r *pgRepository) Update(ctx context.Context, t Tenant) (Tenant, error) {
	query := `
		UPDATE lkfl_platform.tenants
		SET slug = $1, name = $2, status = $3, settings = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, slug, name, status, settings, created_at, updated_at
	`

	var tenant Tenant
	err := r.pool.QueryRow(ctx, query,
		t.Slug, t.Name, t.Status, t.Settings, t.ID,
	).Scan(
		&tenant.ID, &tenant.Slug, &tenant.Name, &tenant.Status,
		&tenant.Settings, &tenant.CreatedAt, &tenant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tenant{}, ErrNotFound
		}
		return Tenant{}, fmt.Errorf("tenant repository: update: %w", err)
	}

	return tenant, nil
}

// Delete удаляет tenant.
func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM lkfl_platform.tenants WHERE id = $1`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("tenant repository: delete: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetBrandConfig возвращает brand config для tenant.
func (r *pgRepository) GetBrandConfig(ctx context.Context, tenantID uuid.UUID) (BrandConfig, error) {
	query := `
		SELECT id, tenant_id, primary_color, secondary_color, logo_url, favicon_url,
		       brand_name, css_variables, meta_title, meta_description, created_at, updated_at
		FROM lkfl_platform.tenant_brand_config
		WHERE tenant_id = $1
	`

	var bc BrandConfig
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&bc.ID, &bc.TenantID, &bc.PrimaryColor, &bc.SecondaryColor,
		&bc.LogoURL, &bc.FaviconURL, &bc.BrandName,
		&bc.CSSVariables, &bc.MetaTitle, &bc.MetaDescription,
		&bc.CreatedAt, &bc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return BrandConfig{}, ErrBrandNotFound
		}
		return BrandConfig{}, fmt.Errorf("tenant repository: get brand config: %w", err)
	}

	return bc, nil
}

// UpsertBrandConfig создаёт или обновляет brand config.
func (r *pgRepository) UpsertBrandConfig(ctx context.Context, bc BrandConfig) (BrandConfig, error) {
	query := `
		INSERT INTO lkfl_platform.tenant_brand_config (
			tenant_id, primary_color, secondary_color, logo_url, favicon_url,
			brand_name, css_variables, meta_title, meta_description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tenant_id) DO UPDATE SET
			primary_color = EXCLUDED.primary_color,
			secondary_color = EXCLUDED.secondary_color,
			logo_url = EXCLUDED.logo_url,
			favicon_url = EXCLUDED.favicon_url,
			brand_name = EXCLUDED.brand_name,
			css_variables = EXCLUDED.css_variables,
			meta_title = EXCLUDED.meta_title,
			meta_description = EXCLUDED.meta_description,
			updated_at = NOW()
		RETURNING id, tenant_id, primary_color, secondary_color, logo_url, favicon_url,
		          brand_name, css_variables, meta_title, meta_description, created_at, updated_at
	`

	var result BrandConfig
	err := r.pool.QueryRow(ctx, query,
		bc.TenantID, bc.PrimaryColor, bc.SecondaryColor,
		bc.LogoURL, bc.FaviconURL, bc.BrandName,
		bc.CSSVariables, bc.MetaTitle, bc.MetaDescription,
	).Scan(
		&result.ID, &result.TenantID, &result.PrimaryColor, &result.SecondaryColor,
		&result.LogoURL, &result.FaviconURL, &result.BrandName,
		&result.CSSVariables, &result.MetaTitle, &result.MetaDescription,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return BrandConfig{}, fmt.Errorf("tenant repository: upsert brand config: %w", err)
	}

	return result, nil
}
