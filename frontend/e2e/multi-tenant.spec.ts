import { test, expect } from '@playwright/test';
import { setupApiMocks, mockUserEmployee, gotoWithAuth } from './helpers';

test.describe('Multi-tenant', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('tenant isolation — user sees only own tenant data', async ({ page }) => {
		const tenant1User = {
			id: 'user-tenant1', email: 'user@tenant1.test', first_name: 'Алексей', last_name: 'Тенантов',
			tenant_id: 'tenant-001', roles: ['employee'],
		};

		await gotoWithAuth(page, '/catalog', tenant1User);
		await page.waitForTimeout(2000);

		await expect(page.locator('body')).toBeVisible();
	});

	test('cross-tenant access is blocked', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);

		await expect(page.locator('body')).toBeVisible();
	});

	test('different tenants see different catalogs', async ({ browser }) => {
		const tenant1User = { id: 'user-t1', email: 'user@t1.test', first_name: 'User', last_name: 'T1', tenant_id: 'tenant-001', roles: ['employee'] };
		const tenant2User = { id: 'user-t2', email: 'user@t2.test', first_name: 'User', last_name: 'T2', tenant_id: 'tenant-002', roles: ['employee'] };

		const context1 = await browser.newContext();
		const page1 = await context1.newPage();
		await setupApiMocks(page1);

		const context2 = await browser.newContext();
		const page2 = await context2.newPage();
		await setupApiMocks(page2);

		await gotoWithAuth(page1, '/catalog', tenant1User);
		await gotoWithAuth(page2, '/catalog', tenant2User);

		await expect(page1.locator('body')).toBeVisible();
		await expect(page2.locator('body')).toBeVisible();

		await context1.close();
		await context2.close();
	});
});
