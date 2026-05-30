# T2204 — E2E тесты (Playwright)

## Веха

M22-f1-hardening

## Тип

code

## Контекст

E2E тестирование всех user journeys F1 с помощью Playwright.
Покрытие всех ключевых сценариев использования в трёх браузерах.

## Что сделать

> **🔴 Критическое требование:** 100% покрытие всех user journeys F1. Каждый путь пользователя от входа до выхода — покрыт E2E тестом.

### Сценарии E2E тестов

- **Login flow:** success, failed credentials, password reset, expired session, concurrent login, mobile login
- **Catalog:** load, filter by category, filter by type, search, pagination, empty state, error state, slow network, offline
- **Multi-tenant:** switch tenant, verify isolation, cross-tenant access blocked
- **Admin CRUD:** create category, create type, update status, delete protection, concurrent admin operations
- **Dashboard:** load, greeting by time of day, stat cards, quick actions, profile loading error
- **Routing:** protected routes redirect to login, admin routes redirect to forbidden, lazy loading, 404
- **Shell:** mobile navigation, admin navigation, user menu, logout flow, role-based visibility

### Конфигурация

- **Config:** `playwright.config.ts`, browsers: Chromium, Firefox, Webkit
- **Visual regression:** screenshot comparison для catalog page, dashboard, card component
- **Accessibility:** a11y audit (WCAG 2.1 AA) для ключевых страниц

**Минимум 30 E2E тестов, каждый user journey покрыт 2+ тестами (happy path + error path).**

## Критерии приёмки

- [ ] Playwright config (Chromium, Firefox, Webkit)
- [ ] Login flow E2E (5+ тестов)
- [ ] Catalog E2E (5+ тестов)
- [ ] Multi-tenant E2E (3+ тестов)
- [ ] Admin CRUD E2E (5+ тестов)
- [ ] Dashboard E2E (3+ тестов)
- [ ] Routing E2E (3+ тестов)
- [ ] Shell E2E (3+ тестов)
- [ ] Visual regression (screenshot comparison)
- [ ] Accessibility audit (WCAG 2.1 AA)
- [ ] **Минимум 30 E2E тестов**
- [ ] Все три браузера: Chromium + Firefox + Webkit
