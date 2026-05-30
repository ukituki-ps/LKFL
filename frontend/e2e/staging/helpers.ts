import { expect, type Page } from '@playwright/test';

/**
 * Хелперы для staging E2E тестов (реальный backend + Keycloak).
 *
 * В отличие от `e2e/helpers.ts` — никаких моков.
 */

// ─── Конфигурация ───

export const STAGING_USERNAME = process.env.E2E_USERNAME || 'admin';
export const STAGING_PASSWORD = process.env.E2E_PASSWORD || 'admin-dev-password';

// ─── localStorage ключи (соответствуют authStore.ts) ───

export const LS_TOKEN = 'lkfl_token';
export const LS_USER = 'lkfl_user';
export const LS_ROLES = 'lkfl_roles';

// ─── Keycloak селекторы ───

/** Keycloak login form — username поле */
export const KC_USERNAME_SELECTOR = 'input[name="username"]';
/** Keycloak login form — password поле */
export const KC_PASSWORD_SELECTOR = 'input[name="password"]';
/** Keycloak login form — форма (стандартный id Keycloak) */
export const KC_FORM_SELECTOR = 'form#kc-form-login';
/** Keycloak login form — submit кнопка (текст "Sign In" в default theme) */
export const KC_SUBMIT_SELECTOR = 'text=Sign In';
/** Keycloak VERIFY_PROFILE required action — firstName поле */
export const KC_FIRSTNAME_SELECTOR = 'input[name="firstName"]';
/** Keycloak VERIFY_PROFILE required action — lastName поле */
export const KC_LASTNAME_SELECTOR = 'input[name="lastName"]';
/** Keycloak VERIFY_PROFILE required action — email поле */
export const KC_EMAIL_SELECTOR = 'input[name="email"]';

// ─── Хелперы ───

/**
 * Ожидать появления Keycloak login формы.
 * Keycloak может загрузиться с задержкой (DNS, TLS, rendering).
 */
export async function expectKeycloakLoginForm(page: Page, timeout = 20_000) {
	await page.waitForSelector(KC_USERNAME_SELECTOR, { timeout });
}

/**
 * Проверить что мы на странице VERIFY_PROFILE required action.
 * Возвращает true если найдено поле firstName (уникально для VERIFY_PROFILE).
 */
export async function isVerifyProfilePage(page: Page): Promise<boolean> {
	try {
		const el = page.locator(KC_FIRSTNAME_SELECTOR);
		await el.waitFor({ state: 'visible', timeout: 2000 });
		return true;
	} catch {
		return false;
	}
}

/**
 * Заполнить и отправить форму VERIFY_PROFILE.
 * Появляется при первом входе пользователя в Keycloak.
 */
export async function submitVerifyProfile(page: Page) {
	// Заполняем profile поля (если есть firstName/lastName)
	const firstNameVisible = await page.locator(KC_FIRSTNAME_SELECTOR).isVisible({ timeout: 2000 }).catch(() => false);
	const lastNameVisible = await page.locator(KC_LASTNAME_SELECTOR).isVisible({ timeout: 2000 }).catch(() => false);
	const emailVisible = await page.locator(KC_EMAIL_SELECTOR).isVisible({ timeout: 2000 }).catch(() => false);

	if (firstNameVisible) {
		await page.fill(KC_FIRSTNAME_SELECTOR, 'Admin');
	}
	if (lastNameVisible) {
		await page.fill(KC_LASTNAME_SELECTOR, 'LKFL');
	}
	if (emailVisible) {
		await page.fill(KC_EMAIL_SELECTOR, 'admin@lkfl.dev');
	}

	// Submit — кнопка "Submit" или "Done" или "Continue"
	await page.getByRole('button', { name: 'Submit' }).click();
}

/**
 * Заполнить и отправить Keycloak login форму.
 * После отправки обрабатывает VERIFY_PROFILE required action если он появился.
 */
export async function submitKeycloakLoginForm(
	page: Page,
	username = STAGING_USERNAME,
	password = STAGING_PASSWORD,
) {
	await page.fill(KC_USERNAME_SELECTOR, username);
	await page.fill(KC_PASSWORD_SELECTOR, password);
	// Клик по кнопке "Sign In" — стандартный Keycloak default theme
	await page.getByRole('button', { name: 'Sign In' }).click();

	// Ждём редиректа Keycloak (после submit формы логина)
	await page.waitForTimeout(2000);

	// Проверяем не появился ли VERIFY_PROFILE required action
	if (await isVerifyProfilePage(page)) {
		await submitVerifyProfile(page);
		// Ждём редиректа после VERIFY_PROFILE
		await page.waitForTimeout(2000);
	}
}

/**
 * Полный login flow: переход на /login → Keycloak → callback → dashboard.
 * Обрабатывает VERIFY_PROFILE required action если он появляется.
 * @returns token из localStorage после логина
 */
export async function loginThroughKeycloak(page: Page): Promise<string> {
	// Шаг 1: Переход на /login → редирект на backend → редирект на Keycloak
	await page.goto('/login', { waitUntil: 'domcontentloaded' });

	// Шаг 2: Ждём Keycloak login form
	await expectKeycloakLoginForm(page);

	// Шаг 3: Заполняем и отправляем форму (обрабатывает VERIFY_PROFILE)
	await submitKeycloakLoginForm(page);

	// Шаг 4: Ожидаем редирект на callback → backend → dashboard
	// Callback может быть коротким, поэтому ждём стабилизации URL
	await page.waitForURL(/\/(callback|$)/, { timeout: 30_000 });

	// Шаг 5: Ждём полного рендера dashboard (появление контента)
	await page.waitForLoadState('networkidle', { timeout: 15_000 });

	// Шаг 6: Проверяем что token сохранён в localStorage
	const token = await page.evaluate((key) => localStorage.getItem(key), LS_TOKEN);
	return token || '';
}

/**
 * Проверить что пользователь на dashboard (не на login, не на callback).
 */
export async function expectOnDashboard(page: Page) {
	const url = page.url();
	expect(url).toMatch(/\/$/);
	expect(url).not.toMatch(/\/login/);
	expect(url).not.toMatch(/\/callback/);
}

/**
 * Получить token из localStorage.
 */
export async function getToken(page: Page): Promise<string | null> {
	return page.evaluate((key) => localStorage.getItem(key), LS_TOKEN);
}

/**
 * Получить user из localStorage.
 */
export async function getUser(page: Page): Promise<Record<string, unknown> | null> {
	return page.evaluate((key) => {
		const raw = localStorage.getItem(key);
		return raw ? JSON.parse(raw) : null;
	}, LS_USER);
}

/**
 * Получить roles из localStorage.
 */
export async function getRoles(page: Page): Promise<string[] | null> {
	return page.evaluate((key) => {
		const raw = localStorage.getItem(key);
		return raw ? JSON.parse(raw) : null;
	}, LS_ROLES);
}

/**
 * Сделать API запрос с токеном из localStorage.
 */
export async function apiRequest(
	page: Page,
	request: import('@playwright/test').APIRequestContext,
	path: string,
): Promise<import('@playwright/test').APIResponse> {
	const token = await getToken(page);
	expect(token).toBeTruthy();

	return request.get(path, {
		headers: {
			Authorization: `Bearer ${token}`,
			Accept: 'application/json',
		},
	});
}

/**
 * Проверить что сессия стабильна (нет стаб-цикла).
 * Ждёт N секунд и проверяет что URL не изменился на /login или /callback.
 */
export async function expectNoStuckLoop(page: Page, waitMs = 5000) {
	const urlBefore = page.url();

	await page.waitForTimeout(waitMs);

	const urlAfter = page.url();
	expect(urlAfter).not.toMatch(/\/login/);
	expect(urlAfter).not.toMatch(/\/callback/);
	// URL должен остаться тем же (dashboard)
	expect(urlAfter).toBe(urlBefore);
}
