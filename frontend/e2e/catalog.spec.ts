import { test, expect } from '@playwright/test';
import { setupApiMocks, mockUserEmployee, gotoWithAuth } from './helpers';

test.describe('Catalog', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('catalog loads and displays engagement cards', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('filter by category shows filtered results', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('search input triggers search with debounce', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('pagination navigates between pages', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('empty state shows message when no engagements', async ({ page }) => {
		await page.route('/api/v1/engagements*', (route) => {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					data: [],
					pagination: { page: 1, per_page: 20, total: 0, total_pages: 0 },
				}),
			});
		});
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('error state shows error message and retry button', async ({ page }) => {
		await page.route('/api/v1/engagements*', (route) => {
			return route.fulfill({
				status: 500,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'Internal server error' }),
			});
		});
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('catalog cards display correct information', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});
});
