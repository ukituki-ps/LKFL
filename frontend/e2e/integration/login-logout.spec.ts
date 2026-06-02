// T2311 — E2E Integration Tests: Login → Dashboard → Logout
// Реальный OIDC flow через Keycloak, без моков и перехвата запросов.
//
// Правила:
// - Ноль page.route() — не перехватываем HTTP-запросы
// - Ноль setupAuthForTest() — только реальный OIDC
// - Чистый context — каждый тест начинается без cookie и localStorage

import { test, expect } from '@playwright/test';

// ─── Credentials ────────────────────────────────────────────────────────────
const USERNAME = process.env.E2E_USERNAME || 'test.employee';
const PASSWORD = process.env.E2E_PASSWORD || 'employee123';

// ─── Health check ───────────────────────────────────────────────────────────
test.beforeAll(async ({ request }, testInfo) => {
	const baseUrl = process.env.E2E_BASE_URL || 'http://localhost:80';

	// Проверяем доступность backend'а
	const health = await request.get(`${baseUrl}/healthz`);
	if (health.status() !== 200) {
		testInfo.skip(`Backend недоступен: ${health.status()}`);
		return;
	}

	// Проверяем доступность Keycloak OIDC discovery
	const oidc = await request.get(
		`${baseUrl}/realms/lkfl-sdek/.well-known/openid-configuration`,
	);
	if (oidc.status() !== 200) {
		testInfo.skip(`Keycloak недоступен: ${oidc.status()}`);
		return;
	}
});

// ─── Helpers ────────────────────────────────────────────────────────────────

/**
 * loginViaKeycloak — выполняет полный OIDC flow через Keycloak:
 * 1. Переход на /login → redirect на Keycloak
 * 2. Ввод username/password на странице Keycloak
 * 3. Submit → Keycloak redirect → callback → token exchange → /
 * 4. Проверка что пользователь на Dashboard
 *
 * Ноль page.route(), ноль apiRequest — чистая навигация браузера.
 */
async function loginViaKeycloak(page: import('@playwright/test').Page): Promise<void> {
	// Шаг 1: Переход на /login
	// RequireAuth перенаправит на /login, если пользователь не авторизован.
	// Login.tsx делает window.location.href = /api/v1/auth/login → backend → 302 → Keycloak
	await page.goto('/login');

	// Шаг 2: Ожидание формы Keycloak
	await page.waitForURL(/\/realms\/lkfl-sdek\/login-form/, { timeout: 30_000 });

	// Шаг 3: Ввод учётных данных
	await page.fill('input[name="username"]', USERNAME);
	await page.fill('input[name="password"]', PASSWORD);
	await page.click('button[type="submit"]');

	// Шаг 4: Browser следует все redirect'ы сам:
	// Keycloak → 302 → /api/v1/auth/callback?code=...
	// → backend token exchange → Redis PKCE validation → set-cookie
	// → 302 → /
	await page.waitForURL(/\/$/, { timeout: 30_000 });

	// Шаг 5: Проверка авторизации
	const cookie = (await page.context().cookies()).find(
		(c) => c.name === 'lkfl_session',
	);
	expect(cookie, 'lkfl_session cookie должна быть установлена после логина').toBeTruthy();

	const user = await page.evaluate(() => localStorage.getItem('lkfl_user'));
	expect(user, 'lkfl_user в localStorage должен быть установлен после логина').toBeTruthy();

	// Шаг 6: Проверка UI
	await expect(page.getByTestId('app-header')).toBeVisible();
}

/**
 * logoutViaUI — выполняет logout через UI:
 * 1. Клик по аватару пользователя
 * 2. Выбор «Выйти» из меню
 * 3. Проверка редиректа на /login
 *
 * Реальный logout: fetch к /api/v1/auth/logout → backend → 302 → /login
 */
async function logoutViaUI(page: import('@playwright/test').Page): Promise<void> {
	// Клик по аватару — открывает меню
	await page.locator('[data-testid="user-avatar"]').click();

	// Выбор пункта «Выйти»
	await page.getByRole('menuitem', { name: 'Выйти' }).click();

	// Ожидание редиректа на /login
	await page.waitForURL(/\/login/, { timeout: 15_000 });

	// Проверка очистки сессии
	const cookie = (await page.context().cookies()).find(
		(c) => c.name === 'lkfl_session',
	);
	expect(cookie, 'lkfl_session cookie должна быть удалена после logout').toBeFalsy();

	const user = await page.evaluate(() => localStorage.getItem('lkfl_user'));
	expect(user, 'lkfl_user в localStorage должен быть очищен после logout').toBeNull();
}

// ─── Tests ──────────────────────────────────────────────────────────────────

test.describe('OIDC Login → Dashboard → Logout', () => {
	// Тест 1: Unauthenticated → Login redirect
	test('unauthenticated user is redirected to /login', async ({ page }) => {
		await page.goto('/');
		await expect(page).toHaveURL(/\/login/);
	});

	// Тест 2: Login → Dashboard (реальный OIDC flow)
	test('login via Keycloak shows dashboard', async ({ page }) => {
		await loginViaKeycloak(page);

		// Дополнительная проверка email пользователя
		const user = await page.evaluate(() => localStorage.getItem('lkfl_user'));
		const parsed = JSON.parse(user);
		expect(parsed.email).toBe('test.employee@sdek.local');
	});

	// Тест 3: Logout → Login page
	test('logout clears session and redirects to login', async ({ page }) => {
		// Сначала логинимся
		await loginViaKeycloak(page);

		// Логаут через UI
		await logoutViaUI(page);

		// Финальная проверка URL
		await expect(page).toHaveURL(/\/login/);
	});

	// Тест 4: Full cycle — Login → Dashboard → Logout → Login → Dashboard
	test('full cycle: login → dashboard → logout → login', async ({ page }) => {
		// Первый логин
		await loginViaKeycloak(page);
		await expect(page.getByTestId('app-header')).toBeVisible();

		// Логаут
		await logoutViaUI(page);
		await expect(page).toHaveURL(/\/login/);

		// Повторный логин
		await loginViaKeycloak(page);
		await expect(page.getByTestId('app-header')).toBeVisible();

		// Проверка что после повторного логина email верный
		const user = await page.evaluate(() => localStorage.getItem('lkfl_user'));
		expect(JSON.parse(user).email).toBe('test.employee@sdek.local');
	});
});
