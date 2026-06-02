# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: login-flow.spec.js >> Login Flow (browser-based logout) >> login → logout → login again
- Location: login-flow.spec.js:26:3

# Error details

```
Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/
Call log:
  - navigating to "http://localhost:5173/", waiting until "domcontentloaded"

```

# Test source

```ts
  1   | // E2E test: Login → Logout → Login (browser-based SSO invalidation)
  2   | //
  3   | // Run: npx playwright test login-flow.spec.js --config=playwright.config.js
  4   | 
  5   | const { test, expect } = require('@playwright/test');
  6   | 
  7   | async function waitForDashboard(page) {
  8   |   await page.waitForURL('**://localhost:5173/callback**', { timeout: 30000 });
  9   |   await page.waitForURL('http://localhost:5173/', { timeout: 15000 });
  10  |   await page.waitForTimeout(2000);
  11  | }
  12  | 
  13  | async function findAvatar(page) {
  14  |   for (const sel of [
  15  |     'div[style*="brand-green"][style*="pointer"]',
  16  |     'div[style*="#00B33C"][style*="pointer"]',
  17  |     'div[style*="border-radius: 50%"]',
  18  |   ]) {
  19  |     const el = page.locator(sel).first();
  20  |     try { if (await el.isVisible({ timeout: 2000 })) return el; } catch (e) {}
  21  |   }
  22  |   return null;
  23  | }
  24  | 
  25  | test.describe('Login Flow (browser-based logout)', () => {
  26  |   test('login → logout → login again', async ({ page }) => {
  27  |     // ═══════════════════════════════════════════════════════════
  28  |     // FIRST LOGIN
  29  |     // ═══════════════════════════════════════════════════════════
  30  |     console.log('=== FIRST LOGIN ===');
  31  | 
  32  |     console.log('Step 1: Navigate to app → Keycloak');
> 33  |     await page.goto('http://localhost:5173/', { waitUntil: 'domcontentloaded', timeout: 10000 });
      |                ^ Error: page.goto: net::ERR_CONNECTION_REFUSED at http://localhost:5173/
  34  |     await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/auth**', { timeout: 15000 });
  35  | 
  36  |     console.log('Step 2: Login with petrova/dev-password');
  37  |     // Wait for Keycloak v2 theme JS-rendered form
  38  |     await page.waitForLoadState('networkidle');
  39  |     await page.locator('input[name="username"]').first().fill('petrova');
  40  |     await page.locator('input[name="password"]').first().fill('dev-password');
  41  |     await page.locator('button[type="submit"]').click();
  42  | 
  43  |     console.log('Step 3: Wait for callback → dashboard');
  44  |     await waitForDashboard(page);
  45  |     console.log('  Dashboard URL:', page.url());
  46  | 
  47  |     const user = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_user') || 'null'));
  48  |     const roles = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_roles') || 'null'));
  49  |     expect(user?.email).toBe('petrova@sdek.local');
  50  |     expect(roles).toContain('employee');
  51  |     console.log('  User:', user?.email, 'Roles:', roles, '✅');
  52  | 
  53  |     // ═══════════════════════════════════════════════════════════
  54  |     // LOGOUT — browser-based SSO invalidation
  55  |     // ═══════════════════════════════════════════════════════════
  56  |     console.log('=== LOGOUT ===');
  57  | 
  58  |     console.log('Step 4: Click avatar → Выйти');
  59  |     const avatar = await findAvatar(page);
  60  |     expect(avatar).not.toBeNull();
  61  |     await avatar.click();
  62  |     await page.waitForTimeout(500);
  63  |     await page.locator('text=Выйти').click();
  64  | 
  65  |     // Flow: window.location.href → backend /logout → 302 Keycloak logout
  66  |     //
  67  |     // After T2309 (hybrid logout) the backend invalidates SSO server-side first.
  68  |     // Keycloak may either:
  69  |     //   A) Show the "Logging out" confirmation page (#kc-logout button visible)
  70  |     //   B) Redirect immediately (SSO already dead in DB, no confirmation needed)
  71  |     // The test must handle both cases.
  72  |     console.log('Step 5: Wait for Keycloak logout page');
  73  |     await page.waitForURL('**://://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/logout**', {
  74  |       timeout: 15000,
  75  |     });
  76  |     console.log('  Keycloak logout page:', page.url());
  77  | 
  78  |     // Click confirmation button ONLY if visible (case A).
  79  |     // If not visible within 3s, SSO was already invalidated server-side (case B)
  80  |     // and Keycloak is redirecting automatically.
  81  |     console.log('Step 6: Handle Keycloak confirmation page (if any)');
  82  |     const kcLogoutBtn = page.locator('#kc-logout');
  83  |     try {
  84  |       await kcLogoutBtn.waitFor({ state: 'visible', timeout: 3000 });
  85  |       await kcLogoutBtn.click();
  86  |       console.log('  Keycloak confirmation page clicked (case A) ✅');
  87  |     } catch {
  88  |       // Button not visible — SSO already invalidated server-side,
  89  |       // Keycloak redirects automatically without confirmation (case B)
  90  |       console.log('  No confirmation page — SSO invalidated server-side (case B) ✅');
  91  |     }
  92  | 
  93  |     // Keycloak redirects to post_logout_redirect_uri → http://localhost:5173/login
  94  |     await page.waitForURL('**://localhost:5173/**', { timeout: 15000 });
  95  |     await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
  96  |     await page.waitForTimeout(1000);
  97  |     console.log('  After logout URL:', page.url());
  98  | 
  99  |     // Verify auth state is cleared
  100 |     const hasUser = await page.evaluate(() => !!localStorage.getItem('lkfl_user'));
  101 |     const hasRoles = await page.evaluate(() => !!localStorage.getItem('lkfl_roles'));
  102 |     const kcCookie = (await page.context().cookies()).find(c => c.name === 'KAUTH_SESSION_ID');
  103 |     console.log('  localStorage cleared:', !hasUser && !hasRoles, '✅');
  104 |     console.log('  KAUTH_SESSION_ID cookie:', kcCookie ? 'PRESENT ❌' : 'ABSENT ✅');
  105 |     expect(hasUser).toBe(false);
  106 |     expect(hasRoles).toBe(false);
  107 |     expect(kcCookie).toBeUndefined();
  108 | 
  109 |     // ═══════════════════════════════════════════════════════════
  110 |     // SECOND LOGIN — SSO must be invalidated
  111 |     // ═══════════════════════════════════════════════════════════
  112 |     console.log('=== SECOND LOGIN ===');
  113 | 
  114 |     console.log('Step 7: Trigger login');
  115 |     await page.goto(
  116 |       'http://localhost:8082/api/v1/auth/login?post_redirect=%2F&retry=0',
  117 |       { waitUntil: 'domcontentloaded', timeout: 10000 },
  118 |     );
  119 |     await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/auth**', {
  120 |       timeout: 15000,
  121 |     });
  122 | 
  123 |     // CRITICAL: Keycloak must show login form (SSO invalidated)
  124 |     const hasUsername = await page.locator('#username').isVisible({ timeout: 10000 });
  125 |     expect(hasUsername).toBe(true);
  126 |     console.log('  Keycloak shows login form (SSO invalidated) ✅');
  127 | 
  128 |     console.log('Step 8: Login again');
  129 |     await page.locator('#username').fill('petrova');
  130 |     await page.locator('#password').fill('dev-password');
  131 |     await page.locator('#kc-login').click();
  132 | 
  133 |     console.log('Step 9: Wait for dashboard');
```