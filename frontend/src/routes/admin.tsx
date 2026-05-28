// Admin routes — константы и типы для маршрутов администратора
// Используется в навигации (T2105) и App.tsx

export const adminRoutes = [
	{ path: '/admin/hr', label: 'HR', icon: 'users', roles: ['hr', 'admin'] as const },
	{ path: '/admin/catalog', label: 'Каталог', icon: 'grid', roles: ['catalog_manager', 'admin'] as const },
	{ path: '/admin/content', label: 'Контент', icon: 'file-text', roles: ['admin'] as const },
] as const

export type AdminRoute = (typeof adminRoutes)[number]
