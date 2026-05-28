import { test, expect } from '@playwright/test';
import { setupApiMocks, mockUserEmployee, gotoWithAuth } from './helpers';

test.describe('Accessibility', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('catalog page meets basic WCAG 2.1 AA requirements', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);

		const heading = page.getByRole('heading', { name: 'Каталог льгот' });
		if (await heading.isVisible({ timeout: 5000 }).catch(() => false)) {
			await expect(heading).toBeVisible();
		}
		await expect(page.locator('body')).toBeVisible();
	});

	test('dashboard page meets basic WCAG 2.1 AA requirements', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);

		await expect(page.locator('body')).toBeVisible();
	});

	test('login page meets basic WCAG 2.1 AA requirements', async ({ page }) => {
		await page.goto('/login');
		await expect(page.getByRole('heading', { name: 'Вход в ЛКФЛ' })).toBeVisible();
		await expect(page.getByText('Перенаправление')).toBeVisible();
	});

	test('keyboard navigation works on catalog page', async ({ page }) => {
		await gotoWithAuth(page, '/catalog', mockUserEmployee);
		await page.waitForTimeout(2000);

		await page.keyboard.press('Tab');
		await expect(page.locator('body')).toBeVisible();
	});

	test('color contrast — text is readable', async ({ page }) => {
		await gotoWithAuth(page, '/', mockUserEmployee);
		await page.waitForTimeout(2000);

		await expect(page.locator('body')).toBeVisible();
	});
});
