import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { apiRequest } from '@/api/client'
import { useAuthStore } from '@/stores/authStore'

describe('apiRequest', () => {
	beforeEach(() => {
		vi.restoreAllMocks()
		useAuthStore.getState().clearAuth()
	})

	afterEach(() => {
		vi.useRealTimers()
	})

	it('добавляет Authorization header с токеном', async () => {
		useAuthStore.getState().setAuth(
			'test-token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			[]
		)

		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ data: 'test' }),
				})) as any
		)

		await apiRequest('/api/v1/test')

		expect(fetch).toHaveBeenCalledWith(
			'/api/v1/test',
			expect.objectContaining({
				headers: expect.objectContaining({
					Authorization: 'Bearer test-token',
				}),
			})
		)
	})

	it('работает без токена если не авторизован', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ data: 'test' }),
				})) as any
		)

		await apiRequest('/api/v1/public')

		expect(fetch).toHaveBeenCalledWith(
			'/api/v1/public',
			expect.objectContaining({
				headers: expect.not.objectContaining({
					Authorization: expect.any(String),
				}),
			})
		)
	})

	it('перенаправляет на /login при 401', async () => {
		const mockLocation = {
			...window.location,
			href: '',
		}
		Object.defineProperty(window, 'location', {
			value: mockLocation,
			writable: true,
			configurable: true,
		})

		vi.spyOn(window, 'fetch').mockImplementation(
			(() => Promise.resolve({ status: 401 })) as any
		)

		await expect(apiRequest('/api/v1/test')).rejects.toThrow('Unauthorized')
		expect(window.location.href).toBe('/login')
		expect(useAuthStore.getState().isAuthenticated).toBe(false)
	})

	it('бросает ошибку Forbidden при 403', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() => Promise.resolve({ status: 403 })) as any
		)

		await expect(apiRequest('/admin/test')).rejects.toThrow('Forbidden')
	})

	it('возвращает null при 204 NoContent', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() => Promise.resolve({ status: 204 })) as any
		)

		const result = await apiRequest('/admin/test/123')
		expect(result).toBeNull()
	})

	it('повторяет запрос при 5xx с exponential backoff', async () => {
		let callCount = 0
		vi.spyOn(window, 'fetch').mockImplementation(
			((() => {
				callCount++
				if (callCount < 3) {
					return Promise.resolve({ status: 500 })
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ success: true }),
				})
			}) as any)
		)

		vi.useFakeTimers()

		const promise = apiRequest('/api/v1/test')

		// Advance through retry delays: 1000ms (attempt 0) + 2000ms (attempt 1)
		await vi.advanceTimersByTimeAsync(1000)
		await vi.advanceTimersByTimeAsync(2000)

		const result = await promise
		expect(callCount).toBe(3)
		expect(result).toEqual({ success: true })
	})

	it('бросает ApiError с status при не-OK ответе', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: false,
					status: 422,
					text: () => Promise.resolve('Validation error'),
				})) as any
		)

		await expect(apiRequest('/api/v1/test')).rejects.toMatchObject({
			message: 'Validation error',
			status: 422,
		})
	})

	it('передаёт custom headers', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ ok: true }),
				})) as any
		)

		await apiRequest('/api/v1/test', {
			method: 'POST',
			headers: { 'X-Custom': 'value' },
			body: JSON.stringify({ key: 'val' }),
		})

		expect(fetch).toHaveBeenCalledWith(
			'/api/v1/test',
			expect.objectContaining({
				method: 'POST',
				headers: expect.objectContaining({
					'X-Custom': 'value',
				}),
				body: JSON.stringify({ key: 'val' }),
			})
		)
	})

	// =============================================================================
	// EDGE CASE TESTS
	// =============================================================================

	it('5xx retry — все 3 попытки исчерпаны, бросает ошибку', async () => {
		let callCount = 0
		vi.spyOn(window, 'fetch').mockImplementation(
			((() => {
				callCount++
				return Promise.resolve({
					status: 500,
					ok: false,
					text: () => Promise.resolve('Internal Server Error'),
				})
			}) as any)
		)

		vi.useFakeTimers()

		let errorThrown: any = null
		const promise = apiRequest('/api/v1/test').catch((err) => {
			errorThrown = err
		})

		// Advance through all retry delays: 1000ms + 2000ms
		await vi.advanceTimersByTimeAsync(1000)
		await vi.advanceTimersByTimeAsync(2000)

		await promise

		expect(errorThrown).toBeTruthy()
		expect(errorThrown!.status).toBe(500)
		expect(callCount).toBe(3) // initial + 2 retries
	})

	it('502 Bad Gateway retry', async () => {
		let callCount = 0
		vi.spyOn(window, 'fetch').mockImplementation(
			((() => {
				callCount++
				if (callCount < 2) {
					return Promise.resolve({ status: 502 })
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ ok: true }),
				})
			}) as any)
		)

		vi.useFakeTimers()

		const promise = apiRequest('/api/v1/test')
		await vi.advanceTimersByTimeAsync(1000)

		const result = await promise
		expect(callCount).toBe(2)
		expect(result).toEqual({ ok: true })
	})

	it('503 Service Unavailable retry', async () => {
		let callCount = 0
		vi.spyOn(window, 'fetch').mockImplementation(
			((() => {
				callCount++
				if (callCount < 2) {
					return Promise.resolve({ status: 503 })
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ recovered: true }),
				})
			}) as any)
		)

		vi.useFakeTimers()

		const promise = apiRequest('/api/v1/test')
		await vi.advanceTimersByTimeAsync(1000)

		const result = await promise
		expect(result).toEqual({ recovered: true })
	})

	it('timeout handling — AbortError бросается', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() => {
				const error = new Error('The operation was aborted')
				error.name = 'AbortError'
				return Promise.reject(error)
			}) as any
		)

		vi.useFakeTimers()

		await expect(apiRequest('/api/v1/test')).rejects.toMatchObject({
			name: 'AbortError',
		})
	})

	it('network error — бросает ошибку', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() => Promise.reject(new Error('Network error'))) as any
		)

		await expect(apiRequest('/api/v1/test')).rejects.toThrow('Network error')
	})

	it('abort signal передаётся в fetch', async () => {
		let capturedSignal: AbortSignal | undefined

		vi.spyOn(window, 'fetch').mockImplementation(
			(((_input: any, _init?: any) => {
				const init = _init as RequestInit | undefined
				capturedSignal = (init?.signal as AbortSignal) ?? undefined
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ ok: true }),
				})
			}) as any)
		)

		await apiRequest('/api/v1/test')

		expect(capturedSignal).toBeDefined()
		expect(capturedSignal).toBeInstanceOf(AbortSignal)
	})

	it('race condition — двойной вызов apiRequest', async () => {
		let callCount = 0
		vi.spyOn(window, 'fetch').mockImplementation(
			((() => {
				callCount++
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ call: callCount }),
				})
			}) as any)
		)

		const [result1, result2] = await Promise.all([
			apiRequest('/api/v1/test'),
			apiRequest('/api/v1/test'),
		])

		expect(callCount).toBe(2)
		// Both requests complete, but call numbers may vary due to concurrency
		expect(result1).toBeDefined()
		expect(result2).toBeDefined()
	})

	it('422 Unprocessable Entity возвращает ApiError', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: false,
					status: 422,
					text: () => Promise.resolve('{"error":"validation failed"}'),
				})) as any
		)

		await expect(apiRequest('/api/v1/test')).rejects.toMatchObject({
			message: '{"error":"validation failed"}',
			status: 422,
		})
	})

	it('404 Not Found возвращает ApiError', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: false,
					status: 404,
					text: () => Promise.resolve('Not found'),
				})) as any
		)

		await expect(apiRequest('/api/v1/nonexistent')).rejects.toMatchObject({
			message: 'Not found',
			status: 404,
		})
	})

	it('401 очищает auth store', async () => {
		const mockLocation = {
			...window.location,
			href: '',
		}
		Object.defineProperty(window, 'location', {
			value: mockLocation,
			writable: true,
			configurable: true,
		})

		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['employee']
		)

		vi.spyOn(window, 'fetch').mockImplementation(
			(() => Promise.resolve({ status: 401 })) as any
		)

		await expect(apiRequest('/api/v1/test')).rejects.toThrow('Unauthorized')

		expect(useAuthStore.getState().isAuthenticated).toBe(false)
		expect(useAuthStore.getState().token).toBeNull()
		expect(useAuthStore.getState().user).toBeNull()
		expect(useAuthStore.getState().userRoles).toEqual([])
	})

	it('403 не очищает auth store', async () => {
		useAuthStore.getState().setAuth(
			'token',
			{ id: '1', email: 't@t.com', first_name: 'T', last_name: 'T' },
			['employee']
		)

		vi.spyOn(window, 'fetch').mockImplementation(
			(() => Promise.resolve({ status: 403 })) as any
		)

		await expect(apiRequest('/api/v1/admin')).rejects.toThrow('Forbidden')

		// Auth state should be preserved on 403
		expect(useAuthStore.getState().isAuthenticated).toBe(true)
		expect(useAuthStore.getState().token).toBe('token')
	})

	it('200 с пустым JSON', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: true,
					json: () => Promise.resolve({}),
				})) as any
		)

		const result = await apiRequest('/api/v1/test')
		expect(result).toEqual({})
	})

	it('200 с массивом данных', async () => {
		const mockData = [
			{ id: '1', name: 'Item 1' },
			{ id: '2', name: 'Item 2' },
		]

		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: true,
					json: () => Promise.resolve(mockData),
				})) as any
		)

		const result = await apiRequest('/api/v1/test')
		expect(result).toEqual(mockData)
	})

	it('200 с null в JSON', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() =>
				Promise.resolve({
					ok: true,
					json: () => Promise.resolve(null),
				})) as any
		)

		const result = await apiRequest('/api/v1/test')
		expect(result).toBeNull()
	})

	it('добавляет Accept и Content-Type headers', async () => {
		let capturedHeaders: Record<string, string> | undefined

		vi.spyOn(window, 'fetch').mockImplementation(
			(((_input: any, _init?: any) => {
				const init = _init as RequestInit | undefined
				const h = init?.headers as Record<string, string> | undefined
				capturedHeaders = h || {}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ ok: true }),
				})
			}) as any)
		)

		await apiRequest('/api/v1/test')

		expect(capturedHeaders?.['Accept']).toBe('application/json')
		expect(capturedHeaders?.['Content-Type']).toBe('application/json')
	})

	it('POST метод передаётся корректно', async () => {
		let capturedMethod: string | undefined

		vi.spyOn(window, 'fetch').mockImplementation(
			(((_input: any, _init?: any) => {
				const init = _init as RequestInit | undefined
				capturedMethod = init?.method
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ ok: true }),
				})
			}) as any)
		)

		await apiRequest('/api/v1/test', { method: 'POST', body: JSON.stringify({ key: 'val' }) })

		expect(capturedMethod).toBe('POST')
	})

	it('DELETE метод с 204', async () => {
		vi.spyOn(window, 'fetch').mockImplementation(
			(() => Promise.resolve({ status: 204 })) as any
		)

		const result = await apiRequest('/api/v1/test/123', { method: 'DELETE' })
		expect(result).toBeNull()
	})
})
