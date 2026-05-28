import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
} from './helpers';

/**
 * Хаос-тест: случайные клики по кнопкам и кликабельным элементам.
 *
 * 10 тестов с разными seed. Каждый тест находит все кликабельные элементы
 * на странице и кликает по ним в случайном порядке.
 */
test.describe('Chaos: Random Button Clicks', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 200 + i * 333;

		test(`random button clicks seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			// Ждём загрузки каталога
			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			// Собираем все кликабельные элементы
			const clickableSelector = 'button, a, [role="button"], [role="link"], input[type="submit"], input[type="button"]';
			const clickables = await page.$$(clickableSelector);

			if (clickables.length === 0) {
				// Если нет кликабельных элементов, просто проверяем отсутствие краша
				await expectNoCrash(page);
				expect(pageErrors).toHaveLength(0);
				return;
			}

			// Перемешиваем порядок кликов
			const indices = clickables.map((_, idx) => idx);
			for (let idx = indices.length - 1; idx > 0; idx--) {
				const j = randomInt(rng, 0, idx);
				[indices[idx], indices[j]] = [indices[j], indices[idx]];
			}

			for (const idx of indices) {
				const element = clickables[idx];

				try {
					// Проверяем видимость перед кликом
					const isVisible = await element.isVisible().catch(() => false);
					if (!isVisible) continue;

					await element.click({ timeout: 3000 }).catch(() => {
						// Клик мог вызвать навигацию или ошибку — это нормально
					});
				} catch {
					// Элемент мог стать невалидным после предыдущего клика
					continue;
				}

				// Небольшая задержка
				await page.waitForTimeout(randomInt(rng, 50, 200));

				// Проверяем отсутствие краша
				await expectNoCrash(page);
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});
