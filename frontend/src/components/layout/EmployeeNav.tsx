import { Link, useLocation } from 'react-router-dom'
import { Group, Text, UnstyledButton } from '@mantine/core'
import { employeeRoutes } from '@/routes/employee'

interface EmployeeNavProps {
	onClose?: () => void
}

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
