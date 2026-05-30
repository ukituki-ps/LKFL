# T2111 — Unit Tests

## Веха

M21-frontend-catalog

## Тип

code

## Контекст

Настройка Vitest и написание unit тестов для ключевых модулей M21.
Зависит от T2104 (auth flow), T2107 (card component), T2109 (API layer).

### Зависимости (из T2101 — package.json)

```json
{
  "devDependencies": {
    "vitest": "^2.0.0",
    "@testing-library/react": "^16.0.0",
    "@testing-library/jest-dom": "^6.4.0",
    "jsdom": "^24.0.0"
  }
}
```

### Роли (из M20 backend — `backend/internal/user/model.go`)

- `employee` — сотрудник
- `catalog_manager` — менеджер каталога
- `hr` — HR-менеджер
- `admin` — полный доступ

### Badge значения (из M20 backend — `backend/internal/engagement/catalog/handler.go`)

- `status == "promo"` → badge = "Промо"
- `status == "active"` → badge = "Доступна"
- остальные статусы → badge = "Доступна"

## Что сделать

### `vitest.config.ts` — конфигурация

```ts
import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: './src/test/setup.ts',
  },
})
```

### `src/test/setup.ts` — setup файл

```ts
import '@testing-library/jest-dom'
```

### `src/stores/authStore.test.ts` — тесты authStore

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from '@/stores/authStore'

describe('authStore', () => {
  beforeEach(() => {
    useAuthStore.setState({
      token: null,
      user: null,
      userRoles: [],
      isAuthenticated: false,
      isLoading: false,
    })
  })

  describe('setAuth', () => {
    it('устанавливает auth state', () => {
      useAuthStore.getState().setAuth(
        'test-token',
        { id: '1', email: 'test@test.com', first_name: 'Test', last_name: 'User' },
        ['employee', 'catalog_manager']
      )

      const state = useAuthStore.getState()
      expect(state.token).toBe('test-token')
      expect(state.isAuthenticated).toBe(true)
      expect(state.userRoles).toEqual(['employee', 'catalog_manager'])
      expect(state.user).toEqual({
        id: '1',
        email: 'test@test.com',
        first_name: 'Test',
        last_name: 'User',
      })
    })
  })

  describe('clearAuth', () => {
    it('очищает auth state', () => {
      useAuthStore.getState().setAuth('token', { id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' }, ['admin'])
      useAuthStore.getState().clearAuth()

      const state = useAuthStore.getState()
      expect(state.token).toBeNull()
      expect(state.isAuthenticated).toBe(false)
      expect(state.userRoles).toEqual([])
      expect(state.user).toBeNull()
    })
  })

  describe('logout', () => {
    it('вызывает POST /api/v1/auth/logout и очищает state', async () => {
      global.fetch = vi.fn(() => Promise.resolve({ ok: true }))

      useAuthStore.getState().setAuth('token', { id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' }, ['admin'])
      await useAuthStore.getState().logout()

      expect(fetch).toHaveBeenCalledWith('/api/v1/auth/logout', {
        method: 'POST',
        headers: { Authorization: 'Bearer token' },
      })

      const state = useAuthStore.getState()
      expect(state.isAuthenticated).toBe(false)
    })
  })
})
```

### `src/api/client.test.ts` — тесты API client

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiRequest } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'

describe('apiRequest', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    useAuthStore.getState().clearAuth()
  })

  it('добавляет Authorization header с токеном', async () => {
    useAuthStore.getState().setAuth('test-token', { id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' }, [])

    global.fetch = vi.fn(() => Promise.resolve({
      ok: true,
      json: () => Promise.resolve({ data: 'test' }),
    }))

    await apiRequest('/api/v1/test')

    expect(fetch).toHaveBeenCalledWith(
      '/api/v1/test',
      expect.objectContaining({
        headers: expect.objectContaining({
          'Authorization': 'Bearer test-token',
        }),
      })
    )
  })

  it('перенаправляет на /login при 401', async () => {
    const originalLocation = window.location
    vi.spyOn(window, 'location', 'get').mockImplementation(() => ({
      ...originalLocation,
      href: '',
    }))

    global.fetch = vi.fn(() => Promise.resolve({ status: 401 }))

    await expect(apiRequest('/api/v1/test')).rejects.toThrow('Unauthorized')
    expect(window.location.href).toBe('/login')
  })

  it('возвращает null при 204 NoContent', async () => {
    global.fetch = vi.fn(() => Promise.resolve({ status: 204 }))

    const result = await apiRequest('/admin/test/123')
    expect(result).toBeNull()
  })

  it('повторяет запрос при 5xx с exponential backoff', async () => {
    let callCount = 0
    global.fetch = vi.fn(() => {
      callCount++
      if (callCount < 3) {
        return Promise.resolve({ status: 500 })
      }
      return Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ success: true }),
      })
    })

    // Mock setTimeout для ускорения тестов
    vi.useFakeTimers()

    const result = await apiRequest('/api/v1/test')
    expect(callCount).toBe(3)
    expect(result).toEqual({ success: true })

    vi.useRealTimers()
  })
})
```

### `src/components/catalog/EngagementCard.test.tsx` — тесты карточки

```ts
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { EngagementCard } from '@/components/catalog/EngagementCard'
import type { EngagementTypeResponse } from '@/api/types'

const mockEngagement: EngagementTypeResponse = {
  id: '550e8400-e29b-41d4-a716-446655440000',
  slug: 'yoga-studio',
  name: 'Йога в студии',
  description: 'Абонемент на йогу',
  type: 'benefit',
  status: 'active',
  cost_cents: 150000,
  provider_name: 'FitLife',
  image_url: 'https://example.com/yoga.jpg',
  category: {
    id: 'cat-1',
    slug: 'fitness',
    name: 'Фитнес',
    icon: '🏋️',
    sort_order: 1,
  },
  offers: [],
  badge: 'Доступна',
}

describe('EngagementCard', () => {
  it('отображает название льготы', () => {
    render(<EngagementCard engagement={mockEngagement} />)
    expect(screen.getByText('Йога в студии')).toBeInTheDocument()
  })

  it('отображает стоимость в рублях', () => {
    render(<EngagementCard engagement={mockEngagement} />)
    expect(screen.getByText('1 500 ₽')).toBeInTheDocument()
  })

  it('отображает бейдж "Промо"', () => {
    const promoEngagement = { ...mockEngagement, badge: 'Промо', status: 'promo' }
    render(<EngagementCard engagement={promoEngagement} />)
    expect(screen.getByText('Промо')).toBeInTheDocument()
  })

  it('отображает бейдж "Доступна"', () => {
    render(<EngagementCard engagement={mockEngagement} />)
    expect(screen.getByText('Доступна')).toBeInTheDocument()
  })

  it('отображает название категории', () => {
    render(<EngagementCard engagement={mockEngagement} />)
    expect(screen.getByText('Фитнес')).toBeInTheDocument()
  })

  it('отображает название провайдера', () => {
    render(<EngagementCard engagement={mockEngagement} />)
    expect(screen.getByText('FitLife')).toBeInTheDocument()
  })

  it('не отображает стоимость если cost_cents не задан', () => {
    const noCostEngagement = { ...mockEngagement, cost_cents: undefined }
    render(<EngagementCard engagement={noCostEngagement} />)
    expect(screen.queryByText(/₽/)).not.toBeInTheDocument()
  })

  it('отображает количество офферов если > 1', () => {
    const multiOfferEngagement = {
      ...mockEngagement,
      offers: [
        { id: '1', name: 'Месяц', description: '', cost_cents: 150000, sort_order: 1 },
        { id: '2', name: '3 месяца', description: '', cost_cents: 400000, sort_order: 2 },
      ],
    }
    render(<EngagementCard engagement={multiOfferEngagement} />)
    expect(screen.getByText('2 варианта')).toBeInTheDocument()
  })
})
```

### `src/components/auth/RequireAuth.test.tsx` — тесты RequireAuth

```ts
import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { RequireAuth } from '@/components/auth/RequireAuth'
import { useAuthStore } from '@/stores/authStore'

function renderWithAuth(path: string, roles: string[]) {
  useAuthStore.getState().setAuth(
    'token',
    { id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
    roles
  )

  render(
    <MemoryRouter initialEntries={[path]}>
      <RequireAuth roles={roles}>
        <div data-testid="protected-content">Protected</div>
      </RequireAuth>
    </MemoryRouter>
  )
}

describe('RequireAuth', () => {
  it('перенаправляет на /login если не авторизован', () => {
    useAuthStore.getState().clearAuth()

    render(
      <MemoryRouter>
        <RequireAuth>
          <div data-testid="protected-content">Protected</div>
        </RequireAuth>
      </MemoryRouter>
    )

    expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
  })

  it('показывает контент если авторизован (без роли)', () => {
    renderWithAuth('/', ['employee'])
    expect(screen.getByTestId('protected-content')).toBeInTheDocument()
  })

  it('показывает контент если есть требуемая роль catalog_manager', () => {
    renderWithAuth('/admin/catalog', ['catalog_manager'])
    expect(screen.getByTestId('protected-content')).toBeInTheDocument()
  })

  it('показывает контент если есть требуемая роль admin', () => {
    renderWithAuth('/admin/catalog', ['admin'])
    expect(screen.getByTestId('protected-content')).toBeInTheDocument()
  })

  it('перенаправляет на /forbidden если нет требуемой роли', () => {
    useAuthStore.getState().setAuth(
      'token',
      { id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
      ['employee']
    )

    render(
      <MemoryRouter>
        <RequireAuth roles={['catalog_manager', 'admin']}>
          <div data-testid="protected-content">Protected</div>
        </RequireAuth>
      </MemoryRouter>
    )

    expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument()
  })
})
```

## Требования

- Vitest 2 + jsdom environment
- @testing-library/react 16
- Тесты authStore: setAuth, clearAuth, logout
- Тесты api/client: Authorization header, 401, 403, 204, 5xx retry
- Тесты EngagementCard: rendering, badge types, cost formatting, optional fields
- Тесты RequireAuth: auth check, role check (catalog_manager, admin, employee)
- Mock fetch для API тестов
- Fake timers для retry тестов

## Критерии приёмки

- [ ] `vitest.config.ts` — конфигурация Vitest
- [ ] `src/test/setup.ts` — setup файл
- [ ] `src/stores/authStore.test.ts` — 3 теста (setAuth, clearAuth, logout)
- [ ] `src/api/client.test.ts` — 4 теста (auth header, 401, 204, 5xx retry)
- [ ] `src/components/catalog/EngagementCard.test.tsx` — 8 тестов (name, cost, badges, category, provider, optional, offers)
- [ ] `src/components/auth/RequireAuth.test.tsx` — 5 тестов (unauth, auth, catalog_manager, admin, forbidden)
- [ ] `npm test` — все тесты зелёные
- [ ] Покрытие ключевых модулей ≥ 80%
