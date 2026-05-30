# T2202 — Unit тесты: Edge cases

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Unit тесты для edge cases всех пакетов F1.
Фокус на boundary conditions и error paths.

## Что сделать

> **🔴 Критическое требование:** 100% покрытие функционала F1, 100% покрытие всех юзеркейсов, все крайние кейсы. Тестов должно быть супермного. Каждая функция, каждый branch, каждый error path — покрыт тестом.

### tenant/

- Create: empty slug, invalid slug format, duplicate slug, max length slug, unicode slug, slug with spaces, slug with special chars
- GetBySlug: suspended tenant, non-existent slug, SQL injection attempt, empty slug, nil context
- List: page 0, negative per_page, max per_page (100+), empty result, pagination boundary (last page, overflow page)
- Update: update suspended tenant, update to existing slug, partial update (only one field), all fields update
- Delete: delete with users, delete non-existent, delete already deleted
- Brand config: empty JSON, invalid JSON, max size JSON, concurrent update

### auth/

- JWT middleware: expired token, invalid signature, missing Bearer, malformed JWT, empty token, wrong algorithm, token with extra claims
- RBAC: no roles, wrong role, multiple roles (one matches), role escalation attempt, empty roles list, nil user
- OIDC verifier: invalid issuer, network error, clock skew, revoked token, token without expected scopes
- Login: missing state, invalid state, state expired, concurrent login attempts
- Callback: missing code, invalid code, code already used, code expired, provider error

### user/

- Profile update: empty name, max length name, invalid email, own profile only, concurrent update, profile not found
- AdminList: search special chars, SQL injection, XSS in search, empty search, unicode search, pagination edge cases
- Deactivate: already deactivated, active user engagements, concurrent deactivate, user not found
- Create: duplicate email, invalid email format, empty required fields, max length fields

### catalog/

- ListTypes: empty filters, all filters combined, non-existent category, promo sort order, empty result, pagination boundary
- GetTypeByID: invalid UUID, non-existent ID, cross-tenant access attempt, hidden type, draft type
- Admin: delete with active engagements, status invalid transition, concurrent status change, delete non-existent
- Categories: duplicate slug, empty name, max length name, sort_order conflicts
- Offers: negative cost, zero cost, max cost, duplicate name, sort_order validation

### frontend/

- authStore: token expiration, refresh failure, logout cleanup, concurrent state updates, store reset, role change
- API client: 401 redirect, 403 handling, 5xx retry (all 3 attempts), timeout, network error, abort signal, race condition
- EngagementCard: null image, null cost, long description truncation, empty name, zero cost, negative cost, no offers, no category, no provider, all fields missing
- RequireAuth: auth check race, role change during render, nested auth guards, concurrent navigation
- Catalog page: empty API response, API error, filter reset, search debounce, pagination overflow, category change
- Shell: role change, mobile/desktop toggle, concurrent route change, no user data

## Требования

- Table-driven tests (Go) — каждый edge case отдельный test case
- Mock repository (interface) — изоляция от БД
- **Test coverage 100% для всех F1 пакетов** (Go + frontend)
- **Каждый юзеркейс покрыт минимум 3 тестами:** happy path, error path, edge case
- **Все крайние кейсы:** nil/empty/zero/max/overflow/concurrent/race
- Frontend: Vitest, React Testing Library, fake timers, memory router
- `go test ./... -race` — 0 failures (race detector)
- **Тестов должно быть супермного** — минимум 50 тестов на каждый пакет

## Критерии приёмки

- [ ] tenant/ edge cases покрыты (минимум 30 тестов)
- [ ] auth/ edge cases покрыты (минимум 40 тестов)
- [ ] user/ edge cases покрыты (минимум 30 тестов)
- [ ] catalog/ edge cases покрыты (минимум 40 тестов)
- [ ] frontend/ edge cases покрыты (минимум 30 тестов)
- [ ] **Coverage 100% для всех F1 пакетов**
- [ ] **Каждый юзеркейс покрыт 3+ тестами (happy + error + edge)**
- [ ] **Все крайние кейсы покрыты (nil/empty/zero/max/overflow/concurrent/race)**
- [ ] `go test ./... -race` — 0 failures
