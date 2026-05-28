package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// --- WithTenantID Tests ---

func TestWithTenantID_AddsWhereClause(t *testing.T) {
	tenantID := uuid.New()
	ctx := TenantContext(context.Background(), tenantID)

	query := "SELECT * FROM lkfl_platform.engagement_types"
	result, args := WithTenantID(ctx, query)

	expectedQuery := "SELECT * FROM lkfl_platform.engagement_types WHERE tenant_id = $1"
	if result != expectedQuery {
		t.Errorf("expected query:\n  %s\n got:\n  %s", expectedQuery, result)
	}

	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != tenantID {
		t.Errorf("expected arg[0] = %s, got %v", tenantID, args[0])
	}
}

func TestWithTenantID_AddsAndClause(t *testing.T) {
	tenantID := uuid.New()
	ctx := TenantContext(context.Background(), tenantID)

	query := "SELECT * FROM lkfl_platform.engagement_types WHERE status = 'active'"
	result, args := WithTenantID(ctx, query)

	expectedQuery := "SELECT * FROM lkfl_platform.engagement_types WHERE status = 'active' AND tenant_id = $1"
	if result != expectedQuery {
		t.Errorf("expected query:\n  %s\n got:\n  %s", expectedQuery, result)
	}

	if len(args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(args))
	}
	if args[0] != tenantID {
		t.Errorf("expected arg[0] = %s, got %v", tenantID, args[0])
	}
}

func TestWithTenantID_NilTenantID_NoFilter(t *testing.T) {
	ctx := context.WithValue(context.Background(), TenantIDKey, uuid.Nil)

	query := "SELECT * FROM lkfl_platform.engagement_types WHERE status = 'active'"
	result, args := WithTenantID(ctx, query)

	if result != query {
		t.Errorf("query should be unchanged, got: %s", result)
	}
	if args != nil {
		t.Errorf("expected nil args, got %v", args)
	}
}

func TestWithTenantID_NoTenantInContext(t *testing.T) {
	ctx := context.Background()

	query := "SELECT * FROM lkfl_platform.engagement_types"
	result, args := WithTenantID(ctx, query)

	if result != query {
		t.Errorf("query should be unchanged, got: %s", result)
	}
	if args != nil {
		t.Errorf("expected nil args, got %v", args)
	}
}

func TestWithTenantID_AdminContext(t *testing.T) {
	ctx := WithAdminTenant(context.Background())

	query := "SELECT * FROM lkfl_platform.tenants"
	result, args := WithTenantID(ctx, query)

	if result != query {
		t.Errorf("query should be unchanged for admin context, got: %s", result)
	}
	if args != nil {
		t.Errorf("expected nil args for admin context, got %v", args)
	}
}

func TestWithTenantID_CaseInsensitiveWhere(t *testing.T) {
	tenantID := uuid.New()
	ctx := TenantContext(context.Background(), tenantID)

	// Проверяем что WHERE определяется case-insensitive
	query := "SELECT * FROM lkfl_platform.engagement_types where status = 'active'"
	result, _ := WithTenantID(ctx, query)

	if result == query {
		t.Error("query should have been modified with AND tenant_id")
	}
	if result != query+" AND tenant_id = $1" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestWithTenantID_TrimsWhitespace(t *testing.T) {
	tenantID := uuid.New()
	ctx := TenantContext(context.Background(), tenantID)

	query := "  SELECT * FROM lkfl_platform.engagement_types  "
	result, _ := WithTenantID(ctx, query)

	// Результат должен быть trimmed
	if result == query {
		t.Error("query should be trimmed")
	}
	if result != "SELECT * FROM lkfl_platform.engagement_types WHERE tenant_id = $1" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestWithTenantID_MultipleWhereInQuery(t *testing.T) {
	tenantID := uuid.New()
	ctx := TenantContext(context.Background(), tenantID)

	// Query с WHERE в подзапросе
	query := "SELECT * FROM lkfl_platform.engagement_types e WHERE e.id IN (SELECT id FROM other WHERE active = true)"
	result, _ := WithTenantID(ctx, query)

	// Должен добавить AND в конец
	if result != query+" AND tenant_id = $1" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestTenantContext(t *testing.T) {
	tenantID := uuid.New()
	ctx := TenantContext(context.Background(), tenantID)

	result := TenantIDFromContext(ctx)
	if result != tenantID {
		t.Errorf("expected %s, got %s", tenantID, result)
	}
}

func TestWithAdminTenant(t *testing.T) {
	ctx := WithAdminTenant(context.Background())

	result := TenantIDFromContext(ctx)
	if result != uuid.Nil {
		t.Errorf("expected uuid.Nil for admin tenant, got %s", result)
	}
}
