// Package catalog — каталог энгейджментов (льготы/активности).
//
// model.go       — EngagementType, EngagementCategory, EngagementOffer, CatalogFilter
// repository.go  — Repository interface + pgx реализация
// service.go     — бизнес-логика (валидация, status transitions)
package catalog

import (
	"time"

	"github.com/google/uuid"
	"lkfl/internal/tenant"
)

// EngagementType — тип энгейджмента (конкретная льгота или активность).
type EngagementType struct {
	ID           uuid.UUID           `json:"id"`
	TenantID     uuid.UUID           `json:"tenant_id"`
	CategoryID   uuid.UUID           `json:"category_id"`
	Slug         string              `json:"slug"`
	Name         string              `json:"name"`
	Description  *string             `json:"description,omitempty"`
	Type         string              `json:"type"`   // benefit, activity
	Status       string              `json:"status"` // draft, active, promo, hidden, completed
	CostCents    *int64              `json:"cost_cents,omitempty"`
	ProviderName *string             `json:"provider_name,omitempty"`
	ImageURL     *string             `json:"image_url,omitempty"`
	Metadata     tenant.JSONB        `json:"metadata,omitempty"`
	Category     *EngagementCategory `json:"category,omitempty"`
	Offers       []EngagementOffer   `json:"offers,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

// EngagementCategory — категория энгейджментов.
type EngagementCategory struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Icon      *string   `json:"icon,omitempty"`
	SortOrder int       `json:"sort_order"`
}

// EngagementOffer — оффер (тариф/план внутри типа).
type EngagementOffer struct {
	ID               uuid.UUID    `json:"id"`
	TenantID         uuid.UUID    `json:"tenant_id"`
	EngagementTypeID uuid.UUID    `json:"engagement_type_id"`
	Name             string       `json:"name"`
	Description      *string      `json:"description,omitempty"`
	CostCents        int64        `json:"cost_cents"`
	BillingRuleID    *uuid.UUID   `json:"billing_rule_id,omitempty"`
	Metadata         tenant.JSONB `json:"metadata,omitempty"`
	SortOrder        int          `json:"sort_order"`
}

// CatalogFilter — фильтр для списка энгейджментов.
type CatalogFilter struct {
	TenantID uuid.UUID
	Type     string // benefit, activity (optional)
	Status   string // active, promo (optional)
	Category string // category slug (optional)
	Search   string // ILIKE name/description (optional)
	Page     int
	PerPage  int
}

// Статусы.
const (
	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusPromo     = "promo"
	StatusHidden    = "hidden"
	StatusCompleted = "completed"
)

// Типы.
const (
	TypeBenefit  = "benefit"
	TypeActivity = "activity"
)
