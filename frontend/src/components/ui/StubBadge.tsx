import { Tooltip } from '@mantine/core'

/**
 * Индикатор заглушки — показывается ТОЛЬКО в dev-режиме.
 * В production — null.
 */
export function StubBadge() {
	if (!import.meta.env.DEV) return null

	return (
		<Tooltip label="Заглушка — данные появятся после подключения API">
			<div
				style={{
					width: 18,
					height: 18,
					borderRadius: '50%',
					background: '#FEE2E2',
					color: '#DC2626',
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center',
					fontSize: 11,
					fontWeight: 700,
					cursor: 'help',
					flexShrink: 0,
				}}
			>
				?
			</div>
		</Tooltip>
	)
}
