import { Card, Badge as MantineBadge, Text, Group, Box, Paper } from '@mantine/core'
import { Link } from 'react-router-dom'
import type { EngagementTypeResponse } from '@/api/types'

// ─── Props ──

interface EngagementCardProps {
	engagement: EngagementTypeResponse
}

// ─── Helpers ──

/** Цвет бейджа из badge_color или по значению badge. */
function getBadgeColor(badgeColor: string, badge: string): string {
	if (badgeColor) return badgeColor
	switch (badge) {
		case 'Промо':
			return 'yellow'
		case 'Активна':
			return 'green'
		case 'Ожидает':
			return 'yellow'
		case 'Новинка':
			return 'blue'
		default:
			return 'gray'
	}
}

/** Склонение слова «вариант/варианта/вариантов». */
function pluralizeOffers(count: number): string {
	const lastTwo = count % 100
	const lastOne = count % 10

	if (lastTwo >= 11 && lastTwo <= 19) {
		return 'вариантов'
	}

	if (lastOne === 1) {
		return 'вариант'
	}

	if (lastOne >= 2 && lastOne <= 4) {
		return 'варианта'
	}

	return 'вариантов'
}

// ─── Lucide icon mapping ──

/** Рендерит Lucide-иконку по имени из metadata.icon_name. */
function renderIcon(name: string, _color: string) {
	const icons: Record<string, string> = {
		'heart-pulse': '❤️',
		'shield-plus': '🛡️',
		users: '👥',
		dumbbell: '🏋️',
		bike: '🚴',
		utensils: '🍴',
		'graduation-cap': '🎓',
		brain: '🧠',
		languages: '🌍',
		'shopping-bag': '🛍️',
		smile: '😁',
		coffee: '☕',
	}
	const emoji = icons[name] || '📋'
	return (
		<Text size="xl" style={{ fontSize: 48 }}>
			{emoji}
		</Text>
	)
}

// ─── Component ──

/** Карточка льготы/активности для каталога. */
export function EngagementCard({ engagement }: EngagementCardProps) {
	const badgeColor = getBadgeColor(engagement.badge_color, engagement.badge)
	const priceDisplay = engagement.price_display || ''

	return (
		<Card withBorder padding="lg" radius="md" shadow="sm">
			<Link
				to={`/catalog/${engagement.slug}`}
				style={{ textDecoration: 'none', color: 'inherit' }}
			>
				<Box>
					{/* Icon placeholder */}
					<Paper
						radius="md"
						mb="md"
						style={{
							height: 160,
							display: 'flex',
							alignItems: 'center',
							justifyContent: 'center',
							backgroundColor: badgeColor === 'green' ? '#f0fdf4' :
								badgeColor === 'blue' ? '#eff6ff' :
								badgeColor === 'yellow' ? '#fefce8' : '#f8fafc',
						}}
					>
						{engagement.icon_name
							? renderIcon(engagement.icon_name, badgeColor)
							: <Text c="dimmed" size="sm">Нет изображения</Text>
						}
					</Paper>

					{/* Badge + Category */}
					<Group justify="space-between" mb="xs">
						<MantineBadge
							color={badgeColor}
							size="sm"
							variant="light"
						>
							{engagement.badge}
						</MantineBadge>

						{engagement.category && (
							<Text size="xs" c="dimmed">
								{engagement.category.name}
							</Text>
						)}
					</Group>

					{/* Name */}
					<Text fw={600} size="lg" mb="xs" lineClamp={1}>
						{engagement.name}
					</Text>

					{/* Description */}
					{engagement.description && (
						<Text size="sm" c="dimmed" mb="md" lineClamp={2}>
							{engagement.description}
						</Text>
					)}

					{/* Footer */}
					<Group justify="space-between" wrap="nowrap">
						{engagement.provider_name && (
							<Text size="xs" c="dimmed" lineClamp={1}>
								{engagement.provider_name}
							</Text>
						)}

						{priceDisplay && (
							<Text fw={600} size="sm" c="var(--mantine-color-blue-7)">
								{priceDisplay}
							</Text>
						)}
					</Group>

					{/* Offers count */}
					{engagement.offers && engagement.offers.length > 1 && (
						<Text size="xs" c="dimmed" mt="xs">
							{engagement.offers.length}{' '}
							{pluralizeOffers(engagement.offers.length)}
						</Text>
					)}
				</Box>
			</Link>
		</Card>
	)
}

// ─── Grid ──

/** Сетка карточек для каталога. */
export function EngagementGrid({
	engagements,
}: {
	engagements: EngagementTypeResponse[]
}) {
	return (
		<div
			style={{
				display: 'grid',
				gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
				gap: '16px',
			}}
		>
			{engagements.map((e) => (
				<EngagementCard key={e.id} engagement={e} />
			))}
		</div>
	)
}
