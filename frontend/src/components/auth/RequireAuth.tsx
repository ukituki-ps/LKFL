import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore, type UserRole } from '@/stores/authStore'

interface RequireAuthProps {
	roles?: string[] // если пустой массив — любой авторизованный
}

export function RequireAuth({ roles = [] }: RequireAuthProps) {
	const { isAuthenticated, userRoles } = useAuthStore()
	const location = useLocation()

	if (!isAuthenticated) {
		// Сохраняем attempted URL для редиректа после логина
		return <Navigate to="/login" state={{ from: location }} replace />
	}

	// Проверка ролей (RBAC)
	if (roles.length > 0) {
		const safeRoles = userRoles ?? []
		const hasRole = roles.some((role) => safeRoles.includes(role as UserRole))
		if (!hasRole) {
			return <Navigate to="/forbidden" replace />
		}
	}

	return <Outlet />
}
