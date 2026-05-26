# T2202 — Unit тесты: Edge cases

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Unit тесты для edge cases всех пакетов F1.
Фокус на boundary conditions и error paths.

## Что сделать

### tenant/

- Create: empty slug, invalid slug format, duplicate slug, max length slug
- GetBySlug: suspended tenant, non-existent slug, SQL injection attempt
- List: page 0, negative per_page, max per_page (100+), empty result

### auth/

- JWT middleware: expired token, invalid signature, missing Bearer, malformed JWT, empty token
- RBAC: no roles, wrong role, multiple roles (one matches), role escalation attempt
- OIDC verifier: invalid issuer, network error, clock skew

### user/

- Profile update: empty name, max length name, invalid email, own profile only
- AdminList: search special chars, SQL injection, XSS in search
- Deactivate: already deactivated, active user engagements

### catalog/

- ListTypes: empty filters, all filters combined, non-existent category, promo sort order
- GetTypeByID: invalid UUID, non-existent ID, cross-tenant access attempt
- Admin: delete with active engagements, status invalid transition

### frontend/

- authStore: token expiration, refresh failure, logout cleanup
- API client: 401 redirect, 403 handling, 5xx retry, timeout
- EngagementCard: null image, null cost, long description truncation

## Требования

- Table-driven tests (Go)
- Mock repository (interface)
- Test coverage > 60% для всех F1 пакетов
- Frontend: Vitest, React Testing Library

## Критерии приёмки

- [ ] tenant/ edge cases покрыты
- [ ] auth/ edge cases покрыты
- [ ] user/ edge cases покрыты
- [ ] catalog/ edge cases покрыты
- [ ] frontend/ edge cases покрыты
- [ ] Coverage > 60%
- [ ] `go test ./... -race` — 0 failures
