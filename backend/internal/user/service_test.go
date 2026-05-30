package user

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"lkfl/internal/tenant"
)

// mockRepository — мока для Repository для unit-тестов Service.
type mockRepository struct {
	users    map[uuid.UUID]User
	accounts map[uuid.UUID]Account
	roles    map[uuid.UUID][]UserRole
	nextID   int
	errOn    map[string]error
	errKeys  map[string]bool
}

func newMockRepo() *mockRepository {
	return &mockRepository{
		users:    make(map[uuid.UUID]User),
		accounts: make(map[uuid.UUID]Account),
		roles:    make(map[uuid.UUID][]UserRole),
		errOn:    make(map[string]error),
		errKeys:  make(map[string]bool),
	}
}

func (m *mockRepository) nextUUID() uuid.UUID {
	m.nextID++
	return uuid.New()
}

func (m *mockRepository) Create(ctx context.Context, u User) (User, error) {
	if m.errKeys["Create"] {
		return User{}, m.errOn["Create"]
	}
	u.ID = m.nextUUID()
	m.users[u.ID] = u
	return u, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	if m.errKeys["GetByID"] {
		return User{}, m.errOn["GetByID"]
	}
	u, ok := m.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (m *mockRepository) GetByKeycloakSub(ctx context.Context, keycloakSub string) (User, error) {
	if m.errKeys["GetByKeycloakSub"] {
		return User{}, m.errOn["GetByKeycloakSub"]
	}
	for _, u := range m.users {
		if u.KeycloakSub == keycloakSub {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}

func (m *mockRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (User, error) {
	if m.errKeys["GetByEmail"] {
		return User{}, m.errOn["GetByEmail"]
	}
	for _, u := range m.users {
		if u.TenantID == tenantID && u.Email == email {
			return u, nil
		}
	}
	return User{}, ErrNotFound
}

func (m *mockRepository) Update(ctx context.Context, u User) (User, error) {
	if m.errKeys["Update"] {
		return User{}, m.errOn["Update"]
	}
	existing, ok := m.users[u.ID]
	if !ok {
		return User{}, ErrNotFound
	}
	existing.Email = u.Email
	existing.FirstName = u.FirstName
	existing.LastName = u.LastName
	existing.Phone = u.Phone
	existing.Metadata = u.Metadata
	m.users[u.ID] = existing
	return existing, nil
}

func (m *mockRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (User, error) {
	if m.errKeys["UpdateStatus"] {
		return User{}, m.errOn["UpdateStatus"]
	}
	u, ok := m.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	u.Status = status
	m.users[id] = u
	return u, nil
}

func (m *mockRepository) List(ctx context.Context, filter UserFilter) ([]User, int64, error) {
	if m.errKeys["List"] {
		return nil, 0, m.errOn["List"]
	}
	var result []User
	for _, u := range m.users {
		if filter.Status != "" && u.Status != filter.Status {
			continue
		}
		result = append(result, u)
	}
	if result == nil {
		result = []User{}
	}
	return result, int64(len(result)), nil
}

func (m *mockRepository) CreateAccount(ctx context.Context, userID uuid.UUID, settings tenant.JSONB) (Account, error) {
	if m.errKeys["CreateAccount"] {
		return Account{}, m.errOn["CreateAccount"]
	}
	a := Account{
		ID:           m.nextUUID(),
		UserID:       userID,
		TotalBalance: 0,
		Settings:     settings,
	}
	m.accounts[userID] = a
	return a, nil
}

func (m *mockRepository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (Account, error) {
	if m.errKeys["GetAccountByUserID"] {
		return Account{}, m.errOn["GetAccountByUserID"]
	}
	a, ok := m.accounts[userID]
	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return a, nil
}

func (m *mockRepository) GetRoles(ctx context.Context, userID uuid.UUID) ([]UserRole, error) {
	if m.errKeys["GetRoles"] {
		return nil, m.errOn["GetRoles"]
	}
	roles, ok := m.roles[userID]
	if !ok {
		return []UserRole{}, nil
	}
	return roles, nil
}

func (m *mockRepository) AddRole(ctx context.Context, userID uuid.UUID, role string, grantedBy *uuid.UUID) (UserRole, error) {
	if m.errKeys["AddRole"] {
		return UserRole{}, m.errOn["AddRole"]
	}
	ur := UserRole{
		ID:        m.nextUUID(),
		UserID:    userID,
		Role:      role,
		GrantedBy: grantedBy,
	}
	m.roles[userID] = append(m.roles[userID], ur)
	return ur, nil
}

func (m *mockRepository) RemoveRole(ctx context.Context, userID uuid.UUID, role string) error {
	if m.errKeys["RemoveRole"] {
		return m.errOn["RemoveRole"]
	}
	roles, ok := m.roles[userID]
	if !ok {
		return ErrRoleNotFound
	}
	newRoles := make([]UserRole, 0, len(roles))
	found := false
	for _, r := range roles {
		if r.Role == role {
			found = true
		} else {
			newRoles = append(newRoles, r)
		}
	}
	if !found {
		return ErrRoleNotFound
	}
	m.roles[userID] = newRoles
	return nil
}

// TestService_Deactivate — тест деактивации пользователя.
func TestService_Deactivate(t *testing.T) {
	t.Run("успешная деактивация", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:     uuid.New(),
			Email:  "test@example.com",
			Status: StatusActive,
		}
		repo.users[user.ID] = user

		result, err := svc.Deactivate(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != StatusDeactivated {
			t.Errorf("expected status %s, got %s", StatusDeactivated, result.Status)
		}
	})

	t.Run("нельзя деактивировать уже деактивированного", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:     uuid.New(),
			Email:  "test@example.com",
			Status: StatusDeactivated,
		}
		repo.users[user.ID] = user

		_, err := svc.Deactivate(context.Background(), user.ID)
		if !errors.Is(err, ErrAlreadyDeactivated) {
			t.Fatalf("expected ErrAlreadyDeactivated, got: %v", err)
		}
	})

	t.Run("нельзя деактивировать удалённого", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:     uuid.New(),
			Email:  "test@example.com",
			Status: StatusDeleted,
		}
		repo.users[user.ID] = user

		_, err := svc.Deactivate(context.Background(), user.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		_, err := svc.Deactivate(context.Background(), uuid.New())
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

// TestService_Activate — тест активации пользователя.
func TestService_Activate(t *testing.T) {
	t.Run("успешная активация", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:     uuid.New(),
			Email:  "test@example.com",
			Status: StatusDeactivated,
		}
		repo.users[user.ID] = user

		result, err := svc.Activate(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Status != StatusActive {
			t.Errorf("expected status %s, got %s", StatusActive, result.Status)
		}
	})

	t.Run("нельзя активировать уже активного", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:     uuid.New(),
			Email:  "test@example.com",
			Status: StatusActive,
		}
		repo.users[user.ID] = user

		_, err := svc.Activate(context.Background(), user.ID)
		if !errors.Is(err, ErrInvalidStatus) {
			t.Fatalf("expected ErrInvalidStatus, got: %v", err)
		}
	})
}

// TestService_UpdateProfile — тест обновления профиля.
func TestService_UpdateProfile(t *testing.T) {
	t.Run("успешное обновление", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:        uuid.New(),
			Email:     "old@example.com",
			FirstName: "Old",
			Status:    StatusActive,
		}
		repo.users[user.ID] = user

		result, err := svc.UpdateProfile(context.Background(), user.ID, "new@example.com", "New", "Name", "+1234567890")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Email != "new@example.com" {
			t.Errorf("expected email new@example.com, got %s", result.Email)
		}
		if result.FirstName != "New" {
			t.Errorf("expected first_name New, got %s", result.FirstName)
		}
	})

	t.Run("нельзя обновить деактивированный профиль", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:     uuid.New(),
			Email:  "test@example.com",
			Status: StatusDeactivated,
		}
		repo.users[user.ID] = user

		_, err := svc.UpdateProfile(context.Background(), user.ID, "", "New", "", "")
		if !errors.Is(err, ErrUserDeactivated) {
			t.Fatalf("expected ErrUserDeactivated, got: %v", err)
		}
	})

	t.Run("дубликат email", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		tid := uuid.New()
		user1 := User{
			ID:       uuid.New(),
			Email:    "existing@example.com",
			Status:   StatusActive,
			TenantID: tid,
		}
		user2 := User{
			ID:       uuid.New(),
			Email:    "other@example.com",
			Status:   StatusActive,
			TenantID: tid,
		}
		repo.users[user1.ID] = user1
		repo.users[user2.ID] = user2

		// Пытаемся поменять email user2 на email user1
		// Устанавливаем tenant ID в context для проверки уникальности
		ctx := tenant.TenantContext(context.Background(), tid)
		_, err := svc.UpdateProfile(ctx, user2.ID, "existing@example.com", "", "", "")
		if !errors.Is(err, ErrDuplicateEmail) {
			t.Fatalf("expected ErrDuplicateEmail, got: %v", err)
		}
	})
}

// TestService_AddRole — тест добавления роли.
func TestService_AddRole(t *testing.T) {
	t.Run("успешное добавление роли", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}
		repo.users[user.ID] = user

		result, err := svc.AddRole(context.Background(), user.ID, RoleHR, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Role != RoleHR {
			t.Errorf("expected role %s, got %s", RoleHR, result.Role)
		}
	})

	t.Run("недопустимая роль", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}
		repo.users[user.ID] = user

		_, err := svc.AddRole(context.Background(), user.ID, "superadmin", nil)
		if !errors.Is(err, ErrInvalidRole) {
			t.Fatalf("expected ErrInvalidRole, got: %v", err)
		}
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		_, err := svc.AddRole(context.Background(), uuid.New(), RoleHR, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got: %v", err)
		}
	})
}

// TestService_RemoveRole — тест удаления роли.
func TestService_RemoveRole(t *testing.T) {
	t.Run("успешное удаление роли", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{ID: uuid.New()}
		repo.users[user.ID] = user
		repo.roles[user.ID] = []UserRole{{Role: RoleHR}}

		err := svc.RemoveRole(context.Background(), user.ID, RoleHR)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("роль не найдена", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		user := User{ID: uuid.New()}
		repo.users[user.ID] = user
		repo.roles[user.ID] = []UserRole{{Role: RoleEmployee}}

		err := svc.RemoveRole(context.Background(), user.ID, RoleHR)
		if !errors.Is(err, ErrRoleNotFound) {
			t.Fatalf("expected ErrRoleNotFound, got: %v", err)
		}
	})
}

// TestService_List — тест списка пользователей.
func TestService_List(t *testing.T) {
	t.Run("с пагинацией по умолчанию", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		// Добавляем пользователей
		for i := 0; i < 5; i++ {
			user := User{
				ID:     uuid.New(),
				Email:  "test" + string(rune('a'+i)) + "@example.com",
				Status: StatusActive,
			}
			repo.users[user.ID] = user
		}

		// Пустой фильтр — должно использовать defaults
		users, total, err := svc.List(context.Background(), UserFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(users) != 5 {
			t.Errorf("expected 5 users, got %d", len(users))
		}
	})

	t.Run("фильтр по статусу", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		for i := 0; i < 3; i++ {
			user := User{
				ID:     uuid.New(),
				Email:  "active" + string(rune('a'+i)) + "@example.com",
				Status: StatusActive,
			}
			repo.users[user.ID] = user
		}
		for i := 0; i < 2; i++ {
			user := User{
				ID:     uuid.New(),
				Email:  "deactivated" + string(rune('a'+i)) + "@example.com",
				Status: StatusDeactivated,
			}
			repo.users[user.ID] = user
		}

		_, total, err := svc.List(context.Background(), UserFilter{Status: StatusActive})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected total 3, got %d", total)
		}
	})

	t.Run("валидация page/per_page", func(t *testing.T) {
		repo := newMockRepo()
		svc := NewService(repo)

		// page=0 → должно стать 1
		// per_page=0 → должно стать 20
		// per_page=200 → должно стать 100
		_, _, err := svc.List(context.Background(), UserFilter{Page: 0, PerPage: 200})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestUser_ToProfile — тест преобразования User в UserProfile.
func TestUser_ToProfile(t *testing.T) {
	user := User{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Email:       "test@example.com",
		FirstName:   "Test",
		LastName:    "User",
		Phone:       "+1234567890",
		Status:      StatusActive,
		KeycloakSub: "secret-subject",
	}

	profile := user.ToProfile()

	if profile.ID != user.ID {
		t.Errorf("profile ID mismatch")
	}
	if profile.Email != user.Email {
		t.Errorf("profile email mismatch")
	}
	// KeycloakSub не должен быть в профиле (json:"-" в User)
	// Это проверяется на уровне json marshalling, но логически:
	// profile не имеет поля KeycloakSub — OK
}

// TestValidRoles — тест валидации ролей.
func TestValidRoles(t *testing.T) {
	roles := []string{RoleEmployee, RoleHR, RoleCatalogManager, RoleAdmin}
	for _, role := range roles {
		if !validRoles[role] {
			t.Errorf("role %s should be valid", role)
		}
	}
	if validRoles["superadmin"] {
		t.Error("superadmin should not be a valid role")
	}
}

// =============================================================================
// EDGE CASE TESTS — расширенные тесты для boundary conditions и error paths
// =============================================================================

// --- UpdateProfile Edge Cases ---

func TestService_UpdateProfile_ProfileNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.UpdateProfile(context.Background(), uuid.New(), "", "", "", "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestService_UpdateProfile_EmptyNameFields(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Original",
		Status:    StatusActive,
	}
	repo.users[user.ID] = user

	// Пустые поля не должны изменять профиль
	result, err := svc.UpdateProfile(context.Background(), user.ID, "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FirstName != "Original" {
		t.Errorf("expected firstName 'Original', got '%s'", result.FirstName)
	}
}

func TestService_UpdateProfile_OnlyEmail(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	tid := uuid.New()
	user := User{
		ID:        uuid.New(),
		Email:     "old@example.com",
		FirstName: "Test",
		Status:    StatusActive,
		TenantID:  tid,
	}
	repo.users[user.ID] = user

	ctx := tenant.TenantContext(context.Background(), tid)
	result, err := svc.UpdateProfile(ctx, user.ID, "new@example.com", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got '%s'", result.Email)
	}
	if result.FirstName != "Test" {
		t.Errorf("expected firstName unchanged, got '%s'", result.FirstName)
	}
}

func TestService_UpdateProfile_OnlyFirstName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Old",
		Status:    StatusActive,
	}
	repo.users[user.ID] = user

	result, err := svc.UpdateProfile(context.Background(), user.ID, "", "New", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FirstName != "New" {
		t.Errorf("expected firstName 'New', got '%s'", result.FirstName)
	}
	if result.Email != "test@example.com" {
		t.Errorf("expected email unchanged, got '%s'", result.Email)
	}
}

func TestService_UpdateProfile_OnlyLastName(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		LastName: "Old",
		Status:   StatusActive,
	}
	repo.users[user.ID] = user

	result, err := svc.UpdateProfile(context.Background(), user.ID, "", "", "New", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LastName != "New" {
		t.Errorf("expected lastName 'New', got '%s'", result.LastName)
	}
}

func TestService_UpdateProfile_OnlyPhone(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusActive,
	}
	repo.users[user.ID] = user

	result, err := svc.UpdateProfile(context.Background(), user.ID, "", "", "", "+79991234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Phone != "+79991234567" {
		t.Errorf("expected phone '+79991234567', got '%s'", result.Phone)
	}
}

func TestService_UpdateProfile_SameEmailNoConflict(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	tid := uuid.New()
	user := User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Status:   StatusActive,
		TenantID: tid,
	}
	repo.users[user.ID] = user

	ctx := tenant.TenantContext(context.Background(), tid)
	result, err := svc.UpdateProfile(ctx, user.ID, "test@example.com", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error when updating to same email: %v", err)
	}
	if result.Email != "test@example.com" {
		t.Errorf("expected email unchanged, got '%s'", result.Email)
	}
}

func TestService_UpdateProfile_EmailUniquenessCheckError(t *testing.T) {
	dbErr := errors.New("connection timeout")
	repo := newMockRepo()
	repo.errKeys["GetByEmail"] = true
	repo.errOn["GetByEmail"] = dbErr
	svc := NewService(repo)

	tid := uuid.New()
	user := User{
		ID:       uuid.New(),
		Email:    "old@example.com",
		Status:   StatusActive,
		TenantID: tid,
	}
	repo.users[user.ID] = user

	ctx := tenant.TenantContext(context.Background(), tid)
	_, err := svc.UpdateProfile(ctx, user.ID, "new@example.com", "", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "check email uniqueness") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestService_UpdateProfile_DeactivatedUser(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusDeactivated,
	}
	repo.users[user.ID] = user

	_, err := svc.UpdateProfile(context.Background(), user.ID, "new@example.com", "", "", "")
	if !errors.Is(err, ErrUserDeactivated) {
		t.Fatalf("expected ErrUserDeactivated, got: %v", err)
	}
}

func TestService_UpdateProfile_GetByIDError(t *testing.T) {
	dbErr := errors.New("db error")
	repo := newMockRepo()
	repo.errKeys["GetByID"] = true
	repo.errOn["GetByID"] = dbErr
	svc := NewService(repo)

	_, err := svc.UpdateProfile(context.Background(), uuid.New(), "", "", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_UpdateProfile_UpdateError(t *testing.T) {
	dbErr := errors.New("update failed")
	repo := newMockRepo()
	repo.errKeys["Update"] = true
	repo.errOn["Update"] = dbErr
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusActive,
	}
	repo.users[user.ID] = user

	_, err := svc.UpdateProfile(context.Background(), user.ID, "", "New", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Deactivate Edge Cases ---

func TestService_Deactivate_UserDeleted(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusDeleted,
	}
	repo.users[user.ID] = user

	_, err := svc.Deactivate(context.Background(), user.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for deleted user, got: %v", err)
	}
}

func TestService_Deactivate_UserInvalidStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: "unknown_status",
	}
	repo.users[user.ID] = user

	_, err := svc.Deactivate(context.Background(), user.ID)
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got: %v", err)
	}
}

func TestService_Deactivate_GetByIDError(t *testing.T) {
	dbErr := errors.New("db error")
	repo := newMockRepo()
	repo.errKeys["GetByID"] = true
	repo.errOn["GetByID"] = dbErr
	svc := NewService(repo)

	_, err := svc.Deactivate(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Deactivate_UpdateStatusError(t *testing.T) {
	dbErr := errors.New("update status failed")
	repo := newMockRepo()
	repo.errKeys["UpdateStatus"] = true
	repo.errOn["UpdateStatus"] = dbErr
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusActive,
	}
	repo.users[user.ID] = user

	_, err := svc.Deactivate(context.Background(), user.ID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Activate Edge Cases ---

func TestService_Activate_UserAlreadyActive(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusActive,
	}
	repo.users[user.ID] = user

	_, err := svc.Activate(context.Background(), user.ID)
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got: %v", err)
	}
}

func TestService_Activate_UserDeleted(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusDeleted,
	}
	repo.users[user.ID] = user

	_, err := svc.Activate(context.Background(), user.ID)
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus for deleted user, got: %v", err)
	}
}

func TestService_Activate_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.Activate(context.Background(), uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestService_Activate_GetByIDError(t *testing.T) {
	dbErr := errors.New("db error")
	repo := newMockRepo()
	repo.errKeys["GetByID"] = true
	repo.errOn["GetByID"] = dbErr
	svc := NewService(repo)

	_, err := svc.Activate(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- CreateAndSetupUser Edge Cases ---

func TestService_CreateAndSetupUser_EmptyEmail(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		FirstName: "Test",
		LastName:  "User",
	}

	_, err := svc.CreateAndSetupUser(context.Background(), user)
	if err == nil {
		t.Fatal("expected error for empty email, got nil")
	}
}

func TestService_CreateAndSetupUser_CreateError(t *testing.T) {
	dbErr := errors.New("create failed")
	repo := newMockRepo()
	repo.errKeys["Create"] = true
	repo.errOn["Create"] = dbErr
	svc := NewService(repo)

	user := User{
		Email:     "test@example.com",
		FirstName: "Test",
	}

	_, err := svc.CreateAndSetupUser(context.Background(), user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_CreateAndSetupUser_DefaultStatus(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		Email:     "test@example.com",
		FirstName: "Test",
	}

	result, err := svc.CreateAndSetupUser(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusActive {
		t.Errorf("expected default status '%s', got '%s'", StatusActive, result.Status)
	}
}

func TestService_CreateAndSetupUser_WithMetadata(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		Email:     "test@example.com",
		FirstName: "Test",
		Metadata: tenant.JSONB{
			"account_settings": map[string]any{
				"theme": "dark",
				"lang":  "ru",
			},
		},
	}

	_, err := svc.CreateAndSetupUser(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_CreateAndSetupUser_AccountCreateNonCritical(t *testing.T) {
	acctErr := errors.New("account creation failed")
	repo := newMockRepo()
	repo.errKeys["CreateAccount"] = true
	repo.errOn["CreateAccount"] = acctErr
	svc := NewService(repo)

	user := User{
		Email:     "test@example.com",
		FirstName: "Test",
	}

	result, err := svc.CreateAndSetupUser(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error (account creation should be non-critical): %v", err)
	}
	if result.Email != "test@example.com" {
		t.Errorf("expected user to be created despite account error")
	}
}

// --- List Edge Cases ---

func TestService_List_PageNegative(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _, err := svc.List(context.Background(), UserFilter{Page: -1})
	if err != nil {
		t.Fatalf("unexpected error for negative page: %v", err)
	}
}

func TestService_List_PerPageNegative(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _, err := svc.List(context.Background(), UserFilter{PerPage: -1})
	if err != nil {
		t.Fatalf("unexpected error for negative per_page: %v", err)
	}
}

func TestService_List_PerPageOverflow(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, _, err := svc.List(context.Background(), UserFilter{PerPage: 1000})
	if err != nil {
		t.Fatalf("unexpected error for large per_page: %v", err)
	}
}

func TestService_List_EmptyResult(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	users, total, err := svc.List(context.Background(), UserFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if users == nil {
		t.Error("expected non-nil empty slice")
	}
}

func TestService_List_StatusFilterNoMatch(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusActive,
	}
	repo.users[user.ID] = user

	users, total, err := svc.List(context.Background(), UserFilter{Status: StatusDeactivated})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

// --- AddRole Edge Cases ---

func TestService_AddRole_AllValidRoles(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{ID: uuid.New()}
	repo.users[user.ID] = user

	validRoleNames := []string{RoleEmployee, RoleHR, RoleCatalogManager, RoleAdmin}
	for _, role := range validRoleNames {
		_, err := svc.AddRole(context.Background(), user.ID, role, nil)
		if err != nil {
			t.Errorf("role %s should be valid, got error: %v", role, err)
		}
	}
}

func TestService_AddRole_EmptyRole(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{ID: uuid.New()}
	repo.users[user.ID] = user

	_, err := svc.AddRole(context.Background(), user.ID, "", nil)
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole for empty role, got: %v", err)
	}
}

func TestService_AddRole_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.AddRole(context.Background(), uuid.New(), RoleHR, nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestService_AddRole_WithGrantedBy(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{ID: uuid.New()}
	repo.users[user.ID] = user

	grantedBy := uuid.New()
	result, err := svc.AddRole(context.Background(), user.ID, RoleHR, &grantedBy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GrantedBy == nil || *result.GrantedBy != grantedBy {
		t.Errorf("expected granted_by %s, got %v", grantedBy, result.GrantedBy)
	}
}

// --- RemoveRole Edge Cases ---

func TestService_RemoveRole_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	err := svc.RemoveRole(context.Background(), uuid.New(), RoleHR)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestService_RemoveRole_RoleNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{ID: uuid.New()}
	repo.users[user.ID] = user

	err := svc.RemoveRole(context.Background(), user.ID, RoleHR)
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got: %v", err)
	}
}

// --- GetByID Edge Cases ---

func TestService_GetByID_NilUUID(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.GetByID(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for nil UUID, got: %v", err)
	}
}

// --- GetRoles Edge Cases ---

func TestService_GetRoles_NoRoles(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{ID: uuid.New()}
	repo.users[user.ID] = user

	roles, err := svc.GetRoles(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if roles == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(roles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(roles))
	}
}

func TestService_GetRoles_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	roles, err := svc.GetRoles(context.Background(), uuid.New())
	// Mock returns empty slice for unknown user, not error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected 0 roles, got %d", len(roles))
	}
}

// --- Concurrent tests ---

func TestService_UpdateProfile_SequentialMultiple(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Original",
		Status:    StatusActive,
	}
	repo.users[user.ID] = user

	// Sequential updates (concurrent would require sync.Mutex in mock)
	names := []string{"Update1", "Update2", "Update3"}
	for _, name := range names {
		result, err := svc.UpdateProfile(context.Background(), user.ID, "", name, "", "")
		if err != nil {
			t.Errorf("sequential update %s error: %v", name, err)
		}
		if result.FirstName != name {
			t.Errorf("expected firstName '%s', got '%s'", name, result.FirstName)
		}
	}
}

func TestService_Deactivate_Concurrent(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	user := User{
		ID:     uuid.New(),
		Email:  "test@example.com",
		Status: StatusActive,
	}
	repo.users[user.ID] = user

	// Sequential deactivate calls (concurrent would require sync.Mutex in mock)
	_, err1 := svc.Deactivate(context.Background(), user.ID)
	if err1 != nil {
		t.Fatalf("first deactivate error: %v", err1)
	}

	_, err2 := svc.Deactivate(context.Background(), user.ID)
	if !errors.Is(err2, ErrAlreadyDeactivated) {
		t.Fatalf("second deactivate should return ErrAlreadyDeactivated, got: %v", err2)
	}
}
