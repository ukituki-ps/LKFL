const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  
  // Capture console & network logs
  const logs = [];
  page.on('console', msg => logs.push(`[CONSOLE ${msg.type()}] ${msg.text()}`));
  
  const navLogs = [];
  page.on('response', resp => {
    const url = resp.url();
    if (url.includes('logout') || url.includes('auth')) {
      navLogs.push(`[RESP] ${resp.status()} ${url}`);
    }
  });
  
  page.on('request', req => {
    const url = req.url();
    if (url.includes('logout') || url.includes('auth')) {
      navLogs.push(`[REQ] ${req.method()} ${url}`);
    }
  });
  
  try {
    // Set up auth state BEFORE navigating
    await page.addInitScript(() => {
      localStorage.setItem('lkfl_user', JSON.stringify({
        id: '1',
        email: 'test@test.com',
        first_name: 'Алексей',
        last_name: 'Тестов'
      }));
      localStorage.setItem('lkfl_roles', JSON.stringify(['employee']));
    });
    
    await page.goto('http://localhost:5173/', { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(3000);
    
    console.log('=== Before logout ===');
    console.log('URL:', page.url());
    console.log('Title:', await page.title());
    
    // Find the green circle avatar and click it
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
      console.log('Found avatar, clicking...');
      await avatar.click();
      await page.waitForTimeout(1000);
      
      // Check if menu is open
      const menuItems = await page.locator('body').innerText();
      console.log('Menu visible text (last 200 chars):', menuItems.slice(-200));
      
      // Find and click "Выйти"
      const exitButton = await page.locator('text=Выйти').first();
      const exitCount = await exitButton.count();
      console.log('Exit button count:', exitCount);
      
      if (exitCount > 0) {
        console.log('Clicking "Выйти"...');
        await exitButton.click();
        await page.waitForTimeout(3000);
        
        console.log('');
        console.log('=== After logout ===');
        console.log('URL:', page.url());
        console.log('Title:', await page.title());
        
        const bodyText = await page.locator('body').innerText();
        console.log('Body text (first 200 chars):', bodyText.slice(0, 200));
        
        // Check localStorage
        const lsUser = await page.evaluate(() => localStorage.getItem('lkfl_user'));
        const lsRoles = await page.evaluate(() => localStorage.getItem('lkfl_roles'));
        console.log('localStorage user:', lsUser);
        console.log('localStorage roles:', lsRoles);
        
      } else {
        console.log('ERROR: "Выйти" button not found');
        await page.screenshot({ path: '/tmp/kilo/logout-menu.png', fullPage: false });
      }
    } else {
      console.log('ERROR: Avatar not found');
    }

  } catch (e) {
    console.error('Error:', e.message);
  } finally {
    console.log('');
    console.log('=== Logs ===');
    logs.forEach(l => console.log(l));
    navLogs.forEach(l => console.log(l));
    await browser.close();
  }
})();
