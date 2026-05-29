import { Card, Text, Group, Stack, Title, Badge, Paper } from '@mantine/core'
import {
	AprilIconCoins,
	AprilIconSuccess,
	AprilIconCalendar,
	AprilIconSearch,
	AprilIconFileText,
	AprilIconHelp,
	AprilIconEdit,
	AprilIconSend,
	AprilIconGift,
	AprilIconDumbbell,
	AprilIconHeart,
	AprilIconUser,
} from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'

/* ─── Types ─── */

interface BenefitItem {
	name: string
	provider: string
	icon: React.ComponentType<{ size?: number | string }>
	status: string
}

/* ─── Mock data (заменится на API в F2) ─── */

const mockActiveBenefits: BenefitItem[] = [
	{ name: 'Онлайн-кинотеатр', provider: 'KION', icon: AprilIconGift, status: 'Активна' },
	{ name: 'Фитнес-клуб', provider: 'World Class', icon: AprilIconDumbbell, status: 'Активна' },
	{ name: 'Страховка ДМС', provider: 'СОГАЗ', icon: AprilIconHeart, status: 'Ожидает' },
]

const mockEvents = [
	{ text: 'Новая льгота: онлайн-кинотеатр', color: '#00B33C' },
	{ text: 'Начислено 500 баллов за опрос', color: '#F59E0B' },
	{ text: 'Обновлены условия программы', color: '#3B82F6' },
]

interface QuickAction {
	label: string
	icon: React.ComponentType<{ size?: number | string }>
}

const mockQuickActions: QuickAction[] = [
	{ label: 'Каталог льгот', icon: AprilIconSearch },
	{ label: 'Мои документы', icon: AprilIconFileText },
	{ label: 'Поддержка', icon: AprilIconHelp },
	{ label: 'Бонусы', icon: AprilIconEdit },
	{ label: 'Обратная связь', icon: AprilIconSend },
	{ label: 'Профиль', icon: AprilIconUser },
]

/* ─── Components ─── */

function StatCard({
	title,
	value,
	subtitle,
	icon: Icon,
}: {
	title: string
	value: string
	subtitle: string
	icon: React.ComponentType<{ size?: number | string }>
}) {
	return (
		<Card
			withBorder
			padding="md"
			style={{
				flex: 1,
				minWidth: 180,
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group justify="space-between" mb="xs">
				<Text size="xs" c="dimmed">
					{title}
				</Text>
				<StubBadge />
			</Group>
			<Group align="center" gap={8}>
				<Icon size={20} />
				<Text fw={700} size="xl" style={{ color: 'var(--brand-text)' }}>
					{value}
				</Text>
			</Group>
			<Text size="xs" c="dimmed" mt="xs">
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

function EventsFeed() {
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
			<Stack gap="sm">
				{mockEvents.map((e, i) => (
					<Group key={i} gap="sm" align="center">
						<div
							style={{
								width: 8,
								height: 8,
								borderRadius: '50%',
								backgroundColor: e.color,
								flexShrink: 0,
							}}
						/>
						<Text size="sm">{e.text}</Text>
					</Group>
				))}
			</Stack>
		</Card>
	)
}

function QuickActionsGrid() {
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
					Быстрые действия
				</Text>
				<StubBadge />
			</Group>
			<div
				style={{
					display: 'grid',
					gridTemplateColumns: 'repeat(3, 1fr)',
					gap: 8,
				}}
			>
				{mockQuickActions.map((a) => (
					<Paper
						key={a.label}
						withBorder
						style={{
							padding: 12,
							borderRadius: 'var(--brand-radius-btn, 6px)',
							textAlign: 'center',
							cursor: 'pointer',
							transition: 'background-color 150ms',
						}}
						onMouseEnter={(e) => {
							e.currentTarget.style.backgroundColor = 'var(--brand-green-light)'
						}}
						onMouseLeave={(e) => {
							e.currentTarget.style.backgroundColor = 'transparent'
						}}
					>
						<a.icon size={20} />
						<Text size="xs" fw={500} mt={4}>
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
 */
export function Dashboard() {
	const today = new Date().toLocaleDateString('ru-RU', {
		day: 'numeric',
		month: 'long',
		year: 'numeric',
	})

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

			{/* Stat cards */}
			<Group gap="md" wrap="wrap">
				<StatCard
					title="Баланс баллов"
					value="1 250"
					subtitle="Достаточно для 2 льгот"
					icon={AprilIconCoins}
				/>
				<StatCard
					title="Активные льготы"
					value="3"
					subtitle="Из 5 доступных"
					icon={AprilIconSuccess}
				/>
				<StatCard
					title="До конца периода"
					value="18 дней"
					subtitle="Сброс 15 июня"
					icon={AprilIconCalendar}
				/>
			</Group>

			{/* Two-column layout */}
			<Group wrap="nowrap" gap="md">
				<div style={{ flex: '1 1 55%' }}>
					<ActiveBenefitsList />
				</div>
				<div style={{ flex: '1 1 45%' }}>
					<Stack gap="lg">
						<EventsFeed />
					</Stack>
				</div>
			</Group>

			{/* Quick actions */}
			<QuickActionsGrid />
		</Stack>
	)
}
