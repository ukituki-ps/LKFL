# ADR-033: Frontend Testing Strategy

**Статус:** Accepted
**Дата:** 2026-05-26
**Контекст:** M15-архитектура-фронтенда, T1505

---

## Контекст

Фронтенд — пользовательский интерфейс платформы для 100 000+ сотрудников. Ошибки напрямую влияют на всех пользователей. Нужна тестовая стратегия.

---

## Решение

**Тестовая пирамида:**

| Уровень | Инструмент | Что покрывает | Покрытие |
|---------|-----------|---------------|----------|
| **Unit** | Vitest + `@testing-library/react` | Компоненты, hooks, utils | 80% компонентов LKFL |
| **Integration** | Vitest + MSW | API client, store actions, точки интеграции LKFL → DS | 90% hooks |
| **E2E** | Playwright | Критические journeys (J01-J10) | 100% критических путей |

---

## Граница тестирования

| Что тестируем | Где |
|---------------|-----|
| **DS-компоненты** (`@april/ui`) | Репозиторий `DisignApril-kilo` (smoke-тесты: `AprilMobileShellBar.smoke.test.tsx`, `CardListColumn.smoke.test.tsx` и др.) |
| **LKFL-компоненты** (pages, modals, store, hooks, api/) | Репозиторий LKFL |
| **Integration-точки** (LKFL → DS) | Репозиторий LKFL: `AprilModal` в wizard, `AprilMobileShellBar` на mobile, `CardListColumn` в admin |

**LKFL пишет тесты ТОЛЬКО для LKFL-компонентов.** DS-компоненты тестируются в DS repo.

---

## Unit-тесты (Vitest + RTL)

**Что тестируем:**
- Компоненты LKFL: `Dashboard`, `Catalog`, `Points`, `Documents`, `Support`
- Shared компоненты: `BenefitCard`, `StatCard`, `TransactionRow`, `DocumentRow`, `EventRow`, `QuickActionBtn`
- Модалки: `Wizard`, `BenefitDetail`
- Layout: `Header`, `Sidebar`, `PageLayout`
- Store actions: `useBalanceStore`, `useCatalogStore`
- Hooks: `useAuth`, `useTenant`, `useNotifications`
- Utils: formatters, validators

**Не тестируем:**
- DS-компоненты (`@april/ui`) — тесты в DS repo
- Mantine-компоненты — тесты в Mantine repo

**Пример:**
```ts
// components/BenefitCard.test.tsx
import { render, screen } from '@testing-library/react'
import { BenefitCard } from './BenefitCard'

it('рендерит карточку льготы с названием и провайдером', () => {
  render(<BenefitCard name="ДМС" provider="АльфаСтрахование" price="Включено" />)
  expect(screen.getByText('ДМС')).toBeInTheDocument()
  expect(screen.getByText('АльфаСтрахование')).toBeInTheDocument()
})
```

---

## Integration-тесты (Vitest + MSW)

**Что тестируем:**
- API client: interceptors, error handling, retry
- Store actions: `fetchBalance`, `fetchCatalog` (с MSW mock)
- Integration-точки LKFL → DS:
  - `Wizard` + `AprilModal` — открытие/закрытие, навигация по шагам
  - `AprilMobileShellBar` на mobile — отображение, жесты
  - `CardListColumn` в admin — CRUD операции
  - `AprilJsonSchemaForm` в admin — рендеринг по JSON Schema

**Mock strategy:**
- MSW для API (intercept fetch → return mock data)
- `vi.mock()` для Go dependencies (не применимо — фронт)

---

## E2E-тесты (Playwright)

**Критические journeys (J01-J10):**
| Journey | Описание |
|---------|----------|
| J01 | Регистрация → Dashboard |
| J02 | Каталог → фильтрация → карточка льготы |
| J03 | Подключение льготы (wizard) |
| J04 | ДМС upgrade (wizard) |
| J05 | Мои баллы → транзакции |
| J06 | Документы → скачивание |
| J07 | Поддержка → FAQ + обращение |
| J08 | Нотификации → открытие |
| J09 | Admin: HR → импорт XLSX |
| J10 | Admin: Catalog → CRUD карточки |

**CI timing:**
- Unit: < 30s
- E2E: < 5min

---

## Подход

**Тесты после реализации** (не TDD). Проект новый, архитектура ещё не закреплена.

**Почему не TDD:**
- Архитектура фронтенда только формируется
- DS-компоненты могут измениться
- Тесты будут ломаться при рефакторинге

---

## Следствия

- `vitest.config.ts` — конфигурация Vitest
- `playwright.config.ts` — конфигурация Playwright
- `__tests__/` — рядом с компонентами (`component.test.tsx`) или в `tests/`
- `tests/msw/handlers.ts` — MSW handlers для API mocking
- `tests/e2e/` — E2E тесты Playwright (по journeys)
- CI: `npm run test` → unit + integration; `npm run test:e2e` → Playwright
