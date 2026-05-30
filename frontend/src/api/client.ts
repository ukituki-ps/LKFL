import { useAuthStore } from '@/stores/authStore'

const DEFAULT_TIMEOUT = 25_000 // 25s (серверный timeout 30s)
const MAX_RETRIES = 2
const RETRY_BASE_DELAY = 1000 // 1s

export interface ApiError extends Error {
	status?: number
}

export async function apiRequest<T>(
	url: string,
	options: RequestInit = {}
): Promise<T> {
	const headers: Record<string, string> = {
		'Accept': 'application/json',
		'Content-Type': 'application/json',
		...(options.headers as Record<string, string> || {}),
	}

	// AbortController для timeout
	const controller = new AbortController()
	const timeoutId = setTimeout(() => controller.abort(), DEFAULT_TIMEOUT)

	let lastError: ApiError | null = null

	for (let attempt = 0; attempt <= MAX_RETRIES; attempt++) {
		try {
			const response = await fetch(url, {
				...options,
				headers,
				signal: controller.signal,
				credentials: 'include', // D2: отправляем httpOnly cookie (lkfl_session)
			})

			clearTimeout(timeoutId)

			// 401 → redirect на login
			if (response.status === 401) {
				useAuthStore.getState().clearAuth()
				window.location.href = '/login'
				const err = new Error('Unauthorized') as ApiError
				err.status = 401
				throw err
			}

			// 403 → forbidden
			if (response.status === 403) {
				const err = new Error('Forbidden') as ApiError
				err.status = 403
				throw err
			}

			// 204 NoContent (delete operations)
			if (response.status === 204) {
				return null as T
			}

			// 5xx → retry с exponential backoff
			if (response.status >= 500 && attempt < MAX_RETRIES) {
				const delay = RETRY_BASE_DELAY * Math.pow(2, attempt)
				await new Promise(resolve => setTimeout(resolve, delay))
				continue
			}

			if (!response.ok) {
				const errorBody = await response.text()
				const err = new Error(errorBody || `HTTP ${response.status}`) as ApiError
				err.status = response.status
				throw err
			}

			const data = await response.json()
			return data as T
		} catch (error) {
			lastError = error as ApiError
			if (attempt < MAX_RETRIES && error instanceof Error && error.name === 'AbortError') {
				continue
			}
			throw error
		}
	}

	throw lastError || new Error('Request failed')
}
