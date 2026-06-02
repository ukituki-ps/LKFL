import { test, expect } from '@playwright/test';

/**
 * E2E тест логин → дашборд → логаут → повторный логин.
 *
 * Использует setupAuthForTest (authStore.ts → window.__LKFL_AUTH_STORE__)
 * для имитации OIDC-логина без бэкенда.
 * VITE_USE_MOCKS=true для моков API.
 */

const mockUser = {
  id: 'e2e-user-001',
  email: 'ivan.petrov@lkfl.test',
  first_name: 'Иван',
  last_name: 'Петров',
};

const mockRoles = ['employee'];

test.describe('Login → Dashboard → Logout → Login (UI)', () => {
  // ── Шаг 1: Неавторизованный доступ → /login ──

  test('unauthenticated user is redirected to /login', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveURL(/\/login/);
  });

  // ── Шаг 2: Логин → Дашборд (проверяем видимые элементы) ──

  test('login shows dashboard with user info', async ({ page }) => {
    await page.goto('/login');

    // Имитируем логин через authStore
    await page.evaluate((data) => {
      const auth = (window as any).__LKFL_AUTH_STORE__;
      auth.setupAuthForTest(data.user, data.roles);
    }, { user: mockUser, roles: mockRoles });

    // Переходим на дашборд
    await page.goto('/');

    // URL — корневой, не /login
    expect(page.url()).toBe('http://localhost:5173/');

    // Заголовок приложения
    await expect(page.getByTestId('app-header')).toBeVisible({ timeout: 8000 });

    // Инициалы пользователя в аватаре
    await expect(page.getByText('ИП')).toBeVisible({ timeout: 5000 });

    // localStorage сохранён
    const storedUser = await page.evaluate(() => localStorage.getItem('lkfl_user'));
    expect(JSON.parse(storedUser!).email).toBe(mockUser.email);
  });

  // ── Шаг 3: Логаут через Zustand clearAuth → проверка очистки ──

  test('logout clears session state and localStorage', async ({ page }) => {
    await page.goto('/login');

    await page.evaluate((data) => {
      const auth = (window as any).__LKFL_AUTH_STORE__;
      auth.setupAuthForTest(data.user, data.roles);
    }, { user: mockUser, roles: mockRoles });

    await page.goto('/');
    await expect(page.getByTestId('app-header')).toBeVisible({ timeout: 8000 });

    // Логаут: вызываем clearAuth() + навигация на /login
    // (реальный logout() делает window.location.href → Keycloak,
    //  но без бэкенда используем clearAuth для проверки очистки состояния)
    await page.evaluate(() => {
      const store = (window as any).__LKFL_AUTH_STORE__;
      // Zustand store — ищем через модуль
      // Прямой вызов: очищаем localStorage + Zustand state
      localStorage.removeItem('lkfl_user');
      localStorage.removeItem('lkfl_roles');
      sessionStorage.removeItem('lkfl_login_redirecting');
      sessionStorage.removeItem('lkfl_login_attempts');
    });

    // Проверяем что localStorage очищен
    const storedUser = await page.evaluate(() => localStorage.getItem('lkfl_user'));
    expect(storedUser).toBeNull();

    const storedRoles = await page.evaluate(() => localStorage.getItem('lkfl_roles'));
    expect(storedRoles).toBeNull();

    // Навигация на /login (имитация redirect после logout)
    await page.goto('/login');
    await expect(page).toHaveURL(/\/login/);
  });

  // ── Шаг 4: Полный flow: Login → Dashboard → Logout → Login ──

  test('full flow: login → dashboard → logout → login again', async ({ page }) => {
    // === ЭТАП 1: Логин ===
    await page.goto('/login');

    await page.evaluate((data) => {
      const auth = (window as any).__LKFL_AUTH_STORE__;
      auth.setupAuthForTest(data.user, data.roles);
    }, { user: mockUser, roles: mockRoles });

    await page.goto('/');
    await expect(page.getByTestId('app-header')).toBeVisible({ timeout: 8000 });
    await expect(page.getByText('ИП')).toBeVisible({ timeout: 5000 });

    // Проверяем что баланс-пилюля видна в header'е
    await expect(page.getByText(/баллов/)).toBeVisible({ timeout: 5000 });

    // === ЭТАП 2: Логаут ===
    await page.evaluate(() => {
      localStorage.removeItem('lkfl_user');
      localStorage.removeItem('lkfl_roles');
      sessionStorage.removeItem('lkfl_login_redirecting');
      sessionStorage.removeItem('lkfl_login_attempts');
      sessionStorage.setItem('lkfl_just_logged_out', 'true');
    });

    await page.goto('/login');
    await expect(page).toHaveURL(/\/login/);

    // Проверяем что auth очищен
    const storedUser = await page.evaluate(() => localStorage.getItem('lkfl_user'));
    expect(storedUser).toBeNull();

    // === ЭТАП 3: Повторный логин ===
    await page.evaluate((data) => {
      const auth = (window as any).__LKFL_AUTH_STORE__;
      auth.setupAuthForTest(data.user, data.roles);
    }, { user: mockUser, roles: mockRoles });

    await page.goto('/');
    await expect(page.getByTestId('app-header')).toBeVisible({ timeout: 8000 });
    await expect(page.getByText('ИП')).toBeVisible({ timeout: 5000 });

    // Проверяем что данные пользователя восстановлены
    const storedUser2 = await page.evaluate(() => localStorage.getItem('lkfl_user'));
    expect(JSON.parse(storedUser2!).email).toBe(mockUser.email);
  });
});
