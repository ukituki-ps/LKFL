# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: local-login-test.spec.ts >> Login → Logout → Login
- Location: e2e/local-login-test.spec.ts:219:1

# Error details

```
Error: expect(received).not.toMatch(expected)

Expected pattern: not /\/callback/
Received string:      "http://localhost:5173/api/v1/auth/callback?state=4d77f8f6cb6be58b8b3d6b7b9500a69d5078a88c8ca3b5a386b937783caa046a&session_state=fb351171-87f4-4ff9-acb1-c4c931f246a7&iss=http%3A%2F%2Flocalhost%3A8081%2Frealms%2Flkfl-sdek&code=4e0deca0-19f0-473a-beec-08135f864d6c.fb351171-87f4-4ff9-acb1-c4c931f246a7.03b6b882-4982-4480-ada8-501816566a08"
```

# Test source

```ts
  102 |     console.log(`  [handleCallback] No code/state in URL`);
  103 |     return false;
  104 |   }
  105 | 
  106 |   console.log(`  [handleCallback] Exchanging code for token`);
  107 |   const backendUrl = `${BACKEND}/api/v1/auth/callback?code=${code}&state=${state}`;
  108 | 
  109 |   const resp = await apiRequest.get(backendUrl, {
  110 |     maxRedirects: 0,
  111 |     headers: { 'Accept': 'application/json' },
  112 |   });
  113 | 
  114 |   if (resp.status() !== 200) {
  115 |     console.log(`  [handleCallback] Backend error: ${resp.status()} ${await resp.text()}`);
  116 |     return false;
  117 |   }
  118 | 
  119 |   const body = await resp.json();
  120 |   console.log(`  [handleCallback] User: ${body.user?.email || 'unknown'}`);
  121 | 
  122 |   // Cookie для localhost
  123 |   const setCookie = resp.headers()['set-cookie'] || '';
  124 |   const cookieMatch = setCookie.match(/lkfl_session=([^;]+)/);
  125 |   if (cookieMatch) {
  126 |     await page.context().addCookies([{
  127 |       name: 'lkfl_session',
  128 |       value: cookieMatch[1],
  129 |       domain: 'localhost',
  130 |       path: '/',
  131 |       httpOnly: true,
  132 |       secure: false,
  133 |       sameSite: 'Lax',
  134 |     }]);
  135 |     console.log(`  [handleCallback] Cookie set ✓`);
  136 |   }
  137 | 
  138 |   // localStorage
  139 |   await page.evaluate((data) => {
  140 |     localStorage.setItem('lkfl_user', JSON.stringify(data.user));
  141 |     localStorage.setItem('lkfl_roles', JSON.stringify(data.roles || []));
  142 |   }, body);
  143 |   console.log(`  [handleCallback] localStorage set ✓`);
  144 | 
  145 |   // Навигируем на dashboard
  146 |   callbackHandled = true;
  147 |   await page.goto(`${FRONTEND}/`, { waitUntil: 'domcontentloaded' });
  148 |   console.log(`  [handleCallback] Navigated to dashboard`);
  149 |   return true;
  150 | }
  151 | 
  152 | async function waitForKeycloak(page, timeout = 30_000) {
  153 |   await page.waitForURL(/\/realms\//, { timeout });
  154 |   await page.waitForSelector(KC_USERNAME, { timeout: 10_000 });
  155 | }
  156 | 
  157 | async function login(page) {
  158 |   callbackHandled = false;
  159 | 
  160 |   console.log('  [1] Переход на /login');
  161 |   await page.goto('/login', { waitUntil: 'domcontentloaded' });
  162 | 
  163 |   console.log('  [2] Ожидание Keycloak');
  164 |   await waitForKeycloak(page);
  165 |   console.log(`  [2a] Keycloak OK`);
  166 | 
  167 |   console.log('  [3] Ввод login/password');
  168 |   await page.fill(KC_USERNAME, USERNAME);
  169 |   await page.fill(KC_PASSWORD, PASSWORD);
  170 | 
  171 |   // После клика "Sign In" ждём навигации на callback URL
  172 |   console.log('  [3a] Submit Keycloak form');
  173 |   const signinBtn = page.getByRole('button', { name: 'Sign In' });
  174 | 
  175 |   // Слушаем навигацию
  176 |   const navPromise = page.waitForEvent('framenavigated').then(() => {
  177 |     const currentUrl = page.url();
  178 |     console.log(`  [3b] Navigated to: ${currentUrl.substring(0, 100)}...`);
  179 |     if (currentUrl.includes('/api/v1/auth/callback') || currentUrl.includes('/callback')) {
  180 |       return handleCallback(page);
  181 |     }
  182 |     return false;
  183 |   });
  184 | 
  185 |   await signinBtn.click();
  186 |   await navPromise.catch(e => console.log(`  [3c] Nav error: ${e.message}`));
  187 | 
  188 |   // Если handleCallback не сработал в promise, проверяем вручную
  189 |   if (!callbackHandled) {
  190 |     await page.waitForTimeout(2000);
  191 |     const url = page.url();
  192 |     console.log(`  [4] Current URL: ${url.substring(0, 100)}...`);
  193 |     if (url.includes('/callback')) {
  194 |       await handleCallback(page);
  195 |     }
  196 |   }
  197 | 
  198 |   await page.waitForLoadState('networkidle', { timeout: 10_000 }).catch(() => {});
  199 |   const url = page.url();
  200 |   console.log(`  [5] Final URL: ${url}`);
  201 |   expect(url).not.toMatch(/\/login/);
> 202 |   expect(url).not.toMatch(/\/callback/);
      |                   ^ Error: expect(received).not.toMatch(expected)
  203 |   expect(url).not.toMatch(/\/realms\//);
  204 | }
  205 | 
  206 | async function logout(page) {
  207 |   console.log('  [1] Аватар');
  208 |   await page.locator('div[style*="borderRadius: 50%"]').click();
  209 | 
  210 |   console.log('  [2] Выйти');
  211 |   await page.getByRole('menuitem', { name: 'Выйти' }).click();
  212 | 
  213 |   console.log('  [3] Ждём /login');
  214 |   await page.waitForURL(/\/login/, { timeout: 10_000 });
  215 |   console.log(`  [4] URL: ${page.url()}`);
  216 |   expect(page.url()).toMatch(/\/login/);
  217 | }
  218 | 
  219 | test('Login → Logout → Login', async ({ page }) => {
  220 |   await setupRouting(page);
  221 | 
  222 |   console.log('\n=== STEP 1: ЛОГИН ===');
  223 |   await login(page);
  224 | 
  225 |   const cookies1 = await page.context().cookies();
  226 |   const session1 = cookies1.find(c => c.name === 'lkfl_session');
  227 |   console.log(`  Cookie: ${session1 ? 'ЕСТЬ ✓' : 'НЕТ ✗'}`);
  228 |   expect(session1).toBeTruthy();
  229 | 
  230 |   const user1 = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  231 |   const ud1 = user1 ? JSON.parse(user1) : null;
  232 |   console.log(`  User: ${ud1?.email || 'НЕТ ✗'}`);
  233 |   expect(user1).toBeTruthy();
  234 | 
  235 |   console.log('\n=== STEP 2: ЛОГАУТ ===');
  236 |   await logout(page);
  237 | 
  238 |   const cookies2 = await page.context().cookies();
  239 |   const session2 = cookies2.find(c => c.name === 'lkfl_session');
  240 |   console.log(`  Cookie: ${session2 ? 'ОСТАЛАСЬ ✗' : 'УДАЛЕНА ✓'}`);
  241 | 
  242 |   const user2 = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  243 |   console.log(`  User: ${user2 ? 'ОСТАЛСЯ ✗' : 'ОЧИЩЕН ✓'}`);
  244 | 
  245 |   console.log('\n=== STEP 3: ПОВТОРНЫЙ ЛОГИН ===');
  246 |   await login(page);
  247 | 
  248 |   const cookies3 = await page.context().cookies();
  249 |   const session3 = cookies3.find(c => c.name === 'lkfl_session');
  250 |   console.log(`  Cookie: ${session3 ? 'ЕСТЬ ✓' : 'НЕТ ✗'}`);
  251 |   expect(session3).toBeTruthy();
  252 | 
  253 |   const user3 = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  254 |   const ud3 = user3 ? JSON.parse(user3) : null;
  255 |   console.log(`  User: ${ud3?.email || 'НЕТ ✗'}`);
  256 |   expect(user3).toBeTruthy();
  257 | 
  258 |   console.log('\n=== ВСЕ ШАГИ ПРОЙДЕНЫ ✓ ===');
  259 | });
  260 | 
```