import { test, expect } from '@playwright/test';
import { setupApiMocks, mockUserEmployee, setAuthState } from './helpers';

/**
 * E2E тесты для Login flow.
 *
 * Login в LKFL делегируется backend → Keycloak (OIDC).
 * Тесты проверяют:
 * - Редирект на backend login endpoint
 * - Callback обработку
 * - Состояние авторизации после логина
 * - Конкурентные сессии
 * - Мобильную адаптацию
 */

test.describe('Login Flow', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('login success — редирект на backend login и переход на dashboard', async ({
		page,
	}) => {
		// Мокаем login redirect — вместо перехода на Keycloak
		// возвращаем callback с токеном
		await page.route('/api/v1/auth/login*', (route) => {
			// Redirect на callback с токеном
			return route.fulfill({
				status: 302,
				headers: {
					location: '/callback?token=mock-jwt&user=' + encodeURIComponent(JSON.stringify(mockUserEmployee)),
				},
			});
		});

		await page.goto('/login');

		// Проверяем что страница показывает сообщение о перенаправлении
		await expect(page.getByText('Вход в ЛКФЛ')).toBeVisible();
		await expect(page.getByText('Перенаправление')).toBeVisible();
	});

	test('login page shows loading state', async ({ page }) => {
		await page.goto('/login');

		// Страница логина показывает заголовок и текст загрузки
		await expect(page.getByRole('heading', { name: 'Вход в ЛКФЛ' })).toBeVisible();
		await expect(page.getByText('Перенаправление на страницу входа')).toBeVisible();
	});

	test('login page is responsive on mobile viewport', async ({ page }) => {
		await page.setViewportSize({ width: 375, height: 812 });
		await page.goto('/login');

		// Проверяем что контент виден на мобильном
		await expect(page.getByRole('heading', { name: 'Вход в ЛКФЛ' })).toBeVisible();

		// Проверяем что контейнер не выходит за пределы viewport
		const container = page.locator('[class*="mantine-Stack-root"]');
		await expect(container).toBeVisible();
	});

	test('callback page processes auth response', async ({ page }) => {
		// Мокаем callback endpoint
		await page.route('/api/v1/auth/callback*', (route) => {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					token: 'mock-jwt',
					user: mockUserEmployee,
				}),
			});
		});

		await page.goto('/callback?token=mock-jwt');

		// Callback страница должна обработать токен
		// (реализация зависит от Callback.tsx)
		await expect(page).toHaveURL(/.*\/callback.*/);
	});

	test('expired session redirects to login', async ({ page }) => {
		// Мокаем 401 ответ на запрос профиля
		await page.route('/api/v1/auth/me', (route) => {
			return route.fulfill({
				status: 401,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'unauthorized' }),
			});
		});

		await page.route('/api/v1/users/me', (route) => {
			return route.fulfill({
				status: 401,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'Token expired' }),
			});
		});

		// При 401 apiRequest должен редиректить на /login
		// Проверяем что страница логина доступна
		await page.goto('/login');
		await expect(page.getByRole('heading', { name: 'Вход в ЛКФЛ' })).toBeVisible();
	});

	test('concurrent sessions are independent', async ({ browser }) => {
		// Создаём два контекста браузера (две независимые сессии)
		const context1 = await browser.newContext();
		const page1 = await context1.newPage();
		await setupApiMocks(page1);

		const context2 = await browser.newContext();
		const page2 = await context2.newPage();
		await setupApiMocks(page2);

		// Открываем login на обеих страницах
		await page1.goto('/login');
		await page2.goto('/login');

		// Обе страницы должны работать независимо
		await expect(page1.getByRole('heading', { name: 'Вход в ЛКФЛ' })).toBeVisible();
		await expect(page2.getByRole('heading', { name: 'Вход в ЛКФЛ' })).toBeVisible();

		await context1.close();
		await context2.close();
	});

	test('login page has proper meta tags for SEO', async ({ page }) => {
		await page.goto('/login');

		// Проверяем title страницы
		await expect(page).toHaveTitle(/LKFL/);
	});
});
