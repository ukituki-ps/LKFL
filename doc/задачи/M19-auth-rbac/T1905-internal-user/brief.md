# T1905 — internal/user/ (CRUD + Profile)

## Веха

M19-auth-rbac

## Тип

code

## Контекст

`backend/internal/user/` — CRUD пользователей, профиль, admin operations.
Исходник: `doc/архитектура/пакеты-platform.md` (строка ~197 — user/).

**Зависимости от M18:**
- `tenant.JSONB` — переиспользовать тип из `internal/tenant/model.go` (или импортировать)
- `tenant.TenantIDFromContext()` — для tenant isolation в query
- `tenant.WithTenantID()` — для автоматического `WHERE tenant_id` в repository
- `shared/pkg/auth.UserIDFromContext()` — для определения текущего пользователя
- `shared/pkg/auth.RolesFromContext()` — для проверки прав (own profile vs admin)

## Что сделать

### Структура

```
internal/user/
├── model.go       # User struct
├── repository.go  # DB operations
├── service.go     # Business logic
└── handler.go     # HTTP handlers
```

### `model.go`

```go
package user

import (
    "time"

    "github.com/google/uuid"
    "lkfl/internal/tenant"  // JSONB тип из M18
)

type User struct {
    ID           uuid.UUID          `json:"id"`
    TenantID     uuid.UUID          `json:"tenant_id"`
    Email        string             `json:"email"`
    FirstName    string             `json:"first_name"`
    LastName     string             `json:"last_name"`
    Phone        string             `json:"phone,omitempty"`
    Status       string             `json:"status"`
    KeycloakSub  string             `json:"-"` // never expose в API
    Metadata     tenant.JSONB       `json:"metadata,omitempty"`
    CreatedAt    time.Time          `json:"created_at"`
    UpdatedAt    time.Time          `json:"updated_at"`
}
```

> **Важно:** `tenant.JSONB` переиспользуется из M18 (`internal/tenant/model.go`).
> Не дублировать тип — один JSONB на весь проект.

### `handler.go`

```go
// Me — профиль текущего пользователя (employee)
// GET /api/v1/users/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request)

// UpdateMe — обновление профиля (employee)
// PUT /api/v1/users/me
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request)

// AdminList — список пользователей (admin/hr)
// GET /admin/users?page=1&per_page=20&status=active&search=email
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request)

// AdminGet — детали пользователя (admin/hr)
// GET /admin/users/:id
func (h *Handler) AdminGet(w http.ResponseWriter, r *http.Request)

// AdminUpdate — обновление пользователя (admin/hr)
// PUT /admin/users/:id
func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request)

// AdminDeactivate — деактивация (admin/hr)
// POST /admin/users/:id/deactivate
func (h *Handler) AdminDeactivate(w http.ResponseWriter, r *http.Request)
```

## Требования

- Repository interface + pgx implementation
- Service layer — business logic (status transitions, validation)
- Handler — HTTP handlers (employee + admin)
- Profile update — только own profile (employee)
- Admin operations — tenant-scoped (admin видит только своих пользователей)
- Search — по email, name (ILIKE query)
- Pagination — page/per_page
- Status transitions: active → deactivated (не reverse)

## Критерии приёмки

- [ ] `model.go` — User struct
- [ ] `repository.go` — Repository interface + pgx impl
- [ ] `service.go` — business logic
- [ ] `handler.go` — Me, UpdateMe, AdminList, AdminGet, AdminUpdate, AdminDeactivate
- [ ] Profile update (own only)
- [ ] Admin tenant-scoped
- [ ] Search + pagination
- [ ] Status transitions
- [ ] Unit tests
