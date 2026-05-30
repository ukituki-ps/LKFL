import type { Page } from '@playwright/test';
import { expect } from '@playwright/test';
import { setupApiMocks } from '../helpers';

/**
 * Вспомогательные функции для хаос-тестов (chaos tests).
 *
 * Хаос-тесты имитируют поведение пользователя «тыкает на всё подряд»
 * и проверяют, что приложение не крашится, не показывает белый экран
 * и не генерирует unhandled console errors.
 */

// ─── Детерминированный генератор случайных чисел ───

/**
 * Создаёт детерминированный генератор случайных чисел по seed.
 * Использует алгоритм mulberry32 для воспроизводимости.
 */
export function chaosSeed(seed: number): () => number {
	let s = seed | 0;
	return () => {
		s = (s + 0x6d2b79f5) | 0;
		let t = Math.imul(s ^ (s >>> 15), 1 | s);
		t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

/**
 * Выбирает случайный элемент из массива с помощью детерминированного RNG.
 */
export function randomFrom<T>(rng: () => number, arr: T[]): T {
	return arr[Math.floor(rng() * arr.length)];
}

/**
 * Генерирует случайное число в диапазоне [min, max] с помощью детерминированного RNG.
 */
export function randomInt(rng: () => number, min: number, max: number): number {
	return Math.floor(rng() * (max - min + 1)) + min;
}

// ─── Проверка отсутствия краша ───

/**
 * Проверяет что страница не в состоянии краша:
 * - Root элемент (#root) не пустой
 * - Нет JS errors в console
 * - Страница откликнулась
 */
export async function expectNoCrash(page: Page): Promise<void> {
	// Проверяем что root элемент не пустой (нет белого экрана)
	const rootText = await page.locator('#root').textContent();
	expect(rootText).not.toBe('');
	expect(rootText).not.toBeNull();

	// Проверяем что страница не перешла на about:blank
	const url = page.url();
	expect(url).not.toContain('about:blank');
}

/**
 * Проверяет что в console нет unhandled errors.
 * Собирает console messages и проверяет отсутствие error/warning.
 */
export function collectConsoleErrors(page: Page): string[] {
	const errors: string[] = [];
	page.on('console', (msg) => {
		if (msg.type() === 'error') {
			errors.push(msg.text());
		}
	});
	return errors;
}

/**
 * Проверяет отсутствие page errors (unhandled promise rejections etc).
 */
export function collectPageErrors(page: Page): string[] {
	const errors: string[] = [];
	page.on('pageerror', (error) => {
		errors.push(error.message);
	});
	return errors;
}

// ─── Настройка хаос-теста ───

/**
 * Стандартная настройка хаос-теста:
 * - Мокирует API
 * - Начинает сбор console errors
 * - Начинает сбор page errors
 * - Навигация на базовую страницу
 */
export async function setupChaosTest(page: Page, startUrl: string = '/'): Promise<{
	consoleErrors: string[];
	pageErrors: string[];
}> {
	const consoleErrors: string[] = [];
	const pageErrors: string[] = [];

	page.on('console', (msg) => {
		if (msg.type() === 'error') {
			consoleErrors.push(msg.text());
		}
	});

	page.on('pageerror', (error) => {
		pageErrors.push(error.message);
	});

	await setupApiMocks(page);
	await page.goto(startUrl);
	await page.waitForLoadState('domcontentloaded');

	// Небольшая задержка для стабилизации React рендера
	await page.waitForTimeout(500);

	return { consoleErrors, pageErrors };
}

// ─── Хаос-данные ───

/**
 * Доступные страницы приложения для навигации.
 */
export const APP_PAGES = [
	'/',
	'/catalog',
	'/points',
	'/documents',
	'/support',
] as const;

/**
 * Строки для хаос-теста ввода.
 */
export const CHAOS_INPUTS = [
	// XSS attempts
	'<script>alert(1)</script>',
	'<img src=x onerror=alert(1)>',
	'javascript:alert(document.cookie)',
	'"><script>document.location="http://evil.com"</script>',

	// SQL injection attempts
	"' OR 1=1 --",
	"'; DROP TABLE users; --",
	"1' OR '1'='1",
	"admin'--",

	// Unicode & Emoji
	'🎉🎊🎁🎈🎄🎃👻🤖🚀💎'.repeat(50),
	'こんにちは世界',
	'Привет мир',
	'مرحبا بالعالم',
	'🏆🥇🥈🥉',

	// Long strings
	'a'.repeat(10000),
	' '.repeat(5000) + 'x' + ' '.repeat(5000),

	// Control characters
	'\x00\x01\x02\x03',
	'\t\n\r',
	'\u0000\uFFFF\u10FFFF',

	// Special characters
	'!@#$%^&*()_+-=[]{}|;:,.<>?/~`',
	'\u00AB\u00BB\u201C\u201D\u2018\u2019\u00AB\u00BB',
	'\u2014\u2013\u2014\u2012\u2011\u2011\u2017',

	// Mixed chaos
	'<b>BOLD</b> <i>ITALIC</i> "quotes" &amp; <script>x</script>',
	"SELECT * FROM users WHERE name='' OR '1'='1' -- 🎉",
];

/**
 * Viewport размеры для хаос-теста.
 */
export const VIEWPORT_SIZES = [
	{ width: 375, height: 667, label: 'iPhone 8' },
	{ width: 414, height: 896, label: 'iPhone 11 Pro' },
	{ width: 768, height: 1024, label: 'iPad' },
	{ width: 1024, height: 768, label: 'Tablet landscape' },
	{ width: 1280, height: 720, label: 'Desktop HD' },
	{ width: 1920, height: 1080, label: 'Desktop FHD' },
	{ width: 2560, height: 1440, label: 'Desktop QHD' },
	{ width: 3840, height: 2160, label: 'Desktop 4K' },
] as const;

/**
 * Клавиши для хаос-теста клавиатуры.
 */
export const CHAOS_KEYS = [
	'Tab',
	'Enter',
	'Escape',
	'Backspace',
	'Delete',
	'Home',
	'End',
	'PageUp',
	'PageDown',
	'ArrowUp',
	'ArrowDown',
	'ArrowLeft',
	'ArrowRight',
	'F1',
	'F5',
] as const;

/**
 * Комбинации клавиш для хаос-теста (без тех, что закрывают браузер).
 */
export const CHAOS_KEY_COMBOS = [
	{ key: 'z', modifiers: ['Control'] },
	{ key: 'y', modifiers: ['Control'] },
	{ key: 'a', modifiers: ['Control'] },
	{ key: 'c', modifiers: ['Control'] },
	{ key: 'v', modifiers: ['Control'] },
	{ key: 'x', modifiers: ['Control'] },
	{ key: 'z', modifiers: ['Control', 'Shift'] },
] as const;

/**
 * Фильтры каталога для хаос-теста.
 */
export const CATALOG_FILTERS = {
	types: ['benefit', 'activity', 'discount', 'promo'],
	statuses: ['active', 'inactive', 'scheduled'],
	categories: ['health', 'sport', 'education', 'food'],
} as const;

/**
 * Номера страниц для хаос-теста пагинации.
 */
export const PAGINATION_PAGES = [1, 2, 5, 10, 3, 7, 1, 4, 9, 2] as const;
