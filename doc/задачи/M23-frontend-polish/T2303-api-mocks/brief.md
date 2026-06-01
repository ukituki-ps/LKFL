# T2303 — API Mocks Layer: mocked endpoints для Dashboard, Points, Documents, Support

## Веха

M23 — Frontend Polish

## Тип

code

## Проблема

Бэкенд не покрывает все эндпоинты, необходимые для фронтенда. Текущий код использует mock данные в компонентах, но нет единого слоя API mocks — при подключении бэкенда придётся переписывать компоненты.

**Недостающие эндпоинты:**

| Эндпоинт | Страница | Сейчас |
|----------|----------|--------|
| `GET /api/v1/dashboard/stats` | Dashboard | Mock в Dashboard.tsx |
| `GET /api/v1/dashboard/benefits` | Dashboard | Mock в Dashboard.tsx |
| `GET /api/v1/dashboard/events` | Dashboard | Mock в Dashboard.tsx |
| `GET /api/v1/points/balance` | Points | Mock в Points.tsx |
| `GET /api/v1/points/transactions` | Points | Mock в Points.tsx |
| `GET /api/v1/documents` | Documents | Mock в Documents.tsx |
| `POST /api/v1/support/tickets` | Support | Mock (setState) в Support.tsx |
| `GET /api/v1/support/faq` | Support | Mock в Support.tsx |
| `POST /api/v1/auth/logout` | All | Реально существует, но фронт не редиректит (T2301) |
| `GET /api/v1/auth/me` | All | Существует |

## Что делать

### 1. Создать mock API layer

Создать `frontend/src/api/mocks.ts` — слой, который перехватывает запросы к отсутствующим эндпоинтам и возвращает данные в том же формате, что и реальный бэкенд.

Подход: расширить `frontend/src/api/client.ts` mock-функциями или использовать MSW (Mock Service Worker) для перехвата fetch.

**Рекомендуемый подход:** добавить mock-функции рядом с реальными API-функциями, переключать через env var `VITE_USE_MOCKS=true`.

```typescript
// frontend/src/api/mocks.ts
export const mockDashboardStats = { points: 1250, activeBenefits: 4, daysLeft: 47 }
export const mockDashboardBenefits = [...] // из прототипа
export const mockDashboardEvents = [...] // из прототипа
export const mockPointsBalance = { total: 1250, categories: [...] }
export const mockTransactions = { all: [...], credits: [...], debits: [...] }
export const mockDocuments = [...] // из прототипа
export const mockFaq = [...] // 6 вопросов из прототипа
```

### 2. Создать typed API functions для каждого эндпоинта

```typescript
// frontend/src/api/dashboard.ts
export interface DashboardStats { points: number; activeBenefits: number; daysLeft: number }
export interface BenefitItem { name: string; provider: string; status: string; ... }
export interface EventItem { text: string; time: string; color: string }

export async function getDashboardStats(): Promise<DashboardStats> { ... }
export async function getActiveBenefits(): Promise<BenefitItem[]> { ... }
export async function getEvents(): Promise<EventItem[]> { ... }
```

### 3. Обновить компоненты — заменить inline mock на API calls

- Dashboard.tsx: `useQuery(['dashboard-stats'], getDashboardStats)`
- Points.tsx: `useQuery(['points-balance'], getPointsBalance)`
- Documents.tsx: `useQuery(['documents'], getDocuments)`
- Support.tsx: `useQuery(['faq'], getFaq)`, `useMutation(postSupportTicket)`

### 4. Fallback на mock если бэкенд недоступен

```typescript
async function getDashboardStats(): Promise<DashboardStats> {
  if (import.meta.env.VITE_USE_MOCKS === 'true') {
    return mockDashboardStats
  }
  return client.get<DashboardStats>('/api/v1/dashboard/stats')
}
```

## Требования

- Все 8 недостающих эндпоинтов покрыты mock + typed API functions
- Компоненты используют React Query (useQuery/useMutation) вместо inline mock
- Переключение mock/real через env var
- Типы данных соответствуют формату, который вернёт бэкенд (согласовано со spec)

## Критерии приёмки

- [ ] `frontend/src/api/mocks.ts` — mock данные для 8 эндпоинтов
- [ ] `frontend/src/api/dashboard.ts` — typed functions + интерфейсы
- [ ] `frontend/src/api/points.ts` — typed functions + интерфейсы
- [ ] `frontend/src/api/documents.ts` — typed functions + интерфейс
- [ ] `frontend/src/api/support.ts` — typed functions + интерфейсы
- [ ] Dashboard.tsx — useQuery вместо inline mock
- [ ] Points.tsx — useQuery вместо inline mock
- [ ] Documents.tsx — useQuery вместо inline mock
- [ ] Support.tsx — useQuery + useMutation
- [ ] VITE_USE_MOCKS=true → mock, false → реальный API
- [ ] `tsc --noEmit` ✅
- [ ] `npm run build` ✅

## Зависимости

- **depends_on:** T2301 (logout fix)
- **touches:** `frontend/src/api/`, `Dashboard.tsx`, `Points.tsx`, `Documents.tsx`, `Support.tsx`
