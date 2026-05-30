import { test } from '@playwright/test';
import {
	chaosSeed,
	randomInt,
	expectNoCrash,
	setupChaosTest,
} from './helpers';

/**
 * Хаос-тест: отключение/восстановление сети во время операций.
 *
 * 10 тестов с разными seed. Каждый тест имитирует потерю сети
 * на разное время и проверяет, что приложение не крашится.
 */
test.describe('Chaos: Network Chaos', () => {
	for (let i = 0; i < 10; i++) {
		const seed = 9000 + i * 111;

		test(`network chaos seed=${seed}`, async ({ page }) => {
			const rng = chaosSeed(seed);
			const { consoleErrors, pageErrors } = await setupChaosTest(page, '/catalog');

			await page.waitForLoadState('networkidle').catch(() => {});
			await page.waitForTimeout(500);

			const numInterruptions = randomInt(rng, 2, 5);

			for (let j = 0; j < numInterruptions; j++) {
				const offlineDuration = randomInt(rng, 1000, 5000);
				const interruptionType = randomInt(rng, 0, 2);

				switch (interruptionType) {
					case 0: {
						// Полное отключение сети
						await page.route('**/*', (route) => {
							if (!route.request().url().includes('playwright')) {
								route.abort('networkoffline');
							} else {
								route.continue();
							}
						});

						await page.waitForTimeout(offlineDuration);

						// Восстанавливаем сеть
						await page.unroute('**/*');
						await page.route('**/*', (route) => route.continue());

						// Пытаемся сделать действие после восстановления
						try {
							await page.reload({ waitUntil: 'domcontentloaded', timeout: 10000 }).catch(() => {});
						} catch {
							// Релоад мог не успеть
						}
						break;
					}

					case 1: {
						// Медленная сеть (throttling через задержку)
						await page.route('**/api/**', (route) => {
							const delay = randomInt(rng, 2000, 8000);
							setTimeout(() => route.continue(), delay);
						});

						// Пытаемся навигацию во время медленной сети
						try {
							await page.goto('/catalog', {
								waitUntil: 'domcontentloaded',
								timeout: 15000,
							}).catch(() => {});
						} catch {
							// Таймаут навигации — это нормально
						}

						await page.unroute('**/api/**');
						break;
					}

					case 2: {
						// Случайные ошибки API (500/503/408)
						const errorCodes = [500, 503, 408, 429] as const;
						let errorCount = 0;

						await page.route('**/api/**', (route) => {
							if (errorCount < 3) {
								errorCount++;
								const status = randomFrom(rng, errorCodes);
								route.fulfill({
									status,
									contentType: 'application/json',
									body: JSON.stringify({ error: `Simulated ${status} error` }),
								});
							} else {
								route.continue();
							}
						});

						// Пытаемся действия во время ошибок
						try {
							await page.reload({ waitUntil: 'domcontentloaded', timeout: 10000 }).catch(() => {});
						} catch {
							// Игнорируем
						}

						await page.unroute('**/api/**');
						break;
					}
				}

				await page.waitForTimeout(500);
				await expectNoCrash(page);
			}

			expect(pageErrors).toHaveLength(0);
		});
	}
});

function randomFrom<T>(rng: () => number, arr: T[]): T {
	return arr[Math.floor(rng() * arr.length)];
}
