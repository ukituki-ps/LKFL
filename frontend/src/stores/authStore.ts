import { create } from 'zustand'

// Профиль пользователя (соответствует response от /api/v1/auth/callback.user)
export interface UserProfile {
	id: string
	email: string
	first_name: string
	last_name: string
}

// Роли пользователя (соответствуют backend/internal/auth)
export type UserRole = 'employee' | 'catalog_manager' | 'hr' | 'admin'

// D2: Token хранится в httpOnly cookie (backend), НЕ в localStorage.
// Ключи localStorage для persist user+roles между перезагрузками (без токена).
const LS_USER = 'lkfl_user'
const LS_ROLES = 'lkfl_roles'

interface AuthState {
	// Состояние
	// D2: token больше не хранится в store — используется httpOnly cookie.
	// Поле оставлено для обратной совместимости API (callback возвращает token).
	token: string | null
	user: UserProfile | null
	userRoles: UserRole[]
	isAuthenticated: boolean
	isLoading: boolean

	// Actions
	setAuth: (token: string, user: UserProfile, roles: UserRole[]) => void
	setUser: (user: UserProfile) => void
	setLoading: (loading: boolean) => void
	logout: () => Promise<void>
	clearAuth: () => void
}

// Восстановление auth-состояния из localStorage (после перезагрузки страницы).
// D2: восстанавливаем только user + roles; token отсутствует (httpOnly cookie).
// После перезагрузки frontend полагается на cookie для авторизации запросов.
function restoreAuth(): { token: string | null; user: UserProfile | null; userRoles: UserRole[]; isAuthenticated: boolean } {
	const userRaw = localStorage.getItem(LS_USER)
	const rolesRaw = localStorage.getItem(LS_ROLES)

	if (userRaw && rolesRaw) {
		try {
			const user = JSON.parse(userRaw) as UserProfile
			const userRoles = JSON.parse(rolesRaw) as UserRole[]
			return { token: null, user, userRoles, isAuthenticated: true }
		} catch {
			// Коррумпированные данные — очистка
			localStorage.removeItem(LS_USER)
			localStorage.removeItem(LS_ROLES)
		}
	}
	return { token: null, user: null, userRoles: [], isAuthenticated: false }
}

// Глобальная функция для E2E тестов: установка auth-состояния
// Вызывается через page.evaluate() в Playwright тестах
export function setupAuthForTest(
	token: string,
	user: UserProfile,
	roles: UserRole[],
): void {
	useAuthStore.setState({ token, user, userRoles: roles, isAuthenticated: true, isLoading: false })
}

const restored = restoreAuth()

export const useAuthStore = create<AuthState>((set, _get) => ({
	token: restored.token,
	user: restored.user,
	userRoles: restored.userRoles,
	isAuthenticated: restored.isAuthenticated,
	isLoading: false,

	// Установка auth-состояния после успешного логина.
	// D2: token НЕ записывается в localStorage — сессия в httpOnly cookie.
	setAuth: (token, user, roles) => {
		// D2: localStorage.setItem(LS_TOKEN, token) — УДАЛЕНО (XSS risk)
		localStorage.setItem(LS_USER, JSON.stringify(user))
		localStorage.setItem(LS_ROLES, JSON.stringify(roles))
		set({ token, user, userRoles: roles, isAuthenticated: true, isLoading: false })
	},

	// Обновление данных пользователя (без смены токена)
	setUser: (user) => {
		localStorage.setItem(LS_USER, JSON.stringify(user))
		set({ user })
	},

	// Управление состоянием загрузки
	setLoading: (loading) => set({ isLoading: loading }),

	// Логаут: инвалидация сессии на backend + очистка store.
	// D2: использует credentials: 'include' (cookie) вместо Bearer token.
	logout: async () => {
		try {
			await fetch('/api/v1/auth/logout', {
				method: 'POST',
				credentials: 'include', // D2: cookie-based session
			})
		} catch {
			// Ignored — очищаем state в любом случае
		}
		// D2: localStorage.removeItem(LS_TOKEN) — УДАЛЕНО (нет токена в LS)
		localStorage.removeItem(LS_USER)
		localStorage.removeItem(LS_ROLES)
		set({ token: null, user: null, userRoles: [], isAuthenticated: false })
	},

	// Мгновенный сброс состояния (без запроса к backend)
	clearAuth: () => {
		// D2: localStorage.removeItem(LS_TOKEN) — УДАЛЕНО (нет токена в LS)
		localStorage.removeItem(LS_USER)
		localStorage.removeItem(LS_ROLES)
		set({ token: null, user: null, userRoles: [], isAuthenticated: false })
	},
}))

// Проверка сессии: вызывает /api/v1/auth/me и возвращает профиль или null.
// D2: использует cookie (credentials: 'include') вместо Bearer token.
export async function checkAuthSession(_token?: string): Promise<UserProfile | null> {
	try {
		const res = await fetch('/api/v1/auth/me', {
			credentials: 'include', // D2: cookie-based auth
		})
		if (res.ok) {
			const data: UserProfile = await res.json()
			return data
		}
	} catch {
		// Ignored — сессия невалидна
	}
	return null
}
