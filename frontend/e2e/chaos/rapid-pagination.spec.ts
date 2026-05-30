import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
	PAGINATION_PAGES,
} from './helpers';

/**
 * Хаос-тест: быстрое переключение страниц пагинации.
 *
 * 10 тестов с разными seed. Каждый тест быстро переключает страницы
 * 1→10→1→5→3 для проверки race conditions.
 */
test.describe('Chaos: Rapid Pagination', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 3000 + i * 444;

		test(`rapid pagination seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			const numSwitches = randomInt(rng, 8, 15);

			for (let j = 0; j < numSwitches; j++) {
				const targetPage = randomFrom(rng, PAGINATION_PAGES);

				try {
					// Пробуем кликнуть на номер страницы
					const pageButton = page.getByRole('button', { name: String(targetPage) }).first();

					if (await pageButton.isVisible().catch(() => false)) {
						await pageButton.click({ timeout: 3000 }).catch(() => {});
					} else {
						// Если конкретная страница не найдена, пробуем "Далее" или "Назад"
						const nextBtn = page.getByRole('button', { name: 'Далее' }).first();
						const prevBtn = page.getByRole('button', { name: 'Назад' }).first();

						if (j % 2 === 0 && await nextBtn.isVisible().catch(() => false)) {
							await nextBtn.click({ timeout: 3000 }).catch(() => {});
						} else if (await prevBtn.isVisible().catch(() => false)) {
							await prevBtn.click({ timeout: 3000 }).catch(() => {});
						}
					}
				} catch {
					// Пагинация может не отображаться
				}

				// Очень короткая задержка — имитация быстрого переключения
				await page.waitForTimeout(randomInt(rng, 30, 100));

				await expectNoCrash(page);
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});
