import { Link, useLocation } from 'react-router-dom'
import { Group, Text, UnstyledButton } from '@mantine/core'
import { useAuthStore } from '@/stores/authStore'
import { adminRoutes } from '@/routes/admin'

interface AdminNavProps {
	onClose?: () => void
}

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
								padding: '8px 12px',
								borderRadius: 8,
								backgroundColor: isActive
									? 'var(--mantine-color-blue-light)'
									: 'transparent',
								color: isActive
									? 'var(--mantine-color-blue-7)'
									: 'var(--mantine-color-dark-6)',
								textDecoration: 'none',
								fontWeight: isActive ? 600 : 400,
							}}
						>
							<Group gap="sm" justify="space-between" style={{ width: '100%' }}>
								<Text size="sm">{item.label}</Text>
							</Group>
						</UnstyledButton>
					</Link>
				)
			})}
		</nav>
	)
}
