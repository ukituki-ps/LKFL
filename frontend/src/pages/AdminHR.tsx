import {
	Card,
	Text,
	Stack,
	Table,
	Badge,
	Title,
	Group,
	Paper,
	Button,
	TextInput,
	Select,
	Modal,
} from '@mantine/core'
import { useState } from 'react'
import {
	AprilIconUsers,
	AprilIconCalendar,
	AprilIconSuccess,
	AprilIconEdit,
} from '@ukituki-ps/april-ui'

/* ─── Types ─── */

interface User {
	id: number
	email: string
	firstName: string
	lastName: string
	status: 'active' | 'inactive' | 'pending'
	roles: string[]
}

interface AccrualPeriod {
	id: number
	label: string
	startDate: string
	endDate: string
	status: 'active' | 'upcoming' | 'closed'
}

/* ─── Mock data ─── */

const mockUsers: User[] = [
	{ id: 1, email: 'ivanov@company.ru', firstName: 'Иван', lastName: 'Иванов', status: 'active', roles: ['employee'] },
	{ id: 2, email: 'petrova@company.ru', firstName: 'Мария', lastName: 'Петрова', status: 'active', roles: ['employee', 'catalog_manager'] },
	{ id: 3, email: 'sidorov@company.ru', firstName: 'Алексей', lastName: 'Сидоров', status: 'pending', roles: ['employee'] },
	{ id: 4, email: 'kuznetsova@company.ru', firstName: 'Елена', lastName: 'Кузнецова', status: 'active', roles: ['hr'] },
	{ id: 5, email: 'volkov@company.ru', firstName: 'Дмитрий', lastName: 'Волков', status: 'inactive', roles: ['employee'] },
]

const mockPeriods: AccrualPeriod[] = [
	{ id: 1, label: 'Текущий период', startDate: '01.01.2025', endDate: '30.06.2025', status: 'active' },
	{ id: 2, label: 'Следующий период', startDate: '01.07.2025', endDate: '31.12.2025', status: 'upcoming' },
	{ id: 3, label: 'Прошлый период', startDate: '01.07.2024', endDate: '31.12.2024', status: 'closed' },
]

/* ─── Helpers ─── */

const statusConfig: Record<string, { color: string; label: string }> = {
	active: { color: 'green', label: 'Активен' },
	inactive: { color: 'gray', label: 'Неактивен' },
	pending: { color: 'yellow', label: 'Ожидает' },
}

const periodStatusConfig: Record<string, { color: string; label: string }> = {
	active: { color: 'green', label: 'Активен' },
	upcoming: { color: 'blue', label: 'Предстоящий' },
	closed: { color: 'gray', label: 'Закрыт' },
}

/* ─── Components ─── */

function UsersTable({ users }: { users: User[] }) {
	const [search, setSearch] = useState('')
	const [editingUser, setEditingUser] = useState<User | null>(null)
	const [editForm, setEditForm] = useState<{ firstName: string; lastName: string; roles: string }>({
		firstName: '',
		lastName: '',
		roles: 'employee',
	})

	const filtered = users.filter(
		(u) =>
			u.email.toLowerCase().includes(search.toLowerCase()) ||
			u.firstName.toLowerCase().includes(search.toLowerCase()) ||
			u.lastName.toLowerCase().includes(search.toLowerCase()),
	)

	const handleEdit = (user: User) => {
		setEditingUser(user)
		setEditForm({
			firstName: user.firstName,
			lastName: user.lastName,
			roles: user.roles[0] || 'employee',
		})
	}

	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group justify="space-between" align="center" mb="md">
				<Group gap="sm" align="center">
					<AprilIconUsers size={18} />
					<Text fw={600} size="md">
						Пользователи
					</Text>
					<Badge variant="light" color="blue" size="xs">
						{users.length}
					</Badge>
				</Group>
				<TextInput
					placeholder="Поиск по email или имени..."
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					size="xs"
					style={{ width: 240 }}
					radius="md"
				/>
			</Group>

			<Table highlightOnHover>
				<thead>
					<tr>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							ID
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Email
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Имя
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Фамилия
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Статус
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Роли
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px', width: 80 }}>
							—
						</th>
					</tr>
				</thead>
				<tbody>
					{filtered.map((user) => (
						<tr key={user.id}>
							<td>
								<Text size="sm" c="dimmed">
									{user.id}
								</Text>
							</td>
							<td>
								<Text size="sm" fw={500}>
									{user.email}
								</Text>
							</td>
							<td>
								<Text size="sm">{user.firstName}</Text>
							</td>
							<td>
								<Text size="sm">{user.lastName}</Text>
							</td>
							<td>
								<Badge
									variant="light"
									color={statusConfig[user.status]?.color || 'gray'}
									size="xs"
								>
									{statusConfig[user.status]?.label || user.status}
								</Badge>
							</td>
							<td>
								<Group gap={4}>
									{user.roles.map((role) => (
										<Badge key={role} variant="light" color="blue" size="xs">
											{role}
										</Badge>
									))}
								</Group>
							</td>
							<td>
								<Group gap={4}>
									<Button
										variant="subtle"
										size="xs"
										leftSection={<AprilIconEdit size={12} />}
										onClick={() => handleEdit(user)}
									>
										Редакт.
									</Button>
								</Group>
							</td>
						</tr>
					))}
				</tbody>
			</Table>

			{/* Edit modal */}
			<Modal
				opened={!!editingUser}
				onClose={() => setEditingUser(null)}
				title={`Редактирование: ${editingUser?.firstName} ${editingUser?.lastName}`}
				size="sm"
			>
				<Stack gap="sm">
					<TextInput
						label="Имя"
						value={editForm.firstName}
						onChange={(e) => setEditForm({ ...editForm, firstName: e.target.value })}
						size="xs"
					/>
					<TextInput
						label="Фамилия"
						value={editForm.lastName}
						onChange={(e) => setEditForm({ ...editForm, lastName: e.target.value })}
						size="xs"
					/>
					<Select
						label="Роль"
						value={editForm.roles}
						onChange={(val) => setEditForm({ ...editForm, roles: val || 'employee' })}
						data={[
							{ value: 'employee', label: 'Сотрудник' },
							{ value: 'hr', label: 'HR' },
							{ value: 'catalog_manager', label: 'Менеджер каталога' },
							{ value: 'admin', label: 'Администратор' },
						]}
						size="xs"
					/>
					<Group justify="flex-end" gap={8}>
						<Button variant="subtle" size="xs" onClick={() => setEditingUser(null)}>
							Отмена
						</Button>
						<Button size="xs" onClick={() => setEditingUser(null)}>
							Сохранить
						</Button>
					</Group>
				</Stack>
			</Modal>
		</Card>
	)
}

function AccrualPeriods({ periods }: { periods: AccrualPeriod[] }) {
	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group gap="sm" align="center" mb="md">
				<AprilIconCalendar size={18} />
				<Text fw={600} size="md">
					Периоды начислений
				</Text>
			</Group>

			<Table highlightOnHover>
				<thead>
					<tr>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Период
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Начало
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Конец
						</th>
						<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
							Статус
						</th>
					</tr>
				</thead>
				<tbody>
					{periods.map((period) => (
						<tr key={period.id}>
							<td>
								<Text size="sm" fw={500}>
									{period.label}
								</Text>
							</td>
							<td>
								<Text size="sm">{period.startDate}</Text>
							</td>
							<td>
								<Text size="sm">{period.endDate}</Text>
							</td>
							<td>
								<Badge
									variant="light"
									color={periodStatusConfig[period.status]?.color || 'gray'}
									size="xs"
								>
									{periodStatusConfig[period.status]?.label || period.status}
								</Badge>
							</td>
						</tr>
					))}
				</tbody>
			</Table>
		</Card>
	)
}

function GamificationStub() {
	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group gap="sm" align="center" mb="md">
				<AprilIconSuccess size={18} />
				<Text fw={600} size="md">
					Геймификация
				</Text>
			</Group>

			<Paper
				style={{
					padding: 24,
					textAlign: 'center',
					borderRadius: 8,
					background: 'var(--brand-row)',
				}}
			>
				<Text size="sm" c="dimmed">
					Настройка ачивок и системы лояльности будет доступна после M25.
				</Text>
				<Text size="xs" c="dimmed" mt={4}>
					Конфигурация ачивок, бейджей и рейтинговых таблиц.
				</Text>
			</Paper>
		</Card>
	)
}

/* ─── Page ─── */

/**
 * Admin HR — управление пользователями, периодами, геймификация.
 */
export function AdminHR() {
	return (
		<Stack gap="lg">
			<div>
				<Title order={1} style={{ fontSize: 24, fontWeight: 800, marginBottom: 4 }}>
					HR — Управление
				</Title>
				<Text size="sm" c="dimmed">
					Пользователи, периоды начислений, геймификация
				</Text>
			</div>

			<UsersTable users={mockUsers} />
			<AccrualPeriods periods={mockPeriods} />
			<GamificationStub />
		</Stack>
	)
}
