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
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tenant — модель tenant'а.
type Tenant struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Settings  JSONB     `json:"settings"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BrandConfig — white-label брендирование tenant'а.
type BrandConfig struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	PrimaryColor    string    `json:"primary_color"`
	SecondaryColor  string    `json:"secondary_color"`
	LogoURL         *string   `json:"logo_url,omitempty"`
	FaviconURL      *string   `json:"favicon_url,omitempty"`
	BrandName       *string   `json:"brand_name,omitempty"`
	CSSVariables    JSONB     `json:"css_variables"`
	MetaTitle       *string   `json:"meta_title,omitempty"`
	MetaDescription *string   `json:"meta_description,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// JSONB — тип для JSONB полей PostgreSQL.
type JSONB map[string]any

// Get возвращает значение по ключу.
func (j JSONB) Get(key string) any {
	if j == nil {
		return nil
	}
	return j[key]
}

// GetString возвращает строковое значение по ключу или пустую строку.
func (j JSONB) GetString(key string) string {
	if j == nil {
		return ""
	}
	if v, ok := j[key].(string); ok {
		return v
	}
	return ""
}

// Scan реализует sql.Scanner для JSONB.
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONB: expected []byte or string, got %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// Value реализует driver.Valuer для JSONB.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		j = JSONB{}
	}
	return json.Marshal(j)
}
