import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
	APP_PAGES,
} from './helpers';

/**
 * Хаос-тест: случайная навигация между страницами.
 *
 * 10 тестов с разными seed. Каждый тест выполняет 20-40 случайных переходов
 * между доступными страницами и проверяет отсутствие крашей.
 */
test.describe('Chaos: Random Navigation', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 42 + i * 1000;

		test(`random navigation seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page);

			const numNavigations = randomInt(rng, 20, 40);

			for (let j = 0; j < numNavigations; j++) {
				const targetPage = randomFrom(rng, APP_PAGES);

				// Игнорируем ошибки навигации (например, если страница ушла)
				try {
					await page.goto(targetPage, {
						waitUntil: 'domcontentloaded',
						timeout: 10000,
					}).catch(() => {
						// Навигация могла прерваться — это нормально для хаос-теста
					});
				} catch {
					// Игнорируем ошибки навигации
				}

				// Небольшая задержка между переходами
				await page.waitForTimeout(randomInt(rng, 100, 300));

				// Проверяем отсутствие краша
				await expectNoCrash(page);
			}

			// Финальная проверка: нет unhandled page errors
			// Console errors от моков/переходов допустимы
			expect(pageErrors).toHaveLength(0);
		});
	}
});
