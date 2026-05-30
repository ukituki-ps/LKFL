import { Outlet, useLocation, useNavigate } from 'react-router-dom'
import { AprilProductHeader, AprilIconPanelLeft } from '@ukituki-ps/april-ui'
import { Drawer, Group, Text, UnstyledButton, Stack } from '@mantine/core'
import { useDisclosure, useMediaQuery } from '@mantine/hooks'
import { useAuthStore } from '@/stores/authStore'
import { employeeRoutes } from '@/routes/employee'
import { adminRoutes } from '@/routes/admin'
import { HeaderNav } from './HeaderNav'
import { HeaderRight } from './HeaderRight'

/**
 * Корневой layout приложения.
 *
 * Горизонтальный layout с AprilProductHeader (DS).
 * Sidebar убран (desktop). Мобильная навигация — Burger → Drawer.
 */
export function Shell() {
	const [drawerOpened, { open: openDrawer, close: closeDrawer }] = useDisclosure(false)
	const { userRoles } = useAuthStore()
	const location = useLocation()
	const navigate = useNavigate()
	const isMobile = useMediaQuery('(max-width: 768px)')

	const isAdminRoute = location.pathname.startsWith('/admin')
	const hasAdminAccess = userRoles.some((r) =>
		['admin', 'catalog_manager', 'hr'].includes(r),
	)

	return (
		<div
			style={{
				minHeight: '100vh',
				backgroundColor: 'var(--brand-bg, #F2F2F2)',
			}}
		>
			{/* Header */}
			<AprilProductHeader
				left={
					<UnstyledButton
						onClick={() => navigate(isAdminRoute ? '/admin/hr' : '/')}
						style={{ display: 'flex', alignItems: 'center', gap: 8, padding: 0 }}
					>
						<Text
							style={{
								fontWeight: 700,
								fontSize: 18,
								color: 'var(--brand-text, #1A1A1A)',
							}}
						>
							ЛКФЛ
						</Text>
					</UnstyledButton>
				}
				center={
					!isMobile && (
						<HeaderNav isAdmin={hasAdminAccess && isAdminRoute} />
					)
				}
				right={
					<Group gap={8}>
						{isMobile && (
							<UnstyledButton
								onClick={openDrawer}
								style={{
									display: 'flex',
									alignItems: 'center',
									justifyContent: 'center',
									width: 34,
									height: 34,
									backgroundColor: 'var(--brand-row, #F9FAFB)',
									borderRadius: 9999,
								}}
							>
								<AprilIconPanelLeft size={18} />
							</UnstyledButton>
						)}
						<HeaderRight />
					</Group>
				}
				sticky
			/>

			{/* Main content */}
			<main
				style={{
					maxWidth: 1100,
					margin: '0 auto',
					padding: '28px 28px 56px',
				}}
			>
				<Outlet />
			</main>

			{/* Mobile Drawer */}
			{isMobile && (
				<Drawer
					opened={drawerOpened}
					onClose={closeDrawer}
					position="left"
					title="Меню"
					size="xs"
					padding="md"
				>
					<Stack gap="xs">
						{(hasAdminAccess && isAdminRoute
							? adminRoutes.filter((item) =>
									item.roles.some((role) =>
										userRoles.includes(role as never),
									),
								)
							: employeeRoutes
						).map((item) => {
							const isActive = location.pathname === item.path
							return (
								<UnstyledButton
									key={item.path}
									onClick={() => {
										navigate(item.path)
										closeDrawer()
									}}
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
										fontWeight: isActive ? 600 : 400,
										fontSize: 14,
										textDecoration: 'none',
									}}
								>
									{item.label}
								</UnstyledButton>
							)
						})}
					</Stack>
				</Drawer>
			)}
		</div>
	)
}
