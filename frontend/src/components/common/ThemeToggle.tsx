import { ActionIcon, Tooltip } from '@mantine/core'
import { useThemeStore } from '@/stores/themeStore'

/**
 * Переключатель темы dark/light.
 *
 * Кнопка ☀️ в dark-режиме (переключить на light), 🌙 в light-режиме (на dark).
 * Состояние сохраняется в localStorage (lkfl_theme).
 */
export function ThemeToggle() {
	const { colorScheme, toggle } = useThemeStore()

	return (
		<Tooltip label={colorScheme === 'light' ? 'Тёмная тема' : 'Светлая тема'}>
			<ActionIcon
				variant="subtle"
				color="dimmed"
				size={34}
				radius="xl"
				onClick={toggle}
				title={colorScheme === 'light' ? 'Тёмная тема' : 'Светлая тема'}
			>
				<span style={{ fontSize: 16, lineHeight: 1 }}>
					{colorScheme === 'light' ? '🌙' : '☀️'}
				</span>
			</ActionIcon>
		</Tooltip>
	)
}
