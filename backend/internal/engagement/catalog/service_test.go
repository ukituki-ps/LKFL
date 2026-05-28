package catalog

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockRepository — мока для Repository для unit-тестов Service.
type mockRepository struct {
	types      map[uuid.UUID]EngagementType
	categories map[uuid.UUID]EngagementCategory
	offers     map[uuid.UUID]EngagementOffer
	deleted    map[uuid.UUID]bool
	nextID     int
	errOn      map[string]error
	errKeys    map[string]bool
}

func newMockRepo() *mockRepository {
	return &mockRepository{
		types:      make(map[uuid.UUID]EngagementType),
		categories: make(map[uuid.UUID]EngagementCategory),
		offers:     make(map[uuid.UUID]EngagementOffer),
		deleted:    make(map[uuid.UUID]bool),
		errOn:      make(map[string]error),
		errKeys:    make(map[string]bool),
	}
}

func (m *mockRepository) nextUUID() uuid.UUID {
	m.nextID++
	return uuid.New()
}

// Public methods

func (m *mockRepository) ListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error) {
	if m.errKeys["ListTypes"] {
		return nil, 0, m.errOn["ListTypes"]
	}
	var result []EngagementType
	for _, t := range m.types {
		if t.TenantID != filter.TenantID {
			continue
		}
		// Статус фильтр (по умолчанию active+promo)
		if filter.Status != "" {
			if t.Status != filter.Status {
				continue
			}
		} else {
			if t.Status != StatusActive && t.Status != StatusPromo {
				continue
			}
		}
		if filter.Type != "" && t.Type != filter.Type {
			continue
		}
		if filter.Category != "" && t.Category != nil && t.Category.Slug != filter.Category {
			continue
		}
		result = append(result, t)
	}
	if result == nil {
		result = []EngagementType{}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) GetTypeByID(ctx context.Context, id uuid.UUID) (EngagementType, error) {
	if m.errKeys["GetTypeByID"] {
		return EngagementType{}, m.errOn["GetTypeByID"]
	}
	t, ok := m.types[id]
	if !ok {
		return EngagementType{}, ErrNotFound
	}
	return t, nil
}

func (m *mockRepository) GetCategories(ctx context.Context, tenantID uuid.UUID) ([]EngagementCategory, error) {
	if m.errKeys["GetCategories"] {
		return nil, m.errOn["GetCategories"]
	}
	var result []EngagementCategory
	for _, c := range m.categories {
		if c.TenantID == tenantID {
			result = append(result, c)
		}
	}
	if result == nil {
		result = []EngagementCategory{}
	}
	return result, nil
}

// Admin Category methods

func (m *mockRepository) AdminCreateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error) {
	if m.errKeys["AdminCreateCategory"] {
		return EngagementCategory{}, m.errOn["AdminCreateCategory"]
	}
	c.ID = m.nextUUID()
	m.categories[c.ID] = c
	return c, nil
}

func (m *mockRepository) AdminUpdateCategory(ctx context.Context, c EngagementCategory) (EngagementCategory, error) {
	if m.errKeys["AdminUpdateCategory"] {
		return EngagementCategory{}, m.errOn["AdminUpdateCategory"]
	}
	_, ok := m.categories[c.ID]
	if !ok {
		return EngagementCategory{}, ErrCategoryNotFound
	}
	m.categories[c.ID] = c
	return c, nil
}

// Admin Type methods

func (m *mockRepository) AdminCreateType(ctx context.Context, t EngagementType) (EngagementType, error) {
	if m.errKeys["AdminCreateType"] {
		return EngagementType{}, m.errOn["AdminCreateType"]
	}
	t.ID = m.nextUUID()
	m.types[t.ID] = t
	return t, nil
}

func (m *mockRepository) AdminUpdateType(ctx context.Context, t EngagementType) (EngagementType, error) {
	if m.errKeys["AdminUpdateType"] {
		return EngagementType{}, m.errOn["AdminUpdateType"]
	}
	_, ok := m.types[t.ID]
	if !ok {
		return EngagementType{}, ErrNotFound
	}
	m.types[t.ID] = t
	return t, nil
}

func (m *mockRepository) AdminDeleteType(ctx context.Context, id uuid.UUID) error {
	if m.errKeys["AdminDeleteType"] {
		return m.errOn["AdminDeleteType"]
	}
	_, ok := m.types[id]
	if !ok {
		return ErrNotFound
	}
	m.deleted[id] = true
	return nil
}

// Offer methods

func (m *mockRepository) GetOffersByType(ctx context.Context, typeID uuid.UUID) ([]EngagementOffer, error) {
	if m.errKeys["GetOffersByType"] {
		return nil, m.errOn["GetOffersByType"]
	}
	var result []EngagementOffer
	for _, o := range m.offers {
		if o.EngagementTypeID == typeID {
			result = append(result, o)
		}
	}
	if result == nil {
		result = []EngagementOffer{}
	}
	return result, nil
}

func (m *mockRepository) AdminCreateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error) {
	if m.errKeys["AdminCreateOffer"] {
		return EngagementOffer{}, m.errOn["AdminCreateOffer"]
	}
	o.ID = m.nextUUID()
	m.offers[o.ID] = o
	return o, nil
}

func (m *mockRepository) AdminUpdateOffer(ctx context.Context, o EngagementOffer) (EngagementOffer, error) {
	if m.errKeys["AdminUpdateOffer"] {
		return EngagementOffer{}, m.errOn["AdminUpdateOffer"]
	}
	_, ok := m.offers[o.ID]
	if !ok {
		return EngagementOffer{}, ErrOfferNotFound
	}
	m.offers[o.ID] = o
	return o, nil
}

func (m *mockRepository) AdminDeleteOffer(ctx context.Context, id uuid.UUID) error {
	if m.errKeys["AdminDeleteOffer"] {
		return m.errOn["AdminDeleteOffer"]
	}
	_, ok := m.offers[id]
	if !ok {
		return ErrOfferNotFound
	}
	delete(m.offers, id)
	return nil
}

// Admin methods — новые

func (m *mockRepository) AdminListTypes(ctx context.Context, filter CatalogFilter) ([]EngagementType, int64, error) {
	if m.errKeys["AdminListTypes"] {
		return nil, 0, m.errOn["AdminListTypes"]
	}
	var result []EngagementType
	for _, t := range m.types {
		if t.TenantID != filter.TenantID {
			continue
		}
		// Admin: по умолчанию все статусы; если задан фильтр — только он
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.Type != "" && t.Type != filter.Type {
			continue
		}
		result = append(result, t)
	}
	if result == nil {
		result = []EngagementType{}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) AdminDeleteCategory(ctx context.Context, id uuid.UUID) error {
	if m.errKeys["AdminDeleteCategory"] {
		return m.errOn["AdminDeleteCategory"]
	}
	_, ok := m.categories[id]
	if !ok {
		return ErrCategoryNotFound
	}
	delete(m.categories, id)
	return nil
}

func (m *mockRepository) GetTypeBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (EngagementType, error) {
	if m.errKeys["GetTypeBySlug"] {
		return EngagementType{}, m.errOn["GetTypeBySlug"]
	}
	for _, t := range m.types {
		if t.TenantID == tenantID && t.Slug == slug {
			return t, nil
		}
	}
	return EngagementType{}, ErrNotFound
}

// --- Tests ---

// TestService_ListTypes — тест списка типов с фильтрами.
func TestService_ListTypes(t *testing.T) {
	t.Run("пустой каталог возвращает пустой слайс", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		types, total, err := svc.ListTypes(context.Background(), CatalogFilter{
			TenantID: uuid.New(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("expected total 0, got %d", total)
		}
		if len(types) != 0 {
			t.Errorf("expected 0 types, got %d", len(types))
		}
		// Проверка empty slice pattern
		if types == nil {
			t.Error("expected non-nil empty slice")
		}
	})

	t.Run("фильтр по типу benefit", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		tid := uuid.New()
		benefit := EngagementType{
			ID:       uuid.New(),
			TenantID: tid,
			Type:     TypeBenefit,
			Status:   StatusActive,
			Slug:     "benefit-1",
			Name:     "Benefit 1",
		}
		activity := EngagementType{
			ID:       uuid.New(),
			TenantID: tid,
			Type:     TypeActivity,
			Status:   StatusActive,
			Slug:     "activity-1",
			Name:     "Activity 1",
		}
		repo.types[benefit.ID] = benefit
		repo.types[activity.ID] = activity

		types, total, err := svc.ListTypes(context.Background(), CatalogFilter{
			TenantID: tid,
			Type:     TypeBenefit,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
		if len(types) != 1 {
			t.Errorf("expected 1 type, got %d", len(types))
		}
	})

	t.Run("недопустимый тип", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		_, _, err := svc.ListTypes(context.Background(), CatalogFilter{
			TenantID: uuid.New(),
			Type:     "invalid",
		})
		if !errors.Is(err, ErrInvalidType) {
			t.Fatalf("expected ErrInvalidType, got: %v", err)
		}
	})

	t.Run("валидация пагинации", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		// page=0, per_page=0 → defaults
		_, _, err := svc.ListTypes(context.Background(), CatalogFilter{
			TenantID: uuid.New(),
			Page:     0,
			PerPage:  0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// per_page=200 → capped to 100
		_, _, err = svc.ListTypes(context.Background(), CatalogFilter{
			TenantID: uuid.New(),
			PerPage:  200,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestService_GetTypeByID — тест получения типа по ID.
func TestService_GetTypeByID(t *testing.T) {
	t.Run("успешное получение с офферами", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		typeID := uuid.New()
		et := EngagementType{
			ID:       typeID,
			TenantID: uuid.New(),
			Type:     TypeBenefit,
			Status:   StatusActive,
			Slug:     "test-benefit",
			Name:     "Test Benefit",
		}
		repo.types[typeID] = et

		offer := EngagementOffer{
			ID:               uuid.New(),
			EngagementTypeID: typeID,
			Name:             "Basic Plan",
			CostCents:        1000,
		}
		repo.offers[offer.ID] = offer

		result, err := svc.GetTypeByID(context.Background(), typeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != typeID {
			t.Errorf("expected type ID %s, got %s", typeID, result.ID)
		}
		if len(result.Offers) != 1 {
			t.Errorf("expected 1 offer, got %d", len(result.Offers))
		}
	})

	t.Run("тип не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		_, err := svc.GetTypeByID(context.Background(), uuid.New())
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

// TestService_GetTypeByTenantID — тест получения типа с tenant isolation.
func TestService_GetTypeByTenantID(t *testing.T) {
	t.Run("успешное получение при совпадении tenant", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		tid := uuid.New()
		typeID := uuid.New()
		et := EngagementType{
			ID:       typeID,
			TenantID: tid,
			Type:     TypeBenefit,
			Status:   StatusActive,
			Slug:     "test-benefit",
			Name:     "Test Benefit",
		}
		repo.types[typeID] = et

		result, err := svc.GetTypeByTenantID(context.Background(), tid, typeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != typeID {
			t.Errorf("expected type ID %s, got %s", typeID, result.ID)
		}
	})

	t.Run("tenant mismatch — ошибка", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		tid := uuid.New()
		typeID := uuid.New()
		et := EngagementType{
			ID:       typeID,
			TenantID: uuid.New(), // другой tenant
			Type:     TypeBenefit,
			Status:   StatusActive,
			Slug:     "test-benefit",
			Name:     "Test Benefit",
		}
		repo.types[typeID] = et

		_, err := svc.GetTypeByTenantID(context.Background(), tid, typeID)
		if !errors.Is(err, ErrTenantMismatch) {
			t.Fatalf("expected ErrTenantMismatch, got: %v", err)
		}
	})

	t.Run("тип не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		_, err := svc.GetTypeByTenantID(context.Background(), uuid.New(), uuid.New())
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

// TestService_AdminCreateCategory — тест создания категории.
func TestService_AdminCreateCategory(t *testing.T) {
	t.Run("успешное создание", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		cat := EngagementCategory{
			TenantID:  uuid.New(),
			Slug:      "health",
			Name:      "Здоровье",
			SortOrder: 1,
		}

		result, err := svc.AdminCreateCategory(context.Background(), cat)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID == uuid.Nil {
			t.Error("expected non-nil ID")
		}
	})

	t.Run("пустой slug", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		cat := EngagementCategory{
			TenantID: uuid.New(),
			Slug:     "",
			Name:     "Здоровье",
		}

		_, err := svc.AdminCreateCategory(context.Background(), cat)
		if !errors.Is(err, ErrInvalidFilter) {
			t.Fatalf("expected ErrInvalidFilter, got: %v", err)
		}
	})

	t.Run("дубликат slug", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		tid := uuid.New()
		existing := EngagementCategory{
			ID:       uuid.New(),
			TenantID: tid,
			Slug:     "health",
			Name:     "Здоровье",
		}
		repo.categories[existing.ID] = existing

		newCat := EngagementCategory{
			TenantID: tid,
			Slug:     "health",
			Name:     "Новое здоровье",
		}

		_, err := svc.AdminCreateCategory(context.Background(), newCat)
		if !errors.Is(err, ErrDuplicateSlug) {
			t.Fatalf("expected ErrDuplicateSlug, got: %v", err)
		}
	})
}

// TestService_AdminCreateType — тест создания типа энгейджмента.
func TestService_AdminCreateType(t *testing.T) {
	t.Run("успешное создание с дефолтными значениями", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		et := EngagementType{
			TenantID: uuid.New(),
			Slug:     "gym-membership",
			Name:     "Абонемент в спортзал",
		}

		result, err := svc.AdminCreateType(context.Background(), et)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Type != TypeBenefit {
			t.Errorf("expected default type %s, got %s", TypeBenefit, result.Type)
		}
		if result.Status != StatusDraft {
			t.Errorf("expected default status %s, got %s", StatusDraft, result.Status)
		}
	})

	t.Run("недопустимый тип", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		et := EngagementType{
			TenantID: uuid.New(),
			Slug:     "test",
			Name:     "Test",
			Type:     "invalid",
		}

		_, err := svc.AdminCreateType(context.Background(), et)
		if !errors.Is(err, ErrInvalidType) {
			t.Fatalf("expected ErrInvalidType, got: %v", err)
		}
	})

	t.Run("недопустимый статус", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		et := EngagementType{
			TenantID: uuid.New(),
			Slug:     "test",
			Name:     "Test",
			Status:   "invalid",
		}

		_, err := svc.AdminCreateType(context.Background(), et)
		if !errors.Is(err, ErrInvalidStatus) {
			t.Fatalf("expected ErrInvalidStatus, got: %v", err)
		}
	})

	t.Run("категория не найдена", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		et := EngagementType{
			TenantID:   uuid.New(),
			CategoryID: uuid.New(),
			Slug:       "test",
			Name:       "Test",
		}

		_, err := svc.AdminCreateType(context.Background(), et)
		if !errors.Is(err, ErrCategoryNotFound) {
			t.Fatalf("expected ErrCategoryNotFound, got: %v", err)
		}
	})

	t.Run("пустой slug", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		et := EngagementType{
			TenantID: uuid.New(),
			Slug:     "",
			Name:     "Test",
		}

		_, err := svc.AdminCreateType(context.Background(), et)
		if !errors.Is(err, ErrInvalidFilter) {
			t.Fatalf("expected ErrInvalidFilter, got: %v", err)
		}
	})
}

// TestService_AdminDeleteType — тест удаления типа (soft delete).
func TestService_AdminDeleteType(t *testing.T) {
	t.Run("успешное удаление", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		typeID := uuid.New()
		et := EngagementType{
			ID:       typeID,
			TenantID: uuid.New(),
			Status:   StatusActive,
			Slug:     "test",
			Name:     "Test",
		}
		repo.types[typeID] = et

		err := svc.AdminDeleteType(context.Background(), typeID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("тип не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		err := svc.AdminDeleteType(context.Background(), uuid.New())
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

// TestService_AdminCreateOffer — тест создания оффера.
func TestService_AdminCreateOffer(t *testing.T) {
	t.Run("успешное создание", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		typeID := uuid.New()
		et := EngagementType{
			ID:       typeID,
			TenantID: uuid.New(),
			Status:   StatusActive,
			Slug:     "test",
			Name:     "Test",
		}
		repo.types[typeID] = et

		offer := EngagementOffer{
			EngagementTypeID: typeID,
			Name:             "Basic Plan",
			CostCents:        1000,
		}

		result, err := svc.AdminCreateOffer(context.Background(), offer)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID == uuid.Nil {
			t.Error("expected non-nil ID")
		}
	})

	t.Run("тип не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		offer := EngagementOffer{
			EngagementTypeID: uuid.New(),
			Name:             "Basic Plan",
			CostCents:        1000,
		}

		_, err := svc.AdminCreateOffer(context.Background(), offer)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("пустое имя", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		offer := EngagementOffer{
			EngagementTypeID: uuid.New(),
			Name:             "",
			CostCents:        1000,
		}

		_, err := svc.AdminCreateOffer(context.Background(), offer)
		if !errors.Is(err, ErrInvalidFilter) {
			t.Fatalf("expected ErrInvalidFilter, got: %v", err)
		}
	})
}

// TestService_AdminDeleteOffer — тест удаления оффера.
func TestService_AdminDeleteOffer(t *testing.T) {
	t.Run("успешное удаление", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		offerID := uuid.New()
		offer := EngagementOffer{
			ID:               offerID,
			EngagementTypeID: uuid.New(),
			Name:             "Basic Plan",
		}
		repo.offers[offerID] = offer

		err := svc.AdminDeleteOffer(context.Background(), offerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("оффер не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		err := svc.AdminDeleteOffer(context.Background(), uuid.New())
		if !errors.Is(err, ErrOfferNotFound) {
			t.Fatalf("expected ErrOfferNotFound, got: %v", err)
		}
	})
}

// TestConstants — проверка констант.
func TestConstants(t *testing.T) {
	t.Run("статусы", func(t *testing.T) {
		expected := map[string]bool{
			StatusDraft:     true,
			StatusActive:    true,
			StatusPromo:     true,
			StatusHidden:    true,
			StatusCompleted: true,
		}
		for status := range expected {
			if !validStatuses[status] {
				t.Errorf("status %s should be valid", status)
			}
		}
	})

	t.Run("типы", func(t *testing.T) {
		if !validTypes[TypeBenefit] {
			t.Error("TypeBenefit should be valid")
		}
		if !validTypes[TypeActivity] {
			t.Error("TypeActivity should be valid")
		}
		if validTypes["invalid"] {
			t.Error("invalid should not be a valid type")
		}
	})
}

// TestService_GetCategories — тест получения категорий.
func TestService_GetCategories(t *testing.T) {
	t.Run("пустой список", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		cats, err := svc.GetCategories(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cats) != 0 {
			t.Errorf("expected 0 categories, got %d", len(cats))
		}
		if cats == nil {
			t.Error("expected non-nil empty slice")
		}
	})

	t.Run("с категориями", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		tid := uuid.New()
		for i := 0; i < 3; i++ {
			cat := EngagementCategory{
				ID:        uuid.New(),
				TenantID:  tid,
				Slug:      "cat" + string(rune('a'+i)),
				Name:      "Category " + string(rune('A'+i)),
				SortOrder: i,
			}
			repo.categories[cat.ID] = cat
		}

		cats, err := svc.GetCategories(context.Background(), tid)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cats) != 3 {
			t.Errorf("expected 3 categories, got %d", len(cats))
		}
	})
}

// TestService_AdminUpdateType — тест обновления типа.
func TestService_AdminUpdateType(t *testing.T) {
	t.Run("успешное обновление статуса", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		typeID := uuid.New()
		et := EngagementType{
			ID:       typeID,
			TenantID: uuid.New(),
			Status:   StatusDraft,
			Slug:     "test",
			Name:     "Test",
		}
		repo.types[typeID] = et

		et.Status = StatusActive
		result, err := svc.AdminUpdateType(context.Background(), et)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != StatusActive {
			t.Errorf("expected status %s, got %s", StatusActive, result.Status)
		}
	})

	t.Run("тип не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		et := EngagementType{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Status:   StatusActive,
			Slug:     "test",
			Name:     "Test",
		}

		_, err := svc.AdminUpdateType(context.Background(), et)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("дубликат slug при обновлении", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		tid := uuid.New()
		typeID1 := uuid.New()
		typeID2 := uuid.New()

		et1 := EngagementType{
			ID:       typeID1,
			TenantID: tid,
			Slug:     "gym",
			Name:     "Спортзал",
			Status:   StatusActive,
		}
		et2 := EngagementType{
			ID:       typeID2,
			TenantID: tid,
			Slug:     "fitness",
			Name:     "Фитнес",
			Status:   StatusActive,
		}
		repo.types[typeID1] = et1
		repo.types[typeID2] = et2

		// Попытка изменить slug et2 на "gym" (занято et1)
		et2.Slug = "gym"
		_, err := svc.AdminUpdateType(context.Background(), et2)
		if !errors.Is(err, ErrDuplicateSlug) {
			t.Fatalf("expected ErrDuplicateSlug, got: %v", err)
		}
	})

	t.Run("обновление slug на тот же — OK", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		typeID := uuid.New()
		et := EngagementType{
			ID:       typeID,
			TenantID: uuid.New(),
			Slug:     "gym",
			Name:     "Спортзал обновлённый",
			Status:   StatusActive,
		}
		repo.types[typeID] = et

		result, err := svc.AdminUpdateType(context.Background(), et)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Name != "Спортзал обновлённый" {
			t.Errorf("expected updated name, got %s", result.Name)
		}
	})
}

// TestService_GetOffersByType — тест получения офферов.
func TestService_GetOffersByType(t *testing.T) {
	t.Run("пустой список", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo, nil)

		offers, err := svc.GetOffersByType(context.Background(), uuid.New())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(offers) != 0 {
			t.Errorf("expected 0 offers, got %d", len(offers))
		}
		if offers == nil {
			t.Error("expected non-nil empty slice")
		}
	})
}

// =============================================================================
// EDGE CASE TESTS — расширенные тесты для boundary conditions и error paths
// =============================================================================

// --- ListTypes Edge Cases ---

func TestService_ListTypes_AllFiltersCombined(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	benefit := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "benefit-1",
		Name:     "Benefit 1",
	}
	repo.types[benefit.ID] = benefit

	// All filters: type=benefit, status=active, page=1, per_page=10
	_, total, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Page:     1,
		PerPage:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestService_ListTypes_NonExistentCategoryFilter(t *testing.T) {
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
		Category: &EngagementCategory{
			ID:   uuid.New(),
			Slug: "existing-category",
			Name: "Existing",
		},
	}
	repo.types[et.ID] = et

	types, total, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		Category: "nonexistent-category",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0 for non-existent category, got %d", total)
	}
	if len(types) != 0 {
		t.Errorf("expected 0 types, got %d", len(types))
	}
}

func TestService_ListTypes_PageNegative(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	_, _, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		Page:     -1,
		PerPage:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error for negative page: %v", err)
	}
}

func TestService_ListTypes_PerPageNegative(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	_, _, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		PerPage:  -5,
	})
	if err != nil {
		t.Fatalf("unexpected error for negative per_page: %v", err)
	}
}

func TestService_ListTypes_PerPageOverflow(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	_, _, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		PerPage:  1000,
	})
	if err != nil {
		t.Fatalf("unexpected error for large per_page: %v", err)
	}
}

func TestService_ListTypes_PerPageExactly100(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	_, _, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		PerPage:  100,
	})
	if err != nil {
		t.Fatalf("unexpected error for per_page=100: %v", err)
	}
}

func TestService_ListTypes_PerPage1(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	_, _, err := svc.ListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		PerPage:  1,
	})
	if err != nil {
		t.Fatalf("unexpected error for per_page=1: %v", err)
	}
}

func TestService_ListTypes_NilContext(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	_, _, err := svc.ListTypes(nil, CatalogFilter{TenantID: uuid.New()})
	// Не должно паниковать
	if err == nil {
		t.Log("ListTypes with nil context returned no error (acceptable)")
	}
}

// --- GetTypeByID Edge Cases ---

func TestService_GetTypeByID_HiddenStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(),
		Type:     TypeBenefit,
		Status:   StatusHidden,
		Slug:     "hidden",
		Name:     "Hidden Type",
	}
	repo.types[typeID] = et

	result, err := svc.GetTypeByID(context.Background(), typeID)
	if err != nil {
		t.Fatalf("unexpected error for hidden type: %v", err)
	}
	if result.Status != StatusHidden {
		t.Errorf("expected status 'hidden', got '%s'", result.Status)
	}
}

func TestService_GetTypeByID_DraftStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(),
		Type:     TypeBenefit,
		Status:   StatusDraft,
		Slug:     "draft",
		Name:     "Draft Type",
	}
	repo.types[typeID] = et

	result, err := svc.GetTypeByID(context.Background(), typeID)
	if err != nil {
		t.Fatalf("unexpected error for draft type: %v", err)
	}
	if result.Status != StatusDraft {
		t.Errorf("expected status 'draft', got '%s'", result.Status)
	}
}

func TestService_GetTypeByID_NilUUID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	_, err := svc.GetTypeByID(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for nil UUID, got: %v", err)
	}
}

func TestService_GetTypeByID_WithOffersError(t *testing.T) {
	offerErr := errors.New("offer query failed")
	repo := newMockRepo()
	repo.errKeys["GetOffersByType"] = true
	repo.errOn["GetOffersByType"] = offerErr
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(),
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "test",
		Name:     "Test",
	}
	repo.types[typeID] = et

	result, err := svc.GetTypeByID(context.Background(), typeID)
	if err != nil {
		t.Fatalf("unexpected error (offer error should be non-critical): %v", err)
	}
	if result.Offers == nil {
		t.Error("expected non-nil empty offers slice")
	}
	if len(result.Offers) != 0 {
		t.Errorf("expected 0 offers, got %d", len(result.Offers))
	}
}

// --- GetTypeByTenantID Edge Cases ---

func TestService_GetTypeByTenantID_TenantMismatch(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(), // другой tenant
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "test",
		Name:     "Test",
	}
	repo.types[typeID] = et

	_, err := svc.GetTypeByTenantID(context.Background(), uuid.New(), typeID)
	if !errors.Is(err, ErrTenantMismatch) {
		t.Fatalf("expected ErrTenantMismatch, got: %v", err)
	}
}

func TestService_GetTypeByTenantID_SameTenant(t *testing.T) {
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

	result, err := svc.GetTypeByTenantID(context.Background(), tid, typeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TenantID != tid {
		t.Errorf("expected tenant_id %s, got %s", tid, result.TenantID)
	}
}

func TestService_GetTypeByTenantID_NilContext(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	_, err := svc.GetTypeByTenantID(nil, uuid.New(), uuid.New())
	// Не должно паниковать
	if err == nil {
		t.Log("GetTypeByTenantID with nil context returned no error (acceptable)")
	}
}

// --- AdminCreateCategory Edge Cases ---

func TestService_AdminCreateCategory_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	cat := EngagementCategory{
		TenantID: uuid.New(),
		Slug:     "health",
		Name:     "",
	}

	_, err := svc.AdminCreateCategory(context.Background(), cat)
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter for empty name, got: %v", err)
	}
}

func TestService_AdminCreateCategory_MaxLengthName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	longName := ""
	for i := 0; i < 255; i++ {
		longName += "a"
	}
	cat := EngagementCategory{
		TenantID: uuid.New(),
		Slug:     "long",
		Name:     longName,
	}

	_, err := svc.AdminCreateCategory(context.Background(), cat)
	if err != nil {
		t.Fatalf("unexpected error for max length name: %v", err)
	}
}

func TestService_AdminCreateCategory_WithIcon(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	cat := EngagementCategory{
		TenantID:  uuid.New(),
		Slug:      "fitness",
		Name:      "Фитнес",
		Icon:      strPtr("dumbbell"),
		SortOrder: 5,
	}

	result, err := svc.AdminCreateCategory(context.Background(), cat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Icon == nil || *result.Icon != "dumbbell" {
		t.Errorf("expected icon 'dumbbell', got %v", result.Icon)
	}
}

func TestService_AdminCreateCategory_GetCategoriesError(t *testing.T) {
	dbErr := errors.New("categories query failed")
	repo := newMockRepo()
	repo.errKeys["GetCategories"] = true
	repo.errOn["GetCategories"] = dbErr
	svc := NewService(repo, nil)

	cat := EngagementCategory{
		TenantID: uuid.New(),
		Slug:     "health",
		Name:     "Здоровье",
	}

	_, err := svc.AdminCreateCategory(context.Background(), cat)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "check category slug") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

// --- AdminUpdateCategory Edge Cases ---

func TestService_AdminUpdateCategory_EmptySlug(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	cat := EngagementCategory{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Slug:     "",
		Name:     "Updated",
	}

	_, err := svc.AdminUpdateCategory(context.Background(), cat)
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter for empty slug, got: %v", err)
	}
}

func TestService_AdminUpdateCategory_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	cat := EngagementCategory{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Slug:     "health",
		Name:     "",
	}

	_, err := svc.AdminUpdateCategory(context.Background(), cat)
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter for empty name, got: %v", err)
	}
}

func TestService_AdminUpdateCategory_DuplicateSlugWithDifferentID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	cat1 := EngagementCategory{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "health",
		Name:     "Здоровье",
	}
	cat2 := EngagementCategory{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "fitness",
		Name:     "Фитнес",
	}
	repo.categories[cat1.ID] = cat1
	repo.categories[cat2.ID] = cat2

	// Попытка обновить cat2 slug на "health" (занято cat1)
	cat2.Slug = "health"
	_, err := svc.AdminUpdateCategory(context.Background(), cat2)
	if !errors.Is(err, ErrDuplicateSlug) {
		t.Fatalf("expected ErrDuplicateSlug, got: %v", err)
	}
}

func TestService_AdminUpdateCategory_SameSlugSameID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	cat := EngagementCategory{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "health",
		Name:     "Здоровье обновлённое",
	}
	repo.categories[cat.ID] = cat

	_, err := svc.AdminUpdateCategory(context.Background(), cat)
	if err != nil {
		t.Fatalf("unexpected error for same slug same ID: %v", err)
	}
}

// --- AdminCreateType Edge Cases ---

func TestService_AdminCreateType_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	et := EngagementType{
		TenantID: uuid.New(),
		Slug:     "test",
		Name:     "",
	}

	_, err := svc.AdminCreateType(context.Background(), et)
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter for empty name, got: %v", err)
	}
}

func TestService_AdminCreateType_ActivityType(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	et := EngagementType{
		TenantID: uuid.New(),
		Slug:     "yoga-event",
		Name:     "Йога-мероприятие",
		Type:     TypeActivity,
	}

	result, err := svc.AdminCreateType(context.Background(), et)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != TypeActivity {
		t.Errorf("expected type 'activity', got '%s'", result.Type)
	}
}

func TestService_AdminCreateType_WithCategory(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	catID := uuid.New()
	cat := EngagementCategory{
		ID:       catID,
		TenantID: tid,
		Slug:     "fitness",
		Name:     "Фитнес",
	}
	repo.categories[catID] = cat

	et := EngagementType{
		TenantID:   tid,
		CategoryID: catID,
		Slug:       "gym",
		Name:       "Спортзал",
	}

	result, err := svc.AdminCreateType(context.Background(), et)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CategoryID != catID {
		t.Errorf("expected category_id %s, got %s", catID, result.CategoryID)
	}
}

func TestService_AdminCreateType_GetCategoriesError(t *testing.T) {
	dbErr := errors.New("categories query failed")
	repo := newMockRepo()
	repo.errKeys["GetCategories"] = true
	repo.errOn["GetCategories"] = dbErr
	svc := NewService(repo, nil)

	et := EngagementType{
		TenantID:   uuid.New(),
		CategoryID: uuid.New(),
		Slug:       "test",
		Name:       "Test",
	}

	_, err := svc.AdminCreateType(context.Background(), et)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "check category") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestService_AdminCreateType_DuplicateSlug(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	existing := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Type:     TypeBenefit,
		Status:   StatusActive,
	}
	repo.types[existing.ID] = existing

	et := EngagementType{
		TenantID: tid,
		Slug:     "gym",
		Name:     "Другой спортзал",
	}

	_, err := svc.AdminCreateType(context.Background(), et)
	if !errors.Is(err, ErrDuplicateSlug) {
		t.Fatalf("expected ErrDuplicateSlug, got: %v", err)
	}
}

func TestService_AdminCreateType_GetTypeBySlugError(t *testing.T) {
	dbErr := errors.New("slug check failed")
	repo := newMockRepo()
	repo.errKeys["GetTypeBySlug"] = true
	repo.errOn["GetTypeBySlug"] = dbErr
	svc := NewService(repo, nil)

	et := EngagementType{
		TenantID: uuid.New(),
		Slug:     "gym",
		Name:     "Спортзал",
	}

	_, err := svc.AdminCreateType(context.Background(), et)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "check type slug") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

// --- AdminUpdateType Edge Cases ---

func TestService_AdminUpdateType_EmptySlug(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	et := EngagementType{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Slug:     "",
		Name:     "Test",
	}

	_, err := svc.AdminUpdateType(context.Background(), et)
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter for empty slug, got: %v", err)
	}
}

func TestService_AdminUpdateType_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	et := EngagementType{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Slug:     "test",
		Name:     "",
	}

	_, err := svc.AdminUpdateType(context.Background(), et)
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter for empty name, got: %v", err)
	}
}

func TestService_AdminUpdateType_CategoryNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: tid,
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	// Попытка обновить с несуществующей категорией
	et.CategoryID = uuid.New()
	_, err := svc.AdminUpdateType(context.Background(), et)
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got: %v", err)
	}
}

func TestService_AdminUpdateType_GetCategoriesError(t *testing.T) {
	dbErr := errors.New("categories query failed")
	repo := newMockRepo()
	repo.errKeys["GetCategories"] = true
	repo.errOn["GetCategories"] = dbErr
	svc := NewService(repo, nil)

	tid := uuid.New()
	typeID := uuid.New()
	et := EngagementType{
		ID:         typeID,
		TenantID:   tid,
		CategoryID: uuid.New(),
		Slug:       "gym",
		Name:       "Спортзал",
		Status:     StatusActive,
	}
	repo.types[typeID] = et

	_, err := svc.AdminUpdateType(context.Background(), et)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- AdminDeleteType Edge Cases ---

func TestService_AdminDeleteType_WithActiveOffers(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(),
		Status:   StatusActive,
		Slug:     "gym",
		Name:     "Спортзал",
	}
	repo.types[typeID] = et

	offer := EngagementOffer{
		ID:               uuid.New(),
		EngagementTypeID: typeID,
		Name:             "Basic",
	}
	repo.offers[offer.ID] = offer

	// TODO: F2 — должно проверять активные активации
	err := svc.AdminDeleteType(context.Background(), typeID)
	if err != nil {
		t.Fatalf("unexpected error (F2 check not implemented yet): %v", err)
	}
}

func TestService_AdminDeleteType_NilUUID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	err := svc.AdminDeleteType(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for nil UUID, got: %v", err)
	}
}

// --- AdminCreateOffer Edge Cases ---

func TestService_AdminCreateOffer_NegativeCost(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(),
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	offer := EngagementOffer{
		EngagementTypeID: typeID,
		Name:             "Negative Cost",
		CostCents:        -100,
	}

	result, err := svc.AdminCreateOffer(context.Background(), offer)
	if err != nil {
		t.Fatalf("unexpected error (negative cost not validated at service level): %v", err)
	}
	if result.CostCents != -100 {
		t.Errorf("expected cost_cents -100, got %d", result.CostCents)
	}
}

func TestService_AdminCreateOffer_ZeroCost(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(),
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	offer := EngagementOffer{
		EngagementTypeID: typeID,
		Name:             "Free",
		CostCents:        0,
	}

	result, err := svc.AdminCreateOffer(context.Background(), offer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CostCents != 0 {
		t.Errorf("expected cost_cents 0, got %d", result.CostCents)
	}
}

func TestService_AdminCreateOffer_MaxCost(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	typeID := uuid.New()
	et := EngagementType{
		ID:       typeID,
		TenantID: uuid.New(),
		Slug:     "gym",
		Name:     "Спортзал",
		Status:   StatusActive,
	}
	repo.types[typeID] = et

	offer := EngagementOffer{
		EngagementTypeID: typeID,
		Name:             "Premium",
		CostCents:        999999999,
	}

	result, err := svc.AdminCreateOffer(context.Background(), offer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CostCents != 999999999 {
		t.Errorf("expected cost_cents 999999999, got %d", result.CostCents)
	}
}

// --- AdminUpdateOffer Edge Cases ---

func TestService_AdminUpdateOffer_EmptyName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	offer := EngagementOffer{
		ID:   uuid.New(),
		Name: "",
	}

	_, err := svc.AdminUpdateOffer(context.Background(), offer)
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("expected ErrInvalidFilter for empty name, got: %v", err)
	}
}

// --- AdminListTypes Edge Cases ---

func TestService_AdminListTypes_AllStatuses(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	for _, status := range []string{StatusDraft, StatusActive, StatusPromo, StatusHidden, StatusCompleted} {
		et := EngagementType{
			ID:       uuid.New(),
			TenantID: tid,
			Slug:     "type-" + status,
			Name:     "Type " + status,
			Status:   status,
		}
		repo.types[et.ID] = et
	}

	types, total, err := svc.AdminListTypes(context.Background(), CatalogFilter{TenantID: tid})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5 (all statuses), got %d", total)
	}
	if len(types) != 5 {
		t.Errorf("expected 5 types, got %d", len(types))
	}
}

func TestService_AdminListTypes_FilterByType(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	benefit := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeBenefit,
		Status:   StatusActive,
		Slug:     "benefit-1",
		Name:     "Benefit 1",
	}
	activity := EngagementType{
		ID:       uuid.New(),
		TenantID: tid,
		Type:     TypeActivity,
		Status:   StatusActive,
		Slug:     "activity-1",
		Name:     "Activity 1",
	}
	repo.types[benefit.ID] = benefit
	repo.types[activity.ID] = activity

	types, total, err := svc.AdminListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		Type:     TypeBenefit,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(types) != 1 {
		t.Errorf("expected 1 type, got %d", len(types))
	}
}

func TestService_AdminListTypes_InvalidType(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	_, _, err := svc.AdminListTypes(context.Background(), CatalogFilter{
		TenantID: uuid.New(),
		Type:     "invalid",
	})
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got: %v", err)
	}
}

func TestService_AdminListTypes_PaginationEdge(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	_, _, err := svc.AdminListTypes(context.Background(), CatalogFilter{
		TenantID: tid,
		Page:     0,
		PerPage:  0,
	})
	if err != nil {
		t.Fatalf("unexpected error for default pagination: %v", err)
	}
}

// --- AdminDeleteCategory Edge Cases ---

func TestService_AdminDeleteCategory_NilUUID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	err := svc.AdminDeleteCategory(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound for nil UUID, got: %v", err)
	}
}

// --- Concurrent tests (sequential to avoid mock race) ---

func TestService_ListTypes_SequentialMultiple(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	for i := 0; i < 10; i++ {
		et := EngagementType{
			ID:       uuid.New(),
			TenantID: tid,
			Type:     TypeBenefit,
			Status:   StatusActive,
			Slug:     "sequential-" + string(rune('a'+i)),
			Name:     "Sequential " + string(rune('A'+i)),
		}
		repo.types[et.ID] = et
	}

	// Sequential calls (concurrent would require sync.Mutex in mock)
	for i := 0; i < 5; i++ {
		types, total, err := svc.ListTypes(context.Background(), CatalogFilter{TenantID: tid})
		if err != nil {
			t.Errorf("list call %d error: %v", i, err)
		}
		if total != 10 {
			t.Errorf("list call %d: expected total 10, got %d", i, total)
		}
		if len(types) != 10 {
			t.Errorf("list call %d: expected 10 types, got %d", i, len(types))
		}
	}
}

func TestService_AdminCreateType_SequentialMultiple(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, nil)

	tid := uuid.New()
	slugBase := "sequential"

	for i := 0; i < 3; i++ {
		et := EngagementType{
			TenantID: tid,
			Slug:     slugBase + string(rune('a'+i)),
			Name:     "Sequential " + string(rune('A'+i)),
		}
		_, err := svc.AdminCreateType(context.Background(), et)
		if err != nil {
			t.Errorf("create call %d error: %v", i, err)
		}
	}
}
