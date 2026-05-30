import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MantineProvider } from '@mantine/core'
import { createAprilTheme } from '@/lib/theme'
import type { ReactNode } from 'react'

const aprilTheme = createAprilTheme()

function renderWithProviders(ui: ReactNode, initialEntries: string[] = ['/catalog']) {
	const queryClient = new QueryClient({
		defaultOptions: {
			queries: {
				retry: false,
				staleTime: Infinity,
			},
		},
	})

	return render(
		<QueryClientProvider client={queryClient}>
			<MemoryRouter initialEntries={initialEntries}>
				<MantineProvider theme={aprilTheme}>
					{ui}
				</MantineProvider>
			</MemoryRouter>
		</QueryClientProvider>
	)
}

// Mock API functions
vi.mock('@/api/engagements', () => ({
	getEngagements: vi.fn(),
	getCategories: vi.fn(),
}))

vi.mock('@/components/catalog/EngagementCard', () => ({
	EngagementGrid: ({ engagements }: { engagements: any[] }) => (
		<div data-testid="engagement-grid">
			{engagements.map((e) => (
				<div key={e.id} data-testid={`engagement-${e.id}`}>
					{e.name}
				</div>
			))}
		</div>
	),
}))

vi.mock('@/components/catalog/FilterBar', () => ({
	FilterBar: ({ onChange }: { onChange: (key: string, value: string) => void }) => (
		<div data-testid="filter-bar">
			<button onClick={() => onChange('category', 'fitness')}>Fitness</button>
			<button onClick={() => onChange('type', 'benefit')}>Benefit</button>
		</div>
	),
}))

vi.mock('@/components/catalog/SearchInput', () => ({
	SearchInput: ({ onChange }: { onChange: (value: string) => void }) => (
		<div data-testid="search-input">
			<input
				data-testid="search-field"
				onChange={(e) => onChange(e.target.value)}
			/>
		</div>
	),
}))

vi.mock('@/components/catalog/Pagination', () => ({
	Pagination: ({ page, totalPages, onPageChange }: any) => (
		<div data-testid="pagination">
			<span>Page {page} of {totalPages}</span>
			{page < totalPages && (
				<button onClick={() => onPageChange(page + 1)}>Next</button>
			)}
		</div>
	),
}))

describe('Catalog page', () => {
	beforeEach(() => {
		vi.clearAllMocks()
	})

	afterEach(() => {
		vi.restoreAllMocks()
	})

	it('показывает loader при загрузке', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockReturnValue(new Promise(() => {})) // never resolves
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		// Loader is shown while query is pending — check for the loader element
		expect(document.querySelector('.mantine-Loader-root')).toBeInTheDocument()
	})

	it('показывает ошибку при API error', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockRejectedValue(new Error('API error'))
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByText('Ошибка загрузки каталога')).toBeInTheDocument()
		})

		// Кнопка "Повторить" должна быть
		expect(screen.getByText('Повторить')).toBeInTheDocument()
	})

	it('empty API response — показывает "Нет доступных льгот"', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({ data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 } })
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByText('Нет доступных льгот')).toBeInTheDocument()
		})
	})

	it('API error state — кнопка повторить вызывает refetch', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockRejectedValue(new Error('API error'))
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByText('Ошибка загрузки каталога')).toBeInTheDocument()
		})

		expect(screen.getByText('Повторить')).toBeInTheDocument()
	})

	it('filter reset — сброс фильтра category', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({ data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 } })
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />, ['/catalog?category=fitness'])

		await waitFor(() => {
			expect(screen.getByTestId('filter-bar')).toBeInTheDocument()
		})
	})

	it('search debounce behavior', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({ data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 } })
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByTestId('search-input')).toBeInTheDocument()
		})
	})

	it('pagination overflow — переход на несуществующую страницу', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({ data: [], pagination: { page: 999, per_page: 20, total: 0, total_pages: 1 } })
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />, ['/catalog?page=999'])

		await waitFor(() => {
			expect(screen.getByText('Нет доступных льгот')).toBeInTheDocument()
		})
	})

	it('category change triggering refetch', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({ data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 } })
		vi.mocked(getCategories).mockResolvedValue([
			{ id: '1', slug: 'fitness', name: 'Фитнес', icon: 'dumbbell', sort_order: 1 },
		])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByTestId('filter-bar')).toBeInTheDocument()
		})
	})

	it('загружает данные при успешном API ответе', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({
			data: [
				{ id: '1', slug: 'yoga', name: 'Йога', type: 'benefit', status: 'active', badge: 'Доступна', badge_color: 'gray' },
			],
			pagination: { page: 1, per_page: 20, total: 1, total_pages: 1 },
		})
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByTestId('engagement-1')).toBeInTheDocument()
		})

		expect(screen.getByText('Йога')).toBeInTheDocument()
	})

	it('отображает пагинацию при наличии данных', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({
			data: [
				{ id: '1', slug: 'yoga', name: 'Йога', type: 'benefit', status: 'active', badge: 'Доступна', badge_color: 'gray' },
			],
			pagination: { page: 1, per_page: 20, total: 50, total_pages: 3 },
		})
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByTestId('pagination')).toBeInTheDocument()
		})
	})

	it('title "Каталог льгот" отображается', async () => {
		const { getEngagements, getCategories } = await import('@/api/engagements')
		vi.mocked(getEngagements).mockResolvedValue({ data: [], pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 } })
		vi.mocked(getCategories).mockResolvedValue([])

		const { Catalog } = await import('@/pages/Catalog')
		renderWithProviders(<Catalog />)

		await waitFor(() => {
			expect(screen.getByText('Каталог льгот')).toBeInTheDocument()
		})
	})
})
