# T3701-T3711 — Frontend F3

## Веха

M37-frontend-f3

## T3701 — /support
- FAQ (из content API)
- Форма обращений (тема + сообщение → ticket)
- Экран успеха

## T3702 — Activity UI в каталоге
- Карточка activity (бонус, регистрация, completion)
- Регистрация на событие
- Completion flow

## T3703 — Survey UI
- Опрос с branching
- Step-by-step
- Optimistic submit

## T3704 — Gamification UI
- Ачивки (лента, бейджи)
- Уровень лояльности, прогресс
- Trigger bonus notification

## T3705 — Admin: /admin/catalog
- CRUD карточек
- Метрики (просмотры, активации, GMV, конверсия)
- Продвижение (promo status)
- Коллекции

## T3706 — Admin: /admin/content
- CRUD FAQ
- CRUD баннеры
- Описания карточек

## T3707 — Admin: /admin/gamification
- CRUD ачивки, уровни лояльности, триггеры
- XLSX импорт

## T3708 — Admin: /admin/notifications
- Templates CRUD
- Mass notification (CEL сегмент)
- Delivery tracking

## T3709 — Admin: /admin/compliance
- Audit logs (таблица, фильтры)
- Retention policies CRUD

## T3710 — In-app notifications
- Bell icon, dropdown
- Read/unread
- Real-time (polling или SSE)

## T3711 — E2E тесты (Playwright)
- Activity flow
- Survey flow (branching)
- Gamification (achievement earn, loyalty upgrade)
- Cascade revoke (admin dismiss → employee blocked)
- Collections (create → activate all)

## Критерии приёмки
- [ ] Все 11 задач
- [ ] Все страницы сотрудника + админа
- [ ] E2E тесты
