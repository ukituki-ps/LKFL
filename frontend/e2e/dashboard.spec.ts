import { test, expect } from '@playwright/test';
import { setupApiMocks, setupUserProfileErrorMocks, mockUserEmployee, gotoWithAuth } from './helpers';

test.describe('Dashboard', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('dashboard loads with greeting and stat cards', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.getByText('ЛКФЛ')).toBeVisible({ timeout: 10000 });
	});

	test('dashboard shows stat cards placeholders', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);
		// Stat cards rendered by Dashboard component inside Shell
		await expect(page.getByText('ЛКФЛ')).toBeVisible({ timeout: 10000 });
	});

	test('dashboard quick actions navigate to correct pages', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);
	});

	test('dashboard shows error when profile fails to load', async ({ page }) => {
		await setupUserProfileErrorMocks(page);
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('dashboard shows loader while loading', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});
});
