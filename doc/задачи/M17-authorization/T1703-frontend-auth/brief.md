# T1703 — Frontend: auth

## Контекст

Реализация фронтенда системы авторизации: Vite bootstrap, интеграция с Keycloak через keycloak-js, Zustand store для состояния auth, state machine для управления переходами между состояниями, layout компоненты.

**Родительский эпик:** T1700 (Полная система авторизации)
**Зависит от:** T1701 (инфраструктура)
**ADR:** ADR-036 (Authorization System), ADR-003 (Keycloak), ADR-009 (Multi-tenancy)

## Что включено

### Инициализация проекта
- Vite bootstrap + package.json (React 18, keycloak-js, @tanstack/react-query, zustand, @ukituki-ps/april-ui, Mantine, vitest)

### Tenant resolution (frontend)
- `src/utils/tenant.ts` — `extractTenantSlug(hostname)` → realm name, `buildKeycloakConfig(slug)` → {url, realm, clientId}

### Keycloak интеграция
- `src/keycloak.ts` — Keycloak instance, init (realm из tenant resolution, не из env)
- `src/keycloak-token-subscribers.ts` — token rotation notification

### API и store
- `src/api/platform.ts` — authorizedFetch (401 retry + token refresh), React Query integration
- `src/store/auth.ts` — Zustand auth store (user, roles, tenant, status, fetchMe, logout)

### Auth state machine
- `src/App.tsx` — auth state machine (guest → transition → authorized/forbidden/no-context)
- `src/main.tsx` — Sentry init, auth init, `<AprilProviders>`, render `<App>`
- `src/components/auth/RequireAuth.tsx` — auth guard с role checking
- `src/components/auth/GuestZone.tsx` — landing page + login button
- `src/components/auth/ForbiddenZone.tsx` — 403 screen
- `src/components/auth/AuthorizedShell.tsx` — full nav + role filtering

### Layout и страницы
- `src/components/layout/Header.tsx`, `Sidebar.tsx`, `PageLayout.tsx` — layout компоненты
- `src/pages/Dashboard.tsx` — stub: баланс, активные льготы, события
- `src/pages/Catalog.tsx` — stub: каталог льгот

### Unit-тесты (vitest)
- `src/utils/tenant.test.ts` — extractTenantSlug: sdek.lkfl.ru→sdek, localhost→demo, buildKeycloakConfig
- `src/store/auth.test.ts` — Zustand store: fetchMe, logout, state transitions
- `src/api/platform.test.ts` — authorizedFetch: 401 retry, token refresh, error handling
- `src/App.test.tsx` — auth state machine: guest → transition → authorized/forbidden/no-context

### Playwright E2E
- `playwright.config.ts` — конфигурация, dev server, auth fixture
- `tests/e2e/login.spec.ts` — login flow E2E (guest → Keycloak login → Dashboard)
- `tests/e2e/forbidden.spec.ts` — 403 screen для неавторизованного доступа

## Результат

- Frontend запускается: `npm run dev` (Vite)
- keycloak-js инициализируется, подключается к Keycloak
- Login flow работает: guest → login → transition → authorized
- authorizedFetch retry при 401 + token refresh
- RequireAuth guard блокирует без ролей
- Layout рендерится с nav (role-filtered)
- Unit-тесты (vitest): store/auth.ts, api/platform.ts, App.tsx — все зелёные
- Playwright E2E: login.spec.ts и forbidden.spec.ts проходят локально и в CI
