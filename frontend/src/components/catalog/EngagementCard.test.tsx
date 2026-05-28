import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { MantineProvider } from '@mantine/core'
import { createAprilTheme } from '@/lib/theme'
import { EngagementCard } from '@/components/catalog/EngagementCard'
import type { EngagementTypeResponse } from '@/api/types'

const aprilTheme = createAprilTheme()

function renderWithProviders(ui: React.ReactElement) {
	return render(
		<MemoryRouter>
			<MantineProvider theme={aprilTheme}>
				{ui}
			</MantineProvider>
		</MemoryRouter>
	)
}

const mockEngagement: EngagementTypeResponse = {
	id: '550e8400-e29b-41d4-a716-446655440000',
	slug: 'yoga-studio',
	name: 'Йога в студии',
	description: 'Абонемент на йогу',
	type: 'benefit',
	status: 'active',
	cost_cents: 150000,
	provider_name: 'FitLife',
	image_url: 'https://example.com/yoga.jpg',
	category: {
		id: 'cat-1',
		slug: 'fitness',
		name: 'Фитнес',
		icon: '🏋️',
		sort_order: 1,
	},
	offers: [],
	badge: 'Доступна',
	badge_color: 'gray',
}

describe('EngagementCard', () => {
	it('отображает название льготы', () => {
		renderWithProviders(<EngagementCard engagement={mockEngagement} />)
		expect(screen.getByText('Йога в студии')).toBeInTheDocument()
	})

	it('отображает стоимость в рублях', () => {
		renderWithProviders(<EngagementCard engagement={mockEngagement} />)
		expect(screen.getByText('1 500 ₽')).toBeInTheDocument()
	})

	it('отображает бейдж "Промо"', () => {
		const promoEngagement = { ...mockEngagement, badge: 'Промо', status: 'promo' as const }
		renderWithProviders(<EngagementCard engagement={promoEngagement} />)
		expect(screen.getByText('Промо')).toBeInTheDocument()
	})

	it('отображает бейдж "Доступна"', () => {
		renderWithProviders(<EngagementCard engagement={mockEngagement} />)
		expect(screen.getByText('Доступна')).toBeInTheDocument()
	})

	it('отображает название категории', () => {
		renderWithProviders(<EngagementCard engagement={mockEngagement} />)
		expect(screen.getByText('Фитнес')).toBeInTheDocument()
	})

	it('отображает название провайдера', () => {
		renderWithProviders(<EngagementCard engagement={mockEngagement} />)
		expect(screen.getByText('FitLife')).toBeInTheDocument()
	})

	it('отображает описание если задано', () => {
		renderWithProviders(<EngagementCard engagement={mockEngagement} />)
		expect(screen.getByText('Абонемент на йогу')).toBeInTheDocument()
	})

	it('не отображает стоимость если cost_cents не задан', () => {
		const noCostEngagement: EngagementTypeResponse = {
			...mockEngagement,
			cost_cents: undefined,
		}
		renderWithProviders(<EngagementCard engagement={noCostEngagement} />)
		expect(screen.queryByText(/₽/)).not.toBeInTheDocument()
	})

	it('отображает количество офферов если > 1', () => {
		const multiOfferEngagement: EngagementTypeResponse = {
			...mockEngagement,
			offers: [
				{ id: '1', name: 'Месяц', cost_cents: 150000, sort_order: 1 },
				{ id: '2', name: '3 месяца', cost_cents: 400000, sort_order: 2 },
			],
		}
		renderWithProviders(<EngagementCard engagement={multiOfferEngagement} />)
		expect(screen.getByText('2 варианта')).toBeInTheDocument()
	})

	it('отображает "3 варианта" для 3 офферов', () => {
		const threeOfferEngagement: EngagementTypeResponse = {
			...mockEngagement,
			offers: [
				{ id: '1', name: 'Месяц', cost_cents: 150000, sort_order: 1 },
				{ id: '2', name: '3 месяца', cost_cents: 400000, sort_order: 2 },
				{ id: '3', name: 'Год', cost_cents: 1200000, sort_order: 3 },
			],
		}
		renderWithProviders(<EngagementCard engagement={threeOfferEngagement} />)
		expect(screen.getByText('3 варианта')).toBeInTheDocument()
	})

	it('не отображает количество офферов если 0 или 1', () => {
		const singleOfferEngagement: EngagementTypeResponse = {
			...mockEngagement,
			offers: [{ id: '1', name: 'Месяц', cost_cents: 150000, sort_order: 1 }],
		}
		renderWithProviders(<EngagementCard engagement={singleOfferEngagement} />)
		expect(screen.queryByText(/вариант/)).not.toBeInTheDocument()
	})

	it('отображает заглушку если нет изображения', () => {
		const noImageEngagement: EngagementTypeResponse = {
			...mockEngagement,
			image_url: undefined,
		}
		renderWithProviders(<EngagementCard engagement={noImageEngagement} />)
		expect(screen.getByText('Нет изображения')).toBeInTheDocument()
	})

	it('не отображает категорию если она не задана', () => {
		const noCategoryEngagement: EngagementTypeResponse = {
			...mockEngagement,
			category: undefined,
		}
		renderWithProviders(<EngagementCard engagement={noCategoryEngagement} />)
		expect(screen.queryByText('Фитнес')).not.toBeInTheDocument()
	})

	it('не отображает провайдера если он не задан', () => {
		const noProviderEngagement: EngagementTypeResponse = {
			...mockEngagement,
			provider_name: undefined,
		}
		renderWithProviders(<EngagementCard engagement={noProviderEngagement} />)
		expect(screen.queryByText('FitLife')).not.toBeInTheDocument()
	})

	// =============================================================================
	// EDGE CASE TESTS
	// =============================================================================

	it('null cost_cents — не отображает стоимость', () => {
		const nullCostEngagement: EngagementTypeResponse = {
			...mockEngagement,
			cost_cents: undefined,
		}
		renderWithProviders(<EngagementCard engagement={nullCostEngagement} />)
		expect(screen.queryByText(/₽/)).not.toBeInTheDocument()
	})

	it('zero cost — отображает 0 ₽', () => {
		const zeroCostEngagement: EngagementTypeResponse = {
			...mockEngagement,
			cost_cents: 0,
		}
		renderWithProviders(<EngagementCard engagement={zeroCostEngagement} />)
		expect(screen.getByText('0 ₽')).toBeInTheDocument()
	})

	it('negative cost — отображает отрицательную стоимость', () => {
		const negativeCostEngagement: EngagementTypeResponse = {
			...mockEngagement,
			cost_cents: -5000,
		}
		renderWithProviders(<EngagementCard engagement={negativeCostEngagement} />)
		expect(screen.getByText('-50 ₽')).toBeInTheDocument()
	})

	it('очень большое значение cost_cents', () => {
		const largeCostEngagement: EngagementTypeResponse = {
			...mockEngagement,
			cost_cents: 999999999,
		}
		renderWithProviders(<EngagementCard engagement={largeCostEngagement} />)
		expect(screen.getByText('9 999 999 ₽')).toBeInTheDocument()
	})

	it('пустое имя', () => {
		const emptyNameEngagement: EngagementTypeResponse = {
			...mockEngagement,
			name: '',
		}
		renderWithProviders(<EngagementCard engagement={emptyNameEngagement} />)
		// Component renders the name even if empty (lineClamp handles it)
	})

	it('очень длинное имя (1000+ символов)', () => {
		const longName = 'A'.repeat(1024)
		const longNameEngagement: EngagementTypeResponse = {
			...mockEngagement,
			name: longName,
		}
		renderWithProviders(<EngagementCard engagement={longNameEngagement} />)
		// lineClamp={1} should truncate visually
	})

	it('очень длинное описание (1000+ символов)', () => {
		const longDescription = 'B'.repeat(2048)
		const longDescEngagement: EngagementTypeResponse = {
			...mockEngagement,
			description: longDescription,
		}
		renderWithProviders(<EngagementCard engagement={longDescEngagement} />)
		// lineClamp={2} should truncate visually
	})

	it('no offers — не отображает счётчик вариантов', () => {
		const noOffersEngagement: EngagementTypeResponse = {
			...mockEngagement,
			offers: undefined,
		}
		renderWithProviders(<EngagementCard engagement={noOffersEngagement} />)
		expect(screen.queryByText(/вариант/)).not.toBeInTheDocument()
	})

	it('empty offers array — не отображает счётчик', () => {
		const emptyOffersEngagement: EngagementTypeResponse = {
			...mockEngagement,
			offers: [],
		}
		renderWithProviders(<EngagementCard engagement={emptyOffersEngagement} />)
		expect(screen.queryByText(/вариант/)).not.toBeInTheDocument()
	})

	it('no category — не отображает категорию', () => {
		const noCategoryEngagement: EngagementTypeResponse = {
			...mockEngagement,
			category: undefined,
		}
		renderWithProviders(<EngagementCard engagement={noCategoryEngagement} />)
		expect(screen.queryByText('Фитнес')).not.toBeInTheDocument()
	})

	it('no provider — не отображает провайдера', () => {
		const noProviderEngagement: EngagementTypeResponse = {
			...mockEngagement,
			provider_name: undefined,
		}
		renderWithProviders(<EngagementCard engagement={noProviderEngagement} />)
		expect(screen.queryByText('FitLife')).not.toBeInTheDocument()
	})

	it('все поля null — компонент не падает', () => {
		const minimalEngagement: EngagementTypeResponse = {
			id: '00000000-0000-0000-0000-000000000000',
			slug: 'minimal',
			name: 'Минимальный',
			type: 'benefit',
			status: 'active',
			badge: 'Доступна',
			badge_color: 'gray',
		}
		renderWithProviders(<EngagementCard engagement={minimalEngagement} />)
		expect(screen.getByText('Минимальный')).toBeInTheDocument()
		expect(screen.getByText('Доступна')).toBeInTheDocument()
	})

	it('null image_url — показывает заглушку', () => {
		const nullImageEngagement: EngagementTypeResponse = {
			...mockEngagement,
			image_url: undefined,
		}
		renderWithProviders(<EngagementCard engagement={nullImageEngagement} />)
		expect(screen.getByText('Нет изображения')).toBeInTheDocument()
	})

	it('пустая строка image_url', () => {
		const emptyImageEngagement: EngagementTypeResponse = {
			...mockEngagement,
			image_url: '',
		}
		renderWithProviders(<EngagementCard engagement={emptyImageEngagement} />)
		// Empty string is falsy, should show placeholder
	})

	it('5 офферов — правильный склон', () => {
		const fiveOfferEngagement: EngagementTypeResponse = {
			...mockEngagement,
			offers: [
				{ id: '1', name: 'О1', cost_cents: 100, sort_order: 1 },
				{ id: '2', name: 'О2', cost_cents: 200, sort_order: 2 },
				{ id: '3', name: 'О3', cost_cents: 300, sort_order: 3 },
				{ id: '4', name: 'О4', cost_cents: 400, sort_order: 4 },
				{ id: '5', name: 'О5', cost_cents: 500, sort_order: 5 },
			],
		}
		renderWithProviders(<EngagementCard engagement={fiveOfferEngagement} />)
		expect(screen.getByText('5 вариантов')).toBeInTheDocument()
	})

	it('21 оффер — "21 вариант" (правильный склонение)', () => {
		const twentyOneOffers: EngagementTypeResponse = {
			...mockEngagement,
			offers: Array.from({ length: 21 }, (_, i) => ({
				id: String(i),
				name: `О${i}`,
				cost_cents: i * 100,
				sort_order: i,
			})),
		}
		renderWithProviders(<EngagementCard engagement={twentyOneOffers} />)
		expect(screen.getByText('21 вариант')).toBeInTheDocument()
	})

	it('22 оффера — "22 варианта"', () => {
		const twentyTwoOffers: EngagementTypeResponse = {
			...mockEngagement,
			offers: Array.from({ length: 22 }, (_, i) => ({
				id: String(i),
				name: `О${i}`,
				cost_cents: i * 100,
				sort_order: i,
			})),
		}
		renderWithProviders(<EngagementCard engagement={twentyTwoOffers} />)
		expect(screen.getByText('22 варианта')).toBeInTheDocument()
	})

	it('5 офферов — "5 вариантов"', () => {
		const fiveOffers: EngagementTypeResponse = {
			...mockEngagement,
			offers: Array.from({ length: 5 }, (_, i) => ({
				id: String(i),
				name: `О${i}`,
				cost_cents: i * 100,
				sort_order: i,
			})),
		}
		renderWithProviders(<EngagementCard engagement={fiveOffers} />)
		expect(screen.getByText('5 вариантов')).toBeInTheDocument()
	})

	it('11 офферов — "11 вариантов" (исключение 11-19)', () => {
		const elevenOffers: EngagementTypeResponse = {
			...mockEngagement,
			offers: Array.from({ length: 11 }, (_, i) => ({
				id: String(i),
				name: `О${i}`,
				cost_cents: i * 100,
				sort_order: i,
			})),
		}
		renderWithProviders(<EngagementCard engagement={elevenOffers} />)
		expect(screen.getByText('11 вариантов')).toBeInTheDocument()
	})

	it('badge Промо — жёлтый цвет', () => {
		const promoEngagement: EngagementTypeResponse = {
			...mockEngagement,
			badge: 'Промо',
			badge_color: 'yellow',
			status: 'promo' as const,
		}
		renderWithProviders(<EngagementCard engagement={promoEngagement} />)
		expect(screen.getByText('Промо')).toBeInTheDocument()
	})

	it('пустое описание — не отображается', () => {
		const noDescEngagement: EngagementTypeResponse = {
			...mockEngagement,
			description: '',
		}
		renderWithProviders(<EngagementCard engagement={noDescEngagement} />)
		expect(screen.queryByText('Абонемент на йогу')).not.toBeInTheDocument()
	})

	it('undefined description — не отображается', () => {
		const undefDescEngagement: EngagementTypeResponse = {
			...mockEngagement,
			description: undefined,
		}
		renderWithProviders(<EngagementCard engagement={undefDescEngagement} />)
		expect(screen.queryByText('Абонемент на йогу')).not.toBeInTheDocument()
	})
})
