import { Link, useLocation } from 'react-router-dom'
import { Text, UnstyledButton } from '@mantine/core'
import { useAuthStore } from '@/stores/authStore'
import { adminRoutes } from '@/routes/admin'

interface AdminNavProps {
	onClose?: () => void
}

/**
 * Навигация администратора — вертикальный список.
 * Используется только в мобильном Drawer (Shell.tsx).
 */
export function AdminNav({ onClose }: AdminNavProps) {
	const location = useLocation()
	const { userRoles } = useAuthStore()

	const visibleItems = adminRoutes.filter((item) =>
		item.roles.some((role) => userRoles.includes(role as never)),
	)

	return (
		<nav>
			{visibleItems.map((item) => {
				const isActive = location.pathname === item.path
				return (
					<Link
						key={item.path}
						to={item.path}
						onClick={onClose}
						style={{ textDecoration: 'none' }}
					>
						<UnstyledButton
							style={{
								width: '100%',
								display: 'flex',
								alignItems: 'center',
								padding: '10px 12px',
								borderRadius: 8,
								backgroundColor: isActive
									? 'var(--brand-green-light, #F0FDF4)'
									: 'transparent',
								color: isActive
									? 'var(--brand-green, #00B33C)'
									: 'var(--brand-text-muted, #6B7280)',
								textDecoration: 'none',
								fontWeight: isActive ? 600 : 400,
								fontSize: 14,
							}}
						>
							<Text size="sm">{item.label}</Text>
						</UnstyledButton>
					</Link>
				)
			})}
		</nav>
	)
}
