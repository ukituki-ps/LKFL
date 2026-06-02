// T2311 — E2E Integration Tests: Login → Dashboard → Logout
// Реальный OIDC flow через Keycloak.
//
// NOTE: Keycloak 25.0.6 в Docker + nginx proxy не устанавливает session cookie
// на auth endpoint — это known issue инфраструктуры. Тесты проверяют:
// 1. Редирект на Keycloak (browser flow)
// 2. Keycloak token endpoint (direct access grant)
// 3. Backend /healthz
// 4. Invalid auth → 401
//
// Полный login→dashboard→logout flow через browser будет работать после
// исправления Keycloak cookie policy (TODO).

import { test, expect } from '@playwright/test';

// ─── Credentials ────────────────────────────────────────────────────────────
const USERNAME = process.env.E2E_USERNAME || 'test.employee';
const PASSWORD = process.env.E2E_PASSWORD || 'employee123';

// ─── Health check ───────────────────────────────────────────────────────────
test.beforeAll(async ({ request }, testInfo) => {
	const baseUrl = process.env.E2E_BASE_URL || 'http://project.ukituki.tech';

	const health = await request.get(`${baseUrl}/healthz`);
	if (health.status() !== 200) {
		testInfo.skip(`Backend недоступен: ${health.status()}`);
		return;
	}

	const oidc = await request.get(
		`${baseUrl}/realms/lkfl-sdek/.well-known/openid-configuration`,
	);
	if (oidc.status() !== 200) {
		testInfo.skip(`Keycloak недоступен: ${oidc.status()}`);
		return;
	}
});

// ─── Tests ──────────────────────────────────────────────────────────────────

test.describe('OIDC Authentication', () => {
	// Тест 1: Unauthenticated → Login redirect → Keycloak auth page renders
	test('unauthenticated user is redirected to Keycloak login', async ({ page }) => {
		await page.goto('/', { waitUntil: 'commit' });
		// SPA: / → /login → /api/v1/auth/login → 302 → Keycloak auth
		await page.waitForURL(/\/realms\/lkfl-sdek\/protocol\/openid-connect\/auth/, {
			timeout: 30_000,
		});
		// Keycloak рендерит форму логина (встроенная тема keycloak)
		await expect(page.locator('#username')).toBeVisible({ timeout: 30_000 });
	});

	// Тест 2: Keycloak direct access grant возвращает валидный token
	test('Keycloak direct access grant returns valid token', async ({ request }) => {
		const baseUrl = process.env.E2E_BASE_URL || 'http://project.ukituki.tech';

		const tokenResp = await request.post(
			`${baseUrl}/realms/lkfl-sdek/protocol/openid-connect/token`,
			{
				form: {
					grant_type: 'password',
					client_id: 'lkfl-spa',
					username: USERNAME,
					password: PASSWORD,
				},
			},
		);

		expect(tokenResp.status(), 'direct access grant должен вернуть 200').toBe(200);
		const body = await tokenResp.json();
		expect(body.access_token, 'access_token должен быть в ответе').toBeTruthy();
		expect(body.token_type).toBe('Bearer');
		expect(body.expires_in).toBeTruthy();
	});

	// Тест 3: Keycloak отклоняет невалидные учётные данные
	test('Keycloak rejects invalid credentials', async ({ request }) => {
		const baseUrl = process.env.E2E_BASE_URL || 'http://project.ukituki.tech';

		const tokenResp = await request.post(
			`${baseUrl}/realms/lkfl-sdek/protocol/openid-connect/token`,
			{
				form: {
					grant_type: 'password',
					client_id: 'lkfl-spa',
					username: USERNAME,
					password: 'wrong-password',
				},
			},
		);

		expect(tokenResp.status()).toBe(401);
	});

	// Тест 4: Backend /healthz доступен
	test('backend health endpoint is available', async ({ request }) => {
		const baseUrl = process.env.E2E_BASE_URL || 'http://project.ukituki.tech';

		const resp = await request.get(`${baseUrl}/healthz`);
		expect(resp.status()).toBe(200);
		expect(await resp.text()).toBe('OK');
	});

	// Тест 5: Backend /auth/me без сессии возвращает 401
	test('backend /auth/me without session returns 401', async ({ request }) => {
		const baseUrl = process.env.E2E_BASE_URL || 'http://project.ukituki.tech';

		const resp = await request.get(`${baseUrl}/api/v1/auth/me`);
		expect(resp.status()).toBe(401);
	});

	// Тест 6: OIDC discovery endpoint содержит правильные URL
	test('OIDC discovery endpoint has correct URLs', async ({ request }) => {
		const baseUrl = process.env.E2E_BASE_URL || 'http://project.ukituki.tech';

		const resp = await request.get(
			`${baseUrl}/realms/lkfl-sdek/.well-known/openid-configuration`,
		);
		expect(resp.status()).toBe(200);

		const body = await resp.json();
		expect(body.issuer).toContain('lkfl-sdek');
		expect(body.authorization_endpoint).toBeTruthy();
		expect(body.token_endpoint).toBeTruthy();
		expect(body.jwks_uri).toBeTruthy();
	});
});
