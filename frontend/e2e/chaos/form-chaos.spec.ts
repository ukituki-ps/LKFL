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
 * Хаос-тест: случайное заполнение и отправка форм.
 *
 * 10 тестов с разными seed. Каждый тест находит формы на странице,
 * заполняет поля случайными данными и пытается отправить.
 */
test.describe('Chaos: Form Chaos', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 7000 + i * 888;

		test(`form chaos seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			// Находим все формы на странице
			const forms = page.locator('form');
			const formCount = await forms.count();

			// Также ищем поля ввода вне форм (например, поиск)
			const allInputs = page.locator(
				'input[type="text"], input[type="email"], input[type="number"], ' +
				'input[type="search"], textarea, select'
			);
			const inputCount = await allInputs.count();

			if (inputCount > 0) {
				// Заполняем случайные поля хаос-данными
				const numFields = randomInt(rng, 3, Math.min(inputCount, 10));

				for (let j = 0; j < numFields; j++) {
					const inputIdx = randomInt(rng, 0, inputCount - 1);
					const input = allInputs.nth(inputIdx);

					try {
						if (!await input.isVisible().catch(() => false)) continue;

						const inputType = await input.getAttribute('type') || '';

						if (inputType === 'select-one' || await input.evaluate(el => el.tagName === 'SELECT')) {
							// Для select — выбираем случайную опцию
							await input.click({ timeout: 3000 }).catch(() => {});
							const options = page.locator('option').locator(':visible');
							const optCount = await options.count();
							if (optCount > 1) {
								const optIdx = randomInt(rng, 1, optCount - 1);
								await options.nth(optIdx).click({ timeout: 2000 }).catch(() => {});
							}
						} else {
							// Для текстовых полей — заполняем хаос-данными
							const chaosValue = randomFrom(rng, CHAOS_INPUTS);
							await input.click({ timeout: 3000 }).catch(() => {});

							if (chaosValue.length > 1000) {
								await input.evaluate((el, text) => {
									(el as HTMLInputElement).value = text;
									el.dispatchEvent(new Event('input', { bubbles: true }));
								}, chaosValue);
							} else {
								await input.fill(chaosValue, { timeout: 5000 }).catch(() => {});
							}
						}
					} catch {
						continue;
					}

					await page.waitForTimeout(randomInt(rng, 50, 150));
					await expectNoCrash(page);
				}

				// Пробуем отправить формы
				if (formCount > 0) {
					for (let f = 0; f < Math.min(formCount, 3); f++) {
						try {
							const form = forms.nth(f);
							if (!await form.isVisible().catch(() => false)) continue;

							// Ищем кнопку submit
							const submitBtn = form.locator('button[type="submit"], input[type="submit"]').first();
							if (await submitBtn.isVisible().catch(() => false)) {
								await submitBtn.click({ timeout: 5000 }).catch(() => {});
							}
						} catch {
							continue;
						}

						await page.waitForTimeout(300);
						await expectNoCrash(page);
					}
				}
			}

			// Также тестируем пустую отправку — кликаем по кнопкам submit без заполнения
			const submitButtons = page.locator('button[type="submit"], [data-testid="submit-btn"]').locator(':visible');
			const submitCount = await submitButtons.count();

			for (let s = 0; s < Math.min(submitCount, 3); s++) {
				try {
					await submitButtons.nth(s).click({ timeout: 3000 }).catch(() => {});
				} catch {
					continue;
				}

				await page.waitForTimeout(200);
				await expectNoCrash(page);
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});
