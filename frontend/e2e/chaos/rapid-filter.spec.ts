import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
	CATALOG_FILTERS,
} from './helpers';

/**
 * Хаос-тест: быстрая смена фильтров каталога.
 *
 * 10 тестов с разными seed. Каждый тест быстро переключает фильтры
 * (тип, статус, категория) 5-10 раз подряд.
 */
test.describe('Chaos: Rapid Filter Changes', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 500 + i * 777;

		test(`rapid filter changes seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			const numChanges = randomInt(rng, 5, 10);

			for (let j = 0; j < numChanges; j++) {
				const filterType = randomFrom(rng, ['type', 'status', 'category'] as const);

				try {
					switch (filterType) {
						case 'type': {
							const select = page.getByLabel('Тип').first();
							if (await select.isVisible().catch(() => false)) {
								await select.click({ timeout: 3000 }).catch(() => {});
								const option = randomFrom(rng, CATALOG_FILTERS.types);
								await page.getByRole('option', { name: option, exact: false }).first()
									.click({ timeout: 2000 }).catch(() => {});
							}
							break;
						}
						case 'status': {
							const select = page.getByLabel('Статус').first();
							if (await select.isVisible().catch(() => false)) {
								await select.click({ timeout: 3000 }).catch(() => {});
								const option = randomFrom(rng, CATALOG_FILTERS.statuses);
								await page.getByRole('option', { name: option, exact: false }).first()
									.click({ timeout: 2000 }).catch(() => {});
							}
							break;
						}
						case 'category': {
							const select = page.getByLabel('Категория').first();
							if (await select.isVisible().catch(() => false)) {
								await select.click({ timeout: 3000 }).catch(() => {});
								const option = randomFrom(rng, CATALOG_FILTERS.categories);
								await page.getByRole('option', { name: option, exact: false }).first()
									.click({ timeout: 2000 }).catch(() => {});
							}
							break;
						}
					}
				} catch {
					// Фильтр мог не найтись или быть недоступным
				}

				// Очень короткая задержка — имитация быстрого переключения
				await page.waitForTimeout(randomInt(rng, 50, 150));

				await expectNoCrash(page);
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});
