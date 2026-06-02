// Этот файл заменён на e2e/integration/login-logout.spec.ts
// Старая реализация использовала page.route() interception — ломала PKCE flow.
// Новая реализация: чистая навигация через nginx, без перехвата.
import { test } from '@playwright/test';

test.skip('deprecated — use e2e/integration/login-logout.spec.ts', () => {});
