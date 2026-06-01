// Real UI logout test — clicks "Выйти" button, not manual clearCookies
import { chromium } from 'playwright';
import { mkdirSync, rmSync } from 'fs';

const SS = '/tmp/kilo/logout-test';
rmSync(SS, { recursive: true, force: true });
mkdirSync(SS, { recursive: true });
let sn = 0;
async function ss(page, name) { sn++; await page.screenshot({ path: `${SS}/${sn}_${name}.png`, fullPage: true }); }

(async () => {
  console.log('=== Real UI Logout Test ===\n');

  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  const logs = [];
  page.on('console', m => { if (m.type() === 'error') logs.push(m.text()); });

  // ─── 1. Login ───
  console.log('1. Login...');
  await page.goto('http://localhost:5173/', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForSelector('input[name="username"]', { timeout: 20000 });
  await page.fill('input[name="username"]', 'ivanov');
  await page.fill('input[name="password"]', 'dev-password');
  await page.click('input[type="submit"]');
  await page.waitForURL(url => url.pathname === '/', { timeout: 15000 });
  await ss(page, 'after-login');

  // Check auth
  const meBefore = await page.evaluate(async () => {
    const r = await fetch('/api/v1/auth/me', { credentials: 'include' });
    return r.status;
  });
  console.log(`   Before logout: /me=${meBefore}, URL=${page.url()}`);

  // ─── 2. Click user avatar ───
  console.log('\n2. Click user avatar...');
  // The avatar is a div with border-radius: 50%, green bg, white text
  const avatar = await page.$('div[style*="border-radius"][style*="50%"]') ||
                  await page.$('div[style*="borderRadius"][style*="50"]') ||
                  await page.$('div[style*="#00B33C"]');
  if (!avatar) {
    console.log('   Avatar not found!');
    await ss(page, 'no-avatar');
  } else {
    await avatar.click();
    await page.waitForTimeout(800);
    await ss(page, 'menu-open');
    console.log('   Menu opened');
  }

  // ─── 3. Click "Выйти" ───
  console.log('\n3. Click Выйти...');
  const exitBtn = await page.$(':text("Выйти")') ||
                   await page.$('.mantine-Menu-item') ||
                   await page.evaluateHandle(() => {
                     const items = document.querySelectorAll('[role="menuitem"], [class*="Menu-item"]');
                     for (const el of items) {
                       if (el.textContent.includes('Выйти')) return el;
                     }
                     return null;
                   }).asElement();

  if (exitBtn) {
    await exitBtn.click();
    console.log('   Clicked');
  } else {
    console.log('   Выйти button not found');
    await ss(page, 'no-exit-btn');
  }

  // Wait for navigation
  await page.waitForLoadState('networkidle', { timeout: 10000 }).catch(() => {});
  await page.waitForTimeout(3000);
  await ss(page, 'after-logout');

  // ─── 4. Results ───
  console.log('\n=== RESULTS ===');
  const urlAfter = page.url();
  console.log(`URL: ${urlAfter}`);

  const meAfter = await page.evaluate(async () => {
    try {
      const r = await fetch('/api/v1/auth/me', { credentials: 'include' });
      return { status: r.status, body: await r.text() };
    } catch(e) { return { error: e.message }; }
  });
  console.log(`/api/v1/auth/me: ${JSON.stringify(meAfter)}`);

  const bodyText = await page.evaluate(() => document.body.innerText.substring(0, 300));
  console.log(`Page text: "${bodyText.substring(0, 150)}..."`);

  const onDashboard = bodyText.includes('Каталог льгот') || bodyText.includes('балл');
  const onLogin = bodyText.includes('Войти') || bodyText.includes('Вход') || bodyText.includes('Вышли');
  const onKeycloak = urlAfter.includes('8081') || urlAfter.includes('keycloak');

  console.log(`\nOn dashboard: ${onDashboard}`);
  console.log(`On login: ${onLogin}`);
  console.log(`On Keycloak: ${onKeycloak}`);

  // Check Zustand store
  const zustandState = await page.evaluate(() => {
    // Try to access Zustand store state from the window
    // Zustand stores are typically not directly accessible, but let's try
    try {
      const els = document.querySelectorAll('[data-testid*="user"], [class*="User"]');
      return els.length > 0 ? 'user elements found' : 'no user elements';
    } catch { return 'error'; }
  });
  console.log(`DOM user elements: ${zustandState}`);

  // Check cookies
  const cookies = await page.context().cookies();
  const session = cookies.find(c => c.name === 'lkfl_session');
  console.log(`Session cookie: ${session ? 'PRESENT' : 'GONE'}`);

  // Check localStorage
  const lsUser = await page.evaluate(() => localStorage.getItem('lkfl_user'));
  console.log(`localStorage lkfl_user: ${lsUser ? 'PRESENT' : 'GONE'}`);

  // Check sessionStorage
  const ssLoggedOut = await page.evaluate(() => sessionStorage.getItem('lkfl_just_logged_out'));
  console.log(`sessionStorage lkfl_just_logged_out: ${ssLoggedOut}`);

  console.log(`\nConsole errors: ${logs.length}`);
  logs.slice(0, 5).forEach(e => console.log(`  - ${e.substring(0, 120)}`));

  console.log(`\nScreenshots: ${SS}/`);

  await browser.close();
})();
