import {
	Card,
	Text,
	Stack,
	Title,
	Group,
	Button,
	TextInput,
	Textarea,
	Modal,
	Paper,
	Badge,
} from '@mantine/core'
import { useState } from 'react'
import {
	AprilIconFileText,
	AprilIconPlus,
	AprilIconEdit,
	AprilIconTrash,
} from '@ukituki-ps/april-ui'

/* ─── Types ─── */

interface FaqEntry {
	id: number
	question: string
	answer: string
}

interface BannerEntry {
	id: number
	title: string
	content: string
	status: 'active' | 'draft' | 'archived'
	position: string
}

/* ─── Mock data ─── */

const mockFaq: FaqEntry[] = [
	{ id: 1, question: 'Как получить льготу?', answer: 'Перейдите в каталог льгот, выберите интересующую карточку и нажмите «Подключить». Льгота активируется в течение 1-2 рабочих дней.' },
	{ id: 2, question: 'Как начисляются баллы?', answer: 'Баллы начисляются автоматически в начале каждого периода. Также можно получить дополнительные баллы за прохождение опросов и активность в системе.' },
	{ id: 3, question: 'Можно ли передать баллы коллеге?', answer: 'Перевод баллов между сотрудниками будет доступен в ближайшем обновлении. Следите за новостями в разделе «События».' },
	{ id: 4, question: 'Как изменить данные профиля?', answer: 'Перейдите в раздел «Профиль» в правом верхнем углу. Вы можете изменить контактные данные и настройки уведомлений.' },
	{ id: 5, question: 'Куда обращаться при проблемах?', answer: 'Используйте раздел «Поддержка» — выберите тему обращения и опишите проблему. Мы ответим в течение 1 рабочего дня.' },
]

const mockBanners: BannerEntry[] = [
	{ id: 1, title: 'Лето 2025 — новые льготы', content: 'Посмотрите обновлённый каталог летних льгот', status: 'active', position: 'main' },
	{ id: 2, title: 'Опрос удовлетворённости', content: 'Пройдите опрос и получите 100 баллов', status: 'active', position: 'sidebar' },
	{ id: 3, title: 'ДМС — расширение покрытия', content: 'Обновлённые условия ДМС для всех сотрудников', status: 'draft', position: 'main' },
]

const statusConfig: Record<string, { color: string; label: string }> = {
	active: { color: 'green', label: 'Активен' },
	draft: { color: 'yellow', label: 'Черновик' },
	archived: { color: 'gray', label: 'Архив' },
}

/* ─── Components ─── */

function FaqSection({
	items,
	onEdit,
	onDelete,
	onAdd,
}: {
	items: FaqEntry[]
	onEdit: (item: FaqEntry) => void
	onDelete: (id: number) => void
	onAdd: () => void
}) {
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
					<AprilIconFileText size={18} />
					<Text fw={600} size="md">
						FAQ — Вопросы и ответы
					</Text>
					<Badge variant="light" color="blue" size="xs">
						{items.length}
					</Badge>
				</Group>
				<Button
					size="xs"
					leftSection={<AprilIconPlus size={12} />}
					onClick={onAdd}
				>
					Добавить
				</Button>
			</Group>

			<Stack gap="sm">
				{items.map((item) => (
					<Paper
						key={item.id}
						withBorder
						style={{
							padding: '12px 16px',
							borderRadius: 8,
						}}
					>
						<Group justify="space-between" align="flex-start">
							<div style={{ flex: 1 }}>
								<Group gap={8} align="center" mb={4}>
									<Text size="xs" fw={600} c="dimmed">
										#{item.id}
									</Text>
									<Text size="sm" fw={600}>
										{item.question}
									</Text>
								</Group>
								<Text size="sm" c="dimmed" style={{ paddingLeft: 24 }}>
									{item.answer}
								</Text>
							</div>
							<Group gap={4}>
								<Button
									variant="subtle"
									size="xs"
									leftSection={<AprilIconEdit size={12} />}
									onClick={() => onEdit(item)}
								>
									Ред.
								</Button>
								<Button
									variant="subtle"
									color="red"
									size="xs"
									leftSection={<AprilIconTrash size={12} />}
									onClick={() => onDelete(item.id)}
								>
									Удал.
								</Button>
							</Group>
						</Group>
					</Paper>
				))}
			</Stack>
		</Card>
	)
}

function BannerSection({
	items,
}: {
	items: BannerEntry[]
}) {
	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Group justify="space-between" align="center" mb="md">
				<Text fw={600} size="md">
					Баннеры
				</Text>
				<Button size="xs" leftSection={<AprilIconPlus size={12} />} onClick={() => {}}>
					Добавить
				</Button>
			</Group>

			{items.length === 0 ? (
				<Paper style={{ padding: 24, textAlign: 'center', borderRadius: 8, background: 'var(--brand-row)' }}>
					<Text size="sm" c="dimmed">
						Баннеры не добавлены.
					</Text>
				</Paper>
			) : (
				<Stack gap="sm">
					{items.map((banner) => (
						<Paper
							key={banner.id}
							withBorder
							style={{
								padding: '12px 16px',
								borderRadius: 8,
							}}
						>
							<Group justify="space-between" align="center">
								<div>
									<Text size="sm" fw={500}>
										{banner.title}
									</Text>
									<Text size="xs" c="dimmed" mt={2}>
										{banner.content}
									</Text>
									<Group gap={8} mt={4}>
										<Badge variant="light" color={statusConfig[banner.status]?.color || 'gray'} size="xs">
											{statusConfig[banner.status]?.label || banner.status}
										</Badge>
										<Badge variant="light" color="blue" size="xs">
											{banner.position}
										</Badge>
									</Group>
								</div>
								<Group gap={4}>
									<Button variant="subtle" size="xs" leftSection={<AprilIconEdit size={12} />}>
										Ред.
									</Button>
									<Button variant="subtle" color="red" size="xs" leftSection={<AprilIconTrash size={12} />}>
										Удал.
									</Button>
								</Group>
							</Group>
						</Paper>
					))}
				</Stack>
			)}
		</Card>
	)
}

function CardDescriptionsStub() {
	return (
		<Card
			withBorder
			style={{
				borderRadius: 'var(--brand-radius-card, 14px)',
				boxShadow: 'var(--brand-shadow-card)',
			}}
		>
			<Text fw={600} size="md" mb="md">
				Описания карточек
			</Text>

			<Paper
				style={{
					padding: 24,
					textAlign: 'center',
					borderRadius: 8,
					background: 'var(--brand-row)',
				}}
			>
				<Text size="sm" c="dimmed">
					Управление описаниями и локализация карточек будет доступна после M24.
				</Text>
				<Text size="xs" c="dimmed" mt={4}>
					SEO-описания, мультиязычность, A/B тесты текстов.
				</Text>
			</Paper>
		</Card>
	)
}

/* ─── FAQ Form ─── */

function FaqForm({
	initial,
	onSave,
	onClose,
}: {
	initial: FaqEntry | null
	onSave: (data: { question: string; answer: string }) => void
	onClose: () => void
}) {
	const [question, setQuestion] = useState(initial?.question ?? '')
	const [answer, setAnswer] = useState(initial?.answer ?? '')

	const canSave = question.trim() && answer.trim()

	return (
		<Stack gap="sm">
			<TextInput
				label="Вопрос"
				placeholder="Введите вопрос..."
				value={question}
				onChange={(e) => setQuestion(e.target.value)}
				size="xs"
			/>
			<Textarea
				label="Ответ"
				placeholder="Введите ответ..."
				value={answer}
				onChange={(e) => setAnswer(e.target.value)}
				minRows={4}
				size="xs"
			/>
			<Group justify="flex-end" gap={8}>
				<Button variant="subtle" size="xs" onClick={onClose}>
					Отмена
				</Button>
				<Button size="xs" disabled={!canSave} onClick={() => onSave({ question, answer })}>
					{initial ? 'Сохранить' : 'Создать'}
				</Button>
			</Group>
		</Stack>
	)
}

/* ─── Page ─── */

/**
 * Admin Content — управление контентом: FAQ, баннеры, описания карточек.
 */
export function AdminContent() {
	const [faqItems, setFaqItems] = useState(mockFaq)
	const [modalMode, setModalMode] = useState<'add' | 'edit' | null>(null)
	const [editingFaq, setEditingFaq] = useState<FaqEntry | null>(null)

	const handleAddFaq = () => {
		setEditingFaq(null)
		setModalMode('add')
	}

	const handleEditFaq = (item: FaqEntry) => {
		setEditingFaq(item)
		setModalMode('edit')
	}

	const handleSaveFaq = (data: { question: string; answer: string }) => {
		if (modalMode === 'add') {
			setFaqItems((prev) => [
				...prev,
				{ ...data, id: Math.max(...prev.map((f) => f.id), 0) + 1 },
			])
		} else if (modalMode === 'edit' && editingFaq) {
			setFaqItems((prev) =>
				prev.map((f) =>
					f.id === editingFaq.id ? { ...f, ...data } : f,
				),
			)
		}
		setModalMode(null)
		setEditingFaq(null)
	}

	const handleDeleteFaq = (id: number) => {
		setFaqItems((prev) => prev.filter((f) => f.id !== id))
	}

	return (
		<Stack gap="lg">
			<div>
				<Title order={1} style={{ fontSize: 24, fontWeight: 800, marginBottom: 4 }}>
					Контент
				</Title>
				<Text size="sm" c="dimmed">
					FAQ, баннеры, описания карточек
				</Text>
			</div>

			<FaqSection
				items={faqItems}
				onAdd={handleAddFaq}
				onEdit={handleEditFaq}
				onDelete={handleDeleteFaq}
			/>

			<BannerSection items={mockBanners} />

			<CardDescriptionsStub />

			{/* FAQ Add/Edit modal */}
			<Modal
				opened={modalMode !== null}
				onClose={() => {
					setModalMode(null)
					setEditingFaq(null)
				}}
				title={modalMode === 'add' ? 'Новый вопрос' : `Редактирование: ${editingFaq?.question}`}
				size="md"
			>
				<FaqForm
					initial={editingFaq}
					onSave={handleSaveFaq}
					onClose={() => {
						setModalMode(null)
						setEditingFaq(null)
					}}
				/>
			</Modal>
		</Stack>
	)
}
