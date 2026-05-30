import { test } from '@playwright/test';
import {
	chaosSeed,
	randomFrom,
	randomInt,
	expectNoCrash,
	setupChaosTest,
	CHAOS_KEYS,
	CHAOS_KEY_COMBOS,
} from './helpers';

/**
 * Хаос-тест: случайные нажатия клавиш.
 *
 * 10 тестов с разными seed. Каждый тест генерирует случайную последовательность
 * нажатий клавиш (Tab, Enter, Escape, Backspace, Ctrl+Z, Ctrl+A и т.д.).
 */
test.describe('Chaos: Keyboard Chaos', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 13000 + i * 333;

		test(`keyboard chaos seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			const numKeystrokes = randomInt(rng, 20, 50);

			for (let j = 0; j < numKeystrokes; j++) {
				const actionType = randomInt(rng, 0, 3);

				switch (actionType) {
					case 0: {
						// Одиночное нажатие клавиши
						const key = randomFrom(rng, CHAOS_KEYS);
						await page.keyboard.press(key, { timeout: 3000 }).catch(() => {});
						break;
					}

					case 1: {
						// Комбинация клавиш
						const combo = randomFrom(rng, CHAOS_KEY_COMBOS);
						const fullKey = combo.modifiers.join('+') + '+' + combo.key;
						await page.keyboard.press(fullKey, { timeout: 3000 }).catch(() => {});
						break;
					}

					case 2: {
						// Ввод случайного символа
						const chars = 'abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*';
						const char = chars[Math.floor(rng() * chars.length)];
						await page.keyboard.type(char, {
							delay: randomInt(rng, 20, 100),
						}).catch(() => {});
						break;
					}

					case 3: {
						// Tab навигация (переход между фокусируемыми элементами)
						const tabCount = randomInt(rng, 3, 8);
						for (let t = 0; t < tabCount; t++) {
							await page.keyboard.press('Tab', { timeout: 2000 }).catch(() => {});
							await page.waitForTimeout(30);
						}
						break;
					}
				}

				await page.waitForTimeout(randomInt(rng, 30, 100));

				// Проверяем после каждой серии действий
				if (j % 5 === 0) {
					await expectNoCrash(page);
				}
			}

			// Финальная проверка
			await expectNoCrash(page);
			expect(pageErrors).toHaveLength(0);
		});
	}
});
