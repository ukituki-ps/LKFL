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

// Ключи localStorage для persist user+roles между перезагрузками (без токена).
// Токен хранится в httpOnly cookie на backend.
const LS_USER = 'lkfl_user'
const LS_ROLES = 'lkfl_roles'

interface AuthState {
	// Состояние
	user: UserProfile | null
	userRoles: UserRole[]
	isAuthenticated: boolean
	isLoading: boolean

	// Actions
	setAuth: (user: UserProfile, roles: UserRole[]) => void
	setUser: (user: UserProfile) => void
	setLoading: (loading: boolean) => void
	logout: () => Promise<void>
	clearAuth: () => void
}

// Восстановление auth-состояния из localStorage (после перезагрузки страницы).
// Токен отсутствует (httpOnly cookie) — frontend полагается на cookie для авторизации запросов.
function restoreAuth(): { user: UserProfile | null; userRoles: UserRole[]; isAuthenticated: boolean } {
	const userRaw = localStorage.getItem(LS_USER)
	const rolesRaw = localStorage.getItem(LS_ROLES)

	if (userRaw && rolesRaw) {
		try {
			const user = JSON.parse(userRaw) as UserProfile
			const userRoles = JSON.parse(rolesRaw) as UserRole[]
			return { user, userRoles, isAuthenticated: true }
		} catch {
			// Коррумпированные данные — очистка
			localStorage.removeItem(LS_USER)
			localStorage.removeItem(LS_ROLES)
		}
	}
	return { user: null, userRoles: [], isAuthenticated: false }
}

// Глобальная функция для E2E тестов: установка auth-состояния
// Вызывается через page.evaluate() в Playwright тестах
export function setupAuthForTest(
	user: UserProfile,
	roles: UserRole[],
): void {
	useAuthStore.setState({ user, userRoles: roles, isAuthenticated: true, isLoading: false })
}

const restored = restoreAuth()

export const useAuthStore = create<AuthState>((set, _get) => ({
	user: restored.user,
	userRoles: restored.userRoles,
	isAuthenticated: restored.isAuthenticated,
	isLoading: false,

	// Установка auth-состояния после успешного логина.
	// Токен НЕ передаётся — сессия в httpOnly cookie.
	setAuth: (user, roles) => {
		localStorage.setItem(LS_USER, JSON.stringify(user))
		localStorage.setItem(LS_ROLES, JSON.stringify(roles))
		set({ user, userRoles: roles, isAuthenticated: true, isLoading: false })
	},

	// Обновление данных пользователя (без смены сессии)
	setUser: (user) => {
		localStorage.setItem(LS_USER, JSON.stringify(user))
		set({ user })
	},

	// Управление состоянием загрузки
	setLoading: (loading) => set({ isLoading: loading }),

	// Логаут: browser-based Keycloak SSO invalidation через redirect.
	//
	// Backend (POST /api/v1/auth/logout) удаляет server-side session (Redis)
	// и session cookie, затем возвращает 302 redirect на Keycloak logout endpoint.
	// Keycloak инвалидирует SSO сессию (удаляет KAUTH_SESSION_ID cookie) и
	// redirect-ит обратно на frontend /login.
	//
	// Это гарантирует, что повторный вход потребует ввода логина/пароля
	// (в отличие от server-side инвалидации, которая не трогает browser cookie).
	//
	// Очистка localStorage + sessionStorage происходит ДО redirect, чтобы
	// при сбое backend/Keycloak пользователь всё равно был выведен.
	logout: async () => {
		// 1. Очистка frontend state (до redirect — на случай сбоев)
		localStorage.removeItem(LS_USER)
		localStorage.removeItem(LS_ROLES)
		set({ user: null, userRoles: [], isAuthenticated: false })

		// Очистка артефактов login-потока:
		// - lkfl_login_redirecting — guard от дублирования redirect (StrictMode)
		// - lkfl_login_attempts — счётчик попыток callback
		sessionStorage.removeItem('lkfl_login_redirecting')
		sessionStorage.removeItem('lkfl_login_attempts')

		// Флаг для Login.tsx: показать кнопку "Войти" вместо авто-редиректа
		sessionStorage.setItem('lkfl_just_logged_out', 'true')

		// 2. Browser-based logout: backend → 302 Keycloak logout → /login
		//
		// Используем window.location.href (не fetch), потому что:
		// - Backend возвращает 302 redirect на Keycloak logout
		// - Keycloak инвалидирует SSO сессию и удаляет KAUTH_SESSION_ID cookie
		// - Только browser redirect может это сделать (fetch не следует redirect'ам
		//   к другим доменам и не обрабатывает Set-Cookie от Keycloak)
		window.location.href = '/api/v1/auth/logout'
	},

	// Мгновенный сброс состояния (без запроса к backend)
	clearAuth: () => {
		localStorage.removeItem(LS_USER)
		localStorage.removeItem(LS_ROLES)
		set({ user: null, userRoles: [], isAuthenticated: false })
	},
}))

// Проверка сессии: вызывает /api/v1/auth/me и возвращает профиль или null.
// Использует cookie (credentials: 'include') для авторизации.
export async function checkAuthSession(): Promise<UserProfile | null> {
	try {
		const res = await fetch('/api/v1/auth/me', {
			credentials: 'include',
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
