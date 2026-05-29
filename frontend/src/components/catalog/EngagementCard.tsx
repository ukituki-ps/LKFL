import { Card, Badge as MantineBadge, Text, Group, Box } from '@mantine/core'
import { Link } from 'react-router-dom'
import type { EngagementTypeResponse } from '@/api/types'
import {
	AprilIconHeart,
	AprilIconSuccess,
	AprilIconPlusCircle,
	AprilIconUsers,
	AprilIconDumbbell,
	AprilIconGift,
	AprilIconCoffee,
	AprilIconGraduationCap,
	AprilIconBrain,
	AprilIconLanguages,
	AprilIconShoppingBag,
	AprilIconDashboard,
	type AprilLucideIcon,
} from '@ukituki-ps/april-ui'

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

// ─── Icon mapping ──

/** Маппинг icon_name → AprilIcon из DS. */
const iconMap: Record<string, AprilLucideIcon> = {
	'heart-pulse': AprilIconHeart,
	'shield-plus': AprilIconPlusCircle,
	'shield-check': AprilIconSuccess,
	users: AprilIconUsers,
	dumbbell: AprilIconDumbbell,
	bike: AprilIconGift,
	utensils: AprilIconCoffee,
	'graduation-cap': AprilIconGraduationCap,
	brain: AprilIconBrain,
	languages: AprilIconLanguages,
	'shopping-bag': AprilIconShoppingBag,
	smile: AprilIconHeart,
	coffee: AprilIconCoffee,
	default: AprilIconDashboard,
}

/** Рендерит AprilIcon по имени из metadata.icon_name. */
function renderIcon(name: string) {
	const Icon = name ? iconMap[name] || iconMap.default : iconMap.default
	return <Icon size={24} style={{ color: 'var(--brand-green)' }} />
}

// ─── Component ──

/** Форматирует cost_cents в строку «X ₽» (с разделителем тысяч). */
function formatPrice(cents: number): string {
	const rubles = Math.round(cents / 100)
	return `${rubles.toLocaleString('ru-RU')} ₽`
}

/**
 * Карточка льготы/активности для каталога.
 *
 * Layout по прототипу:
 * ┌─────────────────────────┐
 * │  [icon 44×44 bg-gray]   │
 * │  Название (14px fw:700) │
 * │  Провайдер (11px muted) │
 * │  Описание (12px muted)  │
 * ├─────────────────────────┤
 * │  Цена        [badge]    │
 * └─────────────────────────┘
 */
export function EngagementCard({ engagement }: EngagementCardProps) {
	const badgeColor = getBadgeColor(engagement.badge_color, engagement.badge)
	const priceDisplay =
		engagement.price_display ||
		(engagement.cost_cents != null ? formatPrice(engagement.cost_cents) : '')

	return (
		<Card
			withBorder
			padding="lg"
			radius="var(--brand-radius-card, 14px)"
			shadow="var(--brand-shadow-card)"
			style={{
				display: 'flex',
				flexDirection: 'column',
			}}
		>
			<Link
				to={`/catalog/${engagement.slug}`}
				style={{ textDecoration: 'none', color: 'inherit', flex: 1, display: 'flex', flexDirection: 'column' }}
			>
				<Box>
					{/* Icon */}
					<Group
						gap={8}
						mb="md"
						style={{
							padding: '12px',
							borderRadius: 'var(--brand-radius-card, 14px)',
							backgroundColor: 'var(--brand-row, #F9FAFB)',
							height: 44,
						}}
					>
						{renderIcon(engagement.icon_name || '')}
					</Group>

					{/* Badge + Category */}
					<Group justify="space-between" mb="xs">
						<MantineBadge color={badgeColor} size="sm" variant="light">
							{engagement.badge}
						</MantineBadge>

						{engagement.category && (
							<Text size="xs" c="dimmed">
								{engagement.category.name}
							</Text>
						)}
					</Group>

					{/* Name */}
					<Text fw={700} size="md" mb="xs" lineClamp={1}>
						{engagement.name}
					</Text>

					{/* Provider */}
					{engagement.provider_name && (
						<Text size="xs" c="dimmed" mb="xs">
							{engagement.provider_name}
						</Text>
					)}

					{/* Description */}
					{engagement.description && (
						<Text size="sm" c="dimmed" mb="md" lineClamp={2}>
							{engagement.description}
						</Text>
					)}
				</Box>

				{/* Footer — price + badge */}
				<Box mt="auto" pt="md" style={{ borderTop: '1px solid var(--brand-border)' }}>
					<Group justify="space-between" wrap="nowrap">
						{priceDisplay && (
							<Text fw={700} size="md" style={{ color: 'var(--brand-green)' }}>
								{priceDisplay}
							</Text>
						)}
						{engagement.badge && (
							<MantineBadge color={badgeColor} size="xs" variant="light">
								{engagement.badge}
							</MantineBadge>
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

/** Сетка карточек для каталога — 3 колонки. */
export function EngagementGrid({
	engagements,
}: {
	engagements: EngagementTypeResponse[]
}) {
	return (
		<div
			style={{
				display: 'grid',
				gridTemplateColumns: 'repeat(3, 1fr)',
				gap: '16px',
			}}
		>
			{engagements.map((e) => (
				<EngagementCard key={e.id} engagement={e} />
			))}
		</div>
	)
}
