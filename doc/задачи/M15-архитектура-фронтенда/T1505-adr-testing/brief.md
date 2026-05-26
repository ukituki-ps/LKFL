# T1505 — ADR-033: Frontend Testing Strategy

## Веха

M15-архитектура-фронтенда

## Контекст

Фронтенд — пользовательский интерфейс платформы. Ошибки напрямую влияют на 100 000+ сотрудников. Нужна тестовая стратегия.

## Что решить

Какие тесты писать и чем покрывать:

| Уровень | Инструмент | Что покрывает |
|---------|-----------|---------------|
| Unit | Vitest + @testing-library/react | Компоненты, hooks, utils |
| Integration | Vitest + MSW | API client, store actions |
| E2E | Playwright | Критические journeys (J01-J10) |

## Критерии

- Покрытие: компоненты LKFL (80%, не DS — DS тестируются в `DisignApril-kilo`), hooks (90%), E2E (крит. пути — 100%)
- **E2E journeys:** критические пути = J01-J10 из `спецификация/journeys/сотрудник.md`. Не все 57 journeys — только критические для сотрудника (регистрация, каталог, ДМС, баллы, документы, wizard, и т.д.).
- Скорость CI: unit < 30s, E2E < 5min
- Скорость CI: unit < 30s, E2E < 5min
- Подход: тесты после реализации (не TDD — проект новый, архитектура ещё не закреплена)
- Mock strategy: MSW для API, vi.mock для Go dependencies
- **DS-компоненты:** тестируются в репозитории `DisignApril-kilo` (smoke-тесты: `AprilMobileShellBar.smoke.test.tsx`, `AprilVaulBottomSheet.smoke.test.tsx`, `CardListColumn.smoke.test.tsx` и др.). LKFL пишет тесты **только для LKFL-компонентов** (страницы, модалки, store, hooks)
- **Integration-тесты:** для точек интеграции LKFL → DS: `AprilModal` в wizard-е, `AprilMobileShellBar` на mobile, `CardListColumn` в admin

## Ожидаемое решение

Рекомендация: **Vitest + RTL** (unit) + **Playwright** (E2E). MSW для API mocking.

**Граница тестирования:**
- DS-компоненты (`@april/ui`) — тестируются в DS repo (smoke-тесты уже есть)
- LKFL-компоненты (pages, modals, store, hooks, api/) — тестируются в LKFL
- Integration-тесты — для точек LKFL → DS (wizard + AprilModal, mobile shell, admin CRUD + CardListColumn)

## Результат

- `архитектура/adr/033-frontend-testing.md` — полный ADR в формате ХАДД
