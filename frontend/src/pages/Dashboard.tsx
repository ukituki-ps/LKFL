import { Card, Text, Group, Stack, Title, Badge, Paper, Skeleton } from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import {
	AprilIconCoins,
	AprilIconSuccess,
	AprilIconCalendar,
	AprilIconDumbbell,
	AprilIconHeart,
	AprilIconUserPlus,
	AprilIconArrowUpCircle,
	AprilIconShoppingBag,
	AprilIconBrain,
	AprilIconBaby,
} from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'
import { getDashboardStats, getActiveBenefits, getEvents } from '@/api/dashboard'

/* ─── Icon mapping (API returns icon name strings) ─── */

const iconMap: Record<string, React.ComponentType<{ size?: number | string }>> = {
	heart: AprilIconHeart,
	dumbbell: AprilIconDumbbell,
	shield: AprilIconSuccess,
	coins: AprilIconCoins,
	calendar: AprilIconCalendar,
}

/* ─── Helpers ─── */

/**
 * Парсит текст события из API, выделяя ключевые части bold-ом.
 * Формат API: "Новая льгота: онлайн-кинотеатр", "Начислено 500 баллов за опрос", "Обновлены условия программы"
 */
function parseEventText(text: string): React.ReactNode {
	// "Новая льгота: <name>"
	const m1 = text.match(/^(Новая льгота:\s*)(.*)$/)
	if (m1) {
		return (
			<>
				{m1[1]}
				<Text component="span" fw={700}>
					{m1[2]}
				</Text>
			</>
		)
	}

	// "Начислено <amount> баллов <reason>"
	const m2 = text.match(/^(Начислено\s*)(\d+\s*\w+)(.*)$/)
	if (m2) {
		return (
			<>
				{m2[1]}
				<Text component="span" fw={700}>
					{m2[2]}
				</Text>
				{m2[3]}
			</>
		)
	}

	// "Обновлены условия <subject>"
	const m3 = text.match(/^(Обновлены условия:\s*)(.*)$/)
	if (m3) {
		return (
			<>
				{m3[1]}
				<Text component="span" fw={700}>
					{m3[2]}
				</Text>
			</>
		)
	}

	return text
}

/* ─── Types ─── */

interface QuickAction {
	label: string
	icon: React.ComponentType<{ size?: number | string }>
}

/* ─── Mock data (статические — не приходят с API) ─── */

/* ГЭП-2: быстрые действия из прототипа (5 элементов, 2 колонки) */
const mockQuickActions: QuickAction[] = [
	{ label: 'Добавить родственника к ДМС', icon: AprilIconUserPlus },
	{ label: 'Апгрейд ДМС', icon: AprilIconArrowUpCircle },
	{ label: 'Купить мерч СДЭК', icon: AprilIconShoppingBag },
	{ label: 'Записаться к психологу', icon: AprilIconBrain },
	{ label: 'Заявка на мат. капитал от компании', icon: AprilIconBaby },
]

/* ─── Components ─── */

/* ГЭП-6: Stat card с поддержкой зелёного фона */
function StatCard({
	title,
	value,
	suffix,
	subtitle,
	icon: Icon,
	green,
}: {
	title: string
	value: string
	suffix?: string
	subtitle: string
	icon: React.ComponentType<{ size?: number | string }>
	green?: boolean
}) {
	const isGreen = green === true

	return (
		<Card
			withBorder
			padding="md"
			style={{
				flex: 1,
				minWidth: 180,
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
				background: isGreen ? 'var(--brand-green, #00B33C)' : 'transparent',
				color: isGreen ? '#FFFFFF' : 'inherit',
			}}
		>
			<Group justify="space-between" mb="xs">
				<Text size="xs" c={isGreen ? 'white' : 'dimmed'} opacity={isGreen ? 0.85 : 1}>
					{title}
				</Text>
				<StubBadge />
			</Group>
			<Group align="flex-end" gap={4} mb="xs">
				<span style={{ opacity: isGreen ? 0.9 : 1 }}>
					<Icon size={20} />
				</span>
				<div>
					<div style={{ display: 'flex', alignItems: 'baseline', gap: 2 }}>
						<Text
							fw={800}
							style={{
								fontSize: suffix ? 26 : undefined,
								lineHeight: 1.1,
								color: isGreen ? '#FFFFFF' : 'var(--brand-text)',
							}}
						>
							{value}
						</Text>
						{suffix && (
							<Text fw={600} size="md" style={{ color: isGreen ? 'rgba(255,255,255,0.8)' : 'var(--brand-text)' }}>
								{suffix}
							</Text>
						)}
					</div>
				</div>
			</Group>
			<Text size="xs" c={isGreen ? 'white' : 'dimmed'} opacity={isGreen ? 0.7 : 1}>
				{subtitle}
			</Text>
		</Card>
	)
}

function ActiveBenefitsList({
	benefits,
	isLoading,
	isError,
}: {
	benefits: ReturnType<typeof getActiveBenefits> extends Promise<infer T> ? T : never
	isLoading: boolean
	isError: boolean
}) {
	if (isLoading) return <Skeleton height={200} />
	if (isError) return <Text c="red">Не удалось загрузить данные. Попробуйте позже.</Text>

	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group justify="space-between" mb="md">
				<Text fw={600} size="md">
					Активные льготы
				</Text>
				<StubBadge />
			</Group>
			<Stack gap="sm">
				{benefits.map((b) => {
					const Icon = iconMap[b.icon] || AprilIconHeart
					return (
						<Group
							key={b.name}
							gap="sm"
							style={{ padding: '8px 0', borderBottom: '1px solid var(--brand-border)' }}
						>
							<div
								style={{
									width: 32,
									height: 32,
									display: 'flex',
									alignItems: 'center',
									justifyContent: 'center',
									borderRadius: 8,
									backgroundColor: 'var(--brand-row, #F9FAFB)',
									flexShrink: 0,
									color: 'var(--brand-green)',
								}}
							>
								<Icon size={16} />
							</div>
							<div style={{ flex: 1 }}>
								<Text size="sm" fw={500}>
									{b.name}
								</Text>
								<Text size="xs" c="dimmed">
									{b.provider}
								</Text>
							</div>
							<Badge
								variant="light"
								color={b.status === 'active' ? 'green' : 'yellow'}
								size="xs"
							>
								{b.status === 'active' ? 'Активна' : 'Ожидает'}
							</Badge>
						</Group>
					)
				})}
			</Stack>
		</Card>
	)
}

/* ГЭП-7: лента событий с иконками в цветных квадратах + временные метки */
function EventsFeed({
	events,
	isLoading,
	isError,
}: {
	events: ReturnType<typeof getEvents> extends Promise<infer T> ? T : never
	isLoading: boolean
	isError: boolean
}) {
	/* Маппинг цветов на иконки */
	const getEventIcon = (color: string) => {
		switch (color) {
			case 'green':
				return AprilIconSuccess
			case 'yellow':
				return AprilIconCoins
			case 'blue':
				return AprilIconCalendar
			default:
				return AprilIconSuccess
		}
	}

	const getEventColors = (color: string) => {
		switch (color) {
			case 'green':
				return { bg: '#DCFCE7', fg: '#16A34A' }
			case 'yellow':
				return { bg: '#FEF9C3', fg: '#CA8A04' }
			case 'blue':
				return { bg: '#DBEAFE', fg: '#2563EB' }
			default:
				return { bg: '#DCFCE7', fg: '#16A34A' }
		}
	}

	if (isLoading) return <Skeleton height={200} />
	if (isError) return <Text c="red">Не удалось загрузить данные. Попробуйте позже.</Text>

	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group justify="space-between" mb="md">
				<Text fw={600} size="md">
					Последние события
				</Text>
				<StubBadge />
			</Group>
			<Stack gap="md">
				{events.map((e, i) => {
					const Icon = getEventIcon(e.color)
					const colors = getEventColors(e.color)
					return (
						<Group key={i} gap="sm" align="flex-start">
							<div
								style={{
									width: 30,
									height: 30,
									display: 'flex',
									alignItems: 'center',
									justifyContent: 'center',
									borderRadius: 8,
									backgroundColor: colors.bg,
									flexShrink: 0,
									color: colors.fg,
								}}
							>
								<Icon size={16} />
							</div>
							<div style={{ flex: 1 }}>
								<Text size="sm">{parseEventText(e.text)}</Text>
								<Text size="xs" c="dimmed" mt={2}>
									{e.time}
								</Text>
							</div>
						</Group>
					)
				})}
			</Stack>
		</Card>
	)
}

/* ГЭП-2: быстрые действия 2 колонки с toast */
function QuickActionsGrid({ onActionClick }: { onActionClick: () => void }) {
	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
				width: 292,
			}}
		>
			<Group justify="space-between" mb="md">
				<Text fw={600} size="md">
					Быстрые действия
				</Text>
				<StubBadge />
			</Group>
			<div
				style={{
					display: 'grid',
					gridTemplateColumns: 'repeat(2, 1fr)',
					gap: 8,
				}}
			>
				{mockQuickActions.map((a) => (
					<Paper
						key={a.label}
						withBorder
						style={{
							padding: 10,
							borderRadius: 'var(--brand-radius-btn, 6px)',
							textAlign: 'center',
							cursor: 'pointer',
							transition: 'background-color 150ms',
						}}
						onClick={onActionClick}
						onMouseEnter={(e) => {
							e.currentTarget.style.backgroundColor = 'var(--brand-green-light)'
						}}
						onMouseLeave={(e) => {
							e.currentTarget.style.backgroundColor = 'transparent'
						}}
					>
						<a.icon size={18} />
						<Text size="xs" fw={500} mt={4} lineClamp={2}>
							{a.label}
						</Text>
					</Paper>
				))}
			</div>
		</Card>
	)
}

/* ─── Page ─── */

function getGreeting(): string {
	const hour = new Date().getHours()
	if (hour < 6) return 'Доброй ночи'
	if (hour < 12) return 'Доброе утро'
	if (hour < 18) return 'Добрый день'
	return 'Добрый вечер'
}

/**
 * Главная страница (Dashboard) — данные через React Query.
 *
 * ГЭП-3: layout 2 колонки — слева льготы + события, справа быстрые действия (292px).
 * ГЭП-6: Stat card 1 — зелёный фон, белый текст.
 * ГЭП-12: Stat card 3 — число + «дн» мелким шрифтом.
 */
export function Dashboard() {
	const today = new Date().toLocaleDateString('ru-RU', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
	})

	// ─── React Query ───

	const { data: stats, isLoading: statsLoading, isError: statsError } = useQuery({
		queryKey: ['dashboard'],
		queryFn: getDashboardStats,
	})

	const { data: benefits, isLoading: benefitsLoading, isError: benefitsError } = useQuery({
		queryKey: ['dashboard', 'benefits'],
		queryFn: getActiveBenefits,
	})

	const { data: events, isLoading: eventsLoading, isError: eventsError } = useQuery({
		queryKey: ['dashboard', 'events'],
		queryFn: getEvents,
	})

	const showF2Toast = () => {
		/* Toast через alert-подобный механизм — @mantine/notifications подключится в F2.
		 * Сейчас используем нативный подход через временный DOM-элемент. */
		const existing = document.getElementById('lkfl-toast')
		if (existing) existing.remove()

		const toast = document.createElement('div')
		toast.id = 'lkfl-toast'
		toast.textContent = 'Функция будет доступна после F2'
		Object.assign(toast.style, {
			position: 'fixed',
			bottom: '24px',
			left: '50%',
			transform: 'translateX(-50%)',
			background: '#1F2937',
			color: '#FFFFFF',
			padding: '10px 20px',
			borderRadius: '10px',
			fontSize: '14px',
			fontWeight: 500,
			zIndex: 9999,
			boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
			opacity: 0,
			transition: 'opacity 200ms',
		})
		document.body.appendChild(toast)
		requestAnimationFrame(() => { toast.style.opacity = '1' })
		setTimeout(() => {
			toast.style.opacity = '0'
			setTimeout(() => toast.remove(), 200)
		}, 2500)
	}

	return (
		<Stack gap="lg">
			{/* Greeting */}
			<Group justify="space-between" align="flex-start">
				<div>
					<Title order={2} style={{ marginBottom: 4 }}>
						{getGreeting()}, Алексей!
					</Title>
					<Text size="sm" c="dimmed">
						{today}
					</Text>
				</div>
				<Badge variant="light" color="green">
					Пакет «Стандарт»
				</Badge>
			</Group>

			{/* ГЭП-6 + ГЭП-12: stat cards — данные с API */}
			<Group gap="md" wrap="wrap">
				{statsLoading ? (
					<>
						<Skeleton height={100} style={{ flex: 1, minWidth: 180 }} />
						<Skeleton height={100} style={{ flex: 1, minWidth: 180 }} />
						<Skeleton height={100} style={{ flex: 1, minWidth: 180 }} />
					</>
				) : statsError ? (
					<Text c="red">Не удалось загрузить данные. Попробуйте позже.</Text>
				) : (
					<>
						<StatCard
							title="Баланс баллов"
							value={String(stats!.points)}
							subtitle="+500 баллов в июне"
							icon={AprilIconCoins}
							green
						/>
						<StatCard
							title="Активные льготы"
							value={String(stats!.activeBenefits)}
							subtitle="Из 5 доступных"
							icon={AprilIconSuccess}
						/>
						<StatCard
							title="До конца периода"
							value={String(stats!.daysLeft)}
							suffix="дн"
							subtitle="Период: янв — июн 2025"
							icon={AprilIconCalendar}
						/>
					</>
				)}
			</Group>

			{/* ГЭП-3: layout 2 колонки */}
			<Group wrap="nowrap" gap="md" align="flex-start">
				{/* Левая колонка: льготы + события */}
				<div style={{ flex: '1 1 auto', minWidth: 0 }}>
					<Stack gap="lg">
						<ActiveBenefitsList
							benefits={benefits ?? []}
							isLoading={benefitsLoading}
							isError={benefitsError}
						/>
						<EventsFeed
							events={events ?? []}
							isLoading={eventsLoading}
							isError={eventsError}
						/>
					</Stack>
				</div>

				{/* Правая колонка: быстрые действия (292px) */}
				<QuickActionsGrid onActionClick={showF2Toast} />
			</Group>
		</Stack>
	)
}
