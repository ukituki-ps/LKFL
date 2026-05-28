import type { Page, Response } from '@playwright/test';

/**
 * Хелперы для E2E тестов LKFL.
 *
 * Все API-запросы мокаются через page.route() — backend не требуется.
 * Auth-состояние устанавливается через Zustand store напрямую в браузере.
 */

// ─── Моковые данные ───

export const mockUserEmployee = {
	id: 'test-user-001',
	email: 'employee@lkfl.test',
	first_name: 'Иван',
	last_name: 'Петров',
	tenant_id: 'test-tenant',
	roles: ['employee'] as string[],
};

export const mockUserAdmin = {
	id: 'test-user-admin',
	email: 'admin@lkfl.test',
	first_name: 'Анна',
	last_name: 'Иванова',
	tenant_id: 'test-tenant',
	roles: ['admin', 'hr', 'catalog_manager'] as string[],
};

export const mockUserHR = {
	id: 'test-user-hr',
	email: 'hr@lkfl.test',
	first_name: 'Мария',
	last_name: 'Сидорова',
	tenant_id: 'test-tenant',
	roles: ['hr'] as string[],
};

export const mockEngagements = {
	data: [
		{
			id: 'eng-001',
			slug: 'medinsurance-basic',
			name: 'ДМС Базовый',
			description: 'Базовая медицинская страховка для сотрудника',
			type: 'benefit',
			status: 'active',
			badge: 'Доступна',
			image_url: '/placeholder-image.svg',
			cost_cents: 0,
			provider_name: 'Согласие',
			category: { slug: 'health', name: 'Здоровье' },
			offers: [
				{
					id: 'off-001',
					provider_id: 'prov-soglasie',
					provider_name: 'Согласие',
					product_name: 'ДМС Базовый',
					cost_cents: 0,
					status: 'active',
				},
			],
		},
		{
			id: 'eng-002',
			slug: 'gym-membership',
			name: 'Абонемент в фитнес',
			description: 'Ежемесячный абонемент в сеть фитнес-клубов',
			type: 'benefit',
			status: 'active',
			badge: 'Доступна',
			image_url: '/placeholder-image.svg',
			cost_cents: 300000,
			provider_name: 'World Class',
			category: { slug: 'sport', name: 'Спорт' },
			offers: [
				{
					id: 'off-002',
					provider_id: 'prov-wc',
					provider_name: 'World Class',
					product_name: 'Абонемент Standard',
					cost_cents: 300000,
					status: 'active',
				},
			],
		},
		{
			id: 'eng-003',
			slug: 'learning-budget',
			name: 'Обучающий бюджет',
			description: 'Бюджет на обучение и курсы',
			type: 'benefit',
			status: 'promo',
			badge: 'Промо',
			image_url: '/placeholder-image.svg',
			cost_cents: 0,
			provider_name: 'ЛКФЛ',
			category: { slug: 'education', name: 'Образование' },
			offers: [
				{
					id: 'off-003',
					provider_id: 'prov-lkfl',
					provider_name: 'ЛКФЛ',
					product_name: 'Обучающий бюджет',
					cost_cents: 0,
					status: 'active',
				},
			],
		},
		{
			id: 'eng-004',
			slug: 'run-club',
			name: 'Клуб бегунов',
			description: 'Еженедельные пробежки с тренером',
			type: 'activity',
			status: 'active',
			badge: 'Доступна',
			image_url: '/placeholder-image.svg',
			cost_cents: null,
			provider_name: 'ЛКФЛ',
			category: { slug: 'sport', name: 'Спорт' },
			offers: [
				{
					id: 'off-004',
					provider_id: 'prov-lkfl',
					provider_name: 'ЛКФЛ',
					product_name: 'Клуб бегунов',
					cost_cents: null,
					status: 'active',
				},
			],
		},
	],
	pagination: {
		page: 1,
		per_page: 20,
		total: 4,
		total_pages: 1,
	},
};

export const mockEngagementsPaginated = {
	data: [
		{
			id: 'eng-005',
			slug: 'page2-item-1',
			name: 'Льгота на странице 2',
			description: 'Тестовая льгота для пагинации',
			type: 'benefit',
			status: 'active',
			badge: 'Доступна',
			image_url: '/placeholder-image.svg',
			cost_cents: 500000,
			provider_name: 'ТестПровайдер',
			category: { slug: 'health', name: 'Здоровье' },
			offers: [
				{
					id: 'off-005',
					provider_id: 'prov-test',
					provider_name: 'ТестПровайдер',
					product_name: 'Льгота на странице 2',
					cost_cents: 500000,
					status: 'active',
				},
			],
		},
	],
	pagination: {
		page: 2,
		per_page: 2,
		total: 4,
		total_pages: 2,
	},
};

export const mockCategories = [
	{ slug: 'health', name: 'Здоровье', icon: 'heart', sort_order: 1 },
	{ slug: 'sport', name: 'Спорт', icon: 'activity', sort_order: 2 },
	{ slug: 'education', name: 'Образование', icon: 'book', sort_order: 3 },
	{ slug: 'food', name: 'Питание', icon: 'coffee', sort_order: 4 },
];

export const mockCategoriesAdmin = [
	{
		id: 'cat-001',
		slug: 'health',
		name: 'Здоровье',
		icon: 'heart',
		sort_order: 1,
		tenant_id: 'test-tenant',
	},
	{
		id: 'cat-002',
		slug: 'sport',
		name: 'Спорт',
		icon: 'activity',
		sort_order: 2,
		tenant_id: 'test-tenant',
	},
];

// ─── Установка auth-состояния ───

/**
 * Устанавливает auth-состояние в Zustand store через page.evaluate.
 * Вызывает setupAuthForTest() из authStore.ts.
 *
 * ВАЖНО: вызывать ПОСЛЕ page.goto() — когда React app уже загружен.
 * @param page — Playwright Page (должна быть загружена)
 * @param user — моковый пользователь
 */
export async function setAuthState(page: Page, user: typeof mockUserEmployee): Promise<void> {
	await page.evaluate(
		({ user }) => {
			// Вызываем setupAuthForTest из authStore.ts
			// Функция экспортирована через window.__LKFL_AUTH_STORE__ в main.tsx
			const authStore = (window as any).__LKFL_AUTH_STORE__;
			if (authStore && authStore.setupAuthForTest) {
				authStore.setupAuthForTest('mock-jwt-token', user, user.roles || ['employee']);
			} else {
				// Fallback: устанавливаем через localStorage + событие
				localStorage.setItem('lkfl_auth_test', JSON.stringify({
					token: 'mock-jwt-token',
					user,
					isAuthenticated: true,
				}));
				window.dispatchEvent(new CustomEvent('lkfl-auth-test', { detail: { user } }));
			}
		},
		{ user },
	);
}

/**
 * Навигация на страницу с установкой auth-состояния.
 * Используется вместо page.goto() для защищённых маршрутов.
 * @param page — Playwright Page
 * @param url — URL для навигации
 * @param user — моковый пользователь
 */
export async function gotoWithAuth(
	page: Page,
	url: string,
	user: typeof mockUserEmployee,
): Promise<void> {
	// Сначала загружаем любую страницу чтобы React app инициализировался
	await page.goto('/login');
	// Устанавливаем auth-состояние в Zustand store
	await setAuthState(page, user);
	// Навигируемся на целевую страницу
	await page.goto(url);
}

/**
 * Устанавливает auth-состояние и перезагружает страницу.
 * Используется когда store поддерживает гидратацию из localStorage.
 */
export async function loginAs(page: Page, user: typeof mockUserEmployee): Promise<void> {
	await setAuthState(page, user);
	await page.reload();
	await page.waitForLoadState('networkidle');
}

// ─── Установка auth через Zustand store напрямую ───

/**
 * Устанавливает auth-состояние через Zustand store напрямую.
 * Работает без перезагрузки страницы.
 */
export async function setAuthViaStore(page: Page, user: typeof mockUserEmployee): Promise<void> {
	await page.evaluate(
		({ user }) => {
			// Ищем Zustand store по прототипу
			// Zustand create() возвращает хук, доступ к которому через import
			// Альтернативный подход: использовать window.__SETUP_AUTH__
			(window as any).__SETUP_AUTH__ = {
				token: 'mock-jwt-token',
				user: user,
				isAuthenticated: true,
			};
		},
		{ user },
	);
}

// ─── Мокирование API ───

/**
 * Настроить мокирование всех API эндпоинтов для страницы.
 */
export async function setupApiMocks(page: Page): Promise<void> {
	// GET /api/v1/users/me — профиль пользователя
	await page.route('/api/v1/users/me', (route) => {
		const authState = (route.request().headers()['authorization'] || '').includes('mock');
		if (!authState) {
			return route.fulfill({
				status: 401,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'unauthorized' }),
			});
		}
		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(mockUserEmployee),
		});
	});

	// GET /api/v1/engagements — список энгейджментов
	await page.route('/api/v1/engagements*', (route) => {
		const url = new URL(route.request().url());
		const pageParam = parseInt(url.searchParams.get('page') || '1', 10);

		if (pageParam === 2) {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockEngagementsPaginated),
			});
		}

		// Empty state for search with no results
		const search = url.searchParams.get('search');
		if (search === 'nonexistent') {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 0 } }),
			});
		}

		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(mockEngagements),
		});
	});

	// GET /api/v1/engagements/categories — категории
	await page.route('/api/v1/engagements/categories', (route) => {
		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(mockCategories),
		});
	});

	// GET /api/v1/auth/me — проверка сессии
	await page.route('/api/v1/auth/me', (route) => {
		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(mockUserEmployee),
		});
	});

	// POST /api/v1/auth/logout — логаут
	await page.route('/api/v1/auth/logout', (route) => {
		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({ success: true }),
		});
	});

	// Admin API — категории
	await page.route('/admin/engagements/categories', (route) => {
		if (route.request().method() === 'GET') {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(mockCategoriesAdmin),
			});
		}
		if (route.request().method() === 'POST') {
			return route.fulfill({
				status: 201,
				contentType: 'application/json',
				body: JSON.stringify({
					id: 'cat-new-001',
					slug: 'new-category',
					name: 'Новая категория',
					icon: 'star',
					sort_order: 99,
					tenant_id: 'test-tenant',
				}),
			});
		}
	});

	// Admin API — типы
	await page.route('/admin/engagements/types*', (route) => {
		if (route.request().method() === 'GET') {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					data: [
						{ id: 'type-001', name: 'ДМС', slug: 'dms', status: 'active', tenant_id: 'test-tenant' },
						{ id: 'type-002', name: 'Фитнес', slug: 'fitness', status: 'active', tenant_id: 'test-tenant' },
					],
					pagination: { page: 1, per_page: 20, total: 2, total_pages: 1 },
				}),
			});
		}
		if (route.request().method() === 'DELETE') {
			// Check if type has active offers
			const url = route.request().url();
			if (url.includes('type-001')) {
				return route.fulfill({
					status: 409,
					contentType: 'application/json',
					body: JSON.stringify({ error: 'Cannot delete type with active offers' }),
				});
			}
			return route.fulfill({ status: 204 });
		}
		if (route.request().method() === 'PATCH') {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					id: 'type-001',
					name: 'ДМС',
					slug: 'dms',
					status: 'inactive',
					tenant_id: 'test-tenant',
				}),
			});
		}
	});

	// Admin API — категории (update/delete)
	await page.route('/admin/engagements/categories/cat-*', (route) => {
		if (route.request().method() === 'PUT') {
			return route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					id: 'cat-001',
					slug: 'health-updated',
					name: 'Здоровье (обн.)',
					icon: 'heart',
					sort_order: 1,
					tenant_id: 'test-tenant',
				}),
			});
		}
		if (route.request().method() === 'DELETE') {
			return route.fulfill({ status: 204 });
		}
	});
}

/**
 * Настроить мокирование API с ошибкой 500 для энгейджментов.
 */
export async function setupApiErrorMocks(page: Page): Promise<void> {
	await page.route('/api/v1/engagements*', (route) => {
		return route.fulfill({
			status: 500,
			contentType: 'application/json',
			body: JSON.stringify({ error: 'Internal server error' }),
		});
	});

	await page.route('/api/v1/engagements/categories', (route) => {
		return route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify(mockCategories),
		});
	});
}

/**
 * Настроить мокирование API с ошибкой для профиля пользователя.
 */
export async function setupUserProfileErrorMocks(page: Page): Promise<void> {
	await page.route('/api/v1/users/me', (route) => {
		return route.fulfill({
			status: 500,
			contentType: 'application/json',
			body: JSON.stringify({ error: 'Internal server error' }),
		});
	});
}

// ─── Утилиты ожидания ───

/**
 * Дождаться загрузки каталога (появления карточек или пустого состояния).
 */
export async function waitForCatalogLoaded(page: Page): Promise<void> {
	await page.waitForSelector(
		'.engagement-card, [data-testid="engagement-grid"], text="Нет доступных льгот", text="Ошибка загрузки каталога"',
		{ timeout: 10000 },
	);
}

/**
 * Дождаться появления loader или его исчезновения.
 */
export async function waitForNoLoader(page: Page): Promise<void> {
	await page.waitForSelector('.mantine-Loader-root, [data-loading="true"]', {
		state: 'detached',
		timeout: 10000,
	}).catch(() => {
		// Loader может не появиться если моки быстрые
	});
}

/**
 * Дождаться навигации на указанный путь.
 */
export async function waitForUrl(page: Page, path: string): Promise<void> {
	await page.waitForURL(`**${path}`);
}

/**
 * Дождаться появления текста на странице.
 */
export async function waitForText(page: Page, text: string): Promise<void> {
	await page.waitForSelector(`text="${text}"`, { timeout: 5000 });
}

/**
 * Создать тестовую категорию через API (для админ тестов).
 */
export async function createTestCategory(page: Page): Promise<void> {
	const response = await page.request.post('/admin/engagements/categories', {
		data: {
			slug: 'test-category-' + Date.now(),
			name: 'Тестовая категория',
			icon: 'test',
			sort_order: 99,
		},
	});
	expectResponseOk(response);
}

/**
 * Проверить что ответ API OK (2xx).
 */
function expectResponseOk(response: Response | null): void {
	if (response) {
		const status = response.status();
		if (status < 200 || status >= 300) {
			throw new Error(`API request failed with status ${status}`);
		}
	}
}
