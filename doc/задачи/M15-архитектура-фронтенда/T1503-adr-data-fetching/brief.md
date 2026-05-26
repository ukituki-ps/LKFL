# T1503 — ADR-031: API Data Fetching Strategy

## Веха

M15-архитектура-фронтенда

## Контекст

118 API endpoints. Фронтенд должен управлять запросами эффективно: caching, error handling, optimistic updates, retry.

**Критические кейсы (из архитектуры бэкенда):**

| Кейс | Характер | Требования |
|------|----------|-----------|
| **Каталог льгот** | Backend кэширует в Redis (`catalog:` prefix). Фронтенд получает snapshot. Инвалидация при admin CRUD | Cache + invalidation на mutation |
| **Уведомления** | In-app + email + push. Polling (30s) или WebSocket (если добавим). Unread count — realtime-ish | Polling interval или SSE/WebSocket |
| **Survey ответы** | `SurveyEngine.SubmitAnswer()` — мгновенный отклик. Optimistic update UI → rollback на error | Optimistic update + rollback |
| **Баланс** | Изменяется при activation/completion/revert. Пользователь видит изменение сразу | Refetch после mutation |
| **Биллинг транзакции** | Append-only список. Новые транзакции — refetch | Windowed fetch + pagination |

## Что решить

Какую стратегию data fetching использовать:

| Вариант | Плюсы | Минусы |
|---------|-------|--------|
| **React Query (tanstack/react-query)** | Caching, background refetch, optimistic updates, devtools, retry | +15KB bundle, learning curve |
| **SWR** | Лёгкий, React-first, caching, revalidation | Меньше ecosystem, нет built-in retry |
| **Zustand + manual refetch** | Нет dependency, полный контроль | Дублирование кода, нет caching |
| **RTK Query** | Caching, codegen, devtools | Требует Redux, +60KB |

## Критерии

- Bundle size overhead
- Dev experience (devtools, debug)
- Caching strategy (особенно каталог — Redis sync)
- Optimistic updates support (survey, balance)
- Error handling + retry
- Polling/SSE support (notifications)
- Learning curve для команды
- **WebSocket:** не требуется для текущего набора use case'ов. Notifications — polling 30s. WebSocket — если понадобится, отдельный ADR

## Ожидаемое решение

Рекомендация: **React Query** (mature, best devtools, 118 endpoints → caching критичен).

**По кейсам:**
- Каталог: `queryKey: ['catalog']`, `staleTime: 5min`, refetch на admin mutation
- Нотификации: `refetchInterval: 30_000` (polling), unread badge — local state
- Survey: `mutation` + `onSuccess: invalidateQueries(['survey'])` → optimistic
- Баланс: `mutation` → `invalidateQueries(['balance'])` → refetch
- Биллинг: `useInfiniteQuery` для транзакций (pagination)

Если bundle size критичен → SWR (лёгкий, но меньше devtools).

## Результат

- `архитектура/adr/031-api-data-fetching.md` — полный ADR в формате ХАДД
