import { Group, Badge, ActionIcon, Tooltip } from '@mantine/core'
import { AprilIconBell, AprilIconCoins } from '@ukituki-ps/april-ui'
import { UserMenu } from './UserMenu'

/**
 * Правая зона header'а: баланс, уведомления, профиль.
 *
 * Баланс — mock-значение «1 250» (подключится в F2).
 * Колокольчик — заглушка (переход на /notifications — TODO F2).
 */
export function HeaderRight() {
	return (
		<Group gap={10}>
			{/* Balance pill — mock до F2 */}
			<Tooltip label="Баланс баллов — подключится в F2">
				<Badge
					leftSection={<AprilIconCoins size={14} />}
					variant="light"
					color="brand"
					size="sm"
					style={{
						cursor: 'default',
						fontWeight: 600,
						fontSize: 12,
					}}
				>
					1 250
				</Badge>
			</Tooltip>

			{/* Bell icon — заглушка */}
			<Tooltip label="Уведомления — подключится в F2">
				<ActionIcon
					variant="subtle"
					color="dimmed"
					size={34}
					radius="xl"
					style={{ backgroundColor: 'var(--brand-row)' }}
				>
					<AprilIconBell size={18} />
				</ActionIcon>
			</Tooltip>

			{/* User menu */}
			<UserMenu />
		</Group>
	)
}
