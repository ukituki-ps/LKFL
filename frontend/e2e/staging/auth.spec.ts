import { test, expect } from '@playwright/test';
import {
	loginThroughKeycloak,
	expectOnDashboard,
	getSessionCookie,
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

		// Шаг 6: Проверка cookie сессии + localStorage (user, roles)
		// После D2 token хранится в httpOnly cookie, не в localStorage.
		await test.step('Шаг 6: проверка cookie сессии + localStorage', async () => {
			// Cookie сессии lkfl_session установлена
			const cookies = await page.context().cookies();
			const sessionCookie = cookies.find((c) => c.name === 'lkfl_session');
			expect(sessionCookie).toBeTruthy();
			expect(sessionCookie!.value).toBeTruthy();

			// /api/v1/auth/me → 200 через page.request с cookie
			// page.request НЕ наследует cookies из browser context — передаём явно
			const baseUrl = (process.env.E2E_BASE_URL || 'https://dev.april.ukituki.tech').replace(/\/$/, '');
			const meResponse = await page.request.get(`${baseUrl}/api/v1/auth/me`, {
				headers: {
					Accept: 'application/json',
					Cookie: cookies.map((c) => `${c.name}=${c.value}`).join('; '),
				},
			});
			expect(meResponse.status()).toBe(200);

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
		// loginThroughKeycloak возвращает boolean (успех/неудача) после D2
		const loginSuccess = await loginThroughKeycloak(page);

		// Логин успешен (cookie lkfl_session установлена + мы на dashboard)
		expect(loginSuccess).toBe(true);

		// Cookie сессии существует
		const sessionCookie = await getSessionCookie(page);
		expect(sessionCookie).toBeTruthy();
		expect(sessionCookie!.value).toBeTruthy();

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

		// Проверяем cookie сессии до reload (cookie-based auth, D2)
		const cookieBefore = await getSessionCookie(page);
		expect(cookieBefore).toBeTruthy();

		// Hard refresh (bypass cache)
		await page.reload({ bypassCache: true });
		await page.waitForLoadState('networkidle', { timeout: 30_000 });

		// Cookie lkfl_session всё ещё существует (httpOnly cookie persist)
		const cookieAfter = await getSessionCookie(page);
		expect(cookieAfter).toBeTruthy();
		expect(cookieAfter!.value).toBe(cookieBefore!.value);

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
		// Создаём контекст с истёкшим cookie lkfl_session (cookie-based auth, D2).
		// localStorage user+roles оставляем — они корректно остаются в LS,
		// но без валидного cookie бэкенд ответит 401.
		const baseOrigin = (process.env.E2E_BASE_URL || 'https://dev.april.ukituki.tech')
			.replace('https://', '');
		const context = await browser.newContext({
			storageState: {
				origins: [{
					origin: process.env.E2E_BASE_URL || 'https://dev.april.ukituki.tech',
					localStorage: [
						{ name: 'lkfl_user', value: JSON.stringify({ id: 'test', email: 'test@test.com' }) },
						{ name: 'lkfl_roles', value: JSON.stringify(['employee']) },
					],
				}],
				cookies: [
					{
						name: 'lkfl_session',
						value: 'expired-token-value-that-is-invalid',
						domain: '.' + baseOrigin,
						path: '/',
						expires: -1, // истёкший
						httpOnly: true,
						secure: true,
					},
				],
			},
			ignoreHTTPSErrors: true,
		});
		const page = await context.newPage();

		// Переходим на dashboard — истёкший cookie должен вызвать redirect к Keycloak
		await page.goto('/', { waitUntil: 'domcontentloaded' });

		// Ждём что появится Keycloak login form (Expired cookie → backend → Keycloak)
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
		// После логина callback уже обработан — проверяем что cookie сессии
		// установлена и может использоваться для API запросов (cookie-based auth, D2)
		await loginThroughKeycloak(page);

		const sessionCookie = await getSessionCookie(page);
		expect(sessionCookie).toBeTruthy();

		// Проверяем что cookie сессии используется для API запросов
		// (users/me уже покрыт E2E-003) — здесь проверяем что cookie не пустой
		expect(sessionCookie!.value).toBeTruthy();
		expect(sessionCookie!.value.length).toBeGreaterThan(10); // cookie value present
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
