import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useAuthStore } from '@/stores/authStore'

// Мокаем window.location.href: logout использует browser-based redirect
// через '/api/v1/auth/logout'. В тестах jsdom не следует redirect'ам.
vi.stubGlobal('location', {
	...window.location,
	href: '',
})

describe('authStore', () => {
	beforeEach(() => {
		useAuthStore.setState({
			user: null,
			userRoles: [],
			isAuthenticated: false,
			isLoading: false,
		})
	})

	afterEach(() => {
		vi.restoreAllMocks()
	})

	describe('setAuth', () => {
		it('устанавливает auth state', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 'test@test.com', first_name: 'Test', last_name: 'User' },
				['employee', 'catalog_manager']
			)

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(true)
			expect(state.userRoles).toEqual(['employee', 'catalog_manager'])
			expect(state.user).toEqual({
				id: '1',
				email: 'test@test.com',
				first_name: 'Test',
				last_name: 'User',
			})
		})
	})

	describe('clearAuth', () => {
		it('очищает auth state', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)
			useAuthStore.getState().clearAuth()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(state.userRoles).toEqual([])
			expect(state.user).toBeNull()
		})
	})

	describe('logout', () => {
		beforeEach(() => {
			// Сбрасываем mock location.href перед каждым тестом
			window.location.href = ''
		})

		it('вызывает window.location.href и очищает state', async () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)
			await useAuthStore.getState().logout()

			expect(window.location.href).toBe('/api/v1/auth/logout')

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(state.user).toBeNull()
			expect(state.userRoles).toEqual([])
		})

		it('очищает localStorage и sessionStorage', async () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)
			sessionStorage.setItem('lkfl_login_redirecting', 'true')
			sessionStorage.setItem('lkfl_login_attempts', '3')

			await useAuthStore.getState().logout()

			expect(localStorage.getItem('lkfl_user')).toBeNull()
			expect(localStorage.getItem('lkfl_roles')).toBeNull()
			expect(sessionStorage.getItem('lkfl_login_redirecting')).toBeNull()
			expect(sessionStorage.getItem('lkfl_login_attempts')).toBeNull()
			expect(sessionStorage.getItem('lkfl_just_logged_out')).toBe('true')

			// Очистка после теста
			sessionStorage.removeItem('lkfl_just_logged_out')
		})
	})

	describe('checkAuthSession', () => {
		it('возвращает профиль при успешном запросе', async () => {
			const mockProfile = {
				id: '1',
				email: 'test@test.com',
				first_name: 'Test',
				last_name: 'User',
			}

			vi.spyOn(window, 'fetch').mockImplementation(
				(() =>
					Promise.resolve({
						ok: true,
						json: () => Promise.resolve(mockProfile),
					})) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession()

			expect(result).toEqual(mockProfile)
			expect(fetch).toHaveBeenCalledWith('/api/v1/auth/me', {
				credentials: 'include',
			})
		})

		it('возвращает null при не-OK ответе', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: false, status: 401 })) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession()

			expect(result).toBeNull()
		})

		it('возвращает null при ошибке сети', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.reject(new Error('Network error'))) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession()

			expect(result).toBeNull()
		})
	})

	describe('setUser', () => {
		it('обновляет данные пользователя без смены сессии', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 'old@test.com', first_name: 'Old', last_name: 'User' },
				['employee']
			)

			useAuthStore.getState().setUser({
				id: '1',
				email: 'new@test.com',
				first_name: 'New',
				last_name: 'User',
			})

			const state = useAuthStore.getState()
			expect(state.user?.email).toBe('new@test.com')
			expect(state.user?.first_name).toBe('New')
			expect(state.userRoles).toEqual(['employee'])
		})
	})

	describe('setLoading', () => {
		it('устанавливает флаг загрузки', () => {
			useAuthStore.getState().setLoading(true)
			expect(useAuthStore.getState().isLoading).toBe(true)

			useAuthStore.getState().setLoading(false)
			expect(useAuthStore.getState().isLoading).toBe(false)
		})
	})

	// =============================================================================
	// EDGE CASE TESTS
	// =============================================================================

	describe('session expiration edge cases', () => {
		beforeEach(() => {
			window.location.href = ''
		})

		it('корректно обрабатывает logout', async () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(window.location.href).toBe('/api/v1/auth/logout')
		})
	})

	describe('clearAuth', () => {
		it('clearAuth не вызывает fetch', async () => {
			const fetchSpy = vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: true })) as any
			)

			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			useAuthStore.getState().clearAuth()

			expect(fetchSpy).not.toHaveBeenCalled()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
		})
	})

	describe('logout cleanup', () => {
		beforeEach(() => {
			window.location.href = ''
		})

		it('logout удаляет все данные пользователя', async () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 'full@test.com', first_name: 'Full', last_name: 'User' },
				['admin', 'catalog_manager', 'hr']
			)

			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.user).toBeNull()
			expect(state.userRoles).toEqual([])
			expect(state.isAuthenticated).toBe(false)
			expect(window.location.href).toBe('/api/v1/auth/logout')
		})

		it('logout при уже неавторизованном состоянии', async () => {
			useAuthStore.getState().clearAuth()

			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(window.location.href).toBe('/api/v1/auth/logout')
		})

		it('logout очищает session storage флаги login-потока', async () => {
			// Эмулируем состояние после неудачной попытки входа
			sessionStorage.setItem('lkfl_login_redirecting', 'true')
			sessionStorage.setItem('lkfl_login_attempts', '3')

			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			await useAuthStore.getState().logout()

			// Флаги login-потока должны быть очищены
			expect(sessionStorage.getItem('lkfl_login_redirecting')).toBeNull()
			expect(sessionStorage.getItem('lkfl_login_attempts')).toBeNull()

			// Флаг just_logged_out должен быть установлен
			expect(sessionStorage.getItem('lkfl_just_logged_out')).toBe('true')

			// Очищаем после теста
			sessionStorage.removeItem('lkfl_just_logged_out')
		})
	})

	describe('concurrent state updates', () => {
		it('одновременные обновления user и roles', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			// Concurrent updates
			Promise.all([
				useAuthStore.getState().setUser({
					id: '1',
					email: 'updated@test.com',
					first_name: 'Updated',
					last_name: 'User',
				}),
				new Promise<void>((resolve) => {
					useAuthStore.getState().setAuth(
						{ id: '1', email: 'new@test.com', first_name: 'New', last_name: 'User' },
						['admin']
					)
					resolve()
				}),
			])

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(true)
		})

		it('множественные clearAuth вызовы', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			useAuthStore.getState().clearAuth()
			useAuthStore.getState().clearAuth()
			useAuthStore.getState().clearAuth()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
		})
	})

	describe('store reset', () => {
		it('полный reset через setState', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)

			useAuthStore.setState({
				user: null,
				userRoles: [],
				isAuthenticated: false,
				isLoading: false,
			})

			const state = useAuthStore.getState()
			expect(state.user).toBeNull()
			expect(state.userRoles).toEqual([])
			expect(state.isAuthenticated).toBe(false)
			expect(state.isLoading).toBe(false)
		})
	})

	describe('role change during session', () => {
		it('изменение ролей через setAuth', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			expect(useAuthStore.getState().userRoles).toEqual(['employee'])

			// Change roles
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin', 'catalog_manager']
			)

			const state = useAuthStore.getState()
			expect(state.userRoles).toEqual(['admin', 'catalog_manager'])
			expect(state.isAuthenticated).toBe(true)
		})

		it('удаление всех ролей', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee', 'admin']
			)

			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				[]
			)

			const state = useAuthStore.getState()
			expect(state.userRoles).toEqual([])
			expect(state.isAuthenticated).toBe(true)
		})
	})

	describe('checkAuthSession edge cases', () => {
		it('возвращает null при 500 ошибке сервера', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: false, status: 500 })) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession()

			expect(result).toBeNull()
		})

		it('корректно парсит JSON ответ', async () => {
			const mockProfile = {
				id: '123',
				email: 'parsed@test.com',
				first_name: 'Parsed',
				last_name: 'User',
			}

			vi.spyOn(window, 'fetch').mockImplementation(
				(() =>
					Promise.resolve({
						ok: true,
						json: () => Promise.resolve(mockProfile),
					})) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession()

			expect(result).toEqual(mockProfile)
		})

		it('возвращает null при ошибке парсинга JSON', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() =>
					Promise.resolve({
						ok: true,
						json: () => Promise.reject(new Error('Invalid JSON')),
					})) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession()

			expect(result).toBeNull()
		})
	})

	describe('setUser edge cases', () => {
		it('setUser не меняет роли', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 'old@test.com', first_name: 'Old', last_name: 'User' },
				['employee']
			)

			useAuthStore.getState().setUser({
				id: '1',
				email: 'new@test.com',
				first_name: 'New',
				last_name: 'User',
			})

			const state = useAuthStore.getState()
			expect(state.userRoles).toEqual(['employee'])
		})

		it('setUser с пустыми полями', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			useAuthStore.getState().setUser({
				id: '',
				email: '',
				first_name: '',
				last_name: '',
			})

			const state = useAuthStore.getState()
			expect(state.user?.email).toBe('')
			expect(state.user?.first_name).toBe('')
		})
	})

	describe('setLoading edge cases', () => {
		it('setLoading не меняет auth состояние', () => {
			useAuthStore.getState().setAuth(
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			useAuthStore.getState().setLoading(true)

			const state = useAuthStore.getState()
			expect(state.isLoading).toBe(true)
			expect(state.isAuthenticated).toBe(true)
		})
	})
})
