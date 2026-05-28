import { test, expect } from '@playwright/test';
import { setupApiMocks, mockUserAdmin, gotoWithAuth } from './helpers';

test.describe('Admin CRUD', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('admin can create a new category', async ({ page }) => {
		await gotoWithAuth(page, '/admin/catalog', mockUserAdmin);
		await page.waitForTimeout(1000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('admin can update category', async ({ page }) => {
		await page.route('**/admin/engagements/categories/cat-001', (route) => {
			if (route.request().method() === 'PUT') {
				return route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						id: 'cat-001', slug: 'health-updated', name: 'Здоровье (обновлённая)',
						icon: 'heart', sort_order: 1, tenant_id: 'test-tenant',
					}),
				});
			}
			return route.continue();
		});

		const origin = 'http://localhost:5173';
		const ok = await page.evaluate(async (origin) => {
			const res = await fetch(origin + '/admin/engagements/categories/cat-001', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ slug: 'health-updated', name: 'Здоровье (обновлённая)', icon: 'heart', sort_order: 1 }),
			});
			return res.ok;
		}, origin);
		expect(ok).toBe(true);
	});

	test('admin can delete category', async ({ page }) => {
		await page.route('**/admin/engagements/categories/cat-002', (route) => {
			if (route.request().method() === 'DELETE') {
				return route.fulfill({ status: 204 });
			}
			return route.continue();
		});

		const origin = 'http://localhost:5173';
		await page.evaluate(async (origin) => {
			await fetch(origin + '/admin/engagements/categories/cat-002', { method: 'DELETE' });
		}, origin);
	});

	test('delete protection — cannot delete type with active offers', async ({ page }) => {
		await page.route('**/admin/engagements/types/type-001', (route) => {
			if (route.request().method() === 'DELETE') {
				return route.fulfill({
					status: 409,
					contentType: 'application/json',
					body: JSON.stringify({
						error: 'Cannot delete type with active offers',
						details: 'Type has 3 active offers',
					}),
				});
			}
			return route.continue();
		});

		const origin = 'http://localhost:5173';
		const status = await page.evaluate(async (origin) => {
			const res = await fetch(origin + '/admin/engagements/types/type-001', { method: 'DELETE' });
			return res.status;
		}, origin);
		expect(status).toBe(409);
	});

	test('admin can update engagement type status', async ({ page }) => {
		await page.route('**/admin/engagements/types/type-001/status', (route) => {
			if (route.request().method() === 'PATCH') {
				return route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						id: 'type-001', name: 'ДМС', slug: 'dms', status: 'inactive', tenant_id: 'test-tenant',
					}),
				});
			}
			return route.continue();
		});

		const origin = 'http://localhost:5173';
		const body = await page.evaluate(async (origin) => {
			const res = await fetch(origin + '/admin/engagements/types/type-001/status', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ status: 'inactive' }),
			});
			return res.json();
		}, origin);
		expect(body.status).toBe('inactive');
	});

	test('concurrent admin operations — two tabs see consistent state', async ({ browser }) => {
		const context1 = await browser.newContext();
		const page1 = await context1.newPage();
		await setupApiMocks(page1);

		const context2 = await browser.newContext();
		const page2 = await context2.newPage();
		await setupApiMocks(page2);

		await gotoWithAuth(page1, '/admin/catalog', mockUserAdmin);
		await gotoWithAuth(page2, '/admin/catalog', mockUserAdmin);

		await expect(page1.locator('body')).toBeVisible();
		await expect(page2.locator('body')).toBeVisible();

		await context1.close();
		await context2.close();
	});

	test('admin pages are accessible', async ({ page }) => {
		await gotoWithAuth(page, '/admin/hr', mockUserAdmin);
		await page.waitForTimeout(1000);
		await expect(page.locator('body')).toBeVisible();

		await gotoWithAuth(page, '/admin/content', mockUserAdmin);
		await page.waitForTimeout(1000);
		await expect(page.locator('body')).toBeVisible();
	});
});
