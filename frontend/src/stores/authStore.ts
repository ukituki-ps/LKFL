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

// Ключи localStorage для persist auth-состояния между перезагрузками страницы
const LS_TOKEN = 'lkfl_token'
const LS_USER = 'lkfl_user'
const LS_ROLES = 'lkfl_roles'

interface AuthState {
	// Состояние
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

// Восстановление auth-состояния из localStorage (после перезагрузки страницы)
function restoreAuth(): { token: string | null; user: UserProfile | null; userRoles: UserRole[]; isAuthenticated: boolean } {
	const token = localStorage.getItem(LS_TOKEN)
	const userRaw = localStorage.getItem(LS_USER)
	const rolesRaw = localStorage.getItem(LS_ROLES)

	if (token && userRaw && rolesRaw) {
		try {
			const user = JSON.parse(userRaw) as UserProfile
			const userRoles = JSON.parse(rolesRaw) as UserRole[]
			return { token, user, userRoles, isAuthenticated: true }
		} catch {
			// Коррумпированные данные — очистка
			localStorage.removeItem(LS_TOKEN)
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

export const useAuthStore = create<AuthState>((set, get) => ({
	token: restored.token,
	user: restored.user,
	userRoles: restored.userRoles,
	isAuthenticated: restored.isAuthenticated,
	isLoading: false,

	// Установка auth-состояния после успешного логина
	setAuth: (token, user, roles) => {
		localStorage.setItem(LS_TOKEN, token)
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

	// Логаут: инвалидация сессии на backend + очистка store
	logout: async () => {
		const { token } = get()
		try {
			await fetch('/api/v1/auth/logout', {
				method: 'POST',
				headers: { Authorization: `Bearer ${token}` },
			})
		} catch {
			// Ignored — очищаем state в любом случае
		}
		localStorage.removeItem(LS_TOKEN)
		localStorage.removeItem(LS_USER)
		localStorage.removeItem(LS_ROLES)
		set({ token: null, user: null, userRoles: [], isAuthenticated: false })
	},

	// Мгновенный сброс состояния (без запроса к backend)
	clearAuth: () => {
		localStorage.removeItem(LS_TOKEN)
		localStorage.removeItem(LS_USER)
		localStorage.removeItem(LS_ROLES)
		set({ token: null, user: null, userRoles: [], isAuthenticated: false })
	},
}))

// Проверка сессии: вызывает /api/v1/auth/me и возвращает профиль или null
export async function checkAuthSession(token: string): Promise<UserProfile | null> {
	try {
		const res = await fetch('/api/v1/auth/me', {
			headers: { Authorization: `Bearer ${token}` },
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
