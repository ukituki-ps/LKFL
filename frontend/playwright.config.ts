import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright конфигурация для E2E тестов LKFL.
 *
 * Тесты запускаются в трёх браузерах: Chromium, Firefox, Webkit.
 * API-запросы мокаются через page.route() — backend не требуется.
 */
export default defineConfig({
	testDir: './e2e',
	testIgnore: ['**/staging/**/*.spec.ts', '**/chaos/**/*.spec.ts'],
	timeout: 30_000,
	expect: {
		timeout: 5000,
	},
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	workers: process.env.CI ? 1 : undefined,
	reporter: process.env.CI ? 'blob' : 'html',
	use: {
		baseURL: 'http://localhost:5173',
		trace: 'on-first-retry',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure',
		locale: 'ru-RU',
		timezoneId: 'Europe/Moscow',
	},

	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] },
			testIgnore: ['**/chaos/**/*.spec.ts', '**/staging/**/*.spec.ts'],
		},
		{
			name: 'firefox',
			use: { ...devices['Desktop Firefox'] },
			testIgnore: ['**/chaos/**/*.spec.ts', '**/staging/**/*.spec.ts'],
		},
		{
			name: 'webkit',
			use: { ...devices['Desktop Safari'] },
			testIgnore: ['**/chaos/**/*.spec.ts', '**/staging/**/*.spec.ts'],
		},
		{
			name: 'chaos',
			testMatch: '**/chaos/**/*.spec.ts',
			use: {
				...devices['Desktop Chrome'],
				video: 'on',
				trace: 'on',
			},
		},
	],
});
