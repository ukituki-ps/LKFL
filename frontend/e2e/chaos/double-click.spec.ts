import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
} from './helpers';

/**
 * Хаос-тест: двойные и тройные клики на кнопках и ссылках.
 *
 * 10 тестов с разными seed. Каждый тест выполняет двойные и тройные клики
 * по кликабельным элементам для проверки повторных отправок и race conditions.
 */
test.describe('Chaos: Double/Triple Clicks', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 5000 + i * 666;

		test(`double/triple click chaos seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			// Собираем кликабельные элементы
			const clickableSelector = 'button, a, [role="button"]';
			const clickables = await page.$$(clickableSelector);

			const numClicks = randomInt(rng, 5, Math.max(5, Math.floor(clickables.length / 2)));

			for (let j = 0; j < numClicks; j++) {
				if (clickables.length === 0) break;

				const element = clickables[randomInt(rng, 0, clickables.length - 1)];
				const clickType = randomFrom(rng, ['double', 'triple', 'rapid'] as const);

				try {
					const isVisible = await element.isVisible().catch(() => false);
					if (!isVisible) continue;

					switch (clickType) {
						case 'double':
							// Двойной клик
							await element.click({ timeout: 3000 }).catch(() => {});
							await page.waitForTimeout(50);
							await element.click({ timeout: 3000 }).catch(() => {});
							break;
						case 'triple':
							// Тройной клик
							await element.click({ timeout: 3000 }).catch(() => {});
							await page.waitForTimeout(30);
							await element.click({ timeout: 3000 }).catch(() => {});
							await page.waitForTimeout(30);
							await element.click({ timeout: 3000 }).catch(() => {});
							break;
						case 'rapid':
							// Очень быстрая серия кликов
							for (let k = 0; k < 5; k++) {
								await element.click({ timeout: 2000, force: true }).catch(() => {});
								await page.waitForTimeout(10);
							}
							break;
					}
				} catch {
					// Элемент мог стать невалидным
					continue;
				}

				await page.waitForTimeout(randomInt(rng, 100, 300));
				await expectNoCrash(page);
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});
