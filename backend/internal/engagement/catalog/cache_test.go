package catalog

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// --- NewCache tests ---

func TestNewCache_NilClient(t *testing.T) {
	c := NewCache(nil, nil)
	if c == nil {
		t.Fatal("expected non-nil Cache")
	}
	if c.client != nil {
		t.Error("expected nil client")
	}
}

// --- Nil client safety tests ---

func TestCache_NilClient_GetList(t *testing.T) {
	c := NewCache(nil, nil)
	data, ok := c.GetList(context.Background(), CatalogFilter{})
	if ok {
		t.Error("expected false for nil client")
	}
	if data != nil {
		t.Error("expected nil data for nil client")
	}
}

func TestCache_NilClient_SetList(t *testing.T) {
	c := NewCache(nil, nil)
	err := c.SetList(context.Background(), CatalogFilter{}, []byte("test"))
	if err != nil {
		t.Errorf("expected nil error for nil client, got: %v", err)
	}
}

func TestCache_NilClient_GetType(t *testing.T) {
	c := NewCache(nil, nil)
	data, ok := c.GetType(context.Background(), "tenant", "type")
	if ok {
		t.Error("expected false for nil client")
	}
	if data != nil {
		t.Error("expected nil data for nil client")
	}
}

func TestCache_NilClient_SetType(t *testing.T) {
	c := NewCache(nil, nil)
	err := c.SetType(context.Background(), "tenant", "type", []byte("test"))
	if err != nil {
		t.Errorf("expected nil error for nil client, got: %v", err)
	}
}

func TestCache_NilClient_GetCategories(t *testing.T) {
	c := NewCache(nil, nil)
	data, ok := c.GetCategories(context.Background(), "tenant")
	if ok {
		t.Error("expected false for nil client")
	}
	if data != nil {
		t.Error("expected nil data for nil client")
	}
}

func TestCache_NilClient_SetCategories(t *testing.T) {
	c := NewCache(nil, nil)
	err := c.SetCategories(context.Background(), "tenant", []byte("test"))
	if err != nil {
		t.Errorf("expected nil error for nil client, got: %v", err)
	}
}

func TestCache_NilClient_Invalidate(t *testing.T) {
	c := NewCache(nil, nil)
	err := c.Invalidate(context.Background(), "tenant")
	if err != nil {
		t.Errorf("expected nil error for nil client, got: %v", err)
	}
}

// --- Key format tests ---

func TestCache_KeyFormat_List(t *testing.T) {
	// Test that key format includes tenant_id for isolation
	_ = uuid.New()
	_ = CatalogFilter{
		TenantID: uuid.New(),
		Type:     TypeBenefit,
		Status:   StatusActive,
		Search:   "yoga",
		Page:     2,
	}

	// We can't test actual Redis without a connection, but we can verify
	// the key format constants are correct
	if cacheListKeyFmt != "catalog:list:%s:%s:%s:%s:%d" {
		t.Errorf("unexpected key format: %s", cacheListKeyFmt)
	}

	// Verify tenant_id is part of the key format (5 format specifiers)
	expectedParts := 5 // tenant_id, type, status, search, page
	actualParts := 0
	for _, c := range cacheListKeyFmt {
		if c == '%' {
			actualParts++
		}
	}
	if actualParts != expectedParts {
		t.Errorf("expected %d format specifiers in list key, got %d", expectedParts, actualParts)
	}
}

func TestCache_KeyFormat_Type(t *testing.T) {
	if cacheTypeKeyFmt != "catalog:type:%s:%s" {
		t.Errorf("unexpected key format: %s", cacheTypeKeyFmt)
	}
}

func TestCache_KeyFormat_Categories(t *testing.T) {
	if cacheCategoriesKeyFmt != "catalog:categories:%s" {
		t.Errorf("unexpected key format: %s", cacheCategoriesKeyFmt)
	}
}

func TestCache_Prefix(t *testing.T) {
	if cachePrefix != "catalog:" {
		t.Errorf("unexpected prefix: %s", cachePrefix)
	}
}

// --- Service cache integration tests (nil cache) ---

func TestService_ListTypes_NoCache(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	et := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "test",
		Name:     "Test",
	}
	repo.types[et.ID] = et

	_, total, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestService_GetTypeByID_NoCache(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "test",
		Name:     "Test",
	}
	repo.types[typeID] = et

	result, err := svc.GetTypeByID(context.Background(), typeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != typeID {
		t.Errorf("expected ID %s, got %s", typeID, result.ID)
	}
}

func TestService_GetCategories_NoCache(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	cat := EngagementCategory{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "fitness",
		Name:     "Фитнес",
	}
	repo.categories[cat.ID] = cat

	cats, err := svc.GetCategories(context.Background(), tid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 1 {
		t.Errorf("expected 1 category, got %d", len(cats))
	}
}

// --- AdminHandler cache invalidation (nil cache) ---

func TestAdminHandler_Invalidate_NilCache(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)
	h := NewAdminHandler(svc, nil)

	// invalidateCache should not panic with nil cache
	h.invalidateCache(context.Background(), uuid.New().String())
}
