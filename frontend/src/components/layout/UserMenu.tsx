import { useAuthStore } from '@/stores/authStore'
import { Text, Menu, Avatar } from '@mantine/core'

export function UserMenu() {
	const { user, logout, userRoles } = useAuthStore()

	const getInitials = () => {
		if (!user) return '?'
		return `${user.first_name?.[0] || ''}${user.last_name?.[0] || ''}`.toUpperCase()
	}

	const getRoleLabel = () => {
		if (userRoles.length === 0) return 'Сотрудник'
		const labels: Record<string, string> = {
			employee: 'Сотрудник',
			catalog_manager: 'Менеджер каталога',
			hr: 'HR',
			admin: 'Администратор',
		}
		return userRoles
			.map((r) => labels[r] || r)
			.join(', ')
	}

	return (
		<Menu shadow="md" width={220} position="bottom-end">
			<Menu.Target>
				<Avatar
					size={34}
					color="brand"
					style={{
						backgroundColor: 'var(--brand-green, #00B33C)',
						fontSize: 11,
						fontWeight: 600,
						borderRadius: '50%',
						color: '#FFFFFF',
					}}
				>
					{getInitials()}
				</Avatar>
			</Menu.Target>
			<Menu.Dropdown>
				<Menu.Item>
					<Text size="sm" fw={500}>
						{user?.first_name} {user?.last_name}
					</Text>
					<Text size="xs" c="dimmed">
						{getRoleLabel()}
					</Text>
				</Menu.Item>
				<Menu.Item>
					<Text size="xs" c="dimmed">
						{user?.email}
					</Text>
				</Menu.Item>
				<Menu.Divider />
				<Menu.Item onClick={logout}>Выйти</Menu.Item>
			</Menu.Dropdown>
		</Menu>
	)
}
