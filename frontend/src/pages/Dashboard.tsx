import { Card, Text, Group, Stack, Title, Badge, Paper } from '@mantine/core'
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

/* ─── Types ─── */

interface BenefitItem {
	name: string
	provider: string
	icon: React.ComponentType<{ size?: number | string }>
	status: string
}

interface EventItem {
	text: React.ReactNode
	color: string
	iconBg: string
	iconColor: string
	time: string
}

interface QuickAction {
	label: string
	icon: React.ComponentType<{ size?: number | string }>
}

/* ─── Mock data (заменится на API в F2) ─── */

const mockActiveBenefits: BenefitItem[] = [
	{ name: 'Онлайн-кинотеатр', provider: 'KION', icon: AprilIconHeart, status: 'Активна' },
	{ name: 'Фитнес-клуб', provider: 'World Class', icon: AprilIconDumbbell, status: 'Активна' },
	{ name: 'Страховка ДМС', provider: 'СОГАЗ', icon: AprilIconHeart, status: 'Ожидает' },
]

/* ГЭП-7: события с временными метками, bold в тексте, иконки */
const mockEvents: EventItem[] = [
	{
		text: (
			<>
				Новая льгота:{' '}
				<Text component="span" fw={700}>
					онлайн-кинотеатр
				</Text>
			</>
		),
		color: 'ev-green',
		iconBg: '#DCFCE7',
		iconColor: '#16A34A',
		time: 'Сегодня, 14:30',
	},
	{
		text: (
			<>
				Начислено{' '}
				<Text component="span" fw={700}>
					500 баллов
				</Text>{' '}
				за опрос
			</>
		),
		color: 'ev-yellow',
		iconBg: '#FEF9C3',
		iconColor: '#CA8A04',
		time: 'Вчера, 18:15',
	},
	{
		text: (
			<>
				Обновлены условия{' '}
				<Text component="span" fw={700}>
					программы
				</Text>
			</>
		),
		color: 'ev-blue',
		iconBg: '#DBEAFE',
		iconColor: '#2563EB',
		time: '20 мая, 10:00',
	},
]

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

function ActiveBenefitsList() {
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
				{mockActiveBenefits.map((b) => {
					const Icon = b.icon
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
								color={b.status === 'Активна' ? 'green' : 'yellow'}
								size="xs"
							>
								{b.status}
							</Badge>
						</Group>
					)
				})}
			</Stack>
		</Card>
	)
}

/* ГЭП-7: лента событий с иконками в цветных квадратах + временные метки */
function EventsFeed() {
	/* Маппинг цветов на иконки */
	const getEventIcon = (color: string) => {
		switch (color) {
			case 'ev-green':
				return AprilIconSuccess
			case 'ev-yellow':
				return AprilIconCoins
			case 'ev-blue':
				return AprilIconCalendar
			default:
				return AprilIconSuccess
		}
	}

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
				{mockEvents.map((e, i) => {
					const Icon = getEventIcon(e.color)
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
									backgroundColor: e.iconBg,
									flexShrink: 0,
									color: e.iconColor,
								}}
							>
								<Icon size={16} />
							</div>
							<div style={{ flex: 1 }}>
								<Text size="sm">{e.text}</Text>
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
 * Главная страница (Dashboard) — моки по прототипу.
 * Данные статические. API подключится в F2.
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

			{/* ГЭП-6 + ГЭП-12: stat cards — первый зелёный, третий с suffix */}
			<Group gap="md" wrap="wrap">
				<StatCard
					title="Баланс баллов"
					value="1 250"
					subtitle="+500 баллов в июне"
					icon={AprilIconCoins}
					green
				/>
				<StatCard
					title="Активные льготы"
					value="3"
					subtitle="Из 5 доступных"
					icon={AprilIconSuccess}
				/>
				<StatCard
					title="До конца периода"
					value="47"
					suffix="дн"
					subtitle="Период: янв — июн 2025"
					icon={AprilIconCalendar}
				/>
			</Group>

			{/* ГЭП-3: layout 2 колонки */}
			<Group wrap="nowrap" gap="md" align="flex-start">
				{/* Левая колонка: льготы + события */}
				<div style={{ flex: '1 1 auto', minWidth: 0 }}>
					<Stack gap="lg">
						<ActiveBenefitsList />
						<EventsFeed />
					</Stack>
				</div>

				{/* Правая колонка: быстрые действия (292px) */}
				<QuickActionsGrid onActionClick={showF2Toast} />
			</Group>
		</Stack>
	)
}
