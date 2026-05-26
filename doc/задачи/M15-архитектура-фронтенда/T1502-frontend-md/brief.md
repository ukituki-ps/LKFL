# T1502 — `архитектура/фронтенд.md`

## Веха

M15-архитектура-фронтенда

## Контекст

Самая объёмная задача вехи. Создаёт единый архитектурный документ фронтенда — аналог `модули.md` + `пакеты-platform.md` для бэкенда.

## Что сделать

Создать `doc/архитектура/фронтенд.md` с разделами:

### Раздел A — Обзор
- Назначение, технологии, multi-tenant, ссылки на ADR (007, 008, 012, 029)
- **DS-интеграция:** `AprilProviders` как корневой провайдер (не голый `MantineProvider`), `createAprilTheme()`, `DensityProvider`/`useDensity()`
- **Density:** Comfortable по умолчанию. Compact — YAGNI (не включать без use case). `useDensity()` для адаптивных размеров
- **Тёмная тема:** DS поддерживает (light + dark). YAGNI — не включать по умолчанию. Запас: `useMantineColorScheme()`

### Раздел B — Routing
- Таблица всех страниц (employee + admin) с путями, компонентами, RBAC guard
- Lazy loading, auth guards, nested routes

### Раздел C — Структура проекта (tree `src/`)
- Актуализированное дерево с результатами T1501
- `main.tsx` → `<AprilProviders>` (токены, плотность, тема, density context)
- `pages/`, `components/`, `modals/`, `store/`, `api/`, `hooks/`, `lib/`, `lib/translations/` (i18n-подготовка), `types/`, `routes/`
- **Новые директории vs `модули.md`:** `lib/translations/`, `types/`, `routes/` — добавлены для архитектуры (i18n-подготовка, API types, route definitions). `api/platform.ts` и `hooks/useAuth.ts` — остаются (из `модули.md`).
- **Иконки:** только `AprilIcon` + `AprilIcon*` из `@april/ui`. Прямой импорт `lucide-react` — запрет (решение зафиксировать в ADR-007 при обновлении)

### Раздел D — API Layer
- Fetch wrapper, interceptors, error handling, retry policy
- Optimistic updates, caching strategy

### Раздел E — State Management (Zustand)
- Список store'ов (user, balance, catalog, notifications, wizards, survey, content)
- Pattern: selector-based, hydration

### Раздел F — Компоненты
- Shared: `BenefitCard`, `StatCard`, `TransactionRow`, `DocumentRow`, `EventRow`, `QuickActionBtn`
- Layout: `Header` (на основе `ProductHeaderToolbar`), `Sidebar` (`ProductSidebarNavigation` для admin), `PageLayout`
- Modal: `Wizard` (JSON-driven), `BenefitDetail` (tabs). Обёртка `AprilModal` (desktop) / `AprilVaulBottomSheet` (mobile)
- **Admin CRUD Pattern:** `CardListColumn` + `ProductHeaderToolbar` + `AprilModal` (create/edit) + `FacetedSearch` (фильтрация)
- DS Gap → ссылка на обновлённый ADR-029 (T1508)

### Раздел G — White-label
- Brand CSS override, tenant resolver hook
- Dark mode — YAGNI (не включать)

### Раздел H — Performance
- Code splitting, bundle analysis, Lighthouse targets
- **Error tracking:** Sentry (`@sentry/react`) — error grouping, source maps, breadcrumbs. Инициализация в `main.tsx` (после `AprilProviders`). DSN из env.
- **PWA:** `vite-plugin-pwa` — service worker для каталога льгот (offline reading). Manifest + cache strategy для `GET /catalog`.

### Раздел I — Mobile (<768px)
- Краткое описание: `AprilMobileShellBar`, `AprilVaulBottomSheet`, breakpoints, safe area, touch-ориентиры
- **Подробности:** ссылка на `фронтенд-mobile-forms.md` (T1509)

### Раздел J — Формы
- Краткое описание: Zod + react-hook-form, `AprilJsonSchemaForm` для admin-форм, wizard, survey
- **Подробности:** ссылка на `фронтенд-mobile-forms.md` (T1509)

## Результат

- `архитектура/фронтенд.md` — единый файл, ~300-400 строк (разделы I+J — краткие, ссылка на `фронтенд-mobile-forms.md`)

## Критерии приёмки

- [ ] Все 10 разделов (A→J) заполнены
- [ ] Ссылки на ADR-007, 008, 012, 029 (обновлённый)
- [ ] `AprilProviders` описан как корневой провайдер
- [ ] Density: Comfortable по умолчанию, Compact — YAGNI
- [ ] Dark mode — YAGNI, запас описан
- [ ] Иконки: `AprilIcon` + запрет прямого `lucide-react`
- [ ] Tree src/ согласован с T1501
- [ ] Routing table покрывает все страницы (5 employee + 3 admin)
- [ ] API Layer описан (error handling, retry, caching)
- [ ] Admin CRUD Pattern: `CardListColumn` + `ProductHeaderToolbar` + `AprilModal`
- [ ] Mobile раздел: краткое описание + ссылка на `фронтенд-mobile-forms.md`
- [ ] Формы раздел: краткое описание + ссылка на `фронтенд-mobile-forms.md`
- [ ] Error tracking (Sentry) описан в §H
- [ ] PWA (offline catalog) описан в §H
