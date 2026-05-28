import { test as base } from '@playwright/test';
import type { Page } from '@playwright/test';
import {
	setupApiMocks,
	mockUserEmployee,
	mockUserAdmin,
	mockUserHR,
	setAuthState,
} from './helpers';

/**
 * Расширенный fixture для E2E тестов LKFL.
 *
 * Каждый тест получает page с настроенными API моками.
 * Auth-состояние устанавливается через Zustand store напрямую.
 */
export const test = base.extend<{
	authPage: Page;
	adminPage: Page;
	hrPage: Page;
}>({
	page: async ({ page }, use) => {
		// Настроим API моки до каждой навигации
		await setupApiMocks(page);
		await use(page);
	},

	// Авторизованная страница (сотрудник)
	authPage: async ({ page }, use) => {
		await setupApiMocks(page);
		await setAuthState(page, mockUserEmployee);
		await use(page);
	},

	// Авторизованная страница (администратор)
	adminPage: async ({ page }, use) => {
		await setupApiMocks(page);
		await setAuthState(page, mockUserAdmin);
		await use(page);
	},

	// Авторизованная страница (HR)
	hrPage: async ({ page }, use) => {
		await setupApiMocks(page);
		await setAuthState(page, mockUserHR);
		await use(page);
	},
});

export { expect } from '@playwright/test';
