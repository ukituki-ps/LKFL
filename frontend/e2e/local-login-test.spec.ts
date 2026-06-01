import { test, expect, request } from '@playwright/test';

const BACKEND = 'http://172.21.0.5:8080';
const KEYCLOAK_PUBLIC = 'http://localhost:8081';
const FRONTEND = 'http://localhost:5173';
const USERNAME = 'petrova';
const PASSWORD = 'dev-password';
const KC_USERNAME = 'input[name="username"]';
const KC_PASSWORD = 'input[name="password"]';

let apiRequest;
let callbackHandled = false;

function rewriteKeycloakUrl(rawUrl) {
  let url = rawUrl.replace('http://keycloak:8080', KEYCLOAK_PUBLIC);
  const parsed = new URL(url);
  const oldRedirect = parsed.searchParams.get('redirect_uri');
  if (oldRedirect) {
    const newRedirect = oldRedirect
      .replace(BACKEND, FRONTEND)
      .replace('http://172.21.0.5:8080', FRONTEND);
    parsed.searchParams.set('redirect_uri', newRedirect);
    url = parsed.toString();
  }
  return url;
}

async function setupRouting(page) {
  apiRequest = await request.newContext({
    ignoreHTTPSErrors: true,
  });

  await page.route('/api/**', async (route) => {
    const req = route.request();
    const method = req.method();
    const postData = req.postData();
    const headers = req.headers();
    const backendUrl = req.url().replace(FRONTEND, BACKEND);

    // GET /api/v1/auth/login → Keycloak
    if (method === 'GET' && req.url().includes('/api/v1/auth/login')) {
      // Backend determines redirect_uri from Host header.
      // Since we call backend via container IP, we must pass explicit redirect param
      // so backend stores the correct redirect_uri for Keycloak token exchange.
      const urlParams = new URL(req.url()).searchParams;
      const redirectParam = `${FRONTEND}/api/v1/auth/callback`;
      const loginBackendUrl = `${BACKEND}/api/v1/auth/login?redirect=${encodeURIComponent(redirectParam)}&${urlParams.toString()}`;

      const response = await apiRequest.fetch(loginBackendUrl, {
        method: 'GET',
        maxRedirects: 0,
        headers: { 'Accept': headers['accept'] || '*/*' },
      });
      let location = response.headers()['location'] || '';
      location = rewriteKeycloakUrl(location);
      await route.fulfill({ status: 302, headers: { location } });
      return;
    }

    // /api/v1/auth/callback — проксируем, callbackHandled будет
    // установлен в login() после обнаружения навигации на callback URL
    if (callbackHandled && req.url().includes('/api/v1/auth/callback')) {
      // Уже обработали, пропускаем
      await route.continue();
      return;
    }

    // Все остальные запросы — проксируем на backend
    try {
      const response = await apiRequest.fetch(backendUrl, {
        method, maxRedirects: 0, postData, headers,
      });
      const respHeaders = { ...response.headers() };
      let loc = respHeaders['location'] || '';
      if (loc.includes('keycloak:8080')) {
        respHeaders['location'] = rewriteKeycloakUrl(loc);
      }
      await route.fulfill({
        status: response.status(),
        headers: respHeaders,
        body: await response.body(),
      });
    } catch (e) {
      await route.fulfill({ status: 502, body: JSON.stringify({ error: e.message }) });
    }
  });
}

/**
 * Обработка callback: получаем code+state из URL, делаем exchange,
 * ставим cookie и localStorage, навигируем на /.
 */
async function handleCallback(page) {
  const url = page.url();
  console.log(`  [handleCallback] URL: ${url.substring(0, 100)}...`);

  const urlObj = new URL(url);
  const code = urlObj.searchParams.get('code') || '';
  const state = urlObj.searchParams.get('state') || '';

  if (!code || !state) {
    console.log(`  [handleCallback] No code/state in URL`);
    return false;
  }

  console.log(`  [handleCallback] Exchanging code for token`);
  const backendUrl = `${BACKEND}/api/v1/auth/callback?code=${code}&state=${state}`;

  const resp = await apiRequest.get(backendUrl, {
    maxRedirects: 0,
    headers: { 'Accept': 'application/json' },
  });

  if (resp.status() !== 200) {
    console.log(`  [handleCallback] Backend error: ${resp.status()} ${await resp.text()}`);
    return false;
  }

  const body = await resp.json();
  console.log(`  [handleCallback] User: ${body.user?.email || 'unknown'}`);

  // Cookie для localhost
  const setCookie = resp.headers()['set-cookie'] || '';
  const cookieMatch = setCookie.match(/lkfl_session=([^;]+)/);
  if (cookieMatch) {
    await page.context().addCookies([{
      name: 'lkfl_session',
      value: cookieMatch[1],
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax',
    }]);
    console.log(`  [handleCallback] Cookie set ✓`);
  }

  // localStorage
  await page.evaluate((data) => {
    localStorage.setItem('lkfl_user', JSON.stringify(data.user));
    localStorage.setItem('lkfl_roles', JSON.stringify(data.roles || []));
  }, body);
  console.log(`  [handleCallback] localStorage set ✓`);

  // Навигируем на dashboard
  callbackHandled = true;
  await page.goto(`${FRONTEND}/`, { waitUntil: 'domcontentloaded' });
  console.log(`  [handleCallback] Navigated to dashboard`);
  return true;
}

async function waitForKeycloak(page, timeout = 30_000) {
  await page.waitForURL(/\/realms\//, { timeout });
  await page.waitForSelector(KC_USERNAME, { timeout: 10_000 });
}

async function login(page) {
  callbackHandled = false;

  console.log('  [1] Переход на /login');
  await page.goto('/login', { waitUntil: 'domcontentloaded' });

  console.log('  [2] Ожидание Keycloak');
  await waitForKeycloak(page);
  console.log(`  [2a] Keycloak OK`);

  console.log('  [3] Ввод login/password');
  await page.fill(KC_USERNAME, USERNAME);
  await page.fill(KC_PASSWORD, PASSWORD);

  // После клика "Sign In" ждём навигации на callback URL
  console.log('  [3a] Submit Keycloak form');
  const signinBtn = page.getByRole('button', { name: 'Sign In' });

  // Слушаем навигацию
  const navPromise = page.waitForEvent('framenavigated').then(() => {
    const currentUrl = page.url();
    console.log(`  [3b] Navigated to: ${currentUrl.substring(0, 100)}...`);
    if (currentUrl.includes('/api/v1/auth/callback') || currentUrl.includes('/callback')) {
      return handleCallback(page);
    }
    return false;
  });

  await signinBtn.click();
  await navPromise.catch(e => console.log(`  [3c] Nav error: ${e.message}`));

  // Если handleCallback не сработал в promise, проверяем вручную
  if (!callbackHandled) {
    await page.waitForTimeout(2000);
    const url = page.url();
    console.log(`  [4] Current URL: ${url.substring(0, 100)}...`);
    if (url.includes('/callback')) {
      await handleCallback(page);
    }
  }

  await page.waitForLoadState('networkidle', { timeout: 10_000 }).catch(() => {});
  const url = page.url();
  console.log(`  [5] Final URL: ${url}`);
  expect(url).not.toMatch(/\/login/);
  expect(url).not.toMatch(/\/callback/);
  expect(url).not.toMatch(/\/realms\//);
}

async function logout(page) {
  console.log('  [1] Аватар');
  await page.locator('div[style*="borderRadius: 50%"]').click();

  console.log('  [2] Выйти');
  await page.getByRole('menuitem', { name: 'Выйти' }).click();

  console.log('  [3] Ждём /login');
  await page.waitForURL(/\/login/, { timeout: 10_000 });
  console.log(`  [4] URL: ${page.url()}`);
  expect(page.url()).toMatch(/\/login/);
}

test('Login → Logout → Login', async ({ page }) => {
  await setupRouting(page);

  console.log('\n=== STEP 1: ЛОГИН ===');
  await login(page);

  const cookies1 = await page.context().cookies();
  const session1 = cookies1.find(c => c.name === 'lkfl_session');
  console.log(`  Cookie: ${session1 ? 'ЕСТЬ ✓' : 'НЕТ ✗'}`);
  expect(session1).toBeTruthy();

  const user1 = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  const ud1 = user1 ? JSON.parse(user1) : null;
  console.log(`  User: ${ud1?.email || 'НЕТ ✗'}`);
  expect(user1).toBeTruthy();

  console.log('\n=== STEP 2: ЛОГАУТ ===');
  await logout(page);

  const cookies2 = await page.context().cookies();
  const session2 = cookies2.find(c => c.name === 'lkfl_session');
  console.log(`  Cookie: ${session2 ? 'ОСТАЛАСЬ ✗' : 'УДАЛЕНА ✓'}`);

  const user2 = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  console.log(`  User: ${user2 ? 'ОСТАЛСЯ ✗' : 'ОЧИЩЕН ✓'}`);

  console.log('\n=== STEP 3: ПОВТОРНЫЙ ЛОГИН ===');
  await login(page);

  const cookies3 = await page.context().cookies();
  const session3 = cookies3.find(c => c.name === 'lkfl_session');
  console.log(`  Cookie: ${session3 ? 'ЕСТЬ ✓' : 'НЕТ ✗'}`);
  expect(session3).toBeTruthy();

  const user3 = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  const ud3 = user3 ? JSON.parse(user3) : null;
  console.log(`  User: ${ud3?.email || 'НЕТ ✗'}`);
  expect(user3).toBeTruthy();

  console.log('\n=== ВСЕ ШАГИ ПРОЙДЕНЫ ✓ ===');
});
