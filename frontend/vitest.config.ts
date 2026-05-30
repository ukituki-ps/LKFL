import { defineConfig } from 'vitest/config'
import path from 'path'

export default defineConfig({
	resolve: {
		alias: {
			'@': path.resolve(__dirname, './src'),
		},
	},
	test: {
		globals: true,
		environment: 'jsdom',
		setupFiles: ['./src/test/setup.ts'],
		include: ['src/**/*.test.{ts,tsx}'],
		exclude: [
			'**/node_modules/**',
			'**/dist/**',
			'**/e2e/**',
			'**/e2e/**/*.spec.ts',
			'**/.kilo/**',
		],
		coverage: {
			provider: 'v8',
			reporter: ['text', 'json', 'html'],
		},
		server: {
			deps: {
				inline: [
					'@ukituki-ps/april-ui',
					'mantine-vaul',
				],
			},
		},
	},
})
