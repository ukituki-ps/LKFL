const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  
  const navLogs = [];
  page.on('response', resp => {
    const url = resp.url();
    if (url.includes('logout')) {
      navLogs.push(`[RESP] ${resp.status()} ${url}`);
    }
  });
  
  try {
    await page.addInitScript(() => {
      localStorage.setItem('lkfl_user', JSON.stringify({
        id: '1', email: 'test@test.com', first_name: 'Алексей', last_name: 'Тестов'
      }));
      localStorage.setItem('lkfl_roles', JSON.stringify(['employee']));
    });
    
    await page.goto('http://localhost:5173/', { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(3000);
    
    console.log('=== Before logout ===');
    console.log('LS user:', await page.evaluate(() => localStorage.getItem('lkfl_user')));
    console.log('LS roles:', await page.evaluate(() => localStorage.getItem('lkfl_roles')));
    
    // Click avatar
    const avatar = await page.evaluateHandle(() => {
      const all = document.querySelectorAll('div');
      for (const el of all) {
        const s = el.style;
        if (s && s.borderRadius === '50%' && s.backgroundColor && s.backgroundColor.includes('#00B33C')) {
          return el;
        }
      }
      return null;
    });
    
    if (avatar.asElement()) {
      await avatar.click();
      await page.waitForTimeout(1000);
      
      // Click "Выйти" - intercept the navigation
      const navPromise = page.waitForNavigation({ timeout: 5000 }).catch(() => {});
      await page.locator('text=Выйти').first().click();
      
      // Check immediately after click
      await page.waitForTimeout(500);
      
      console.log('');
      console.log('=== 500ms after logout click ===');
      console.log('URL:', page.url());
      console.log('LS user:', await page.evaluate(() => localStorage.getItem('lkfl_user')));
      console.log('LS roles:', await page.evaluate(() => localStorage.getItem('lkfl_roles')));
      
      await navPromise;
      await page.waitForTimeout(2000);
      
      console.log('');
      console.log('=== After navigation ===');
      console.log('URL:', page.url());
      console.log('LS user:', await page.evaluate(() => localStorage.getItem('lkfl_user')));
      console.log('LS roles:', await page.evaluate(() => localStorage.getItem('lkfl_roles')));
      
      console.log('');
      console.log('=== Network logs ===');
      navLogs.forEach(l => console.log(l));
    }

  } catch (e) {
    console.error('Error:', e.message);
  } finally {
    await browser.close();
  }
})();
