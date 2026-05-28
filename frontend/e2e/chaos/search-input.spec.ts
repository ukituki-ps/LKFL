import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
	CHAOS_INPUTS,
} from './helpers';

/**
 * Хаос-тест: ввод случайных строк в поля поиска.
 *
 * 10 тестов с разными seed. Каждый тест вводит хаос-строки
 * (XSS, SQL injection, emoji, unicode, длинные строки) в поля ввода.
 */
test.describe('Chaos: Search Input', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 1000 + i * 555;

		test(`search input chaos seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			// Находим все поля ввода на странице
			const inputs = page.locator('input[type="text"], input[type="search"], textarea');
			const inputCount = await inputs.count();

			// Если есть поля ввода, тестируем их
			if (inputCount > 0) {
				const numInputs = randomInt(rng, 3, Math.min(CHAOS_INPUTS.length, 8));

				for (let j = 0; j < numInputs; j++) {
					const chaosInput = CHAOS_INPUTS[j % CHAOS_INPUTS.length];
					const inputIndex = randomInt(rng, 0, inputCount - 1);

					try {
						const input = inputs.nth(inputIndex);
						if (await input.isVisible().catch(() => false)) {
							await input.click({ timeout: 3000 }).catch(() => {});
							await input.fill('', { timeout: 3000 }).catch(() => {});
							await input.pressSequentially(chaosInput, {
								delay: randomInt(rng, 10, 50),
							}).catch(() => {
								// fill может не справиться с очень длинными строками
								// пробуем через evaluate
								input.evaluate((el, text) => {
									(el as HTMLInputElement).value = text;
									el.dispatchEvent(new Event('input', { bubbles: true }));
								}, chaosInput);
							});
						}
					} catch {
						// Поле ввода могло стать невалидным
					}

					await page.waitForTimeout(randomInt(rng, 100, 300));
					await expectNoCrash(page);
				}
			}

			// Также тестируем поле поиска напрямую
			const searchInput = page.getByPlaceholder('Поиск льгот...').first();
			if (await searchInput.isVisible().catch(() => false)) {
				for (let k = 0; k < Math.min(5, CHAOS_INPUTS.length); k++) {
					const chaosInput = CHAOS_INPUTS[k];
					try {
						await searchInput.click({ timeout: 3000 }).catch(() => {});
						await searchInput.fill('', { timeout: 3000 }).catch(() => {});

						// Для очень длинных строк используем evaluate
						if (chaosInput.length > 1000) {
							await searchInput.evaluate((el, text) => {
								(el as HTMLInputElement).value = text;
								el.dispatchEvent(new Event('input', { bubbles: true }));
							}, chaosInput);
						} else {
							await searchInput.fill(chaosInput, { timeout: 5000 }).catch(() => {});
						}
					} catch {
						// Игнорируем
					}

					await page.waitForTimeout(200);
					await expectNoCrash(page);
				}
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});
