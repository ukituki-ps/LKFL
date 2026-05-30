# T1703 — Отчёт: Frontend Auth

## Статус

✅ Завершено

## Что сделано

### Vite project
- `frontend/package.json` — Vite, React 18, April UI, Mantine, Zustand
- `frontend/vite.config.ts` — dev server, proxy к backend

### Страницы
- `src/pages/Login.tsx` — redirect на Keycloak login
- `src/pages/Callback.tsx` — OIDC callback handler
  - D1: realm из `VITE_KEYCLOAK_REALM` (не hard-coded)
  - D2: `credentials: 'include'` для cookie
  - D11: fallback `'token-from-backend'` → error handling
- `src/pages/Dashboard.tsx` — dashboard stub с auth guard
- `src/pages/Profile.tsx` — profile page stub
- `src/pages/LoginPage.tsx` — login page с form

### Store
- `src/stores/authStore.ts` — Zustand store
  - D2: token НЕ в localStorage (httpOnly cookie)
  - user + roles в localStorage (без токена)
  - `credentials: 'include'` для logout

### API client
- `src/api/client.ts` — apiRequest с timeout, retry, error handling
  - D2: `credentials: 'include'` (cookie-based auth)

### Routing
- `src/App.tsx` — BrowserRouter с protected routes
- `src/components/ProtectedRoute.tsx` — auth guard

## Проблемы

- D1: hard-coded realm `lkfl-sdek` → `VITE_KEYCLOAK_REALM` (исправлено T1709)
- D2: token в localStorage → httpOnly cookie (исправлено T1709)
- D11: fallback `'token-from-backend'` → error (исправлено T1709)

## Следующие шаги

Н/Д — задача завершена. E2E тесты отложены на M19.
