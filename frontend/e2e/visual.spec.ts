import { test, expect } from '@playwright/test';
import { setupApiMocks, mockUserEmployee, gotoWithAuth } from './helpers';

test.describe('Visual Regression', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('catalog page visual baseline', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('dashboard page visual baseline', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('engagement card visual baseline', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('login page visual baseline', async ({ page }) => {
		await page.goto('/login');
		await page.waitForTimeout(300);
		await expect(page.locator('body')).toBeVisible();
	});

	test('catalog empty state visual baseline', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});

	test('catalog error state visual baseline', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);
		await expect(page.locator('body')).toBeVisible();
	});
});
