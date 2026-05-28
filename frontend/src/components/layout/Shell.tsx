import { Outlet, useLocation } from 'react-router-dom'
import { AppShell, Group, Burger } from '@mantine/core'
import { useDisclosure, useMediaQuery } from '@mantine/hooks'
import { useAuthStore } from '@/stores/authStore'
import { EmployeeNav } from './EmployeeNav'
import { AdminNav } from './AdminNav'
import { UserMenu } from './UserMenu'

export function Shell() {
	const [opened, { toggle, close }] = useDisclosure(false)
	const { userRoles } = useAuthStore()
	const location = useLocation()
	const isMobile = useMediaQuery('(max-width: 768px)')

	const isAdminRoute = location.pathname.startsWith('/admin')
	const hasAdminAccess = userRoles.some((r) =>
		['admin', 'catalog_manager', 'hr'].includes(r),
	)

	return (
		<AppShell
			header={{ height: 60 }}
			navbar={{ width: 250, breakpoint: 'sm' }}
			padding="md"
		>
			{/* Header */}
			<AppShell.Header>
				<Group h="100%" px="md" justify="space-between">
					<Group>
						{isMobile && <Burger opened={opened} onClick={toggle} size="sm" />}
						<span style={{ fontWeight: 600, fontSize: 20 }}>ЛКФЛ</span>
					</Group>
					<UserMenu />
				</Group>
			</AppShell.Header>

			{/* Sidebar */}
			<AppShell.Navbar p="md">
				{hasAdminAccess && isAdminRoute ? (
					<AdminNav onClose={close} />
				) : (
					<EmployeeNav onClose={close} />
				)}
			</AppShell.Navbar>

			{/* Main content */}
			<AppShell.Main>
				<Outlet />
			</AppShell.Main>
		</AppShell>
	)
}
