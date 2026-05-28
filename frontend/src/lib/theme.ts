import { createTheme, MantineThemeOverride } from '@mantine/core'

/**
 * Создаёт Mantine тему на основе April tokens.
 *
 * April tokens (@ukituki-ps/april-tokens) предоставляют CSS переменные:
 * --april-color-primary, --april-color-secondary, --april-font-family, etc.
 *
 * White-label: CSS переменные бренда загружаются с backend
 * (GET /admin/tenants/{id}/brand) и переопределяют default April tokens.
 *
 * В M21 используем default April tokens.
 */
export function createAprilTheme(): MantineThemeOverride {
	return createTheme({
		// April tokens подтягиваются из CSS variables
		// White-label brand CSS будет переопределять эти значения
		primaryColor: 'blue',
		fontFamily:
			'var(--april-font-family, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif)',
		fontFamilyMonospace: 'var(--april-font-mono, "Fira Code", monospace)',
		headings: {
			fontFamily: 'var(--april-font-family, inherit)',
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
