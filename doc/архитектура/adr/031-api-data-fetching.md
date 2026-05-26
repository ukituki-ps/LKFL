# ADR-031: API Data Fetching Strategy

**Статус:** Accepted
**Дата:** 2026-05-26
**Контекст:** M15-архитектура-фронтенда, T1503

---

## Контекст

118 API endpoints. Фронтенд — React SPA для 100 000+ сотрудников. Нужно управлять запросами эффективно: caching, error handling, optimistic updates, retry, polling.

**Критические кейсы:**

| Кейс | Характер | Требования |
|------|----------|-----------|
| Каталог льгот | Backend кэширует в Redis (`catalog:`). Фронтенд получает snapshot | Cache + invalidation на mutation |
| Уведомления | In-app + email + push. Unread count — realtime-ish | Polling 30s или SSE |
| Survey ответы | `SurveyEngine.SubmitAnswer()` — мгновенный отклик | Optimistic update + rollback |
| Баланс | Изменяется при activation/completion/revert | Refetch после mutation |
| Биллинг транзакции | Append-only список | Windowed fetch + pagination |

---

## Рассмотренные варианты

### Вариант А: React Query (`@tanstack/react-query`)

**Bundle:** +15KB (gzipped). **Maturity:** production, 20M+ weekly downloads.

| Плюсы | Минусы |
|-------|--------|
| Built-in caching + staleTime | +15KB bundle |
| Background refetch + deduplication | Learning curve (queryKeys, mutation hooks) |
| Optimistic updates (onMutate → onSuccess/onError) | Ещё один абстракционный слой |
| DevTools (React DevTools extension) | — |
| Infinite queries (pagination) | — |
| Polling (`refetchInterval`) | — |
| Query invalidation (cache sync после mutation) | — |

### Вариант Б: SWR (`swr`)

**Bundle:** +3KB. **Maturity:** production, Vercel-backed.

| Плюсы | Минусы |
|-------|--------|
| Лёгкий (+3KB) | Меньше devtools |
| React-first API | Нет built-in retry (нужен `@tanstack/query-core` или кастом) |
| Revalidation on focus/reenable | Меньше ecosystem |
| Deduplication | — |

### Вариант В: Zustand + manual refetch

**Bundle:** 0KB (Zustand уже есть).

| Плюсы | Минусы |
|-------|--------|
| Нет новой dependency | Дублирование кода (caching logic в каждом store) |
| Полный контроль | Нет deduplication |
| Простая отладка | Нет background refetch, optimistic updates — ручная реализация |

### Вариант Г: RTK Query

**Bundle:** +60KB. Требует Redux Toolkit.

| Плюсы | Минусы |
|-------|--------|
| Codegen-friendly | Требует миграцию с Zustand на Redux |
| Caching + devtools | +60KB, overkill |

---

## Решение

**React Query (`@tanstack/react-query`)**

Обоснование:
1. **118 endpoints → caching критичен.** React Query решает caching out-of-the-box.
2. **Optimistic updates** для survey, balance — native support (`onMutate`, `onSuccess`, `onError`).
3. **DevTools** — критично для отладки 100K+ пользователей.
4. **Polling** для нотификаций — `refetchInterval: 30_000` одной строкой.
5. **Infinite queries** для биллинг-транзакций — `useInfiniteQuery`.
6. **+15KB** — приемлемо (PWA offline + Sentry уже добавляют больше).

---

## Стратегия по кейсам

| Кейс | Паттерн | Конфиг |
|------|---------|--------|
| **Каталог** | `useQuery(['catalog'], fetchCatalog)` | `staleTime: 5 * 60_000`, `gcTime: 10 * 60_000` |
| **Нотификации** | `useQuery(['notifications'], fetchNotifications)` | `refetchInterval: 30_000`, unread badge — local state |
| **Survey submit** | `useMutation(['survey.submit'], submitAnswer)` | `onSuccess: invalidateQueries(['survey'])` → optimistic |
| **Баланс** | `useMutation(['balance'], mutation)` → `invalidateQueries(['balance'])` | Refetch после mutation |
| **Биллинг** | `useInfiniteQuery(['transactions'], fetchTransactions)` | Windowed fetch + pagination |

**Invalidation после admin-мутейшнов:**
```ts
// После создания/редактирования карточки в admin:
await queryClient.invalidateQueries({ queryKey: ['catalog'] })
```

---

## WebSocket

Не требуется для текущего набора use case'ов. Notifications — polling 30s. WebSocket — отдельный ADR при необходимости.

---

## Следствия

- `api/queryClient.ts` — создание QueryClient + React Query DevTools
- `api/useCatalogQuery.ts` — кастомные hooks для catalog queries
- `api/useNotificationsQuery.ts` — polling hook для нотификаций
- `api/useSurveyMutation.ts` — optimistic update hook для survey
- `api/useBalanceMutation.ts` — refetch-on-mutation hook для баланса
- `api/useTransactionsQuery.ts` — infinite query для транзакций
- `api/useAdminMutations.ts` — invalidateQueries для admin CRUD

## Альтернативы

SWR рассматривался как более лёгкий вариант. Отклонён: отсутствие devtools и ecosystem для 118 endpoints — неприемлемый компромисс.
