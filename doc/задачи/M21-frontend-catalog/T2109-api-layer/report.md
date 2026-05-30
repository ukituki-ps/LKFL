# Отчёт T2109 — API Layer

## Статус
выполнено

## Дата
2026-05-26

## Что сделано

### 1. `src/api/types.ts` — TypeScript типы
- `EngagementCategoryResponse` — соответствует Go struct
- `EngagementOfferResponse` — соответствует Go struct
- `EngagementTypeResponse` — соответствует Go struct (type, status, badge)
- `PaginationResponse` — page, per_page, total, total_pages
- `ListResponse` — data + pagination

### 2. `src/api/client.ts` — fetch wrapper
- `apiRequest<T>(url, options)` — универсальный typed fetch
- Authorization header инъекция (Bearer token из authStore)
- 401 → clearAuth() + redirect /login
- 403 → throw Error('Forbidden') со статусом
- 204 → return null (для DELETE operations)
- 5xx → retry с exponential backoff (max 2 retry, base 1s)
- Timeout 25s (AbortController)
- ApiError interface (extends Error, status?: number)

### 3. `src/api/engagements.ts` — typed API functions
**Public API:**
- `getCategories()` → EngagementCategoryResponse[]
- `getEngagements(params)` → ListResponse (с query params: type, status, category, search, page, per_page)
- `getEngagement(id)` → EngagementTypeResponse

**Admin API:**
- `createCategory(req)` → EngagementCategoryResponse
- `updateCategory(id, req)` → EngagementCategoryResponse
- `deleteCategory(id)` → null
- `updateEngagementStatus(id, status)` → EngagementTypeResponse

### 4. `src/api/user.ts` — API пользователя
- `UserProfile` interface (id, email, first_name, last_name, tenant_id, keycloak_sub, status, created_at)
- `getUserProfile()` → UserProfile

### 5. `src/api/index.ts` — barrel export
- Экспорт всех публичных типов и функций

## Проверка
- `npm run build` — ✅ успешно (tsc + vite build)
- `npm run lint` — ✅ без ошибок

## Изменённые файлы
- `frontend/src/api/types.ts` — переписано (placeholder → полная реализация)
- `frontend/src/api/client.ts` — переписано (placeholder → полная реализация)
- `frontend/src/api/engagements.ts` — переписано (placeholder → полная реализация)
- `frontend/src/api/user.ts` — переписано (placeholder → полная реализация; UserProfile ре-экспортирован из types.generated.ts)
- `frontend/src/api/index.ts` — создано (barrel export)

## Пост-аудит (M21 audit fix)
- `api/user.ts` — UserProfile заменён на ре-экспорт из `types.ts` (types.generated.ts) для устранения дублирования
- `plan.yaml` — исправлено `max 3 retries` → `max 2 retries` (соответствие коду)

## Замечания
- Типы соответствуют Go struct из backend handler.go
- uuid.UUID → string, *int64 → number | undefined, int64 → number
- Vite proxy (/api → :8080, /admin → :8080) из T2101 используется прозрачно
