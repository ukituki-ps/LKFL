# Отчёт T1501 — Маппинг прототип → архитектура

## Статус

✅ Завершена

## 1. Сверка страниц сотрудника

| Страница | Прототип | `модули.md` строка 177 | Статус |
|----------|----------|----------------------|--------|
| Главная `/` | `page-dashboard` — баланс (3 stat cards), активные льготы (4 benefit rows), лента событий (3 events), быстрые действия (5 quick actions) | Dashboard — баланс, активные льготы, лента событий | ✅ согласовано |
| Каталог `/catalog` | `page-catalog` — search box + filter pills (6 категорий), grid 3×4 (12 benefit cards) | Catalog — поиск, N фильтры, карточки льгот | ✅ согласовано |
| Мои баллы `/points` | `page-points` — balance card (green gradient), категории (3 прогресс-бара), транзакции (8 tx rows) с filter pills (all/plus/minus) | Points — баланс по категориям, история транзакций | ✅ согласовано |
| Документы `/documents` | `page-documents` — таблица: 5 строк (doc name + meta, badge type, date, badge status, download btn) | Documents — заявления, согласия, полисы + download | ✅ согласовано |
| Поддержка `/support` | `page-support` — FAQ accordion (6 вопросов) + форма обращения (select + textarea + submit) | Support — FAQ (accordion) + форма обращения | ✅ согласовано |

**Расхождения:** 0

## 2. Сверка модалок

| Модалка | Прототип | `модули.md` строка 193 | Статус |
|---------|----------|----------------------|--------|
| DMS Wizard (`openDmsWizard('upgrade')`) | 4 шага: Опция (option cards) → Оплата (pay options) → Подтверждение (confirm doc + checkbox) → Готово (success) | Wizard — JSON-driven generic renderer | ✅ согласовано |
| DMS Relative (`openDmsWizard('relative')`) | Аналогичная структура (4 шага) | Wizard — тот же generic renderer | ✅ согласовано |
| MatCapital (`openMatWizard()`) | 4 шага: Условия (info banner + mat-amount + conditions) → Данные (формы) → Подтверждение → Готово | Wizard — тот же generic renderer | ✅ согласовано |
| BenefitDetail (`openBenefitModal('dms-base')`) | 2 режима: Generic (price, provider, btn) / DMS (3 вкладки: Условия, Полис, Клиники) | BenefitDetail — 2 режима: общий / ДМС (3 вкладки) | ✅ согласовано |

**Расхождения:** 0

## 3. Описание admin-страниц (отсутствуют в прототипе)

### HR (`/admin/hr`)
- **Layout:** `ProductSidebarNavigation` + `ProductHeaderToolbar`
- **Секции:** Пользователи (таблица + import XLSX), Периоды (создание/редактирование), Геймификация (ачивки, уровни лояльности), Апрув (очередь подтверждений льгот), Аналитика (метрики использования)
- **Данные:** `GET /admin/users/`, `GET /admin/periods/`, `GET /admin/gamification/`, `GET /admin/approval/`

### Catalog Admin (`/admin/catalog`)
- **Layout:** `ProductSidebarNavigation` + `ProductHeaderToolbar` + `CardListColumn`
- **CRUD:** Карточки льгот (создание, редактирование, жизненный цикл: draft → published → archived)
- **Метрики:** Просмотры, активации, conversion по карточкам
- **Продвижение:** Featured flag, порядок в каталоге
- **Наборы:** EngagementCollection CRUD
- **Данные:** `GET /admin/engagements/`, `GET /admin/categories/`, `GET /admin/engagement-types/`, `GET /admin/collections/`

### Content Admin (`/admin/content`)
- **Layout:** `ProductSidebarNavigation` + `ProductHeaderToolbar` + `CardListColumn`
- **CRUD:** FAQ (create/edit/delete), Баннеры (создание, расписание показа), Описания карточек (rich text)
- **Данные:** `GET /admin/content/`, `POST /admin/content/faq/`, `POST /admin/content/banners/`

## 4. Недостающие компоненты в tree `src/`

| Компонент | Статус | Решение |
|-----------|--------|---------|
| `Header` | ❌ отсутствует | **Добавить** → `components/layout/Header.tsx` (на основе `ProductHeaderToolbar` + навигация из прототипа: sticky nav, logo, nav-links, balance-pill, bell, avatar) |
| `Sidebar` | ❌ отсутствует | **Добавить** → `components/layout/Sidebar.tsx` (`ProductSidebarNavigation` для admin) |
| `PageLayout` | ❌ отсутствует | **Добавить** → `components/layout/PageLayout.tsx` (обёртка: header + main content area + padding) |
| `Footer` | ❌ не в прототипе | **YAGNI** — не нужен для текущего UI |
| `NotificationsPanel` | ❌ отсутствует | **Добавить** → `components/NotificationsPanel.tsx` (dropdown по клику на bell; polling 30s через React Query) |
| `BalancePill` | ✅ есть в прототипе, ❌ нет в tree | **Добавить** → `components/BalancePill.tsx` (compact pill: icon + value + unit; Mantine Badge) |

## 5. Дополненное дерево `src/`

```
src/
├── main.tsx                    # entry point, <AprilProviders> (токены, плотность, тема)
├── App.tsx                     # Router, layout
├── theme/
│   ├── brand-sdek.css          # white-label override (после @april/tokens)
│   └── theme.ts                # createAprilTheme() + brand config
├── pages/
│   ├── Dashboard.tsx
│   ├── Catalog.tsx
│   ├── Points.tsx
│   ├── Documents.tsx
│   ├── Support.tsx
│   └── admin/
│       ├── HR.tsx
│       ├── CatalogAdmin.tsx
│       └── ContentAdmin.tsx
├── components/
│   ├── layout/
│   │   ├── Header.tsx          # sticky header (logo + nav-links + right-section)
│   │   ├── Sidebar.tsx         # admin sidebar (ProductSidebarNavigation)
│   │   └── PageLayout.tsx      # header + main content wrapper
│   ├── BenefitCard.tsx         # карточка льготы (catalog grid)
│   ├── BenefitRow.tsx          # строка льготы (dashboard active benefits)
│   ├── TransactionRow.tsx      # строка транзакции (points page)
│   ├── DocumentRow.tsx         # строка документа (documents table)
│   ├── StatCard.tsx            # карточка статистики (dashboard)
│   ├── EventRow.tsx            # строка события (dashboard events feed)
│   ├── QuickActionBtn.tsx      # кнопка быстрого действия (dashboard)
│   ├── BalancePill.tsx         # компактный pill баланса (header)
│   ├── NotificationsPanel.tsx  # dropdown уведомлений (bell icon)
│   └── FilterPills.tsx         # горизонтальные filter pills (catalog)
├── modals/
│   ├── Wizard.tsx              # JSON-driven wizard renderer (Mantine Stepper + AprilModal)
│   ├── BenefitDetail.tsx       # модалка деталей льготы (tabs: условия / полис / клиники)
│   └── wizard-configs/         # JSON configs — dms-upgrade.json, dms-relative.json, matkapital.json
├── store/                      # Zustand stores
│   ├── user.ts
│   ├── balance.ts
│   ├── catalog.ts
│   ├── notifications.ts
│   ├── wizards.ts              # WizardEngine store
│   └── content.ts              # FAQ, banners (admin)
├── api/                        # API client
│   └── platform.ts             # fetch wrapper + endpoints (118 endpoints)
├── hooks/
│   ├── useAuth.ts              # Keycloak integration
│   ├── useTenant.ts            # tenant resolution
│   └── useNotifications.ts     # polling 30s (или React Query)
├── lib/
│   ├── constants.ts            # категории, статусы, категории льгот
│   └── translations/           # i18n-подготовка
│       └── ru.ts               # strings object (YAGNI: не i18next)
├── types/                      # API types (codegen из OpenAPI)
│   └── api.ts                  # generated
├── routes/                     # route definitions + guards
│   ├── employee.tsx            # public routes
│   └── admin.tsx               # admin routes (RBAC guard)
└── lib/
    └── json-schema.ts          # JSON Schema types для AprilJsonSchemaForm
```

## 6. Расхождения

| # | Описание | Действие |
|---|----------|----------|
| R01 | `модули.md` строка 231: `Admin.tsx` + `HR.tsx` на одном уровне → в T1502 перенести admin-страницы в `pages/admin/` | ✅ зафиксировано |
| R02 | `модули.md` не упоминает `NotificationsPanel`, `BalancePill`, `FilterPills` как отдельные компоненты | ✅ добавлены в tree |
| R03 | `модули.md` строка 186: `RecommendationsAdmin` — ⏸️ TODO M15 → не включать в architecture (stub) | ✅ зафиксировано |
| R04 | `модули.md` строка 218: `main.tsx` — «Mantine provider, April theme» → заменить на `<AprilProviders>` (DS root provider) | ✅ зафиксировано |

## 7. Вывод

- Все 5 страниц сотрудника согласованы с прототипом: **0 расхождений**
- Все 4 модалки согласованы: **0 расхождений**
- 3 admin-страницы описаны (HR, Catalog, Content)
- Tree `src/` дополнено: `components/layout/` (Header, Sidebar, PageLayout), `NotificationsPanel`, `BalancePill`, `FilterPills`, `lib/translations/`, `types/`, `routes/`
- `main.tsx` → `<AprilProviders>` (не голый `MantineProvider`)
