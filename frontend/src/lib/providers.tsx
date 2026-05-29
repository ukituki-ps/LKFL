import { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AprilProviders as AprilProvidersDS } from '@ukituki-ps/april-ui'
import { createAprilTheme } from './theme'

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

interface LKFLProvidersProps {
	children: ReactNode
}

/**
 * Корневой провайдер приложения.
 *
 * Объединяет:
 * - AprilProviders (DS) — density context + Mantine + color scheme
 * - QueryClientProvider для React Query
 *
 * White-label brand CSS variables применяются через CSS injection
 * из backend (TODO M22+).
 */
export function LKFLProviders({ children }: LKFLProvidersProps) {
	return (
		<AprilProvidersDS theme={createAprilTheme()}>
			<QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
		</AprilProvidersDS>
	)
}

// Экспортируем queryClient для использования в тестах и devtools
export { queryClient }
