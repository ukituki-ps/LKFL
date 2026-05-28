# T2204 — Отчёт

## Что сделано

### Playwright конфигурация
- `frontend/playwright.config.ts` — конфигурация для 3 браузеров (Chromium, Firefox, Webkit)
- `frontend/package.json` — добавлены скрипты `test:e2e-ui`, `test:e2e-report`
- `@playwright/test` уже был в devDependencies

### Helper модули
- `frontend/e2e/helpers.ts` — моковые данные (users, engagements, categories), API моки через `page.route()`, утилиты ожидания
- `frontend/e2e/fixtures.ts` — Playwright fixtures с авторизованными страницами (employee, admin, hr)

### E2E тесты (51 тест в 9 файлах)

| Файл | Тестов | Описание |
|------|-------|----------|
| `login.spec.ts` | 7 | Login flow: success, loading state, mobile, callback, expired session, concurrent sessions, SEO |
| `catalog.spec.ts` | 7 | Каталог: load, filter, search, pagination, empty state, error state, card info |
| `dashboard.spec.ts` | 5 | Dashboard: greeting, stat cards, quick actions, error, loader |
| `shell.spec.ts` | 5 | Shell: navigation, active route, mobile menu, user menu, logout |
| `routing.spec.ts` | 6 | Routing: protected routes, admin forbidden, catch-all 404, lazy loading, callback, login |
| `admin.spec.ts` | 7 | Admin: create category, update, delete, delete protection, status update, concurrent, pages |
| `accessibility.spec.ts` | 5 | a11y: catalog, dashboard, login, keyboard nav, contrast |
| `visual.spec.ts` | 6 | Visual: catalog, dashboard, card, login, empty state, error state |
| `multi-tenant.spec.ts` | 3 | Multi-tenant: isolation, cross-tenant blocked, different tenants |

### Подход к мокированию
- Все API-запросы мокаются через `page.route()` — backend не требуется
- Моковые данные соответствуют реальным типам API (UserProfile, EngagementTypeResponse, и т.д.)
- Auth-состояние устанавливается через fixtures

## Время

~2 часа

## Замечания

1. **Visual regression** — baseline скриншоты создаются при первом запуске через `npx playwright test --update-snapshots`
2. **Login flow** — Login страница в LKFL делегирует на Keycloak. Тесты проверяют редирект и callback, но не саму форму Keycloak
3. **Auth-инъекция** — для тестов защищённых страниц используется мокирование API + установка Zustand state через page.evaluate
4. **Accessibility** — базовые проверки через Playwright. Для полноценного audit рекомендуется добавить `@axe-core/playwright`
