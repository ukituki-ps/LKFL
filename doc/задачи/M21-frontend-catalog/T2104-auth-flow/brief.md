# T2104 — Auth Flow

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Интеграция с Keycloak OIDC через backend auth endpoints (M20).
Зависит от T2101 (Vite bootstrap).

### Backend auth endpoints (из `backend/internal/auth/handler.go`)

| Метод | Endpoint | Описание | Auth |
|-------|----------|----------|------|
| GET | `/api/v1/auth/login` | Редирект на Keycloak authorize (302) | Нет |
| GET | `/api/v1/auth/callback` | Keycloak callback, state validation, token verification | Нет |
| POST | `/api/v1/auth/logout` | Инвалидация сессии в Redis, редирект на Keycloak logout | JWT |
| GET | `/api/v1/auth/me` | Профиль текущего пользователя | JWT |
| GET | `/api/v1/users/me` | Профиль пользователя (tenant-isolated) | JWT + tenant |

### Flow логина (из handler.go)

1. `GET /api/v1/auth/login` → генерирует state (32 байта hex), сохраняет в Redis (TTL 10 мин)
2. Редирект 302 на Keycloak authorize endpoint с `response_type=code&scope=openid+profile+email`
3. `GET /api/v1/auth/callback?code=...&state=...` → валидация state, верификация id_token
4. Извлечение claims и ролей через `sharedauth.ExtractClaims(idToken)`
5. Создание/обновление пользователя в БД
6. Сессия в Redis: `auth:session:{userID}` → id_token (TTL 24 часа)
7. Response: `{ "user": {...}, "roles": [...] }`

### Flow логаута (из handler.go)

1. `POST /api/v1/auth/logout` → удаление сессии из Redis (`auth:session:{userID}`)
2. Редирект 302 на Keycloak logout endpoint

### Response от `/api/v1/auth/callback`

```json
{
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "Иван",
    "last_name": "Петров",
    "keycloak_sub": "keycloak-uuid"
  },
  "roles": ["employee", "catalog_manager"]
}
```

### Response от `/api/v1/auth/me`

```json
{
  "id": "uuid",
  "email": "user@example.com",
  "first_name": "Иван",
  "last_name": "Петров"
}
```

## Что сделать

### `src/stores/authStore.ts` — Zustand store

```ts
import { create } from 'zustand'

interface UserProfile {
  id: string
  email: string
  first_name: string
  last_name: string
}

interface AuthState {
  // Состояние
  token: string | null
  user: UserProfile | null
  userRoles: string[]
  isAuthenticated: boolean
  isLoading: boolean

  // Actions
  setAuth: (token: string, user: UserProfile, roles: string[]) => void
  setUser: (user: UserProfile) => void
  logout: () => Promise<void>
  clearAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  userRoles: [],
  isAuthenticated: false,
  isLoading: false,

  setAuth: (token, user, roles) =>
    set({ token, user, userRoles: roles, isAuthenticated: true, isLoading: false }),

  setUser: (user) => set({ user }),

  logout: async () => {
    // POST /api/v1/auth/logout
    await fetch('/api/v1/auth/logout', {
      method: 'POST',
      headers: { Authorization: `Bearer ${useAuthStore.getState().token}` },
    })
    set({ token: null, user: null, userRoles: [], isAuthenticated: false })
  },

  clearAuth: () => set({ token: null, user: null, userRoles: [], isAuthenticated: false }),
}))
```

### Token storage — память (НЕ localStorage)

Токен хранится в состоянии Zustand (in-memory). При перезагрузке страницы пользователь должен авторизоваться заново.
Это требование безопасности (152-ФЗ, ФСТЭК).

### `src/pages/Login.tsx` — страница логина

```tsx
import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'

export function Login() {
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    // Получаем attempted URL из state
    const from = (location.state as any)?.from?.pathname || '/'

    // Редирект на backend login endpoint
    // Backend сам сделает 302 на Keycloak
    window.location.href = `/api/v1/auth/login?redirect=http://localhost:5173/api/v1/auth/callback?post_redirect=${encodeURIComponent(from)}`
  }, [navigate, location])

  return <div>Перенаправление на страницу входа...</div>
}
```

### `src/pages/Callback.tsx` — страница обработки callback

```tsx
import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

export function Callback() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const setAuth = useAuthStore((state) => state.setAuth)

  useEffect(() => {
    // Backend callback уже обработал token и вернул { user, roles }
    // Но для SPA flow нам нужно получить токен из URL hash или cookies
    // TODO: после реализации token flow — извлечь токен и вызвать setAuth

    // Пока: редирект на Dashboard
    navigate('/', { replace: true })
  }, [navigate])

  return <div>Вход выполнен...</div>
}
```

### `/api/v1/auth/me` — проверка сессии при загрузке

```ts
async function checkAuth(): Promise<void> {
  try {
    const res = await fetch('/api/v1/auth/me', {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (res.ok) {
      const user = await res.json()
      useAuthStore.getState().setUser(user)
    }
  } catch {
    useAuthStore.getState().clearAuth()
  }
}
```

## Требования

- Token storage: in-memory (Zustand state), НЕ localStorage, НЕ sessionStorage
- Logout: POST `/api/v1/auth/logout` + очистка store
- Auth check при загрузке: GET `/api/v1/auth/me`
- Redirect на Keycloak через backend (не прямой)
- State parameter для CSRF protection (генерируется backend)

## Критерии приёмки

- [ ] `src/stores/authStore.ts` — Zustand store с auth state
- [ ] Token хранится в памяти (не localStorage)
- [ ] `setAuth(token, user, roles)` — установка auth state
- [ ] `logout()` — POST `/api/v1/auth/logout` + clear
- [ ] `clearAuth()` — сброс состояния
- [ ] `src/pages/Login.tsx` — редирект на `/api/v1/auth/login`
- [ ] `src/pages/Callback.tsx` — обработка callback
- [ ] isAuthenticated вычисляется из наличия token
- [ ] userRoles доступен для RequireAuth (T2103)
