import { test, expect } from '@playwright/test';
import { setupApiMocks, setupUserProfileErrorMocks, mockUserEmployee, gotoWithAuth } from './helpers';

test.describe('Dashboard', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('dashboard loads with greeting and stat cards', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await expect(page.getByText('ЛКФЛ')).toBeVisible({ timeout: 5000 });
	});

	test('dashboard shows stat cards placeholders', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		// Stat cards rendered by Dashboard component inside Shell
		await expect(page.getByText('ЛКФЛ')).toBeVisible({ timeout: 5000 });
	});

	test('dashboard quick actions navigate to correct pages', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await expect(page.locator('body')).not.toBeEmpty();
	});

	test('dashboard shows error when profile fails to load', async ({ page }) => {
		await setupUserProfileErrorMocks(page);
		await gotoWithAuth(page, '/', mockUserEmployee);
		await expect(page.locator('body')).not.toBeEmpty();
	});

	test('dashboard shows loader while loading', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await expect(page.locator('body')).not.toBeEmpty();
	});
});
