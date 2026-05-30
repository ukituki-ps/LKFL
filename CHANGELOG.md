# Changelog

All notable changes to the LKFL platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [F1] Working Catalog — 2026-05-26

### Added

- **Multi-tenant platform** — tenant CRUD, brand config, resolver middleware, tenant isolation
- **Auth + RBAC** — Keycloak OIDC, JWT verification, role-based access control (admin, hr, catalog_manager, employee), middleware
- **User management** — CRUD пользователей, профиль сотрудника, admin-список, tenant isolation
- **Engagement catalog** — engagement types, categories, offers, public API + admin API
- **Redis cache** — catalog list caching, invalidation on admin write
- **Frontend SPA** — Vite + React 18, April UI (`@ukituki-ps/april-ui`), Mantine, Zustand state management
- **Frontend pages** — `/catalog`, `/dashboard`, `/login`, `/admin/*` (categories, types, offers, users)
- **OpenAPI codegen** — openapi-typescript для генерации TypeScript типов
- **Unit tests** — Go (backend packages) + frontend (Vitest)
- **Integration tests** — testcontainers (PostgreSQL 17 + Redis 7)
- **E2E tests** — Playwright, 3 browsers (Chromium, Firefox, Webkit)
- **Chaos tests** — 100 тестов отказоустойчивости
- **Load testing** — k6 (loadtest/)
- **Monitoring** — Prometheus metrics + Grafana dashboards
- **Logging** — Loki + Promtail
- **CI pipeline** — GitHub Actions, 9 jobs (lint, test, build, security, deploy)
- **Docker production images** — multi-stage build, distroless base, cosign signing
- **Staging deployment** — docker-compose.staging.yml
- **Security** — rate limiting, CORS policy, OWASP audit

### Security

- OWASP Top 10 audit passed
- Rate limiting on auth/catalog/admin endpoints
- CORS policy configured
- Dependency audit clean (govulncheck, npm audit, trivy)

### Changed

- N/A (первый релиз)

### Deprecated

- N/A

### Removed

- N/A

### Fixed

- N/A
