// E2E test: Login → Logout → Login (browser-based SSO invalidation)
//
// Run: npx playwright test login-flow.spec.js --config=playwright.config.js

const { test, expect } = require('@playwright/test');

async function waitForDashboard(page) {
  await page.waitForURL('**://localhost:5173/callback**', { timeout: 30000 });
  await page.waitForURL('http://localhost:5173/', { timeout: 15000 });
  await page.waitForTimeout(2000);
}

async function findAvatar(page) {
  for (const sel of [
    'div[style*="brand-green"][style*="pointer"]',
    'div[style*="#00B33C"][style*="pointer"]',
    'div[style*="border-radius: 50%"]',
  ]) {
    const el = page.locator(sel).first();
    try { if (await el.isVisible({ timeout: 2000 })) return el; } catch (e) {}
  }
  return null;
}

test.describe('Login Flow (browser-based logout)', () => {
  test('login → logout → login again', async ({ page }) => {
    // ═══════════════════════════════════════════════════════════
    // FIRST LOGIN
    // ═══════════════════════════════════════════════════════════
    console.log('=== FIRST LOGIN ===');

    console.log('Step 1: Navigate to app → Keycloak');
    await page.goto('http://localhost:5173/', { waitUntil: 'domcontentloaded', timeout: 10000 });
    await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/auth**', { timeout: 15000 });

    console.log('Step 2: Login with petrova/dev-password');
    await page.locator('#username').fill('petrova');
    await page.locator('#password').fill('dev-password');
    await page.locator('#kc-login').click();

    console.log('Step 3: Wait for callback → dashboard');
    await waitForDashboard(page);
    console.log('  Dashboard URL:', page.url());

    const user = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_user') || 'null'));
    const roles = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_roles') || 'null'));
    expect(user?.email).toBe('petrova@sdek.local');
    expect(roles).toContain('employee');
    console.log('  User:', user?.email, 'Roles:', roles, '✅');

    // ═══════════════════════════════════════════════════════════
    // LOGOUT — browser-based SSO invalidation
    // ═══════════════════════════════════════════════════════════
    console.log('=== LOGOUT ===');

    console.log('Step 4: Click avatar → Выйти');
    const avatar = await findAvatar(page);
    expect(avatar).not.toBeNull();
    await avatar.click();
    await page.waitForTimeout(500);
    await page.locator('text=Выйти').click();

    // Flow: window.location.href → backend /logout → 302 Keycloak logout
    // Keycloak shows "Logging out" confirmation page
    console.log('Step 5: Wait for Keycloak logout confirmation page');
    await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/logout**', {
      timeout: 15000,
    });
    console.log('  Keycloak logout page:', page.url());

    // Click the "Log out" confirmation button on Keycloak page
    console.log('Step 6: Click Keycloak "Log out" confirmation button');
    await page.locator('#kc-logout').waitFor({ state: 'visible', timeout: 10000 });
    await page.locator('#kc-logout').click();
    console.log('  Keycloak logout confirmed ✅');

    // Keycloak redirects to post_logout_redirect_uri → http://localhost:5173/login
    await page.waitForURL('**://localhost:5173/**', { timeout: 15000 });
    await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
    await page.waitForTimeout(1000);
    console.log('  After logout URL:', page.url());

    // Verify auth state is cleared
    const hasUser = await page.evaluate(() => !!localStorage.getItem('lkfl_user'));
    const hasRoles = await page.evaluate(() => !!localStorage.getItem('lkfl_roles'));
    const kcCookie = (await page.context().cookies()).find(c => c.name === 'KAUTH_SESSION_ID');
    console.log('  localStorage cleared:', !hasUser && !hasRoles, '✅');
    console.log('  KAUTH_SESSION_ID cookie:', kcCookie ? 'PRESENT ❌' : 'ABSENT ✅');
    expect(hasUser).toBe(false);
    expect(hasRoles).toBe(false);
    expect(kcCookie).toBeUndefined();

    // ═══════════════════════════════════════════════════════════
    // SECOND LOGIN — SSO must be invalidated
    // ═══════════════════════════════════════════════════════════
    console.log('=== SECOND LOGIN ===');

    console.log('Step 7: Trigger login');
    await page.goto(
      'http://localhost:8082/api/v1/auth/login?post_redirect=%2F&retry=0',
      { waitUntil: 'domcontentloaded', timeout: 10000 },
    );
    await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/auth**', {
      timeout: 15000,
    });

    // CRITICAL: Keycloak must show login form (SSO invalidated)
    const hasUsername = await page.locator('#username').isVisible({ timeout: 10000 });
    expect(hasUsername).toBe(true);
    console.log('  Keycloak shows login form (SSO invalidated) ✅');

    console.log('Step 8: Login again');
    await page.locator('#username').fill('petrova');
    await page.locator('#password').fill('dev-password');
    await page.locator('#kc-login').click();

    console.log('Step 9: Wait for dashboard');
    await waitForDashboard(page);
    console.log('  Dashboard URL:', page.url());

    const user2 = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_user') || 'null'));
    const roles2 = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_roles') || 'null'));
    expect(user2?.email).toBe('petrova@sdek.local');
    expect(roles2).toContain('employee');

    const avatar2 = await findAvatar(page);
    expect(!!avatar2).toBe(true);
    console.log('  Avatar visible ✅');

    console.log('═══════════════════════════════════════════════════════════');
    console.log('ALL PASSED: Login → Logout (SSO invalidated) → Login ✅');
    console.log('Final URL:', page.url());
  });
});
