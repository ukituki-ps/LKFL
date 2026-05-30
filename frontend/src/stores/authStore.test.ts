import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { useAuthStore } from '@/stores/authStore'

describe('authStore', () => {
	beforeEach(() => {
		useAuthStore.setState({
			token: null,
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
				'test-token',
				{ id: '1', email: 'test@test.com', first_name: 'Test', last_name: 'User' },
				['employee', 'catalog_manager']
			)

			const state = useAuthStore.getState()
			expect(state.token).toBe('test-token')
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
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)
			useAuthStore.getState().clearAuth()

			const state = useAuthStore.getState()
			expect(state.token).toBeNull()
			expect(state.isAuthenticated).toBe(false)
			expect(state.userRoles).toEqual([])
			expect(state.user).toBeNull()
		})
	})

	describe('logout', () => {
		it('вызывает POST /api/v1/auth/logout и очищает state', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() =>
					Promise.resolve({
						ok: true,
						status: 204,
					})) as any
			)

			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)
			await useAuthStore.getState().logout()

			expect(fetch).toHaveBeenCalledWith('/api/v1/auth/logout', {
				method: 'POST',
				credentials: 'include', // D2: cookie-based session
			})

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(state.token).toBeNull()
			expect(state.user).toBeNull()
			expect(state.userRoles).toEqual([])
		})

		it('очищает state даже если fetch бросает ошибку', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.reject(new Error('Network error'))) as any
			)

			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)
			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(state.token).toBeNull()
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
			const result = await checkAuthSession('valid-token')

			expect(result).toEqual(mockProfile)
			expect(fetch).toHaveBeenCalledWith('/api/v1/auth/me', {
				credentials: 'include', // D2: cookie-based auth
			})
		})

		it('возвращает null при не-OK ответе', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: false, status: 401 })) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession('invalid-token')

			expect(result).toBeNull()
		})

		it('возвращает null при ошибке сети', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.reject(new Error('Network error'))) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession('any-token')

			expect(result).toBeNull()
		})
	})

	describe('setUser', () => {
		it('обновляет данные пользователя без смены токена', () => {
			useAuthStore.getState().setAuth(
				'existing-token',
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
			expect(state.token).toBe('existing-token')
			expect(state.user?.email).toBe('new@test.com')
			expect(state.user?.first_name).toBe('New')
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

	describe('token expiration edge cases', () => {
		it('корректно обрабатывает null токен при logout', async () => {
			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: true })) as any
			)

			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.token).toBeNull()
			expect(state.isAuthenticated).toBe(false)
		})

		it('пустой токен строка обрабатывается корректно', () => {
			useAuthStore.getState().setAuth(
				'',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			const state = useAuthStore.getState()
			expect(state.token).toBe('')
			expect(state.isAuthenticated).toBe(true)
		})
	})

	describe('refresh failure', () => {
		it('clearAuth не вызывает fetch', async () => {
			const fetchSpy = vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: true })) as any
			)

			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			useAuthStore.getState().clearAuth()

			expect(fetchSpy).not.toHaveBeenCalled()

			const state = useAuthStore.getState()
			expect(state.token).toBeNull()
			expect(state.isAuthenticated).toBe(false)
		})
	})

	describe('logout cleanup', () => {
		it('logout удаляет все данные пользователя', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: true })) as any
			)

			useAuthStore.getState().setAuth(
				'long-token-value',
				{ id: '1', email: 'full@test.com', first_name: 'Full', last_name: 'User' },
				['admin', 'catalog_manager', 'hr']
			)

			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.token).toBeNull()
			expect(state.user).toBeNull()
			expect(state.userRoles).toEqual([])
			expect(state.isAuthenticated).toBe(false)
		})

		it('logout сохраняет cleanup при ошибке fetch', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.reject(new Error('Network error'))) as any
			)

			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)

			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(state.token).toBeNull()
			expect(state.user).toBeNull()
		})

		it('logout при уже неавторизованном состоянии', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: true })) as any
			)

			useAuthStore.getState().clearAuth()

			await useAuthStore.getState().logout()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(state.token).toBeNull()
		})
	})

	describe('concurrent state updates', () => {
		it('одновременные обновления user и roles', () => {
			useAuthStore.getState().setAuth(
				'token',
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
						'new-token',
						{ id: '1', email: 'new@test.com', first_name: 'New', last_name: 'User' },
						['admin']
					)
					resolve()
				}),
			])

			const state = useAuthStore.getState()
			// State should be consistent (last write wins in Zustand)
			expect(state.token).toBeTruthy()
			expect(state.isAuthenticated).toBe(true)
		})

		it('множественные clearAuth вызовы', () => {
			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			useAuthStore.getState().clearAuth()
			useAuthStore.getState().clearAuth()
			useAuthStore.getState().clearAuth()

			const state = useAuthStore.getState()
			expect(state.isAuthenticated).toBe(false)
			expect(state.token).toBeNull()
		})
	})

	describe('store reset', () => {
		it('полный reset через setState', () => {
			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin']
			)

			useAuthStore.setState({
				token: null,
				user: null,
				userRoles: [],
				isAuthenticated: false,
				isLoading: false,
			})

			const state = useAuthStore.getState()
			expect(state.token).toBeNull()
			expect(state.user).toBeNull()
			expect(state.userRoles).toEqual([])
			expect(state.isAuthenticated).toBe(false)
			expect(state.isLoading).toBe(false)
		})
	})

	describe('role change during session', () => {
		it('изменение ролей через setAuth', () => {
			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			expect(useAuthStore.getState().userRoles).toEqual(['employee'])

			// Change roles
			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['admin', 'catalog_manager']
			)

			const state = useAuthStore.getState()
			expect(state.userRoles).toEqual(['admin', 'catalog_manager'])
			expect(state.isAuthenticated).toBe(true)
		})

		it('удаление всех ролей', () => {
			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee', 'admin']
			)

			useAuthStore.getState().setAuth(
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				[]
			)

			const state = useAuthStore.getState()
			expect(state.userRoles).toEqual([])
			expect(state.isAuthenticated).toBe(true)
		})
	})

	describe('checkAuthSession edge cases', () => {
		it('возвращает null при пустом токене', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: false, status: 401 })) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession('')

			expect(result).toBeNull()
		})

		it('возвращает null при 500 ошибке сервера', async () => {
			vi.spyOn(window, 'fetch').mockImplementation(
				(() => Promise.resolve({ ok: false, status: 500 })) as any
			)

			const { checkAuthSession } = await import('@/stores/authStore')
			const result = await checkAuthSession('valid-token')

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
			const result = await checkAuthSession('valid-token')

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
			const result = await checkAuthSession('valid-token')

			expect(result).toBeNull()
		})
	})

	describe('setUser edge cases', () => {
		it('setUser не меняет токен', () => {
			useAuthStore.getState().setAuth(
				'original-token',
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
			expect(state.token).toBe('original-token')
			expect(state.userRoles).toEqual(['employee'])
		})

		it('setUser с пустыми полями', () => {
			useAuthStore.getState().setAuth(
				'token',
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
				'token',
				{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
				['employee']
			)

			useAuthStore.getState().setLoading(true)

			const state = useAuthStore.getState()
			expect(state.isLoading).toBe(true)
			expect(state.isAuthenticated).toBe(true)
			expect(state.token).toBe('token')
		})
	})
})
