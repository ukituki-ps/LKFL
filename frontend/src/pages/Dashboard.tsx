import { useQuery } from '@tanstack/react-query'
import { getUserProfile } from '@/api/user'
import { Card, Text, Group, Stack, Title, Loader } from '@mantine/core'
import { Link } from 'react-router-dom'

/**
 * Определяет приветствие по текущему часу суток.
 */
function getGreeting(): string {
	const hour = new Date().getHours()
	if (hour < 6) return 'Доброй ночи'
	if (hour < 12) return 'Доброе утро'
	if (hour < 18) return 'Добрый день'
	return 'Добрый вечер'
}

/**
 * Карточка статистики-заглушка.
 * Значение «—» — подключится в F2 (M23+).
 */
function StatCard({
	title,
	value,
	description,
}: {
	title: string
	value: string
	description: string
}) {
	return (
		<Card withBorder padding="md" style={{ flex: 1, minWidth: 180 }}>
			<Text size="xs" c="dimmed" mb="xs">
				{title}
			</Text>
			<Text fw={600} size="xl">
				{value}
			</Text>
			<Text size="xs" c="dimmed" mt="xs">
				{description}
			</Text>
		</Card>
	)
}

/**
 * Ссылка-карточка для быстрых действий.
 */
function QuickActionLink({ to, label }: { to: string; label: string }) {
	return (
		<Link to={to} style={{ textDecoration: 'none' }}>
			<Card withBorder padding="sm" style={{ cursor: 'pointer' }}>
				<Text size="sm">
					{label} →
				</Text>
			</Card>
		</Link>
	)
}

/**
 * Главная страница (Dashboard) — заглушка с приветствием
 * и placeholder-блоками. Данные подключаются в F2.
 */
export function Dashboard() {
	const { data: user, isLoading, isError } = useQuery({
		queryKey: ['user-profile'],
		queryFn: () => getUserProfile(),
	})

	if (isLoading) {
		return (
			<div
				style={{
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center',
					padding: 40,
				}}
			>
				<Loader />
			</div>
		)
	}

	if (isError) {
		return (
			<Text c="red" ta="center" style={{ marginTop: 40 }}>
				Ошибка загрузки профиля
			</Text>
		)
	}

	const greeting = getGreeting()
	const userName = user?.first_name || 'Сотрудник'

	return (
		<Stack gap="lg">
			{/* Greeting */}
			<Title order={2}>
				{greeting}, {userName}!
			</Title>

			{/* Stat cards — placeholder, подключится в F2 */}
			<Group gap="md" wrap="wrap">
				<StatCard
					title="Баланс баллов"
					value="—"
					description="Подключится в F2"
				/>
				<StatCard
					title="Активные льготы"
					value="—"
					description="Подключится в F2"
				/>
				<StatCard
					title="Доступные льготы"
					value="—"
					description="Подключится в F2"
				/>
			</Group>

			{/* Event feed placeholder */}
			<Card withBorder>
				<Text fw={600} mb="md">
					Последние события
				</Text>
				<Text c="dimmed" size="sm">
					События появятся после активации льгот
				</Text>
			</Card>

			{/* Quick actions */}
			<Card withBorder>
				<Text fw={600} mb="md">
					Быстрые действия
				</Text>
				<Stack gap="sm">
					<QuickActionLink
						to="/catalog"
						label="Просмотреть каталог льгот"
					/>
					<QuickActionLink to="/documents" label="Мои документы" />
					<QuickActionLink to="/support" label="Обратиться в поддержку" />
				</Stack>
			</Card>
		</Stack>
	)
}
