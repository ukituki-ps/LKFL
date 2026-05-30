import { Group, ActionIcon, Tooltip } from '@mantine/core'
import { AprilIconBell, AprilIconCoins } from '@ukituki-ps/april-ui'
import { UserMenu } from './UserMenu'

/**
 * Правая зона header'а: баланс, уведомления, профиль.
 *
 * ГЭП-5: Balance pill — зелёная пилюля по стилям прототипа:
 *   background: #F0FDF4; border: 1px solid #BBF7D0; border-radius: 20px; color: #166534
 *   текст «N баллов»
 */
export function HeaderRight() {
	return (
		<Group gap={10}>
			{/* ГЭП-5: Balance pill — кастомная пилюля по прототипу */}
			<Tooltip label="Баланс баллов — подключится в F2">
				<div
					style={{
						display: 'flex',
						alignItems: 'center',
						gap: 6,
						background: '#F0FDF4',
						border: '1px solid #BBF7D0',
						borderRadius: 20,
						padding: '4px 12px',
						cursor: 'default',
					}}
				>
					<AprilIconCoins size={14} style={{ color: '#166534' }} />
					<Text
						style={{
							fontSize: 12,
							fontWeight: 600,
							color: '#166534',
						}}
					>
						1 250 баллов
					</Text>
				</div>
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

import { Text } from '@mantine/core'
