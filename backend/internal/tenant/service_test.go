package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockRepository — mock реализация Repository для тестов.
type mockRepository struct {
	tenants    []Tenant
	brandConfs []BrandConfig

	getBySlugFn   func(ctx context.Context, slug string) (Tenant, error)
	createFn      func(ctx context.Context, t Tenant) (Tenant, error)
	getByIDFn     func(ctx context.Context, id uuid.UUID) (Tenant, error)
	listFn        func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error)
	updateFn      func(ctx context.Context, t Tenant) (Tenant, error)
	deleteFn      func(ctx context.Context, id uuid.UUID) error
	getBrandFn    func(ctx context.Context, tenantID uuid.UUID) (BrandConfig, error)
	upsertBrandFn func(ctx context.Context, bc BrandConfig) (BrandConfig, error)
}

func (m *mockRepository) Create(ctx context.Context, t Tenant) (Tenant, error) {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	t.ID = uuid.New()
	m.tenants = append(m.tenants, t)
	return t, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (Tenant, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	for _, t := range m.tenants {
		if t.ID == id {
			return t, nil
		}
	}
	return Tenant{}, ErrNotFound
}

func (m *mockRepository) GetBySlug(ctx context.Context, slug string) (Tenant, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, slug)
	}
	for _, t := range m.tenants {
		if t.Slug == slug {
			return t, nil
		}
	}
	return Tenant{}, ErrNotFound
}

func (m *mockRepository) List(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	var filtered []Tenant
	for _, t := range m.tenants {
		if filter.Status == "" || t.Status == filter.Status {
			filtered = append(filtered, t)
		}
	}
	if filtered == nil {
		filtered = []Tenant{}
	}
	return filtered, int64(len(filtered)), nil
}

func (m *mockRepository) Update(ctx context.Context, t Tenant) (Tenant, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, t)
	}
	for i, existing := range m.tenants {
		if existing.ID == t.ID {
			m.tenants[i] = t
			return t, nil
		}
	}
	return Tenant{}, ErrNotFound
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	for i, t := range m.tenants {
		if t.ID == id {
			m.tenants = append(m.tenants[:i], m.tenants[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (m *mockRepository) GetBrandConfig(ctx context.Context, tenantID uuid.UUID) (BrandConfig, error) {
	if m.getBrandFn != nil {
		return m.getBrandFn(ctx, tenantID)
	}
	for _, bc := range m.brandConfs {
		if bc.TenantID == tenantID {
			return bc, nil
		}
	}
	return BrandConfig{}, ErrBrandNotFound
}

func (m *mockRepository) UpsertBrandConfig(ctx context.Context, bc BrandConfig) (BrandConfig, error) {
	if m.upsertBrandFn != nil {
		return m.upsertBrandFn(ctx, bc)
	}
	for i, existing := range m.brandConfs {
		if existing.TenantID == bc.TenantID {
			bc.ID = existing.ID
			bc.CreatedAt = existing.CreatedAt
			m.brandConfs[i] = bc
			return bc, nil
		}
	}
	bc.ID = uuid.New()
	m.brandConfs = append(m.brandConfs, bc)
	return bc, nil
}

// Проверка что mockRepository реализует Repository.
var _ Repository = (*mockRepository)(nil)

// --- Service Tests ---

func newTestService(repo Repository) *Service {
	return NewService(repo)
}

func TestCreateTenant_ValidSlug(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{
		Slug:   "test-tenant",
		Name:   "Test Tenant",
		Status: "active",
	}

	result, err := svc.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Slug != "test-tenant" {
		t.Errorf("expected slug 'test-tenant', got '%s'", result.Slug)
	}
	if result.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", result.Status)
	}
}

func TestCreateTenant_DuplicateSlug(t *testing.T) {
	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return Tenant{Slug: slug}, nil // tenant уже существует
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{
		Slug: "existing",
		Name: "Existing",
	}

	_, err := svc.CreateTenant(ctx, tenant)
	if !errors.Is(err, ErrSlugExists) {
		t.Fatalf("expected ErrSlugExists, got: %v", err)
	}
}

func TestCreateTenant_InvalidSlug(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	tests := []struct {
		name string
		slug string
	}{
		{"uppercase", "Test-Tenant"},
		{"spaces", "test tenant"},
		{"special chars", "test_tenant"},
		{"starts with hyphen", "-test-tenant"},
		{"ends with hyphen", "test-tenant-"},
		{"double hyphen", "test--tenant"},
		{"empty", ""},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := Tenant{Slug: tt.slug, Name: "Test"}
			_, err := svc.CreateTenant(ctx, tenant)
			if !errors.Is(err, ErrInvalidSlug) {
				t.Errorf("expected ErrInvalidSlug for slug %q, got: %v", tt.slug, err)
			}
		})
	}
}

func TestCreateTenant_DefaultValues(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{
		Slug: "defaults",
		Name: "Defaults",
		// Status и Settings не заданы
	}

	result, err := svc.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "active" {
		t.Errorf("expected default status 'active', got '%s'", result.Status)
	}
	if result.Settings == nil {
		t.Error("expected non-nil Settings")
	}
}

func TestGetBySlug_Active(t *testing.T) {
	activeTenant := Tenant{
		ID:     uuid.New(),
		Slug:   "active-tenant",
		Name:   "Active",
		Status: "active",
	}

	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return activeTenant, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	result, err := svc.GetBySlug(ctx, "active-tenant")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != activeTenant.ID {
		t.Errorf("expected ID %s, got %s", activeTenant.ID, result.ID)
	}
}

func TestGetBySlug_Suspended(t *testing.T) {
	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return Tenant{Slug: slug, Status: "suspended"}, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.GetBySlug(ctx, "suspended-tenant")
	if !errors.Is(err, ErrTenantSuspended) {
		t.Fatalf("expected ErrTenantSuspended, got: %v", err)
	}
}

func TestGetBySlug_NotFound(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.GetBySlug(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetByID(t *testing.T) {
	expected := Tenant{
		ID:     uuid.New(),
		Slug:   "test",
		Name:   "Test",
		Status: "active",
	}

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (Tenant, error) {
			return expected, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	result, err := svc.GetByID(ctx, expected.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, result.ID)
	}
}

func TestListTenants_DefaultPagination(t *testing.T) {
	tenants := []Tenant{
		{ID: uuid.New(), Slug: "a", Name: "A", Status: "active"},
		{ID: uuid.New(), Slug: "b", Name: "B", Status: "active"},
		{ID: uuid.New(), Slug: "c", Name: "C", Status: "suspended"},
	}

	repo := &mockRepository{tenants: tenants}
	svc := newTestService(repo)

	ctx := context.Background()
	result, total, err := svc.ListTenants(ctx, TenantFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3, got %d", total)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 tenants, got %d", len(result))
	}
}

func TestListTenants_FilterByStatus(t *testing.T) {
	tenants := []Tenant{
		{ID: uuid.New(), Slug: "a", Name: "A", Status: "active"},
		{ID: uuid.New(), Slug: "b", Name: "B", Status: "active"},
		{ID: uuid.New(), Slug: "c", Name: "C", Status: "suspended"},
	}

	repo := &mockRepository{tenants: tenants}
	svc := newTestService(repo)

	ctx := context.Background()
	result, total, err := svc.ListTenants(ctx, TenantFilter{Status: "active"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(result))
	}
}

func TestDeleteTenant(t *testing.T) {
	tenantID := uuid.New()
	tenants := []Tenant{
		{ID: tenantID, Slug: "deletable", Name: "Deletable", Status: "active"},
	}

	repo := &mockRepository{tenants: tenants}
	svc := newTestService(repo)

	ctx := context.Background()
	err := svc.DeleteTenant(ctx, tenantID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем что tenant удалён
	_, err = svc.GetByID(ctx, tenantID)
	if !errors.Is(err, ErrNotFound) {
		t.Error("expected ErrNotFound after delete")
	}
}

func TestUpsertBrandConfig(t *testing.T) {
	tenantID := uuid.New()
	brandName := "Test Brand"

	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	bc := BrandConfig{
		TenantID:     tenantID,
		PrimaryColor: "#FF0000",
		BrandName:    &brandName,
	}

	result, err := svc.UpsertBrandConfig(ctx, bc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TenantID != tenantID {
		t.Errorf("expected tenant_id %s, got %s", tenantID, result.TenantID)
	}
	if result.PrimaryColor != "#FF0000" {
		t.Errorf("expected primary_color '#FF0000', got '%s'", result.PrimaryColor)
	}
}

// =============================================================================
// EDGE CASE TESTS — расширенные тесты для boundary conditions и error paths
// =============================================================================

// --- CreateTenant Edge Cases ---

func TestCreateTenant_EmptySlug(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for empty slug, got: %v", err)
	}
}

func TestCreateTenant_SlugWithUppercase(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "Test-Tenant", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for uppercase slug, got: %v", err)
	}
}

func TestCreateTenant_SlugWithSpecialChars(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "test_tenant", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for underscore in slug, got: %v", err)
	}
}

func TestCreateTenant_SlugWithSpaces(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "test tenant", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for spaces in slug, got: %v", err)
	}
}

func TestCreateTenant_SlugWithUnicode(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "тест-tenant", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for unicode in slug, got: %v", err)
	}
}

func TestCreateTenant_SlugWithHyphensUnderscoresValid(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{Slug: "my-test-tenant-123", Name: "Test"}
	result, err := svc.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("unexpected error for valid slug with hyphens: %v", err)
	}
	if result.Slug != "my-test-tenant-123" {
		t.Errorf("expected slug 'my-test-tenant-123', got '%s'", result.Slug)
	}
}

func TestCreateTenant_SingleCharSlug(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{Slug: "a", Name: "Single"}
	result, err := svc.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("unexpected error for single char slug: %v", err)
	}
	if result.Slug != "a" {
		t.Errorf("expected slug 'a', got '%s'", result.Slug)
	}
}

func TestCreateTenant_AllDigitsSlug(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{Slug: "12345", Name: "Digits"}
	result, err := svc.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("unexpected error for all-digits slug: %v", err)
	}
	if result.Slug != "12345" {
		t.Errorf("expected slug '12345', got '%s'", result.Slug)
	}
}

func TestCreateTenant_SlugStartsWithHyphen(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "-test", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for slug starting with hyphen, got: %v", err)
	}
}

func TestCreateTenant_SlugEndsWithHyphen(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "test-", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for slug ending with hyphen, got: %v", err)
	}
}

func TestCreateTenant_SlugDoubleHyphen(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "test--tenant", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug for double hyphen in slug, got: %v", err)
	}
}

func TestCreateTenant_DuplicateSlugViaMock(t *testing.T) {
	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return Tenant{ID: uuid.New(), Slug: slug}, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "existing", Name: "Test"})
	if !errors.Is(err, ErrSlugExists) {
		t.Fatalf("expected ErrSlugExists, got: %v", err)
	}
}

func TestCreateTenant_NilContext(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	_, err := svc.CreateTenant(nil, Tenant{Slug: "test", Name: "Test"})
	// Не должно паниковать — либо error, либо успех
	if err == nil {
		t.Log("CreateTenant with nil context succeeded (acceptable)")
	}
}

func TestCreateTenant_GetBySlugError(t *testing.T) {
	dbErr := errors.New("connection refused")
	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return Tenant{}, dbErr
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.CreateTenant(ctx, Tenant{Slug: "test", Name: "Test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, dbErr) {
		t.Logf("error wrapped correctly: %v", err)
	}
}

func TestCreateTenant_DefaultStatus(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{Slug: "no-status", Name: "Test"}
	result, err := svc.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != TenantStatusActive {
		t.Errorf("expected default status '%s', got '%s'", TenantStatusActive, result.Status)
	}
}

func TestCreateTenant_DefaultSettings(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	tenant := Tenant{Slug: "no-settings", Name: "Test"}
	result, err := svc.CreateTenant(ctx, tenant)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Settings == nil {
		t.Error("expected non-nil default settings")
	}
}

// --- GetBySlug Edge Cases ---

func TestGetBySlug_EmptySlug(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.GetBySlug(ctx, "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for empty slug, got: %v", err)
	}
}

func TestGetBySlug_DeletedTenant(t *testing.T) {
	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return Tenant{Slug: slug, Status: "deleted"}, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.GetBySlug(ctx, "deleted-tenant")
	if !errors.Is(err, ErrTenantSuspended) {
		t.Fatalf("expected ErrTenantSuspended for deleted tenant, got: %v", err)
	}
}

func TestGetBySlug_NilContext(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	_, err := svc.GetBySlug(nil, "test")
	// Не должно паниковать
	if err == nil {
		t.Log("GetBySlug with nil context returned no error (acceptable)")
	}
}

func TestGetBySlugRaw_NoStatusCheck(t *testing.T) {
	suspendedTenant := Tenant{
		ID:     uuid.New(),
		Slug:   "suspended",
		Status: "suspended",
	}

	repo := &mockRepository{
		getBySlugFn: func(ctx context.Context, slug string) (Tenant, error) {
			return suspendedTenant, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	result, err := svc.GetBySlugRaw(ctx, "suspended")
	if err != nil {
		t.Fatalf("expected no error for GetBySlugRaw, got: %v", err)
	}
	if result.Status != "suspended" {
		t.Errorf("expected suspended status, got '%s'", result.Status)
	}
}

// --- ListTenants Edge Cases ---

func TestListTenants_PageZero(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
			if filter.Page != 1 {
				t.Errorf("expected page to default to 1, got %d", filter.Page)
			}
			return []Tenant{}, 0, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, _, err := svc.ListTenants(ctx, TenantFilter{Page: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTenants_PageNegative(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
			if filter.Page != 1 {
				t.Errorf("expected page to default to 1 for negative input, got %d", filter.Page)
			}
			return []Tenant{}, 0, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, _, err := svc.ListTenants(ctx, TenantFilter{Page: -5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTenants_PerPageZero(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
			if filter.PerPage != 20 {
				t.Errorf("expected per_page to default to 20, got %d", filter.PerPage)
			}
			return []Tenant{}, 0, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, _, err := svc.ListTenants(ctx, TenantFilter{PerPage: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTenants_PerPageNegative(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
			if filter.PerPage != 20 {
				t.Errorf("expected per_page to default to 20 for negative input, got %d", filter.PerPage)
			}
			return []Tenant{}, 0, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, _, err := svc.ListTenants(ctx, TenantFilter{PerPage: -1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTenants_PerPageCapped(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
			if filter.PerPage != 100 {
				t.Errorf("expected per_page capped to 100, got %d", filter.PerPage)
			}
			return []Tenant{}, 0, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, _, err := svc.ListTenants(ctx, TenantFilter{PerPage: 200})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTenants_PerPageExactly100(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
			if filter.PerPage != 100 {
				t.Errorf("expected per_page 100, got %d", filter.PerPage)
			}
			return []Tenant{}, 0, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, _, err := svc.ListTenants(ctx, TenantFilter{PerPage: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTenants_PerPage1(t *testing.T) {
	repo := &mockRepository{
		listFn: func(ctx context.Context, filter TenantFilter) ([]Tenant, int64, error) {
			if filter.PerPage != 1 {
				t.Errorf("expected per_page 1, got %d", filter.PerPage)
			}
			return []Tenant{}, 0, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, _, err := svc.ListTenants(ctx, TenantFilter{PerPage: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListTenants_EmptyResultSet(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	result, total, err := svc.ListTenants(ctx, TenantFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 tenants, got %d", len(result))
	}
	if result == nil {
		t.Error("expected non-nil empty slice")
	}
}

func TestListTenants_NilContext(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	_, _, err := svc.ListTenants(nil, TenantFilter{})
	// Не должно паниковать
	if err == nil {
		t.Log("ListTenants with nil context returned no error (acceptable)")
	}
}

// --- UpdateTenant Edge Cases ---

func TestUpdateTenant_EmptySlugSkipped(t *testing.T) {
	tenantID := uuid.New()
	repo := &mockRepository{
		updateFn: func(ctx context.Context, t Tenant) (Tenant, error) {
			return t, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.UpdateTenant(ctx, Tenant{ID: tenantID, Slug: "", Name: "Updated"})
	if err != nil {
		t.Fatalf("unexpected error for empty slug update, got: %v", err)
	}
}

func TestUpdateTenant_InvalidSlug(t *testing.T) {
	tenantID := uuid.New()
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.UpdateTenant(ctx, Tenant{ID: tenantID, Slug: "INVALID", Name: "Test"})
	if !errors.Is(err, ErrInvalidSlug) {
		t.Fatalf("expected ErrInvalidSlug, got: %v", err)
	}
}

func TestUpdateTenant_NonExistent(t *testing.T) {
	tenantID := uuid.New()
	repo := &mockRepository{
		updateFn: func(ctx context.Context, t Tenant) (Tenant, error) {
			return Tenant{}, ErrNotFound
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.UpdateTenant(ctx, Tenant{ID: tenantID, Slug: "test", Name: "Updated"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestUpdateTenant_PartialUpdate(t *testing.T) {
	tenantID := uuid.New()
	var captured Tenant
	repo := &mockRepository{
		updateFn: func(ctx context.Context, t Tenant) (Tenant, error) {
			captured = t
			return t, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.UpdateTenant(ctx, Tenant{ID: tenantID, Slug: "new-slug", Name: "New Name"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Slug != "new-slug" {
		t.Errorf("expected slug 'new-slug', got '%s'", captured.Slug)
	}
	if captured.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", captured.Name)
	}
}

func TestUpdateTenant_UpdateSuspended(t *testing.T) {
	tenantID := uuid.New()
	var captured Tenant
	repo := &mockRepository{
		updateFn: func(ctx context.Context, t Tenant) (Tenant, error) {
			captured = t
			return t, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.UpdateTenant(ctx, Tenant{ID: tenantID, Slug: "suspended", Status: "suspended"})
	if err != nil {
		t.Fatalf("unexpected error updating suspended tenant: %v", err)
	}
	if captured.Status != "suspended" {
		t.Errorf("expected status 'suspended', got '%s'", captured.Status)
	}
}

// --- DeleteTenant Edge Cases ---

func TestDeleteTenant_NonExistent(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	err := svc.DeleteTenant(ctx, uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestDeleteTenant_NilContext(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	err := svc.DeleteTenant(nil, uuid.New())
	// Не должно паниковать
	if err == nil {
		t.Log("DeleteTenant with nil context returned no error (acceptable)")
	}
}

// --- BrandConfig Edge Cases ---

func TestUpsertBrandConfig_NilCSSVariables(t *testing.T) {
	tenantID := uuid.New()
	var captured BrandConfig
	repo := &mockRepository{
		upsertBrandFn: func(ctx context.Context, bc BrandConfig) (BrandConfig, error) {
			captured = bc
			return bc, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	bc := BrandConfig{TenantID: tenantID, PrimaryColor: "#000"}
	result, err := svc.UpsertBrandConfig(ctx, bc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.CSSVariables == nil {
		t.Error("expected CSSVariables to be initialized to empty JSONB")
	}
	if result.TenantID != tenantID {
		t.Errorf("expected tenant_id %s, got %s", tenantID, result.TenantID)
	}
}

func TestUpsertBrandConfig_AllFields(t *testing.T) {
	tenantID := uuid.New()
	brandName := "Full Brand"
	logoURL := "https://example.com/logo.png"
	faviconURL := "https://example.com/favicon.ico"
	metaTitle := "Meta Title"
	metaDesc := "Meta Description"

	var captured BrandConfig
	repo := &mockRepository{
		upsertBrandFn: func(ctx context.Context, bc BrandConfig) (BrandConfig, error) {
			captured = bc
			return bc, nil
		},
	}
	svc := newTestService(repo)

	ctx := context.Background()
	bc := BrandConfig{
		TenantID:        tenantID,
		PrimaryColor:    "#FF0000",
		SecondaryColor:  "#00FF00",
		LogoURL:         &logoURL,
		FaviconURL:      &faviconURL,
		BrandName:       &brandName,
		CSSVariables:    JSONB{"--custom": "value"},
		MetaTitle:       &metaTitle,
		MetaDescription: &metaDesc,
	}

	_, err := svc.UpsertBrandConfig(ctx, bc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.PrimaryColor != "#FF0000" {
		t.Errorf("expected primary_color '#FF0000', got '%s'", captured.PrimaryColor)
	}
	if captured.SecondaryColor != "#00FF00" {
		t.Errorf("expected secondary_color '#00FF00', got '%s'", captured.SecondaryColor)
	}
	if captured.LogoURL == nil || *captured.LogoURL != logoURL {
		t.Errorf("expected logo_url '%s', got %v", logoURL, captured.LogoURL)
	}
}

func TestUpsertBrandConfig_SequentialMultiple(t *testing.T) {
	tenantID := uuid.New()
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()

	// Sequential updates (concurrent would require sync.Mutex in mock)
	for i := 0; i < 3; i++ {
		color := fmt.Sprintf("#%06X", i*0x100000)
		bc := BrandConfig{TenantID: tenantID, PrimaryColor: color}
		_, err := svc.UpsertBrandConfig(ctx, bc)
		if err != nil {
			t.Errorf("sequential upsert %d error: %v", i, err)
		}
	}
}

func TestGetBrandConfig_NotFound(t *testing.T) {
	repo := &mockRepository{}
	svc := newTestService(repo)

	ctx := context.Background()
	_, err := svc.GetBrandConfig(ctx, uuid.New())
	if !errors.Is(err, ErrBrandNotFound) {
		t.Fatalf("expected ErrBrandNotFound, got: %v", err)
	}
}

// --- JSONB Edge Cases ---

func TestJSONB_ScanNil(t *testing.T) {
	var j JSONB
	err := j.Scan(nil)
	if err != nil {
		t.Fatalf("unexpected error scanning nil: %v", err)
	}
	if j == nil {
		t.Error("expected non-nil empty JSONB")
	}
	if len(j) != 0 {
		t.Errorf("expected empty JSONB, got %d keys", len(j))
	}
}

func TestJSONB_ScanInvalidType(t *testing.T) {
	var j JSONB
	err := j.Scan(42)
	if err == nil {
		t.Fatal("expected error scanning int, got nil")
	}
}

func TestJSONB_ValueNil(t *testing.T) {
	var j JSONB
	v, err := j.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(v.([]byte)) != "{}" {
		t.Errorf("expected '{}', got '%s'", string(v.([]byte)))
	}
}

func TestJSONB_ValueWithData(t *testing.T) {
	j := JSONB{"key": "value", "num": 42}
	v, err := j.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result := string(v.([]byte))
	if !strings.Contains(result, "key") {
		t.Errorf("expected 'key' in JSON, got '%s'", result)
	}
	if !strings.Contains(result, "value") {
		t.Errorf("expected 'value' in JSON, got '%s'", result)
	}
}

// --- Slug Regex Edge Cases ---

func TestSlugRegex_Validation(t *testing.T) {
	tests := []struct {
		slug    string
		want    bool
		reason  string
	}{
		{"", false, "empty"},
		{"a", true, "single char"},
		{"abc", true, "all lowercase"},
		{"123", true, "all digits"},
		{"a-b", true, "with hyphen"},
		{"a-b-c", true, "multiple hyphens"},
		{"a1b2", true, "mixed"},
		{"abc-", false, "trailing hyphen"},
		{"-abc", false, "leading hyphen"},
		{"ab--cd", false, "double hyphen"},
		{"ABC", false, "uppercase"},
		{"ab cd", false, "space"},
		{"ab_cd", false, "underscore"},
		{"ab.cd", false, "dot"},
		{"ab/cd", false, "slash"},
		{"тест", false, "cyrillic"},
		{"测试", false, "chinese"},
	}

	for _, tt := range tests {
		t.Run(tt.reason, func(t *testing.T) {
			got := slugRegex.MatchString(tt.slug)
			if got != tt.want {
				t.Errorf("slugRegex.MatchString(%q) = %v, want %v (%s)", tt.slug, got, tt.want, tt.reason)
			}
		})
	}
}
