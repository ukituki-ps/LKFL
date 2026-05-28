import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
	VIEWPORT_SIZES,
	APP_PAGES,
} from './helpers';

/**
 * Хаос-тест: переключение viewport в процессе сессии.
 *
 * 10 тестов с разными seed. Каждый тест переключает viewport между
 * mobile/tablet/desktop размерами во время навигации.
 */
test.describe('Chaos: Viewport Changes', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 11000 + i * 222;

		test(`viewport chaos seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			const numChanges = randomInt(rng, 8, 15);

			for (let j = 0; j < numChanges; j++) {
				const viewport = randomFrom(rng, VIEWPORT_SIZES);

				// Меняем viewport
				await page.setViewportSize({
					width: viewport.width,
					height: viewport.height,
				});

				// После изменения viewport навигируем на случайную страницу
				const targetPage = randomFrom(rng, APP_PAGES);
				try {
					await page.goto(targetPage, {
						waitUntil: 'domcontentloaded',
						timeout: 10000,
					}).catch(() => {});
				} catch {
					// Игнорируем
				}

				await page.waitForTimeout(randomInt(rng, 100, 300));

				// Скроллим страницу для проверки рендера
				try {
					await page.evaluate(() => {
						window.scrollBy(0, Math.random() * 500);
					});
				} catch {
					// Игнорируем
				}

				await expectNoCrash(page);
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});
