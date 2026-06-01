# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: login-flow.spec.js >> Login Flow (browser-based logout) >> login → logout → login again
- Location: login-flow.spec.js:26:3

# Error details

```
TimeoutError: page.waitForURL: Timeout 15000ms exceeded.
=========================== logs ===========================
waiting for navigation to "**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/logout**" until "load"
  navigated to "http://localhost:5173/callback?code=ecd06f39-065d-468b-bb84-671a09142107.4b370588-2475-4422-90d6-3dd549cfd5ee.03b6b882-4982-4480-ada8-501816566a08&state=058e5fcd933c943edf001b9b9a5158b173c71d38c27a20dc0a448e6dfe293d1c"
  navigated to "http://localhost:5173/"
  navigated to "http://localhost:5173/"
============================================================
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - generic [ref=e6]:
      - generic [ref=e7]:
        - button [ref=e8] [cursor=pointer]:
          - img [ref=e9]
        - navigation [ref=e11]:
          - link "Главная" [ref=e12] [cursor=pointer]:
            - /url: /
          - link "Каталог льгот" [ref=e13] [cursor=pointer]:
            - /url: /catalog
          - link "Мои баллы" [ref=e14] [cursor=pointer]:
            - /url: /points
          - link "Документы" [ref=e15] [cursor=pointer]:
            - /url: /documents
          - link "Поддержка" [ref=e16] [cursor=pointer]:
            - /url: /support
      - generic [ref=e19]:
        - generic [ref=e20]:
          - img [ref=e21]
          - paragraph [ref=e26]: 1 250 моих баллов
        - button [ref=e27] [cursor=pointer]:
          - img [ref=e29]
        - button "🌙" [ref=e32] [cursor=pointer]:
          - generic [ref=e34]: 🌙
        - generic [ref=e35] [cursor=pointer]: МП
  - main [ref=e36]:
    - generic [ref=e37]:
      - generic [ref=e38]:
        - generic [ref=e39]:
          - heading "Добрый день, Алексей!" [level=1] [ref=e40]
          - paragraph [ref=e41]: понедельник, 1 июня 2026 г.
        - generic [ref=e43]: Пакет «Стандарт»
      - generic [ref=e44]:
        - generic [ref=e45]:
          - paragraph [ref=e47]: Баланс баллов
          - generic [ref=e48]:
            - img [ref=e50]
            - paragraph [ref=e57]: 1 250
          - paragraph [ref=e58]: +500 баллов в июне
        - generic [ref=e59]:
          - paragraph [ref=e61]: Активные льготы
          - generic [ref=e62]:
            - img [ref=e64]
            - paragraph [ref=e69]: "3"
          - paragraph [ref=e70]: Из 5 доступных
        - generic [ref=e71]:
          - paragraph [ref=e73]: До конца периода
          - generic [ref=e74]:
            - img [ref=e76]
            - generic [ref=e79]:
              - paragraph [ref=e80]: "47"
              - paragraph [ref=e81]: дн
          - paragraph [ref=e82]: "Период: янв — июн 2025"
      - generic [ref=e83]:
        - generic [ref=e84]:
          - generic [ref=e85]:
            - generic [ref=e86]:
              - generic [ref=e87]:
                - img [ref=e88]
                - paragraph [ref=e91]: Активные льготы
              - generic [ref=e92] [cursor=pointer]: Весь каталог →
            - generic [ref=e93]:
              - generic [ref=e94] [cursor=pointer]:
                - img [ref=e96]
                - generic [ref=e98]:
                  - paragraph [ref=e99]: Онлайн-кинотеатр
                  - paragraph [ref=e100]: KION · до 31.12.2025
                - generic [ref=e102]: Активна
              - generic [ref=e103] [cursor=pointer]:
                - img [ref=e105]
                - generic [ref=e111]:
                  - paragraph [ref=e112]: Фитнес-клуб
                  - paragraph [ref=e113]: World Class · до 31.12.2025
                - generic [ref=e115]: Активна
              - generic [ref=e116] [cursor=pointer]:
                - img [ref=e118]
                - generic [ref=e120]:
                  - paragraph [ref=e121]: Страховка ДМС
                  - paragraph [ref=e122]: СОГАЗ · до 31.12.2025
                - generic [ref=e124]: Ожидает
          - generic [ref=e125]:
            - generic [ref=e127]:
              - img [ref=e128]
              - paragraph [ref=e131]: Последние события
            - generic [ref=e132]:
              - generic [ref=e133]:
                - img [ref=e135]
                - generic [ref=e138]:
                  - paragraph [ref=e139]: "Новая льгота: онлайн-кинотеатр"
                  - paragraph [ref=e140]: Сегодня, 14:30
              - generic [ref=e141]:
                - img [ref=e143]
                - generic [ref=e148]:
                  - paragraph [ref=e149]: Начислено 500 баллов за опрос
                  - paragraph [ref=e150]: Вчера, 18:15
              - generic [ref=e151]:
                - img [ref=e153]
                - generic [ref=e155]:
                  - paragraph [ref=e156]: Обновлены условия программы
                  - paragraph [ref=e157]: 20 мая, 10:00
        - generic [ref=e158]:
          - generic [ref=e160]:
            - img [ref=e161]
            - paragraph [ref=e163]: Быстрые действия
          - generic [ref=e164]:
            - generic [ref=e165] [cursor=pointer]:
              - img [ref=e167]
              - paragraph [ref=e170]: Добавить родственника к ДМС
            - generic [ref=e171] [cursor=pointer]:
              - img [ref=e173]
              - paragraph [ref=e176]: Апгрейд ДМС
            - generic [ref=e177] [cursor=pointer]:
              - img [ref=e179]
              - paragraph [ref=e182]: Купить мерч СДЭК
            - generic [ref=e183] [cursor=pointer]:
              - img [ref=e185]
              - paragraph [ref=e195]: Записаться к психологу
            - generic [ref=e196] [cursor=pointer]:
              - img [ref=e198]
              - paragraph [ref=e201]: Заявка на мат. капитал от компании
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
  33  |     await page.goto('http://localhost:5173/', { waitUntil: 'domcontentloaded', timeout: 10000 });
  34  |     await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/auth**', { timeout: 15000 });
  35  | 
  36  |     console.log('Step 2: Login with petrova/dev-password');
  37  |     await page.locator('#username').fill('petrova');
  38  |     await page.locator('#password').fill('dev-password');
  39  |     await page.locator('#kc-login').click();
  40  | 
  41  |     console.log('Step 3: Wait for callback → dashboard');
  42  |     await waitForDashboard(page);
  43  |     console.log('  Dashboard URL:', page.url());
  44  | 
  45  |     const user = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_user') || 'null'));
  46  |     const roles = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_roles') || 'null'));
  47  |     expect(user?.email).toBe('petrova@sdek.local');
  48  |     expect(roles).toContain('employee');
  49  |     console.log('  User:', user?.email, 'Roles:', roles, '✅');
  50  | 
  51  |     // ═══════════════════════════════════════════════════════════
  52  |     // LOGOUT — browser-based SSO invalidation
  53  |     // ═══════════════════════════════════════════════════════════
  54  |     console.log('=== LOGOUT ===');
  55  | 
  56  |     console.log('Step 4: Click avatar → Выйти');
  57  |     const avatar = await findAvatar(page);
  58  |     expect(avatar).not.toBeNull();
  59  |     await avatar.click();
  60  |     await page.waitForTimeout(500);
  61  |     await page.locator('text=Выйти').click();
  62  | 
  63  |     // Flow: window.location.href → backend /logout → 302 Keycloak logout
  64  |     //
  65  |     // After T2309 (hybrid logout) the backend invalidates SSO server-side first.
  66  |     // Keycloak may either:
  67  |     //   A) Show the "Logging out" confirmation page (#kc-logout button visible)
  68  |     //   B) Redirect immediately (SSO already dead in DB, no confirmation needed)
  69  |     // The test must handle both cases.
  70  |     console.log('Step 5: Wait for Keycloak logout page');
> 71  |     await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/logout**', {
      |                ^ TimeoutError: page.waitForURL: Timeout 15000ms exceeded.
  72  |       timeout: 15000,
  73  |     });
  74  |     console.log('  Keycloak logout page:', page.url());
  75  | 
  76  |     // Click confirmation button ONLY if visible (case A).
  77  |     // If not visible within 3s, SSO was already invalidated server-side (case B)
  78  |     // and Keycloak is redirecting automatically.
  79  |     console.log('Step 6: Handle Keycloak confirmation page (if any)');
  80  |     const kcLogoutBtn = page.locator('#kc-logout');
  81  |     try {
  82  |       await kcLogoutBtn.waitFor({ state: 'visible', timeout: 3000 });
  83  |       await kcLogoutBtn.click();
  84  |       console.log('  Keycloak confirmation page clicked (case A) ✅');
  85  |     } catch {
  86  |       // Button not visible — SSO already invalidated server-side,
  87  |       // Keycloak redirects automatically without confirmation (case B)
  88  |       console.log('  No confirmation page — SSO invalidated server-side (case B) ✅');
  89  |     }
  90  | 
  91  |     // Keycloak redirects to post_logout_redirect_uri → http://localhost:5173/login
  92  |     await page.waitForURL('**://localhost:5173/**', { timeout: 15000 });
  93  |     await page.waitForLoadState('domcontentloaded', { timeout: 10000 });
  94  |     await page.waitForTimeout(1000);
  95  |     console.log('  After logout URL:', page.url());
  96  | 
  97  |     // Verify auth state is cleared
  98  |     const hasUser = await page.evaluate(() => !!localStorage.getItem('lkfl_user'));
  99  |     const hasRoles = await page.evaluate(() => !!localStorage.getItem('lkfl_roles'));
  100 |     const kcCookie = (await page.context().cookies()).find(c => c.name === 'KAUTH_SESSION_ID');
  101 |     console.log('  localStorage cleared:', !hasUser && !hasRoles, '✅');
  102 |     console.log('  KAUTH_SESSION_ID cookie:', kcCookie ? 'PRESENT ❌' : 'ABSENT ✅');
  103 |     expect(hasUser).toBe(false);
  104 |     expect(hasRoles).toBe(false);
  105 |     expect(kcCookie).toBeUndefined();
  106 | 
  107 |     // ═══════════════════════════════════════════════════════════
  108 |     // SECOND LOGIN — SSO must be invalidated
  109 |     // ═══════════════════════════════════════════════════════════
  110 |     console.log('=== SECOND LOGIN ===');
  111 | 
  112 |     console.log('Step 7: Trigger login');
  113 |     await page.goto(
  114 |       'http://localhost:8082/api/v1/auth/login?post_redirect=%2F&retry=0',
  115 |       { waitUntil: 'domcontentloaded', timeout: 10000 },
  116 |     );
  117 |     await page.waitForURL('**://localhost:8081/realms/lkfl-sdek/protocol/openid-connect/auth**', {
  118 |       timeout: 15000,
  119 |     });
  120 | 
  121 |     // CRITICAL: Keycloak must show login form (SSO invalidated)
  122 |     const hasUsername = await page.locator('#username').isVisible({ timeout: 10000 });
  123 |     expect(hasUsername).toBe(true);
  124 |     console.log('  Keycloak shows login form (SSO invalidated) ✅');
  125 | 
  126 |     console.log('Step 8: Login again');
  127 |     await page.locator('#username').fill('petrova');
  128 |     await page.locator('#password').fill('dev-password');
  129 |     await page.locator('#kc-login').click();
  130 | 
  131 |     console.log('Step 9: Wait for dashboard');
  132 |     await waitForDashboard(page);
  133 |     console.log('  Dashboard URL:', page.url());
  134 | 
  135 |     const user2 = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_user') || 'null'));
  136 |     const roles2 = await page.evaluate(() => JSON.parse(localStorage.getItem('lkfl_roles') || 'null'));
  137 |     expect(user2?.email).toBe('petrova@sdek.local');
  138 |     expect(roles2).toContain('employee');
  139 | 
  140 |     const avatar2 = await findAvatar(page);
  141 |     expect(!!avatar2).toBe(true);
  142 |     console.log('  Avatar visible ✅');
  143 | 
  144 |     console.log('═══════════════════════════════════════════════════════════');
  145 |     console.log('ALL PASSED: Login → Logout (SSO invalidated) → Login ✅');
  146 |     console.log('Final URL:', page.url());
  147 |   });
  148 | });
  149 | 
```