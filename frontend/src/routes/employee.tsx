// Employee routes — константы и типы для маршрутов сотрудника
// Используется в навигации (T2105) и App.tsx

export const employeeRoutes = [
	{ path: '/', label: 'Главная', icon: 'home' },
	{ path: '/catalog', label: 'Каталог льгот', icon: 'grid' },
	{ path: '/points', label: 'Баллы', icon: 'star' },
	{ path: '/documents', label: 'Документы', icon: 'file' },
	{ path: '/support', label: 'Поддержка', icon: 'help' },
] as const

export type EmployeeRoute = (typeof employeeRoutes)[number]
