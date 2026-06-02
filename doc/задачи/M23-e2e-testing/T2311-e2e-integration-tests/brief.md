# T2311 — E2E Integration Tests: Login → Dashboard → Logout через реальный OIDC flow

## Веха

M23 — E2E Testing

## Тип

code

## Проблема

Текущие e2e-тесты (`e2e/login-logout-ui.spec.ts`) используют `setupAuthForTest()` — обход Keycloak через прямой доступ к Zustand store. Это **не проверяет реальный OIDC flow** — именно то, через что проходит пользователь.

В сессии `ses_177b8285cffe3AtEud2VirtmiH` (02.06.2026) была предпринята попытка написать полноценный OIDC-тест. Результат: **~50 итераций, ни одного прохода**.

### Корневые причины провала

| Причина | Что происходило | Почему ломает flow |
|---------|----------------|-------------------|
| **`page.route()` interception** | Тест перехватывал HTTP-запросы и подменял ответы (`route.abort()`, `route.fulfill()`) | PKCE verifier хранился в Redis для «запроса» который никогда не дошёл до Keycloak |
| **Сетевая топология** | Playwright → `localhost:19081` (Keycloak), Backend → `keycloak:8080` (Docker DNS) | Три разных пути к Keycloak, code не валидируется |
| **Дублирование конфига** | `.env` vs `.env.staging` — разные `KEYCLOAK_ISSUER`, `KEYCLOAK_PUBLIC_URL` | Backend discovery URL не совпадает с тем, что видит браузер |
| **`apiRequest` vs Browser** | Тест использовал `apiRequest.fetch()` для PKCE, а browser — для навигации | PKCE verifier привязан к HTTP-контексту, cookie не совпадают |

### Что видит реальный пользователь

```
Браузер → https://dev.april.ukituki.tech/
  → nginx: / → frontend SPA (RequireAuth)
  → /login → Login.tsx → window.location.href = /api/v1/auth/login
  → nginx: /api → lkfl-server → 302 → Keycloak URL
  → nginx: /realms → keycloak → login form
  → Ввод login/password → Sign In
  → Keycloak → 302 → /api/v1/auth/callback?code=...&state=...
  → nginx → lkfl-server → token exchange → Redis PKCE → set-cookie → 302 → /
  → Dashboard с cookie lkfl_session
```

**Ноль перехватов, ноль моков, один путь через nginx.**

## Что делать

### 1. Добавить пользователя e2e в Keycloak realm

Файл: `infra/keycloak/realm-lkfl-sdek.json`

Добавить пользователя `test.employee`:

```json
{
  "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaah1",
  "username": "test.employee",
  "enabled": true,
  "email": "test.employee@sdek.local",
  "firstName": "Тест",
  "lastName": "Сотрудник",
  "credentials": [
    {
      "type": "password",
      "value": "employee123",
      "temporary": false
    }
  ],
  "realmRoles": ["employee"]
}
```

Убедиться что `http://localhost:80/*` есть в `redirectUris` (уже есть).

### 2. Добавить проект `integration` в `playwright.config.ts`

```typescript
{
  name: 'integration',
  testDir: './e2e/integration',
  use: {
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:80',
    ignoreHTTPSErrors: true,
    storageState: { cookies: [], origins: [] },
    screenshot: 'on',
    trace: 'retain-on-failure',
    video: 'retain-on-failure',
    locale: 'ru-RU',
    timezoneId: 'Europe/Moscow',
  },
  timeout: 60_000,
}
```

**`E2E_BASE_URL` из env:**
- Dev: `E2E_BASE_URL=http://localhost:80` (через nginx docker compose)
- Staging: `E2E_BASE_URL=https://dev.april.ukituki.tech` (через external nginx)

### 3. Создать `e2e/integration/login-logout.spec.ts`

**4 теста, ноль `page.route()`:**

#### Тест 1: Unauthenticated → Login

```typescript
test('unauthenticated user is redirected to /login', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveURL(/\/login/);
});
```

#### Тест 2: Login → Dashboard (реальный OIDC flow)

```typescript
test('login via Keycloak shows dashboard', async ({ page }) => {
  // Переход на login — RequireAuth уже перенаправит,
  // но Login.tsx делает window.location.href = /api/v1/auth/login
  await page.goto('/login');

  // Browser следует redirect → Keycloak login form
  await page.waitForURL(/\/realms\/lkfl-sdek\/login-form/);

  // Ввод данных
  await page.fill('input[name="username"]', 'test.employee');
  await page.fill('input[name="password"]', 'employee123');
  await page.click('button[type="submit"]');

  // Keycloak → 302 → /api/v1/auth/callback?code=... → backend → token exchange → 302 → /
  // Browser следует все redirect'ы сам
  await page.waitForURL(/\/$/, { timeout: 30_000 });

  // Проверки
  const cookie = (await page.context().cookies()).find(c => c.name === 'lkfl_session');
  expect(cookie).toBeTruthy();

  const user = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  expect(user).toBeTruthy();
  expect(JSON.parse(user).email).toBe('test.employee@sdek.local');

  // UI элементы
  await expect(page.getByTestId('app-header')).toBeVisible();
});
```

#### Тест 3: Logout → Login page

```typescript
test('logout clears session and redirects to login', async ({ page }) => {
  // Сначала логинимся (как в тесте 2)
  await loginViaKeycloak(page);

  // Логаут через UI — кликаем аватар → «Выйти»
  // Реальный logout делает fetch к /api/v1/auth/logout → backend → 302 → /login
  await page.locator('[data-testid="user-avatar"]').click();
  await page.getByRole('menuitem', { name: 'Выйти' }).click();

  await page.waitForURL(/\/login/);

  const cookie = (await page.context().cookies()).find(c => c.name === 'lkfl_session');
  expect(cookie).toBeFalsy();

  const user = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  expect(user).toBeNull();
});
```

#### Тест 4: Full cycle: Login → Dashboard → Logout → Login

```typescript
test('full cycle: login → dashboard → logout → login', async ({ page }) => {
  // Login 1
  await loginViaKeycloak(page);
  await expect(page.getByTestId('app-header')).toBeVisible();

  // Logout
  await logoutViaUI(page);
  await expect(page).toHaveURL(/\/login/);

  // Login 2 — повторный логин
  await loginViaKeycloak(page);
  await expect(page.getByTestId('app-header')).toBeVisible();
});
```

### 4. Health check в `beforeAll`

```typescript
test.beforeAll(async ({ request }, testInfo) => {
  const baseUrl = process.env.E2E_BASE_URL || 'http://localhost:80';

  // Backend
  const health = await request.get(`${baseUrl}/healthz`);
  if (health.status() !== 200) {
    testInfo.skip('Backend недоступен: ' + health.status());
    return;
  }

  // Keycloak
  const oidc = await request.get(`${baseUrl}/realms/lkfl-sdek/.well-known/openid-configuration`);
  if (oidc.status() !== 200) {
    testInfo.skip('Keycloak недоступен: ' + oidc.status());
    return;
  }
});
```

### 5. Обновить `.env.staging`

Добавить:
```
E2E_BASE_URL=https://dev.april.ukituki.tech
```

### 6. Обновить `e2e/local-login-test.spec.ts`

Заменить содержимое на placeholder:

```typescript
// Этот файл заменён на e2e/integration/login-logout.spec.ts
// Старая реализация использовала page.route() interception — ломала PKCE flow.
// Новая реализация: чистая навигация через nginx, без перехвата.
import { test } from '@playwright/test';

test.skip('deprecated — use e2e/integration/login-logout.spec.ts', () => {});
```

### 7. CI integration

Добавить в `.github/workflows/build.yml` новый job `e2e-integration`:

```yaml
e2e-integration:
  needs: [deploy-staging]
  runs-on: self-hosted
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with: { node-version: 20 }
    - run: cd frontend && npm ci
    - run: cd frontend && npx playwright install chromium
    - run: cd frontend && E2E_BASE_URL=https://dev.april.ukituki.tech npx playwright test --project=integration
```

## Требования

- **Ноль `page.route()`** в тестах integration
- **Ноль `setupAuthForTest()`** — реальный OIDC flow через Keycloak
- **Ноль `apiRequest.fetch()`** для PKCE — браузер следует redirect'ы сам
- URL из `E2E_BASE_URL` — одинаковые тесты на dev и staging
- Каждый тест — чистый browser context (`storageState: { cookies: [], origins: [] }`)
- Health check в `beforeAll` — тесты skip'аются если backend/keycloak недоступны
- Timeout 60s — OIDC flow медленнее моков

## Критерии приёмки

- [ ] Пользователь `test.employee` добавлен в `realm-lkfl-sdek.json`
- [ ] `E2E_BASE_URL` добавлен в `.env.staging`
- [ ] Проект `integration` добавлен в `playwright.config.ts`
- [ ] `e2e/integration/login-logout.spec.ts` — 4 теста
- [ ] Тест 1: Unauthenticated → Login — PASS на dev (`E2E_BASE_URL=http://localhost:80`)
- [ ] Тест 2: Login → Dashboard — PASS на dev (реальный OIDC flow, cookie, localStorage)
- [ ] Тест 3: Logout → Login — PASS на dev (cookie удалена, localStorage очищен)
- [ ] Тест 4: Full cycle — PASS на dev
- [ ] Все 4 теста PASS на staging (`E2E_BASE_URL=https://dev.april.ukituki.tech`)
- [ ] `e2e/local-login-test.spec.ts` — placeholder с `test.skip`
- [ ] CI job `e2e-integration` добавлен в `build.yml`
- [ ] Время выполнения всех 4 тестов — < 2 минуты на staging

## Зависимости

- **depends_on:** T2310 (CI/CD optimize) — чтобы CI job не тормозил от 20m pipeline
- **touches:** `playwright.config.ts`, `infra/keycloak/realm-lkfl-sdek.json`, `.env.staging`, `.github/workflows/build.yml`
- **creates:** `e2e/integration/login-logout.spec.ts`
- **modifies:** `e2e/local-login-test.spec.ts`
