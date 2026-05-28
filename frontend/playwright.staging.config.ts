import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright конфигурация для STAGING E2E тестов.
 *
 * Эти тесты бьют в РЕАЛЬНЫЙ backend (dev-стенд).
 * Никаких моков — тестируем полный flow: browser → Keycloak → backend → frontend.
 *
 * Запуск:
 *   npx playwright test --config=playwright.staging.config.ts
 *   npx playwright test --config=playwright.staging.config.ts --grep "login"
 *
 * Переменные окружения:
 *   E2E_BASE_URL     — URL стенда (default: https://dev.april.ukituki.tech)
 *   E2E_USERNAME     — логин для Keycloak (default: admin)
 *   E2E_PASSWORD     — пароль для Keycloak (default: admin-dev-password)
 *   E2E_TIMEOUT      — общий таймаут в мс (default: 60000)
 */
export default defineConfig({
	testDir: './e2e/staging',
	testMatch: '*.spec.ts',
	timeout: parseInt(process.env.E2E_TIMEOUT || '60000', 10),
	expect: {
		timeout: 10000,
	},
	fullyParallel: false, // login flow не параллелен — одна сессия Keycloak
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 3 : 1,
	workers: 1,
	reporter: process.env.CI ? [['blob'], ['html']] : 'html',
	use: {
		baseURL: process.env.E2E_BASE_URL || 'https://dev.april.ukituki.tech',
		trace: 'retain-on-failure',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure',
		locale: 'ru-RU',
		timezoneId: 'Europe/Moscow',
		ignoreHTTPSErrors: true,
	},
	projects: [
		{
			name: 'staging',
			use: {
				// Встроенный Chromium от Playwright (не требует системный Chrome)
				...devices['Desktop Chrome'],
			},
		},
	],
});
