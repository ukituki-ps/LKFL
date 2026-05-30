import { Link, useLocation } from 'react-router-dom'
import { Text, UnstyledButton } from '@mantine/core'
import { employeeRoutes } from '@/routes/employee'

interface EmployeeNavProps {
	onClose?: () => void
}

/**
 * Навигация сотрудника — вертикальный список.
 * Используется только в мобильном Drawer (Shell.tsx).
 */
export function EmployeeNav({ onClose }: EmployeeNavProps) {
	const location = useLocation()

	return (
		<nav>
			{employeeRoutes.map((item) => {
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
