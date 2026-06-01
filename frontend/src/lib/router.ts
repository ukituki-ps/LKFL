/**
 * Глобальный навигатор для использования вне React-компонентов.
 *
 * Позволяет Zustand store (authStore) использовать React Router навигацию
 * вместо window.location.href, что предотвращает полный перезагрузок страницы
 * и корректно работает с SPA-маршрутизацией.
 *
 * Инициализируется в Shell.tsx через setGlobalNavigator().
 */
import type { NavigateFunction } from 'react-router-dom'

let globalNavigator: NavigateFunction | null = null

export function setGlobalNavigator(nav: NavigateFunction) {
	globalNavigator = nav
}

/**
 * Навигация вне React-компонента (например, из Zustand store).
 * Если глобальный навигатор не установлен — fallback на window.location.
 */
export function navigateOutside(path: string) {
	if (globalNavigator) {
		globalNavigator(path)
	} else {
		// Fallback: полный перезагрузок (должен срабатывать только
		// в тестах или до инициализации Shell)
		window.location.href = path
	}
}
