# T2102-T2111 — Frontend F1 (оставшиеся задачи)

## Веха

M21-frontend-catalog

## Тип

code

## Краткое описание оставшихся задач

### T2102 — AprilProviders
- `src/main.tsx` — `AprilProviders` корневой провайдер, `createAprilTheme()`
- DensityProvider, ColorScheme
- Brand CSS variables override из API

### T2103 — Routing
- `src/routes/employee.tsx` — employee routes с guards
- `src/routes/admin.tsx` — admin routes с RBAC guards
- `RequireAuth` component (role check)
- Lazy loading (`React.lazy` + `Suspense`)

### T2104 — Auth Flow
- Keycloak redirect login
- Token storage (memory, не localStorage — security)
- `useAuthStore` (Zustand): user, roles, token, login, logout
- Token refresh logic
- Logout → Keycloak end_session

### T2105 — Layout
- `AprilShellBar` (sidebar + header)
- Navigation menu (employee: Dashboard, Catalog, Points, Documents, Support)
- Navigation menu (admin: HR, Catalog, Content, Billing)
- Responsive breakpoints (desktop first)

### T2106 — /catalog страница
- Список карточек.engagements
- Фильтры: category, type, status
- Поиск (debounced)
- Pagination
- Loading states, empty states, error states
- React Query для data fetching

### T2107 — Карточка льготы
- `EngagementCard` компонент
- Название, описание, стоимость
- Бейдж статуса (Активна/Доступна/Новинка/Промо)
- Кнопка «Подробнее»
- Provider name, image

### T2108 — / Dashboard stub
- Greeting (Hello, {name})
- Placeholder stat cards (баланс, активные льготы)
- Placeholder event feed
- Placeholder quick actions

### T2109 — API Layer
- `src/api/client.ts` — fetch wrapper
- Error handling (401 → redirect login, 403 → forbidden page)
- Retry logic (exponential backoff для 5xx)
- Tenant header injection
- Timeout (30s)

### T2110 — OpenAPI Codegen
- `openapi-typescript` → типизация API responses
- Script в package.json: `npm run generate-types`
- Types в `src/api/types.ts`

### T2111 — Unit Tests
- Vitest setup
- `authStore.test.ts` — login, logout, token refresh
- `api/client.test.ts` — error handling, retry
- `EngagementCard.test.tsx` — rendering, badge display
- `RequireAuth.test.tsx` — role check

## Критерии приёмки

- [ ] Все 10 задач реализованы
- [ ] `npm run dev` → SPA работает
- [ ] Login через Keycloak
- [ ] Каталог загружается
- [ ] Фильтры работают
- [ ] Unit tests зелёные
