# T2212 — Пакет F1

## Веха

M22-f1-hardening

## Тип

code

## Контекст

Финализация F1: git tag, changelog, API docs, release notes.
Подготовка к релизу и передаче на QA.

## Что сделать

### Git tag

- Tag: `f1-complete`
- Message: `F1 Complete: Working Catalog — Multi-tenant catalog with auth, RBAC, frontend`
- Annotated tag с GPG signature

### Changelog

`CHANGELOG.md`:

```markdown
## [F1] Working Catalog — YYYY-MM-DD

### Added
- Multi-tenant platform (tenant CRUD, brand config, resolver middleware, isolation)
- Auth + RBAC (Keycloak OIDC, JWT, roles, middleware)
- User management (CRUD, profile, admin list, tenant isolation)
- Catalog (engagement types, categories, offers, public + admin API)
- Redis cache (catalog list, invalidation on admin write)
- Frontend SPA (Vite + React 18, April UI, Mantine, Zustand)
- Frontend pages: /catalog, /dashboard, /login, /admin/*
- OpenAPI codegen (openapi-typescript)
- Unit tests (Go + frontend)
- Integration tests (testcontainers)
- E2E tests (Playwright, 3 browsers)
- Chaos tests (100 tests)
- Load testing (k6)
- Monitoring (Prometheus + Grafana)
- Logging (Loki + Promtail)
- CI pipeline (GitHub Actions)
- Docker production images (multi-stage, distroless, cosign)
- Staging deployment

### Security
- OWASP Top 10 audit passed
- Rate limiting on auth/catalog endpoints
- CORS policy configured
- Dependency audit clean (govulncheck, npm audit, trivy)
```

### API documentation

- **Redocly** — `redocly build-docs` → `docs/api/index.html`
- OpenAPI spec: `docs/api/openapi.yaml`
- Endpoints documented: все F1 endpoints (catalog, auth, user, tenant)

### Release notes

`releases/F1.md`:

- Summary
- Features list
- Known limitations
- Upgrade instructions (если применимо)
- Screenshots

### Release artifacts

- Docker images: `lkfl-server:f1`, `lkfl-integration-proxy:f1`
- Frontend build: `frontend/dist/`
- OpenAPI spec: `openapi.yaml`
- Changelog: `CHANGELOG.md`
- Release notes: `releases/F1.md`

## Критерии приёмки

- [ ] git tag `f1-complete` (annotated, GPG signed)
- [ ] CHANGELOG.md создан
- [ ] API documentation (Redocly, openapi.yaml)
- [ ] Release notes (releases/F1.md)
- [ ] Docker images tagged (`:f1`)
- [ ] Release artifacts собраны
