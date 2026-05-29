import { useAuthStore } from '@/stores/authStore'
import { Group, Text, Menu, UnstyledButton, Avatar } from '@mantine/core'
import { useMediaQuery } from '@mantine/hooks'

export function UserMenu() {
	const { user, logout, userRoles } = useAuthStore()
	const isMobile = useMediaQuery('(max-width: 768px)')

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
		<Menu shadow="md" width={200} position="bottom-end">
			<Menu.Target>
				<UnstyledButton>
					<Group gap="sm">
						<Avatar
							size={34}
							radius="xl"
							color="brand"
							style={{
								backgroundColor: 'var(--brand-green, #00B33C)',
								fontSize: 11,
								fontWeight: 600,
							}}
						>
							{getInitials()}
						</Avatar>
						{!isMobile && (
							<div>
								<Text size="sm" fw={500}>
									{user?.first_name} {user?.last_name}
								</Text>
								<Text size="xs" c="dimmed">
									{getRoleLabel()}
								</Text>
							</div>
						)}
					</Group>
				</UnstyledButton>
			</Menu.Target>
			<Menu.Dropdown>
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
