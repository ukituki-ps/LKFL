import { Link, useLocation } from 'react-router-dom'
import { employeeRoutes } from '@/routes/employee'
import { adminRoutes } from '@/routes/admin'
import { useAuthStore } from '@/stores/authStore'

interface HeaderNavProps {
	isAdmin: boolean
}

/**
 * Горизонтальная навигация в header'е.
 * Для employee — 5 ссылок. Для admin — HR, Каталог, Контент.
 */
export function HeaderNav({ isAdmin }: HeaderNavProps) {
	const location = useLocation()
	const { userRoles } = useAuthStore()

	const routes = isAdmin
		? adminRoutes.filter((item) =>
				item.roles.some((role) => userRoles.includes(role as never)),
			)
		: employeeRoutes

	return (
		<nav style={{ display: 'flex', gap: '2px' }}>
			{routes.map((item) => {
				const isActive = location.pathname === item.path
				return (
					<Link
						key={item.path}
						to={item.path}
						style={{
							padding: '0 14px',
							fontSize: '13px',
							fontWeight: isActive ? 600 : 500,
							color: isActive
								? 'var(--brand-text)'
								: 'var(--brand-text-muted)',
							borderBottom: isActive
								? '2px solid var(--brand-green)'
								: '2px solid transparent',
							textDecoration: 'none',
							cursor: 'pointer',
							transition: 'color 150ms, border-color 150ms',
							display: 'flex',
							alignItems: 'center',
							height: '56px', // AprilProductHeader comfortable
						}}
						onMouseEnter={(e) => {
							if (!isActive) {
								(e.currentTarget.style.color =
									'var(--brand-green)'),
									(e.currentTarget.style.borderBottomColor =
										'transparent')
							}
						}}
						onMouseLeave={(e) => {
							if (!isActive) {
								(e.currentTarget.style.color =
									'var(--brand-text-muted)'),
									(e.currentTarget.style.borderBottomColor =
										'transparent')
							}
						}}
					>
						{item.label}
					</Link>
				)
			})}
		</nav>
	)
}
