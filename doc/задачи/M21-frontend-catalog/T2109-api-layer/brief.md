# T2109 — API Layer

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Слой API для взаимодействия с backend M20. Включает fetch wrapper с обработкой ошибок, авторизацией и таймаутом.
Зависит от T2101 (Vite bootstrap — proxy config) и T2104 (auth flow — токен из authStore).

### Backend endpoints (из `backend/internal/app/server.go`)

**Public API (JWT + tenant middleware):**
- `GET /api/v1/engagements/categories` — список категорий
- `GET /api/v1/engagements` — список энгейджментов (с фильтрами)
- `GET /api/v1/engagements/{id}` — детали энгейджмента
- `GET /api/v1/users/me` — профиль пользователя
- `GET /api/v1/auth/me` — профиль через auth endpoint

**Admin API (JWT + RBAC + admin tenant middleware):**
- `POST /admin/engagements/categories` — создать категорию
- `PUT /admin/engagements/categories/{id}` — обновить категорию
- `DELETE /admin/engagements/categories/{id}` — удалить категорию (204)
- `POST /admin/engagements/types` — создать тип
- `GET /admin/engagements/types` — admin список (все статусы)
- `GET /admin/engagements/types/{id}` — получить тип
- `PUT /admin/engagements/types/{id}` — обновить тип
- `DELETE /admin/engagements/types/{id}` — удалить тип (204)
- `PATCH /admin/engagements/types/{id}/status` — смена статуса
- `POST /admin/engagements/types/{typeId}/offers` — создать оффер
- `PUT /admin/engagements/types/{typeId}/offers/{id}` — обновить оффер
- `DELETE /admin/engagements/types/{typeId}/offers/{id}` — удалить оффер (204)

### CORS headers (из `server.go` — corsMiddleware)

```
Access-Control-Allow-Origin: * (dev)
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Accept, Authorization, Content-Type, X-Tenant-ID
Access-Control-Expose-Headers: Link
Access-Control-Max-Age: 300
```

### Timeout (из `server.go`)

```go
r.Use(middleware.Timeout(30 * time.Second))
```

Серверный таймаут: 30 секунд. Клиентский таймаут должен быть меньше (25s).

### Vite proxy (из `vite.config.ts` — T2101)

```ts
proxy: {
  '/api': { target: 'http://localhost:8080', changeOrigin: true },
  '/admin': { target: 'http://localhost:8080', changeOrigin: true },
}
```

## Что сделать

### `src/api/client.ts` — fetch wrapper

```ts
import { useAuthStore } from '@/stores/authStore'

const DEFAULT_TIMEOUT = 25_000 // 25s (серверный timeout 30s)
const MAX_RETRIES = 2
const RETRY_BASE_DELAY = 1000 // 1s

export interface ApiError {
  message: string
  status: number
}

export async function apiRequest<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  const token = useAuthStore.getState().token

  const headers: Record<string, string> = {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  }

  // Инъекция Authorization header
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  // AbortController для timeout
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT)

  let lastError: ApiError | null = null

  for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
    try {
      const response = await fetch(url, {
        ...options,
        headers,
        signal: controller.signal,
      })

      clearTimeout(timeoutId)

      // 401 → redirect на login
      if (response.status === 401) {
        useAuthStore.getState().clearAuth()
        window.location.href = '/login'
        throw new Error('Unauthorized')
      }

      // 403 → forbidden
      if (response.status === 403) {
        throw new Error('Forbidden')
      }

      // 204 NoContent (delete operations)
      if (response.status === 204) {
        return null as T
      }

      // 5xx → retry с exponential backoff
      if (response.status >= 500 && attempt < MAX_RETRIES) {
        const delay = RETRY_BASE_DELAY * Math.pow(2, attempt)
        await new Promise(resolve => setTimeout(resolve, delay))
        continue
      }

      if (!response.ok) {
        const errorBody = await response.text()
        throw new Error(errorBody || `HTTP ${response.status}`)
      }

      const data = await response.json()
      return data as T
    } catch (error) {
      lastError = error as ApiError
      if (attempt < MAX_RETRIES && error instanceof Error && error.name === 'AbortError') {
        continue
      }
      throw error
    }
  }

  throw lastError || new Error('Request failed')
}
```

### `src/api/engagements.ts` — typed API functions

```ts
import { apiRequest } from './client'
import type {
  EngagementTypeResponse,
  EngagementCategoryResponse,
  ListResponse,
} from './types'

// Public API

export async function getCategories(): Promise<EngagementCategoryResponse[]> {
  return apiRequest<EngagementCategoryResponse[]>('/api/v1/engagements/categories')
}

export interface GetEngagementsParams {
  type?: string
  status?: string
  category?: string
  search?: string
  page?: number
  per_page?: number
}

export async function getEngagements(
  params: GetEngagementsParams = {}
): Promise<ListResponse> {
  const searchParams = new URLSearchParams()
  if (params.type) searchParams.set('type', params.type)
  if (params.status) searchParams.set('status', params.status)
  if (params.category) searchParams.set('category', params.category)
  if (params.search) searchParams.set('search', params.search)
  if (params.page) searchParams.set('page', String(params.page))
  if (params.per_page) searchParams.set('per_page', String(params.per_page))

  const query = searchParams.toString()
  return apiRequest<ListResponse>(
    `/api/v1/engagements${query ? '?' + query : ''}`
  )
}

export async function getEngagement(id: string): Promise<EngagementTypeResponse> {
  return apiRequest<EngagementTypeResponse>(`/api/v1/engagements/${id}`)
}

// Admin API

export interface CreateCategoryRequest {
  slug: string
  name: string
  icon: string
  sort_order: number
}

export async function createCategory(req: CreateCategoryRequest): Promise<EngagementCategoryResponse> {
  return apiRequest<EngagementCategoryResponse>('/admin/engagements/categories', {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export async function updateCategory(id: string, req: CreateCategoryRequest): Promise<EngagementCategoryResponse> {
  return apiRequest<EngagementCategoryResponse>(`/admin/engagements/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(req),
  })
}

export async function deleteCategory(id: string): Promise<null> {
  return apiRequest<null>(`/admin/engagements/categories/${id}`, {
    method: 'DELETE',
  })
}

export async function updateEngagementStatus(
  id: string,
  status: string
): Promise<EngagementTypeResponse> {
  return apiRequest<EngagementTypeResponse>(`/admin/engagements/types/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  })
}
```

### `src/api/user.ts` — API пользователя

```ts
import { apiRequest } from './client'

export interface UserProfile {
  id: string
  email: string
  first_name: string
  last_name: string
  tenant_id: string
  keycloak_sub: string
  status: string
  created_at: string
}

export async function getUserProfile(): Promise<UserProfile> {
  return apiRequest<UserProfile>('/api/v1/users/me')
}
```

### `src/api/types.ts` — TypeScript типы (соответствуют Go struct)

```ts
// EngagementCategoryResponse — соответствует Go EngagementCategoryResponse
export interface EngagementCategoryResponse {
  id: string
  slug: string
  name: string
  icon?: string
  sort_order: number
}

// EngagementOfferResponse — соответствует Go EngagementOfferResponse
export interface EngagementOfferResponse {
  id: string
  name: string
  description?: string
  cost_cents: number
  sort_order: number
}

// EngagementTypeResponse — соответствует Go EngagementTypeResponse
export interface EngagementTypeResponse {
  id: string
  slug: string
  name: string
  description?: string
  type: 'benefit' | 'activity'
  status: 'draft' | 'active' | 'promo' | 'hidden' | 'completed'
  cost_cents?: number
  provider_name?: string
  image_url?: string
  category?: EngagementCategoryResponse
  offers?: EngagementOfferResponse[]
  badge: string // "Промо" | "Доступна"
}

// PaginationResponse — соответствует Go PaginationResponse
export interface PaginationResponse {
  page: number
  per_page: number
  total: number
  total_pages: number
}

// ListResponse — соответствует Go ListResponse
export interface ListResponse {
  data: EngagementTypeResponse[]
  pagination: PaginationResponse
}
```

## Требования

- Timeout 25s (меньше серверного 30s из `middleware.Timeout(30 * time.Second)`)
- Authorization header: `Bearer {token}` из authStore
- 401 → clearAuth + redirect /login
- 403 → forbidden page
- 204 → return null (для DELETE operations)
- 5xx → retry с exponential backoff (max 2 retry)
- Типы точно соответствуют Go response types из handler.go
- Vite proxy для dev (/api → :8080, /admin → :8080)

## Критерии приёмки

- [ ] `src/api/client.ts` — fetch wrapper с auth, error handling, retry
- [ ] Authorization header инъекция (Bearer token из authStore)
- [ ] 401 → clearAuth + redirect /login
- [ ] 403 → forbidden page
- [ ] 204 → null (delete operations)
- [ ] 5xx → retry с exponential backoff (max 2)
- [ ] Timeout 25s
- [ ] `src/api/engagements.ts` — typed functions для всех engagement endpoints
- [ ] `src/api/user.ts` — getUserProfile
- [ ] `src/api/types.ts` — типы соответствуют Go struct из handler.go
- [ ] ListResponse, PaginationResponse, EngagementTypeResponse точно маппятся
