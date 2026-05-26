# T2801-T2810 — Frontend F2

## Веха

M28-frontend-f2

## T2801 — Dashboard (/)
- Greeting: "Hello, {firstName}"
- Stat cards: баланс (общий), активные льготы, доступные льготы
- Active items: список активных engagement
- Event feed: последние транзакции
- Quick actions: каталог, баллы, документы
- React Query для data fetching

## T2802 — /points (Баланс)
- Общий баланс + по категориям
- Прогресс-бар сгорания (days_until_expiration)
- Баннер предупреждения (days < N)
- Транзакции: список с фильтрами (credit/debit), pagination
- Real-time update после транзакции

## T2803 — Flow activation UI
- Модальный wizard в карточке льготы
- Step-by-step navigation
- Step types: info, confirm, form
- Optimistic updates
- Loading states
- Error handling (eligibility fail, insufficient balance)

## T2804 — /documents
- Таблица документов (название, тип, дата, статус)
- Скачивание PDF
- Полис ДМС из модалки карточки льготы

## T2805 — Consent signature UI
- Модалка с текстом согласия
- Чекбокс "Я согласен"
- Подпись (timestamp + IP)
- Экран успеха

## T2806 — Admin: /admin/hr
- CRUD пользователей (таблица, поиск, фильтры)
- CRUD периодов
- Массовое начисление (кнопка → Asynq job → progress bar)
- User status management

## T2807 — Admin: /admin/billing
- CRUD billing rules
- Метрики по периодам (расход, конверсия)
- Transaction log

## T2808 — React Query
- QueryClient configuration
- Cache strategies per endpoint
- Optimistic updates для flow
- Error boundaries

## T2809 — Zustand stores
- `useBalanceStore` — баланс, транзакции
- `useEngagementsStore` — активные, в процессе
- `usePeriodsStore` — периоды (admin)

## T2810 — E2E тесты (Playwright)
- Login → catalog → eligibility → activate → debit → balance update
- Consent sign → flow unblock
- Admin period distribute → balance update

## Критерии приёмки

- [ ] Все 10 задач реализованы
- [ ] Dashboard, Points, Documents, Flow, Consent UI
- [ ] Admin: HR, Billing
- [ ] React Query + Zustand
- [ ] E2E тесты
