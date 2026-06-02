# T2303 — Report

## Выполнено

- [x] mocks.ts — mock данные для 8 эндпоинтов (Dashboard, Points, Documents, Support)
- [x] dashboard.ts — typed functions: getDashboardStats, getActiveBenefits, getEvents
- [x] points.ts — typed functions: getPointsBalance, getTransactions
- [x] documents.ts — typed functions: getDocuments
- [x] support.ts — typed functions: getFaq, postSupportTicket
- [x] index.ts — barrel exports обновлены (функции + типы)
- [x] .env.dev — VITE_USE_MOCKS=true
- [x] Dashboard.tsx — миграция на useQuery (3 запроса: stats, benefits, events)
- [x] Points.tsx — миграция на useQuery (2 запроса: balance, transactions)
- [x] Documents.tsx — миграция на useQuery (1 запрос: documents)
- [x] Support.tsx — миграция на useQuery (faq) + useMutation (postSupportTicket)

## Изменённые файлы

| Файл | Действие | Описание |
|------|----------|----------|
| `frontend/src/api/mocks.ts` | создан | Интерфейсы + mock данные для 8 эндпоинтов |
| `frontend/src/api/dashboard.ts` | создан | getDashboardStats, getActiveBenefits, getEvents |
| `frontend/src/api/points.ts` | создан | getPointsBalance, getTransactions |
| `frontend/src/api/documents.ts` | создан | getDocuments |
| `frontend/src/api/support.ts` | создан | getFaq, postSupportTicket |
| `frontend/src/api/index.ts` | изменён | Добавлены barrel exports для новых модулей |
| `frontend/.env.dev` | создан | VITE_USE_MOCKS=true |
| `frontend/src/pages/Dashboard.tsx` | изменён | useQuery для stats/benefits/events, Skeleton + error handling |
| `frontend/src/pages/Points.tsx` | изменён | useQuery для balance/transactions, Skeleton + error handling |
| `frontend/src/pages/Documents.tsx` | изменён | useQuery для documents, Skeleton + error handling |
| `frontend/src/pages/Support.tsx` | изменён | useQuery для FAQ, useMutation для тикетов, Skeleton + error handling |

## Валидация

- tsc --noEmit: ✅ (без ошибок)
- npm run build: ✅ (сборка успешна, 3249 модулей)

## Замечания

- Каждая страница использует `<Skeleton height={200} />` при загрузке и `<Text c="red">` при ошибке.
- Dashboard: stat cards показывают 3 inline-скелетона; ActiveBenefitsList и EventsFeed — обёрточные компоненты с props isLoading/isError.
- Points: клиентская фильтрация транзакций (all/credits/debits) поверх данных из API. API возвращает `amount` всегда положительным, знак выставляется при рендере (`+`/`-`).
- Support: форма создаёт тикет через `useMutation`, при успехе — success state с кнопкой «Новое обращение». Кнопка отправки показывает `loading` индикатор и заблокирована во время отправки.
- StubBadge компоненты сохранены на всех страницах.
- Каждый API-модуль содержит дублированную функцию `mockDelay` — это намеренно: модули независимы и не зависят друг от друга.
- Задержка mock-запросов: 200ms для имитации сетевой задержки.
