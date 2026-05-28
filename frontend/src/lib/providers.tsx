import { ReactNode } from 'react'
import { MantineProvider } from '@mantine/core'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createAprilTheme } from './theme'

// Создаём тему на основе April tokens
const aprilTheme = createAprilTheme()

// React Query клиент с дефолтными настройками
const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			staleTime: 1000 * 60 * 5, // 5 min
			retry: 1,
			refetchOnWindowFocus: false,
		},
		mutations: {
			retry: 0,
		},
	},
})

interface AprilProvidersProps {
	children: ReactNode
}

/**
 * Корневой провайдер приложения.
 *
 * Объединяет:
 * - MantineProvider с April темой
 * - QueryClientProvider для React Query
 *
 * White-label brand CSS variables применяются через CSS injection
 * из backend (TODO M22+).
 */
export function AprilProviders({ children }: AprilProvidersProps) {
	return (
		<MantineProvider theme={aprilTheme}>
			<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
		</MantineProvider>
	)
}

// Экспортируем queryClient для использования в тестах и devtools
export { queryClient }
