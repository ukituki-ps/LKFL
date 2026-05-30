import { createTheme, MantineThemeOverride } from '@mantine/core'

/**
 * Создаёт Mantine тему на основе brand tokens.
 *
 * White-label: CSS переменные бренда загружаются с backend
 * (GET /admin/tenants/{id}/brand) и переопределяют default tokens.
 */
export function createAprilTheme(): MantineThemeOverride {
	return createTheme({
		/* Кастомная зелёная шкала бренда (#00B33C) */
		colors: {
			brand: [
				'#F0FDF4', // 0
				'#DCFCE7', // 1
				'#BBF7D0', // 2
				'#86EFAC', // 3
				'#4ADE80', // 4
				'#22C55E', // 5
				'#16A34A', // 6
				'#00B33C', // 7
				'#009A33', // 8
				'#00651E', // 9
			],
		},
		primaryColor: 'brand',
		fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
		fontFamilyMonospace: 'var(--april-font-mono, "Fira Code", monospace)',
		headings: {
			fontFamily: 'Inter, sans-serif',
			fontWeight: '800',
		},
		defaultRadius: 'md',
		cursorType: 'default',
		components: {
			// Global button styling
			Button: {
				defaultProps: {
					radius: 'md',
				},
			},
			Card: {
				defaultProps: {
					padding: 'lg',
					radius: 'md',
				},
			},
		},
	})
}
