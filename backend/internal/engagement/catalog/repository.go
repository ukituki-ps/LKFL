package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound — тип энгейджмента не найден.
var ErrNotFound = errors.New("engagement type not found")

// ErrCategoryNotFound — категория не найдена.
var ErrCategoryNotFound = errors.New("category not found")

// ErrOfferNotFound — оффер не найден.
var ErrOfferNotFound = errors.New("offer not found")

// ErrDuplicateSlug — slug уже занят.
var ErrDuplicateSlug = errors.New("slug already exists")

// Repository — интерфейс для операций с каталогом энгейджментов.
type Repository interface {
	// Public
	ListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error)
	GetTypeByID(ctx context.Context, id uuid.UUID) (EngagementType, error)
	GetCategories(ctx context.Context, tenantID uuid.UUID) ([]EngagementCategory, error)

	// Admin Categories
	AdminCreateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error)
	AdminUpdateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error)
	AdminDeleteCategory(ctx context.Context, id uuid.UUID) error

	// Admin Types
	AdminCreateType(ctx context.Context, t EngagementType) (EngagementType, error)
	AdminUpdateType(ctx context.Context, t EngagementType) (EngagementType, error)
	AdminDeleteType(ctx context.Context, id uuid.UUID) error
	AdminListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error)

	// Slug lookup
	GetTypeBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (EngagementType, error)

	// Offers
	GetOffersByType(ctx context.Context, typeID uuid.UUID) ([]EngagementOffer, error)
	AdminCreateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error)
	AdminUpdateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error)
	AdminDeleteOffer(ctx context.Context, id uuid.UUID) error
}

// pgRepository — pgx реализация Repository.
type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт pgx repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// ListTypes возвращает список типов энгейджментов с фильтрами и пагинацией.
// По умолчанию фильтрует status IN ('active', 'promo').
// Promo items сортируются первыми.
func (r *pgRepository) ListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error) {
	// Базовый query с LEFT JOIN категорий
	baseQuery := `
		SELECT
			et.id, et.tenant_id, et.category_id, et.slug, et.name, et.description,
			et.type, et.status, et.cost_cents, et.provider_name, et.image_url,
			et.metadata, et.created_at, et.updated_at,
			ec.id, ec.tenant_id, ec.slug, ec.name, ec.icon, ec.sort_order
		FROM lkfl_platform.engagement_types et
		LEFT JOIN lkfl_platform.engagement_categories ec ON et.category_id = ec.id
		WHERE et.tenant_id = $1
	`

	args := []any{filter.TenantID}
	argNum := 2

	// Статусы по умолчанию: active, promo
	statuses := []string{StatusActive, StatusPromo}
	if filter.Status != "" {
		statuses = []string{filter.Status}
	}
	statusPlaceholders := make([]string, len(statuses))
	for i, s := range statuses {
		statusPlaceholders[i] = fmt.Sprintf("$%d", argNum)
		args = append(args, s)
		argNum++
	}
	query := baseQuery + fmt.Sprintf(" AND et.status IN (%s)", strings.Join(statusPlaceholders, ", "))

	// Фильтр по типу (benefit/activity)
	if filter.Type != "" {
		query += fmt.Sprintf(" AND et.type = $%d", argNum)
		args = append(args, filter.Type)
		argNum++
	}

	// Фильтр по категории (slug)
	if filter.Category != "" {
		query += fmt.Sprintf(" AND ec.slug = $%d", argNum)
		args = append(args, filter.Category)
		argNum++
	}

	// Поиск по name/description (ILIKE)
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query += fmt.Sprintf(" AND (et.name ILIKE $%d OR et.description ILIKE $%d)", argNum, argNum)
		args = append(args, searchPattern)
		argNum++
	}

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM lkfl_platform.engagement_types et
		LEFT JOIN lkfl_platform.engagement_categories ec ON et.category_id = ec.id
		WHERE et.tenant_id = $1
	`
	countArgs := []any{filter.TenantID}
	cArgNum := 2

	countStatusPlaceholders := make([]string, len(statuses))
	for i, s := range statuses {
		countStatusPlaceholders[i] = fmt.Sprintf("$%d", cArgNum)
		countArgs = append(countArgs, s)
		cArgNum++
	}
	countQuery += fmt.Sprintf(" AND et.status IN (%s)", strings.Join(countStatusPlaceholders, ", "))

	if filter.Type != "" {
		countQuery += fmt.Sprintf(" AND et.type = $%d", cArgNum)
		countArgs = append(countArgs, filter.Type)
		cArgNum++
	}
	if filter.Category != "" {
		countQuery += fmt.Sprintf(" AND ec.slug = $%d", cArgNum)
		countArgs = append(countArgs, filter.Category)
		cArgNum++
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		countQuery += fmt.Sprintf(" AND (et.name ILIKE $%d OR et.description ILIKE $%d)", cArgNum, cArgNum)
		countArgs = append(countArgs, searchPattern)
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("catalog repository: count: %w", err)
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

	// ORDER BY: promo first, then by category sort_order, then name
	query += fmt.Sprintf(`
		ORDER BY
			CASE et.status WHEN 'promo' THEN 0 ELSE 1 END,
			COALESCE(ec.sort_order, 9999),
			et.name
		LIMIT $%d OFFSET $%d
	`, argNum, argNum+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("catalog repository: list: %w", err)
	}
	defer rows.Close()

	var types []EngagementType
	for rows.Next() {
		var t EngagementType
		var cat EngagementCategory
		var catID *uuid.UUID
		var catSlug *string
		var catName *string
		var catIcon *string
		var catSortOrder *int
		var descPtr *string

		err := rows.Scan(
			&t.ID, &t.TenantID, &t.CategoryID, &t.Slug, &t.Name, &descPtr,
			&t.Type, &t.Status, &t.CostCents, &t.ProviderName, &t.ImageURL,
			&t.Metadata, &t.CreatedAt, &t.UpdatedAt,
			&catID, &cat.TenantID, &catSlug, &catName, &catIcon, &catSortOrder,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("catalog repository: scan type: %w", err)
		}
		t.Description = descPtr

		// Распаковываем nullable поля категории (LEFT JOIN → NULL если нет категории)
		if catID != nil {
			cat.ID = *catID
			if catSlug != nil {
				cat.Slug = *catSlug
			}
			if catName != nil {
				cat.Name = *catName
			}
			if catIcon != nil {
				cat.Icon = catIcon
			}
			if catSortOrder != nil {
				cat.SortOrder = *catSortOrder
			}
			t.Category = &cat
		}

		types = append(types, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("catalog repository: iterate types: %w", err)
	}

	if types == nil {
		types = []EngagementType{}
	}

	return types, total, nil
}

// GetTypeByID возвращает тип энгейджмента по ID.
func (r *pgRepository) GetTypeByID(ctx context.Context, id uuid.UUID) (EngagementType, error) {
	query := `
		SELECT id, tenant_id, category_id, slug, name, description,
			type, status, cost_cents, provider_name, image_url,
			metadata, created_at, updated_at
		FROM lkfl_platform.engagement_types
		WHERE id = $1
	`

	var t EngagementType
	var descPtr *string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.TenantID, &t.CategoryID, &t.Slug, &t.Name, &descPtr,
		&t.Type, &t.Status, &t.CostCents, &t.ProviderName, &t.ImageURL,
		&t.Metadata, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EngagementType{}, ErrNotFound
		}
		return EngagementType{}, fmt.Errorf("catalog repository: get type by id: %w", err)
	}
	t.Description = descPtr

	return t, nil
}

// GetCategories возвращает список категорий tenant'а, отсортированных по sort_order.
func (r *pgRepository) GetCategories(ctx context.Context, tenantID uuid.UUID) ([]EngagementCategory, error) {
	query := `
		SELECT id, tenant_id, slug, name, icon, sort_order
		FROM lkfl_platform.engagement_categories
		WHERE tenant_id = $1
		ORDER BY sort_order
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("catalog repository: get categories: %w", err)
	}
	defer rows.Close()

	var categories []EngagementCategory
	for rows.Next() {
		var c EngagementCategory
		var iconPtr *string
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Slug, &c.Name, &iconPtr, &c.SortOrder); err != nil {
			return nil, fmt.Errorf("catalog repository: scan category: %w", err)
		}
		c.Icon = iconPtr
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repository: iterate categories: %w", err)
	}

	if categories == nil {
		categories = []EngagementCategory{}
	}

	return categories, nil
}

// AdminCreateCategory создаёт новую категорию.
func (r *pgRepository) AdminCreateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error) {
	query := `
		INSERT INTO lkfl_platform.engagement_categories (tenant_id, slug, name, icon, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, tenant_id, slug, name, icon, sort_order
	`

	var cat EngagementCategory
	var iconPtr *string
	err := r.pool.QueryRow(ctx, query,
		c.TenantID, c.Slug, c.Name, c.Icon, c.SortOrder,
	).Scan(&cat.ID, &cat.TenantID, &cat.Slug, &cat.Name, &iconPtr, &cat.SortOrder)
	if err != nil {
		return EngagementCategory{}, fmt.Errorf("catalog repository: create category: %w", err)
	}
	cat.Icon = iconPtr

	return cat, nil
}

// AdminUpdateCategory обновляет категорию.
func (r *pgRepository) AdminUpdateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error) {
	query := `
		UPDATE lkfl_platform.engagement_categories
		SET slug = $1, name = $2, icon = $3, sort_order = $4
		WHERE id = $5
		RETURNING id, tenant_id, slug, name, icon, sort_order
	`

	var cat EngagementCategory
	var iconPtr *string
	err := r.pool.QueryRow(ctx, query,
		c.Slug, c.Name, c.Icon, c.SortOrder, c.ID,
	).Scan(&cat.ID, &cat.TenantID, &cat.Slug, &cat.Name, &iconPtr, &cat.SortOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EngagementCategory{}, ErrCategoryNotFound
		}
		return EngagementCategory{}, fmt.Errorf("catalog repository: update category: %w", err)
	}
	cat.Icon = iconPtr

	return cat, nil
}

// AdminCreateType создаёт новый тип энгейджмента.
func (r *pgRepository) AdminCreateType(ctx context.Context, t EngagementType) (EngagementType, error) {
	query := `
		INSERT INTO lkfl_platform.engagement_types
			(tenant_id, category_id, slug, name, description, type, status, cost_cents,
			 provider_name, image_url, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, tenant_id, category_id, slug, name, description, type, status,
			cost_cents, provider_name, image_url, metadata, created_at, updated_at
	`

	var et EngagementType
	var descPtr *string
	err := r.pool.QueryRow(ctx, query,
		t.TenantID, t.CategoryID, t.Slug, t.Name, t.Description, t.Type, t.Status,
		t.CostCents, t.ProviderName, t.ImageURL, t.Metadata,
	).Scan(
		&et.ID, &et.TenantID, &et.CategoryID, &et.Slug, &et.Name, &descPtr,
		&et.Type, &et.Status, &et.CostCents, &et.ProviderName, &et.ImageURL,
		&et.Metadata, &et.CreatedAt, &et.UpdatedAt,
	)
	if err != nil {
		return EngagementType{}, fmt.Errorf("catalog repository: create type: %w", err)
	}
	et.Description = descPtr

	return et, nil
}

// AdminUpdateType обновляет тип энгейджмента.
func (r *pgRepository) AdminUpdateType(ctx context.Context, t EngagementType) (EngagementType, error) {
	query := `
		UPDATE lkfl_platform.engagement_types
		SET category_id = $1, slug = $2, name = $3, description = $4, type = $5,
			status = $6, cost_cents = $7, provider_name = $8, image_url = $9,
			metadata = $10, updated_at = NOW()
		WHERE id = $11
		RETURNING id, tenant_id, category_id, slug, name, description, type, status,
			cost_cents, provider_name, image_url, metadata, created_at, updated_at
	`

	var et EngagementType
	var descPtr *string
	err := r.pool.QueryRow(ctx, query,
		t.CategoryID, t.Slug, t.Name, t.Description, t.Type, t.Status,
		t.CostCents, t.ProviderName, t.ImageURL, t.Metadata, t.ID,
	).Scan(
		&et.ID, &et.TenantID, &et.CategoryID, &et.Slug, &et.Name, &descPtr,
		&et.Type, &et.Status, &et.CostCents, &et.ProviderName, &et.ImageURL,
		&et.Metadata, &et.CreatedAt, &et.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EngagementType{}, ErrNotFound
		}
		return EngagementType{}, fmt.Errorf("catalog repository: update type: %w", err)
	}
	et.Description = descPtr

	return et, nil
}

// AdminDeleteCategory удаляет категорию.
func (r *pgRepository) AdminDeleteCategory(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM lkfl_platform.engagement_categories
		WHERE id = $1
	`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("catalog repository: delete category: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

// AdminListTypes возвращает список всех типов энгейджментов (все статусы).
// Admin метод: показывает все статусы, не только active+promo.
func (r *pgRepository) AdminListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error) {
	// Базовый query с LEFT JOIN категорий
	baseQuery := `
		SELECT
			et.id, et.tenant_id, et.category_id, et.slug, et.name, et.description,
			et.type, et.status, et.cost_cents, et.provider_name, et.image_url,
			et.metadata, et.created_at, et.updated_at,
			ec.id, ec.tenant_id, ec.slug, ec.name, ec.icon, ec.sort_order
		FROM lkfl_platform.engagement_types et
		LEFT JOIN lkfl_platform.engagement_categories ec ON et.category_id = ec.id
		WHERE et.tenant_id = $1
	`

	args := []any{filter.TenantID}
	argNum := 2

	// Admin: по умолчанию все статусы; если задан фильтр — только он
	var statusPlaceholders []string
	if filter.Status != "" {
		statusPlaceholders = []string{fmt.Sprintf("$%d", argNum)}
		args = append(args, filter.Status)
		argNum++
	}

	if len(statusPlaceholders) > 0 {
		baseQuery += fmt.Sprintf(" AND et.status IN (%s)", strings.Join(statusPlaceholders, ", "))
	}

	// Фильтр по типу (benefit/activity)
	if filter.Type != "" {
		baseQuery += fmt.Sprintf(" AND et.type = $%d", argNum)
		args = append(args, filter.Type)
		argNum++
	}

	// Фильтр по категории (slug)
	if filter.Category != "" {
		baseQuery += fmt.Sprintf(" AND ec.slug = $%d", argNum)
		args = append(args, filter.Category)
		argNum++
	}

	// Поиск по name/description (ILIKE)
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		baseQuery += fmt.Sprintf(" AND (et.name ILIKE $%d OR et.description ILIKE $%d)", argNum, argNum)
		args = append(args, searchPattern)
		argNum++
	}

	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM lkfl_platform.engagement_types et
		LEFT JOIN lkfl_platform.engagement_categories ec ON et.category_id = ec.id
		WHERE et.tenant_id = $1
	`
	countArgs := []any{filter.TenantID}
	cArgNum := 2

	if filter.Status != "" {
		countQuery += fmt.Sprintf(" AND et.status = $%d", cArgNum)
		countArgs = append(countArgs, filter.Status)
		cArgNum++
	}
	if filter.Type != "" {
		countQuery += fmt.Sprintf(" AND et.type = $%d", cArgNum)
		countArgs = append(countArgs, filter.Type)
		cArgNum++
	}
	if filter.Category != "" {
		countQuery += fmt.Sprintf(" AND ec.slug = $%d", cArgNum)
		countArgs = append(countArgs, filter.Category)
		cArgNum++
	}
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		countQuery += fmt.Sprintf(" AND (et.name ILIKE $%d OR et.description ILIKE $%d)", cArgNum, cArgNum)
		countArgs = append(countArgs, searchPattern)
	}

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("catalog repository: count: %w", err)
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

	// ORDER BY: promo first, then by category sort_order, then name
	baseQuery += fmt.Sprintf(`
		ORDER BY
			CASE et.status WHEN 'promo' THEN 0 ELSE 1 END,
			COALESCE(ec.sort_order, 9999),
			et.name
		LIMIT $%d OFFSET $%d
	`, argNum, argNum+1)
	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("catalog repository: list: %w", err)
	}
	defer rows.Close()

	var types []EngagementType
	for rows.Next() {
		var t EngagementType
		var cat EngagementCategory
		var catID *uuid.UUID
		var catSlug *string
		var catName *string
		var catIcon *string
		var catSortOrder *int
		var descPtr *string

		err := rows.Scan(
			&t.ID, &t.TenantID, &t.CategoryID, &t.Slug, &t.Name, &descPtr,
			&t.Type, &t.Status, &t.CostCents, &t.ProviderName, &t.ImageURL,
			&t.Metadata, &t.CreatedAt, &t.UpdatedAt,
			&catID, &cat.TenantID, &catSlug, &catName, &catIcon, &catSortOrder,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("catalog repository: scan type: %w", err)
		}
		t.Description = descPtr

		if catID != nil {
			cat.ID = *catID
			if catSlug != nil {
				cat.Slug = *catSlug
			}
			if catName != nil {
				cat.Name = *catName
			}
			if catIcon != nil {
				cat.Icon = catIcon
			}
			if catSortOrder != nil {
				cat.SortOrder = *catSortOrder
			}
			t.Category = &cat
		}

		types = append(types, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("catalog repository: iterate types: %w", err)
	}

	if types == nil {
		types = []EngagementType{}
	}

	return types, total, nil
}

// AdminDeleteType удаляет тип энгейджмента (soft delete через status=hidden).
func (r *pgRepository) AdminDeleteType(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE lkfl_platform.engagement_types
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	res, err := r.pool.Exec(ctx, query, StatusHidden, id)
	if err != nil {
		return fmt.Errorf("catalog repository: delete type: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetOffersByType возвращает все офферы для типа энгейджмента.
func (r *pgRepository) GetOffersByType(ctx context.Context, typeID uuid.UUID) ([]EngagementOffer, error) {
	query := `
		SELECT id, tenant_id, engagement_type_id, name, description, cost_cents,
			billing_rule_id, metadata, sort_order
		FROM lkfl_platform.engagement_offers
		WHERE engagement_type_id = $1
		ORDER BY sort_order
	`

	rows, err := r.pool.Query(ctx, query, typeID)
	if err != nil {
		return nil, fmt.Errorf("catalog repository: get offers: %w", err)
	}
	defer rows.Close()

	var offers []EngagementOffer
	for rows.Next() {
		var o EngagementOffer
		var descPtr *string
		if err := rows.Scan(
			&o.ID, &o.TenantID, &o.EngagementTypeID, &o.Name, &descPtr,
			&o.CostCents, &o.BillingRuleID, &o.Metadata, &o.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("catalog repository: scan offer: %w", err)
		}
		o.Description = descPtr
		offers = append(offers, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog repository: iterate offers: %w", err)
	}

	if offers == nil {
		offers = []EngagementOffer{}
	}

	return offers, nil
}

// AdminCreateOffer создаёт новый оффер.
func (r *pgRepository) AdminCreateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error) {
	query := `
		INSERT INTO lkfl_platform.engagement_offers
			(tenant_id, engagement_type_id, name, description, cost_cents,
			 billing_rule_id, metadata, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, tenant_id, engagement_type_id, name, description, cost_cents,
			billing_rule_id, metadata, sort_order
	`

	var offer EngagementOffer
	var descPtr *string
	err := r.pool.QueryRow(ctx, query,
		o.TenantID, o.EngagementTypeID, o.Name, o.Description, o.CostCents,
		o.BillingRuleID, o.Metadata, o.SortOrder,
	).Scan(
		&offer.ID, &offer.TenantID, &offer.EngagementTypeID, &offer.Name, &descPtr,
		&offer.CostCents, &offer.BillingRuleID, &offer.Metadata, &offer.SortOrder,
	)
	if err != nil {
		return EngagementOffer{}, fmt.Errorf("catalog repository: create offer: %w", err)
	}
	offer.Description = descPtr

	return offer, nil
}

// AdminUpdateOffer обновляет оффер.
func (r *pgRepository) AdminUpdateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error) {
	query := `
		UPDATE lkfl_platform.engagement_offers
		SET name = $1, description = $2, cost_cents = $3, billing_rule_id = $4,
			metadata = $5, sort_order = $6
		WHERE id = $7
		RETURNING id, tenant_id, engagement_type_id, name, description, cost_cents,
			billing_rule_id, metadata, sort_order
	`

	var offer EngagementOffer
	var descPtr *string
	err := r.pool.QueryRow(ctx, query,
		o.Name, o.Description, o.CostCents, o.BillingRuleID,
		o.Metadata, o.SortOrder, o.ID,
	).Scan(
		&offer.ID, &offer.TenantID, &offer.EngagementTypeID, &offer.Name, &descPtr,
		&offer.CostCents, &offer.BillingRuleID, &offer.Metadata, &offer.SortOrder,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EngagementOffer{}, ErrOfferNotFound
		}
		return EngagementOffer{}, fmt.Errorf("catalog repository: update offer: %w", err)
	}
	offer.Description = descPtr

	return offer, nil
}

// AdminDeleteOffer удаляет оффер.
func (r *pgRepository) AdminDeleteOffer(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM lkfl_platform.engagement_offers
		WHERE id = $1
	`

	res, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("catalog repository: delete offer: %w", err)
	}

	if res.RowsAffected() == 0 {
		return ErrOfferNotFound
	}

	return nil
}

// GetTypeBySlug возвращает тип энгейджмента по tenant_id + slug.
// Используется для проверки уникальности slug.
func (r *pgRepository) GetTypeBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (EngagementType, error) {
	query := `
		SELECT id, tenant_id, category_id, slug, name, description,
			type, status, cost_cents, provider_name, image_url,
			metadata, created_at, updated_at
		FROM lkfl_platform.engagement_types
		WHERE tenant_id = $1 AND slug = $2
	`

	var t EngagementType
	var descPtr *string
	err := r.pool.QueryRow(ctx, query, tenantID, slug).Scan(
		&t.ID, &t.TenantID, &t.CategoryID, &t.Slug, &t.Name, &descPtr,
		&t.Type, &t.Status, &t.CostCents, &t.ProviderName, &t.ImageURL,
		&t.Metadata, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EngagementType{}, ErrNotFound
		}
		return EngagementType{}, fmt.Errorf("catalog repository: get type by slug: %w", err)
	}
	t.Description = descPtr

	return t, nil
}

// Compile-time check: pgRepository реализует Repository.
var _ Repository = (*pgRepository)(nil)
