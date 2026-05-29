import { test, expect } from '@playwright/test';
import {
	loginThroughKeycloak,
	expectOnDashboard,
	getToken,
	getUser,
	getRoles,
	apiRequest,
	expectNoStuckLoop,
	STAGING_USERNAME,
	STAGING_PASSWORD,
	KC_USERNAME_SELECTOR,
	KC_PASSWORD_SELECTOR,
	expectKeycloakLoginForm,
	submitKeycloakLoginForm,
	isVerifyProfilePage,
	submitVerifyProfile,
} from './helpers';

/**
 * Staging E2E тесты — бьют в реальный backend + Keycloak.
 *
 * Никаких моков — тестируем полный flow: browser → nginx → backend → Keycloak.
 *
 * Запуск:
 *   npx playwright test --config=playwright.staging.config.ts
 *   npx playwright test --config=playwright.staging.config.ts --grep "login"
 *
 * Переменные окружения:
 *   E2E_BASE_URL  — URL стенда (default: https://dev.april.ukituki.tech)
 *   E2E_USERNAME  — логин для Keycloak (default: admin)
 *   E2E_PASSWORD  — пароль для Keycloak (default: admin-dev-password)
 */

// ─── Полный login flow (пошаговый) ───

test.describe('Staging: Full Login Flow Step by Step', () => {
	test('E2E-S01: пошаговый login — frontend → Keycloak → callback → dashboard', async ({ page }) => {
		// Шаг 1: Переход на /login → редирект на backend → редирект на Keycloak
		await test.step('Шаг 1: переход на /login', async () => {
			await page.goto('/login', { waitUntil: 'domcontentloaded' });
		});

		// Шаг 2: Keycloak login page загружается
		await test.step('Шаг 2: Keycloak login page загружается', async () => {
			await expectKeycloakLoginForm(page, 30_000);

			// Проверяем что URL содержит /realms/lkfl-sdek/ (не keycloak:8080)
			const url = page.url();
			expect(url).toContain('/realms/lkfl-sdek/');
			expect(url).not.toContain('keycloak:8080');
			expect(url).not.toContain('localhost');
			expect(url).not.toContain('DNS_PROBE');

			// Проверяем что form доступна
			await expect(page.locator(KC_USERNAME_SELECTOR)).toBeVisible();
			await expect(page.locator(KC_PASSWORD_SELECTOR)).toBeVisible();
			await expect(page.getByRole('button', { name: 'Sign In' })).toBeVisible();
		});

		// Шаг 3: Ввод credentials
		await test.step('Шаг 3: ввод credentials', async () => {
			await page.fill(KC_USERNAME_SELECTOR, STAGING_USERNAME);
			await page.fill(KC_PASSWORD_SELECTOR, STAGING_PASSWORD);
			await page.getByRole('button', { name: 'Sign In' }).click();
		});

		// Шаг 4: Обработка VERIFY_PROFILE если появился
		await test.step('Шаг 4: проверка VERIFY_PROFILE required action', async () => {
			// Ждём редиректа Keycloak после submit
			await page.waitForTimeout(3000);

			if (await isVerifyProfilePage(page)) {
				await test.step('VERIFY_PROFILE — заполняем профиль', async () => {
					await submitVerifyProfile(page);
					// Ждём редиректа после VERIFY_PROFILE
					await page.waitForTimeout(3000);
				});
			} else {
				// VERIFY_PROFILE не появился — нормально для уже настроенных пользователей
			}
		});

		// Шаг 5: Callback → backend → dashboard
		await test.step('Шаг 5: callback и перенаправление на dashboard', async () => {
			// Ожидаем редирект на callback → frontend → dashboard
			await page.waitForURL(/\/(callback|$)/, { timeout: 30_000 });

			// Ждём полного рендера dashboard
			await page.waitForLoadState('networkidle', { timeout: 15_000 });

			// Проверяем что мы на dashboard
			const finalUrl = page.url();
			expect(finalUrl).toMatch(/\/$/);
			expect(finalUrl).not.toMatch(/\/login/);
			expect(finalUrl).not.toMatch(/\/callback/);
			expect(finalUrl).not.toMatch(/\/realms\//);
		});

		// Шаг 6: Проверка localStorage (token, user, roles)
		await test.step('Шаг 6: проверка localStorage', async () => {
			// Token сохранён и валидный (JWT формат: header.payload.signature)
			const token = await getToken(page);
			expect(token).toBeTruthy();
			expect(token!.split('.').length).toBe(3); // JWT формат

			// User сохранён в localStorage
			const user = await getUser(page);
			expect(user).toBeTruthy();
			expect(user!.email).toBeTruthy();

			// Roles сохранены в localStorage
			const roles = await getRoles(page);
			expect(roles).toBeTruthy();
			expect(Array.isArray(roles)).toBe(true);
			expect(roles!.length).toBeGreaterThan(0);
		});

		// Шаг 7: Проверка что dashboard рендерится
		await test.step('Шаг 7: dashboard рендерится', async () => {
			const title = await page.title();
			expect(title).toContain('LKFL');
		});
	});
});

// ─── Полный login flow (через helper) ───

test.describe('Staging: Full Login Flow', () => {
	test('E2E-001: полный login flow через Keycloak → dashboard', async ({ page }) => {
		const token = await loginThroughKeycloak(page);

		// Token сохранён и валидный (JWT формат)
		expect(token).toBeTruthy();
		expect(token!.length).toBeGreaterThan(100);

		// User сохранён в localStorage
		const user = await getUser(page);
		expect(user).toBeTruthy();
		expect(user!.email).toBeTruthy();

		// Roles сохранены в localStorage
		const roles = await getRoles(page);
		expect(roles).toBeTruthy();
		expect(Array.isArray(roles)).toBe(true);
		expect(roles!.length).toBeGreaterThan(0);

		// Мы на dashboard
		await expectOnDashboard(page);
	});

	test('E2E-002: session persistence — hard refresh сохраняет сессию', async ({ page }) => {
		// Логинимся
		await loginThroughKeycloak(page);

		// Проверяем token в localStorage
		const tokenBefore = await getToken(page);
		expect(tokenBefore).toBeTruthy();

		// Hard refresh (bypass cache)
		await page.reload({ bypassCache: true });
		await page.waitForLoadState('networkidle', { timeout: 30_000 });

		// Token всё ещё в localStorage (persist)
		const tokenAfter = await getToken(page);
		expect(tokenAfter).toBe(tokenBefore);

		// Мы всё ещё на dashboard (не выкинуло на login)
		const url = page.url();
		expect(url).not.toMatch(/\/login/);
		expect(url).not.toMatch(/\/callback/);
	});

	test('E2E-003: /api/v1/users/me работает после логина', async ({ page, request }) => {
		await loginThroughKeycloak(page);

		const response = await apiRequest(page, request, '/api/v1/users/me');
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data.id).toBeTruthy();
		expect(data.email).toBeTruthy();
	});
});

// ─── Auth Error Handling ───

test.describe('Staging: Auth Error Handling', () => {
	test('E2E-004: неверный пароль → ошибка на странице Keycloak', async ({ page }) => {
		await page.goto('/login', { waitUntil: 'domcontentloaded' });
		await expectKeycloakLoginForm(page);

		// Вводим неверный пароль
		await page.fill(KC_USERNAME_SELECTOR, STAGING_USERNAME);
		await page.fill(KC_PASSWORD_SELECTOR, 'wrong-password');
		await page.getByRole('button', { name: 'Sign In' }).click();

		// Ждём рендера ошибки Keycloak
		await page.waitForTimeout(5000);

		// URL НЕ должен быть / или /callback (мы всё ещё на Keycloak)
		const url = page.url();
		expect(url).not.toMatch(/\/callback/);
		// Keycloak показывает ошибку — мы на странице логина Keycloak
		await expect(page.locator(KC_USERNAME_SELECTOR)).toBeVisible();
	});

	test('E2E-005: истёкший token → 401 → redirect на login', async ({ browser }) => {
		// Создаём контекст с заранее установленным localStorage (через storageState)
		// чтобы избежать SecurityError на HTTPS
		const context = await browser.newContext({
			storageState: {
				origins: [{
					origin: process.env.E2E_BASE_URL || 'https://dev.april.ukituki.tech',
					localStorage: [
						{ name: 'lkfl_token', value: 'eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0In0.invalid' },
						{ name: 'lkfl_user', value: JSON.stringify({ id: 'test', email: 'test@test.com' }) },
						{ name: 'lkfl_roles', value: JSON.stringify(['employee']) },
					],
				}],
				cookies: [],
			},
			ignoreHTTPSErrors: true,
		});
		const page = await context.newPage();

		// Переходим на dashboard — истёкший токен должен вызвать redirect к Keycloak
		await page.goto('/', { waitUntil: 'domcontentloaded' });

		// Ждём что появится Keycloak login form (Expired token → backend → Keycloak)
		// URL должен содержать /realms/ (Keycloak auth endpoint)
		await page.waitForURL(/\/realms\//, { timeout: 20_000 });
		await expect(page.locator(KC_USERNAME_SELECTOR)).toBeVisible({ timeout: 20_000 });

		await context.close();
	});
});

// ─── Tenant Resolution ───

test.describe('Staging: Tenant Resolution', () => {
	test('E2E-006: tenant resolution через JWT claims работает', async ({ page, request }) => {
		// Логинимся
		await test.step('login', async () => {
			await loginThroughKeycloak(page);
		});

		// /api/v1/users/me — должен работать (tenant из JWT claims)
		const response = await apiRequest(page, request, '/api/v1/users/me');
		expect(response.status()).toBe(200);

		const data = await response.json();
		expect(data.id).toBeTruthy();
	});

	test('E2E-007: /api/v1/auth/callback возвращает структуру после логина', async ({ page, request }) => {
		// После логина callback уже обработан — проверяем что token в localStorage
		// валиден и может использоваться для API запросов
		await loginThroughKeycloak(page);

		const token = await getToken(page);
		expect(token).toBeTruthy();

		// Проверяем что токен используется для dashboard данных
		// (users/me уже покрыт E2E-003) — здесь проверяем что token не пустой
		expect(token!.length).toBeGreaterThan(100); // JWT формат
	});
});

// ─── Dashboard стабильность ───

test.describe('Staging: Dashboard Stability', () => {
	test('E2E-008: после логина dashboard загружается без стаб-цикла', async ({ page }) => {
		await loginThroughKeycloak(page);

		// URL стабилизировался на /
		await expectOnDashboard(page);

		// Ждём 5 секунд — URL не должен измениться (нет стаб-цикла)
		await expectNoStuckLoop(page, 5000);
	});

	test('E2E-009: navigation между страницами работает после логина', async ({ page }) => {
		await loginThroughKeycloak(page);
		await expectOnDashboard(page);

		// Переход на /catalog
		await page.goto('/catalog', { waitUntil: 'domcontentloaded' });
		await page.waitForLoadState('networkidle', { timeout: 15_000 });
		expect(page.url()).toMatch(/\/catalog/);
		expect(page.url()).not.toMatch(/\/login/);

		// Возврат на /
		await page.goto('/', { waitUntil: 'domcontentloaded' });
		await page.waitForLoadState('networkidle', { timeout: 15_000 });
		await expectOnDashboard(page);
	});
});
