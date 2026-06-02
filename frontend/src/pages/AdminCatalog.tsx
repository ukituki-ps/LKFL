import {
	Card,
	Text,
	Stack,
	Table,
	Badge,
	Title,
	Group,
	Button,
	TextInput,
	Textarea,
	Select,
	Modal,
	Paper,
	NumberInput,
} from '@mantine/core'
import { useState } from 'react'
import {
	AprilIconPlus,
	AprilIconEdit,
	AprilIconTrash,
	AprilIconCoins,
	AprilIconFileText,
} from '@ukituki-ps/april-ui'

/* ─── Types ─── */

interface CatalogCard {
	id: number
	title: string
	provider: string
	category: string
	description: string
	cost: number
	status: 'active' | 'draft' | 'archived'
	slug: string
}

/* ─── Mock data ─── */

const mockCards: CatalogCard[] = [
	{ id: 1, title: 'ДМС Стандарт', provider: 'Согласие', category: 'health', description: 'Базовое медицинское страхование', cost: 500, status: 'active', slug: 'dms-standard' },
	{ id: 2, title: 'Фитнес-клуб', provider: 'World Class', category: 'sport', description: 'Абонемент в фитнес-клуб', cost: 300, status: 'active', slug: 'fitness-club' },
	{ id: 3, title: 'Онлайн-кинотеатр', provider: 'Иви', category: 'entertainment', description: 'Подписка на онлайн-кинотеатр', cost: 100, status: 'active', slug: 'online-cinema' },
	{ id: 4, title: 'Психолог', provider: 'Yana', category: 'health', description: 'Консультации психолога', cost: 200, status: 'draft', slug: 'psychologist' },
	{ id: 5, title: 'Мерч', provider: 'Внутренний', category: 'merch', description: 'Корпоративный мерч', cost: 150, status: 'archived', slug: 'merch' },
]

const categoryConfig: Record<string, { label: string; color: string }> = {
	health: { label: 'Здоровье', color: 'red' },
	sport: { label: 'Спорт', color: 'green' },
	entertainment: { label: 'Развлечения', color: 'violet' },
	merch: { label: 'Мерч', color: 'blue' },
	education: { label: 'Образование', color: 'orange' },
}

const statusConfig: Record<string, { color: string; label: string }> = {
	active: { color: 'green', label: 'Активна' },
	draft: { color: 'yellow', label: 'Черновик' },
	archived: { color: 'gray', label: 'Архив' },
}

/* ─── Form defaults ─── */

const emptyCard: Omit<CatalogCard, 'id'> = {
	title: '',
	provider: '',
	category: 'health',
	description: '',
	cost: 0,
	status: 'draft',
	slug: '',
}

/* ─── Components ─── */

function CardForm({
	initial,
	onSave,
	onClose,
}: {
	initial: CatalogCard | null
	onSave: (card: Partial<CatalogCard>) => void
	onClose: () => void
}) {
	const [form, setForm] = useState({
		title: initial?.title ?? emptyCard.title,
		provider: initial?.provider ?? emptyCard.provider,
		category: initial?.category ?? emptyCard.category,
		description: initial?.description ?? emptyCard.description,
		cost: initial?.cost ?? emptyCard.cost,
		status: initial?.status ?? emptyCard.status,
		slug: initial?.slug ?? emptyCard.slug,
	})

	const canSave = form.title.trim() && form.provider.trim()

	return (
		<Stack gap="sm">
			<TextInput
				label="Название"
				placeholder="Например: ДМС Стандарт"
				value={form.title}
				onChange={(e) => setForm({ ...form, title: e.target.value })}
				size="xs"
			/>
			<TextInput
				label="Провайдер"
				placeholder="Например: Согласие"
				value={form.provider}
				onChange={(e) => setForm({ ...form, provider: e.target.value })}
				size="xs"
			/>
			<Select
				label="Категория"
				value={form.category}
				onChange={(val) => setForm({ ...form, category: val || 'health' })}
				data={Object.entries(categoryConfig).map(([value, { label }]) => ({ value, label }))}
				size="xs"
			/>
			<Textarea
				label="Описание"
				placeholder="Краткое описание льготы..."
				value={form.description}
				onChange={(e) => setForm({ ...form, description: e.target.value })}
				minRows={3}
				size="xs"
			/>
			<NumberInput
				label="Стоимость (баллы)"
				value={form.cost}
				onChange={(val) => setForm({ ...form, cost: typeof val === 'number' ? val : 0 })}
				min={0}
				size="xs"
			/>
			<Select
				label="Статус"
				value={form.status}
				onChange={(val) => setForm({ ...form, status: (val || 'draft') as CatalogCard['status'] })}
				data={[
					{ value: 'active', label: 'Активна' },
					{ value: 'draft', label: 'Черновик' },
					{ value: 'archived', label: 'Архив' },
				]}
				size="xs"
			/>
			<Group justify="flex-end" gap={8}>
				<Button variant="subtle" size="xs" onClick={onClose}>
					Отмена
				</Button>
				<Button size="xs" disabled={!canSave} onClick={() => onSave(form)}>
					{initial ? 'Сохранить' : 'Создать'}
				</Button>
			</Group>
		</Stack>
	)
}

function MetricsStub() {
	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group gap="sm" align="center" mb="md">
				<AprilIconCoins size={18} />
				<Text fw={600} size="md">
					Метрики каталога
				</Text>
			</Group>

			<div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12 }}>
				<Paper withBorder style={{ padding: 16, borderRadius: 8, textAlign: 'center' }}>
					<Text size="xl" fw={800} c="green">
						{mockCards.filter((c) => c.status === 'active').length}
					</Text>
					<Text size="xs" c="dimmed" mt={4}>
						Активных карточек
					</Text>
				</Paper>
				<Paper withBorder style={{ padding: 16, borderRadius: 8, textAlign: 'center' }}>
					<Text size="xl" fw={800}>
						{mockCards.length}
					</Text>
					<Text size="xs" c="dimmed" mt={4}>
						Всего карточек
					</Text>
				</Paper>
				<Paper withBorder style={{ padding: 16, borderRadius: 8, textAlign: 'center' }}>
					<Text size="xl" fw={800} c="orange">
						—
					</Text>
					<Text size="xs" c="dimmed" mt={4}>
						Конверсия (F2)
					</Text>
				</Paper>
			</div>
		</Card>
	)
}

/* ─── Page ─── */

/**
 * Admin Catalog — CRUD карточек каталога льгот.
 */
export function AdminCatalog() {
	const [cards, setCards] = useState(mockCards)
	const [modalMode, setModalMode] = useState<'add' | 'edit' | null>(null)
	const [editingCard, setEditingCard] = useState<CatalogCard | null>(null)

	const handleAdd = () => {
		setEditingCard(null)
		setModalMode('add')
	}

	const handleEdit = (card: CatalogCard) => {
		setEditingCard(card)
		setModalMode('edit')
	}

	const handleSave = (data: Partial<CatalogCard>) => {
		if (modalMode === 'add') {
			setCards((prev) => [
				...prev,
				{ ...(data as CatalogCard), id: Math.max(...prev.map((c) => c.id), 0) + 1 },
			])
		} else if (modalMode === 'edit' && editingCard) {
			setCards((prev) =>
				prev.map((c) =>
					c.id === editingCard.id ? { ...c, ...data } : c,
				),
			)
		}
		setModalMode(null)
		setEditingCard(null)
	}

	const handleDelete = (id: number) => {
		setCards((prev) => prev.filter((c) => c.id !== id))
	}

	return (
		<Stack gap="lg">
			<div>
				<Title order={1} style={{ fontSize: 24, fontWeight: 800, marginBottom: 4 }}>
					Каталог льгот
				</Title>
				<Text size="sm" c="dimmed">
					Управление карточками льгот и метрики
				</Text>
			</div>

			{/* Metrics */}
			<MetricsStub />

			{/* Cards table */}
			<Card
				withBorder
				style={{
					borderRadius: 'var(--brand-radius-card, 14px)',
					boxShadow: 'var(--brand-shadow-card)',
				}}
			>
				<Group justify="space-between" align="center" mb="md">
					<Group gap="sm" align="center">
						<AprilIconFileText size={18} />
						<Text fw={600} size="md">
							Карточки льгот
						</Text>
						<Badge variant="light" color="blue" size="xs">
							{cards.length}
						</Badge>
					</Group>
					<Button
						size="xs"
						leftSection={<AprilIconPlus size={12} />}
						onClick={handleAdd}
					>
						Добавить
					</Button>
				</Group>

				<Table highlightOnHover>
					<thead>
						<tr>
							<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
								Название
							</th>
							<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
								Провайдер
							</th>
							<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
								Категория
							</th>
							<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
								Стоимость
							</th>
							<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px' }}>
								Статус
							</th>
							<th style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px', color: 'var(--brand-text-subtle)', padding: '8px 12px', width: 120 }}>
								—
							</th>
						</tr>
					</thead>
					<tbody>
						{cards.map((card) => (
							<tr key={card.id}>
								<td>
									<Text size="sm" fw={500}>
										{card.title}
									</Text>
									<Text size="xs" c="dimmed" mt={2}>
										{card.description}
									</Text>
								</td>
								<td>
									<Text size="sm">{card.provider}</Text>
								</td>
								<td>
									<Badge
										variant="light"
										color={categoryConfig[card.category]?.color || 'gray'}
										size="xs"
									>
										{categoryConfig[card.category]?.label || card.category}
									</Badge>
								</td>
								<td>
									<Text size="sm">{card.cost} балл.</Text>
								</td>
								<td>
									<Badge
										variant="light"
										color={statusConfig[card.status]?.color || 'gray'}
										size="xs"
									>
										{statusConfig[card.status]?.label || card.status}
									</Badge>
								</td>
								<td>
									<Group gap={4}>
										<Button
											variant="subtle"
											size="xs"
											leftSection={<AprilIconEdit size={12} />}
											onClick={() => handleEdit(card)}
										>
											Ред.
										</Button>
										<Button
											variant="subtle"
											color="red"
											size="xs"
											leftSection={<AprilIconTrash size={12} />}
											onClick={() => handleDelete(card.id)}
										>
											Удал.
										</Button>
									</Group>
								</td>
							</tr>
						))}
					</tbody>
				</Table>
			</Card>

			{/* Add/Edit modal */}
			<Modal
				opened={modalMode !== null}
				onClose={() => {
					setModalMode(null)
					setEditingCard(null)
				}}
				title={modalMode === 'add' ? 'Новая карточка льготы' : `Редактирование: ${editingCard?.title}`}
				size="md"
			>
				<CardForm
					initial={editingCard}
					onSave={handleSave}
					onClose={() => {
						setModalMode(null)
						setEditingCard(null)
					}}
				/>
			</Modal>
		</Stack>
	)
}
