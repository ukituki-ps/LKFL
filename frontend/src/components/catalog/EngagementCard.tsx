import { useState } from 'react'
import { Card, Text, Group, Box } from '@mantine/core'
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

// ─── Badge colors ──

/** Кастомные цвета бейджей по прототипу. */
const badgeColors: Record<string, { bg: string; color: string }> = {
	green: { bg: '#DCFCE7', color: '#166534' },
	yellow: { bg: '#FEF9C3', color: '#854D0E' },
	gray: { bg: '#F3F4F6', color: '#4B5563' },
	blue: { bg: '#DBEAFE', color: '#1D4ED8' },
}

/** Цвет бейджа из badge_color или по значению badge. */
function getBadgeColorKey(badgeColor: string, badge: string): string {
	if (badgeColor && badgeColors[badgeColor]) return badgeColor
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

/** Рендерит бейдж с кастомными цветами из прототипа. */
function Badge({
	children,
	colorKey,
	size = 'sm',
}: {
	children: React.ReactNode
	colorKey: string
	size?: 'xs' | 'sm'
}) {
	const colors = badgeColors[colorKey] || badgeColors.gray
	const fontSize = size === 'xs' ? '10px' : '11px'
	const padding = size === 'xs' ? '2px 6px' : '3px 8px'

	return (
		<span
			style={{
				display: 'inline-block',
				backgroundColor: colors.bg,
				color: colors.color,
				fontSize,
				fontWeight: 600,
				padding,
				borderRadius: '6px',
				lineHeight: '1.4',
				whiteSpace: 'nowrap',
			}}
		>
			{children}
		</span>
	)
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
 *
 * Hover: translateY(-2px) + box-shadow 0 4px 16px.
 */
export function EngagementCard({ engagement }: EngagementCardProps) {
	const badgeColorKey = getBadgeColorKey(
		engagement.badge_color,
		engagement.badge,
	)
	const priceDisplay =
		engagement.price_display ||
		(engagement.cost_cents != null ? formatPrice(engagement.cost_cents) : '')

	/* Hover state */
	const [hovered, setHovered] = useState(false)

	return (
		<Card
			withBorder
			padding="lg"
			radius="var(--brand-radius-card, 14px)"
			shadow="var(--brand-shadow-card)"
			style={{
				display: 'flex',
				flexDirection: 'column',
				transition: 'transform 0.15s, box-shadow 0.15s',
				transform: hovered ? 'translateY(-2px)' : 'none',
				boxShadow: hovered
					? '0 4px 16px rgba(0,0,0,0.1)'
					: 'var(--brand-shadow-card)',
			}}
			onMouseEnter={() => setHovered(true)}
			onMouseLeave={() => setHovered(false)}
		>
			<Link
				to={`/catalog/${engagement.slug}`}
				style={{
					textDecoration: 'none',
					color: 'inherit',
					flex: 1,
					display: 'flex',
					flexDirection: 'column',
				}}
			>
				<Box>
					{/* Icon — 44×44, borderRadius 12 */}
					<Group
						gap={8}
						mb="md"
						style={{
							width: 44,
							height: 44,
							padding: '10px',
							borderRadius: 12,
							backgroundColor: 'var(--brand-row, #F9FAFB)',
							justifyContent: 'center',
						}}
					>
						{renderIcon(engagement.icon_name || '')}
					</Group>

					{/* Badge + Category */}
					<Group justify="space-between" mb="xs">
						<Badge colorKey={badgeColorKey}>{engagement.badge}</Badge>

						{engagement.category && (
							<Text size="xs" c="dimmed">
								{engagement.category.name}
							</Text>
						)}
					</Group>

					{/* Name — 14px, fontWeight 700 */}
					<Text
						fw={700}
						style={{ fontSize: '14px' }}
						mb="xs"
						lineClamp={1}
					>
						{engagement.name}
					</Text>

					{/* Provider — 11px, text-subtle */}
					{engagement.provider_name && (
						<Text
							style={{
								fontSize: '11px',
								color: 'var(--brand-text-subtle)',
							}}
							mb="xs"
						>
							{engagement.provider_name}
						</Text>
					)}

					{/* Description — 12px, text-muted, lineHeight 1.5 */}
					{engagement.description && (
						<Text
							style={{
								fontSize: '12px',
								color: 'var(--brand-text-muted)',
								lineHeight: 1.5,
							}}
							mb="md"
							lineClamp={2}
						>
							{engagement.description}
						</Text>
					)}
				</Box>

				{/* Footer — border-top: 1px solid var(--brand-row) */}
				<Box
					mt="auto"
					pt="md"
					style={{ borderTop: '1px solid var(--brand-row)' }}
				>
					<Group justify="space-between" wrap="nowrap">
						{priceDisplay && (
							<Text fw={700} size="md" style={{ color: 'var(--brand-green)' }}>
								{priceDisplay}
							</Text>
						)}
						{engagement.badge && (
							<Badge colorKey={badgeColorKey} size="xs">
								{engagement.badge}
							</Badge>
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

/** Сетка карточек для каталога — 3 колонки, gap 14px. */
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
				gap: '14px',
			}}
		>
			{engagements.map((e) => (
				<EngagementCard key={e.id} engagement={e} />
			))}
		</div>
	)
}
