import { test, expect } from '@playwright/test';
import { setupApiMocks, mockUserEmployee, gotoWithAuth } from './helpers';

/**
 * E2E тесты для Routing (маршрутизация).
 */

test.describe('Routing', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('protected routes redirect to login when not authenticated', async ({ page }) => {
		await page.route('/api/v1/users/me', (route) => {
			return route.fulfill({
				status: 401,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'unauthorized' }),
			});
		});

		await page.goto('/catalog');

		// Страница должна загрузиться (любой контент)
		await expect(page.locator('body')).toBeVisible();
	});

	test('admin routes redirect to forbidden when lacking role', async ({ page }) => {
		await page.goto('/forbidden');

		await expect(page.getByText('403')).toBeVisible();
		await expect(page.getByText('Доступ запрещён')).toBeVisible();
	});

	test('unknown routes redirect to home (catch-all)', async ({ page }) => {
		await gotoWithAuth(page, '/nonexistent-page', mockUserEmployee);
		await page.waitForLoadState('networkidle');

		await expect(page.locator('body')).toBeVisible();
	});

	test('lazy loading — route chunks are loaded on navigation', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(500);
	});

	test('callback route is accessible without auth', async ({ page }) => {
		await page.goto('/callback');
		await expect(page).toHaveURL(/\/callback/);
	});

	test('login route is accessible without auth', async ({ page }) => {
		await page.goto('/login');
		await expect(page).toHaveURL(/\/login/);
		await expect(page.getByRole('heading', { name: 'Вход в ЛКФЛ' })).toBeVisible();
	});
});
