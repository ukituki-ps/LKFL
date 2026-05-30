# T1710 — E2E браузерные тесты (Playwright)

> **Тип:** code. Интеграция Playwright E2E тестов в CI/CD + фикс устаревших staging тестов после D2 (cookie-based auth).
> **Цель:** E2E тесты запускаются в CI, покрывают полный login flow через браузер на staging.
> **Зависит от:** T1709 (D2: cookie-based auth)

## Контекст

После аудита T1709 выяснилось, что E2E тесты **существуют в коде** но **никогда не запускались**:
- 11 spec-файлов, ~60 тестов (local mock-based) — **проходят локально** (`npx playwright test` → 4 passed)
- 10 staging тестов (real Keycloak + backend) — **биты** после D2 (проверяют `lkfl_token` в localStorage, которого больше нет)
- `build.yml` — **нет job'а** для Playwright

### Текущее состояние

| Компонент | Файл | Статус |
|-----------|------|--------|
| Playwright local config | `frontend/playwright.config.ts` | ✅ OK, 3 браузера + chaos |
| Playwright staging config | `frontend/playwright.staging.config.ts` | ✅ OK |
| Local test specs | `frontend/e2e/*.spec.ts` (11 файлов) | ✅ 4 passed локально |
| Staging test spec | `frontend/e2e/staging/auth.spec.ts` | 🔴 **бит** — проверяет `localStorage.getItem('lkfl_token')` |
| Staging helpers | `frontend/e2e/staging/helpers.ts` | 🔴 **бит** — `getToken()` читает `lkfl_token` из LS |
| CI integration | `.github/workflows/build.yml` | ❌ **отсутствует** job для Playwright |
| `setupAuthForTest` | `frontend/src/main.tsx:14-16` | ✅ экспортирован через `window.__LKFL_AUTH_STORE__` |

### Критические проблемы staging тестов после D2

D2 (T1709) убрал token из localStorage → все staging тесты, читающие `lkfl_token`, ломаются:

```typescript
// frontend/e2e/staging/helpers.ts:133 — БИТ
const token = await page.evaluate((key) => localStorage.getItem(key), LS_TOKEN);
// LS_TOKEN = 'lkfl_token' — больше не существует после D2

// frontend/e2e/staging/auth.spec.ts:103-105 — БИТ
const token = await getToken(page);
expect(token).toBeTruthy(); // ← FAIL: null после D2
expect(token!.split('.').length).toBe(3); // ← FAIL

// frontend/e2e/staging/auth.spec.ts:157 — БИТ
const tokenBefore = await getToken(page);
expect(tokenBefore).toBeTruthy(); // ← FAIL: null
```

## Что делать

### 1. Фикс staging helpers и тестов (cookie-based auth)

| Файл | Что менять |
|------|-----------|
| `e2e/staging/helpers.ts` | Убрать `LS_TOKEN = 'lkfl_token'`. `getToken()` → проверять cookie `lkfl_session` вместо LS. `apiRequest()` → использовать cookie context, не Bearer |
| `e2e/staging/auth.spec.ts` | Шаг 6 «localStorage» → заменить на «cookie `lkfl_session` установлен» + «`/api/v1/auth/me` → 200». `E2E-002` persistence → cookie persist check. `E2E-005` expired token → проверить cookie absence |

### 2. CI integration

Добавить job в `.github/workflows/build.yml`:

```
Job 4.5: E2E Local Tests (mock-based)
  - Запуск после lint-test
  - npx playwright install --with-deps chromium
  - npm run dev & (Vite dev server)
  - npx playwright test --config=playwright.config.ts --project=chromium
  - continue-on-error: false

Job 4.6: E2E Staging Tests (real browser)
  - Запуск после smoke-test-staging (есть живой staging)
  - Только на main push
  - npx playwright test --config=playwright.staging.config.ts
  - continue-on-error: false
```

### 3. Запуск и валидация

- Прогнать локальные тесты — убедиться что 4 passed (уже OK)
- Прогнать staging тесты на `https://dev.april.ukituki.tech` — убедиться что все 10 тестов проходят
- CI pipeline — зелёный

## Что НЕ входит

- Мобильные E2E тесты (отложено)
- Visual regression в CI (snapshots — только локально)
- Chaos tests в CI (только вручную)

## Критерии приёмки

- [ ] `e2e/staging/helpers.ts` — `getToken()` проверяет cookie `lkfl_session`, не `localStorage`
- [ ] `e2e/staging/auth.spec.ts` — все 10 тестов проходят на staging без ошибок
- [ ] `build.yml` — job `e2e-local` добавлен в pipeline (после lint-test)
- [ ] `build.yml` — job `e2e-staging` добавлен в pipeline (после smoke-test-staging)
- [ ] CI pipeline — все job'ы PASS (включая новые E2E)
- [ ] Smoke test staging — 6/6 чекпоинтов (не ломается)
- [ ] Локальные тесты — 4 passed (не ломается)
