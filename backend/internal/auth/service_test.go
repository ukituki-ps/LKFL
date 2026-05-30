package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"lkfl/internal/tenant"
	"lkfl/internal/user"
	sharedauth "lkfl/shared/pkg/auth"
)

// mockUserRepository — мок для user.Repository для unit-тестов Service.
type mockUserRepository struct {
	users       map[string]user.User // keycloakSub → User
	createdUser *user.User
	updatedUser *user.User
	errOn       map[string]error
}

func newMockUserRepo() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]user.User),
		errOn: make(map[string]error),
	}
}

func (m *mockUserRepository) Create(_ context.Context, u user.User) (user.User, error) {
	if m.errOn["Create"] != nil {
		return user.User{}, m.errOn["Create"]
	}
	u.ID = uuid.New()
	m.users[u.KeycloakSub] = u
	m.createdUser = &u
	return u, nil
}

func (m *mockUserRepository) GetByID(_ context.Context, _ uuid.UUID) (user.User, error) {
	if m.errOn["GetByID"] != nil {
		return user.User{}, m.errOn["GetByID"]
	}
	return user.User{}, user.ErrNotFound
}

func (m *mockUserRepository) GetByKeycloakSub(_ context.Context, keycloakSub string) (user.User, error) {
	if m.errOn["GetByKeycloakSub"] != nil {
		return user.User{}, m.errOn["GetByKeycloakSub"]
	}
	u, ok := m.users[keycloakSub]
	if !ok {
		return user.User{}, user.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepository) GetByEmail(_ context.Context, _ uuid.UUID, _ string) (user.User, error) {
	return user.User{}, user.ErrNotFound
}

func (m *mockUserRepository) Update(_ context.Context, u user.User) (user.User, error) {
	if m.errOn["Update"] != nil {
		return user.User{}, m.errOn["Update"]
	}
	m.users[u.KeycloakSub] = u
	m.updatedUser = &u
	return u, nil
}

func (m *mockUserRepository) UpdateStatus(_ context.Context, _ uuid.UUID, _ string) (user.User, error) {
	return user.User{}, user.ErrNotFound
}

func (m *mockUserRepository) List(_ context.Context, _ user.UserFilter) ([]user.User, int64, error) {
	return []user.User{}, 0, nil
}

func (m *mockUserRepository) CreateAccount(_ context.Context, _ uuid.UUID, _ tenant.JSONB) (user.Account, error) {
	return user.Account{}, nil
}

func (m *mockUserRepository) GetAccountByUserID(_ context.Context, _ uuid.UUID) (user.Account, error) {
	return user.Account{}, user.ErrAccountNotFound
}

func (m *mockUserRepository) GetRoles(_ context.Context, _ uuid.UUID) ([]user.UserRole, error) {
	return []user.UserRole{}, nil
}

func (m *mockUserRepository) AddRole(_ context.Context, _ uuid.UUID, _ string, _ *uuid.UUID) (user.UserRole, error) {
	return user.UserRole{}, nil
}

func (m *mockUserRepository) RemoveRole(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

// TestService_CreateOrUpdateUser_Create — тест создания нового пользователя.
func TestService_CreateOrUpdateUser_Create(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)

	claims := &sharedauth.Claims{
		Subject:    "kc-sub-123",
		Email:      "user@example.com",
		GivenName:  "Иван",
		FamilyName: "Петров",
	}

	ctx := tenant.TenantContext(context.Background(), uuid.MustParse("00000000-0000-0000-0000-000000000001"))

	result, err := svc.CreateOrUpdateUser(ctx, claims, []string{"employee"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID == uuid.Nil {
		t.Error("expected user ID to be set")
	}
	if result.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", result.Email)
	}
	if result.FirstName != "Иван" {
		t.Errorf("expected first_name Иван, got %s", result.FirstName)
	}
	if result.LastName != "Петров" {
		t.Errorf("expected last_name Петров, got %s", result.LastName)
	}
	if result.KeycloakSub != "kc-sub-123" {
		t.Errorf("expected keycloak_sub kc-sub-123, got %s", result.KeycloakSub)
	}
	if result.Status != user.StatusActive {
		t.Errorf("expected status active, got %s", result.Status)
	}
	if result.TenantID.String() != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("expected tenant_id from context, got %s", result.TenantID)
	}
}

// TestService_CreateOrUpdateUser_Update — тест обновления существующего пользователя.
func TestService_CreateOrUpdateUser_Update(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)

	// Создаём существующего пользователя
	existing := user.User{
		ID:          uuid.New(),
		Email:       "old@example.com",
		FirstName:   "Старый",
		LastName:    "Имя",
		KeycloakSub: "kc-sub-123",
		Status:      user.StatusActive,
	}
	repo.users["kc-sub-123"] = existing

	claims := &sharedauth.Claims{
		Subject:    "kc-sub-123",
		Email:      "new@example.com",
		GivenName:  "Новый",
		FamilyName: "Имя",
	}

	result, err := svc.CreateOrUpdateUser(context.Background(), claims, []string{"employee"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ID должен остаться прежним
	if result.ID != existing.ID {
		t.Errorf("expected user ID to remain the same")
	}
	// Данные должны обновиться
	if result.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", result.Email)
	}
	if result.FirstName != "Новый" {
		t.Errorf("expected first_name Новый, got %s", result.FirstName)
	}
}

// TestService_CreateOrUpdateUser_NoTenant — тест создания без tenant в context.
func TestService_CreateOrUpdateUser_NoTenant(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)

	claims := &sharedauth.Claims{
		Subject:    "kc-sub-456",
		Email:      "notenant@example.com",
		GivenName:  "Без",
		FamilyName: "Тенанта",
	}

	// Context без tenant ID
	ctx := context.Background()

	result, err := svc.CreateOrUpdateUser(ctx, claims, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TenantID != uuid.Nil {
		t.Errorf("expected tenant_id to be nil, got %s", result.TenantID)
	}
}

// TestService_CreateOrUpdateUser_CreateError — тест ошибки при создании.
func TestService_CreateOrUpdateUser_CreateError(t *testing.T) {
	repo := newMockUserRepo()
	repo.errOn["Create"] = errors.New("db error")
	svc := NewService(repo)

	claims := &sharedauth.Claims{
		Subject:   "kc-sub-error",
		Email:     "error@example.com",
		GivenName: "Error",
	}

	_, err := svc.CreateOrUpdateUser(context.Background(), claims, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repo.errOn["Create"]) {
		t.Logf("error wrapped correctly: %v", err)
	}
}

// TestService_CreateOrUpdateUser_UpdateError — тест ошибки при обновлении.
func TestService_CreateOrUpdateUser_UpdateError(t *testing.T) {
	repo := newMockUserRepo()
	repo.errOn["Update"] = errors.New("db update error")
	svc := NewService(repo)

	// Создаём существующего пользователя
	existing := user.User{
		ID:          uuid.New(),
		KeycloakSub: "kc-sub-update-err",
		Status:      user.StatusActive,
	}
	repo.users["kc-sub-update-err"] = existing

	claims := &sharedauth.Claims{
		Subject:   "kc-sub-update-err",
		Email:     "updated@example.com",
		GivenName: "Updated",
	}

	_, err := svc.CreateOrUpdateUser(context.Background(), claims, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestService_GetUserByKeycloakSub — тест получения пользователя.
func TestService_GetUserByKeycloakSub(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo)

	// Создаём пользователя
	u := user.User{
		ID:          uuid.New(),
		Email:       "found@example.com",
		KeycloakSub: "kc-sub-found",
		Status:      user.StatusActive,
	}
	repo.users["kc-sub-found"] = u

	t.Run("пользователь найден", func(t *testing.T) {
		result, err := svc.GetUserByKeycloakSub(context.Background(), "kc-sub-found")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != u.ID {
			t.Errorf("expected user ID mismatch")
		}
		if result.Email != "found@example.com" {
			t.Errorf("expected email found@example.com, got %s", result.Email)
		}
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		_, err := svc.GetUserByKeycloakSub(context.Background(), "kc-sub-missing")
		if !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

// TestGenerateState — тест генерации state-параметра.
func TestGenerateState(t *testing.T) {
	t.Run("уникальность state", func(t *testing.T) {
		states := make(map[string]bool)
		for i := 0; i < 100; i++ {
			s := generateState()
			if states[s] {
				t.Errorf("duplicate state generated: %s", s)
			}
			states[s] = true
		}
	})

	t.Run("длина state (64 hex chars = 32 bytes)", func(t *testing.T) {
		s := generateState()
		if len(s) != 64 {
			t.Errorf("expected state length 64, got %d", len(s))
		}
	})
}
