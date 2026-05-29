import {
	Card,
	Text,
	Group,
	Stack,
	Progress,
	Paper,
	SegmentedControl,
} from '@mantine/core'
import { AprilIconSuccess, AprilIconClose } from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'
import { useState } from 'react'

/* ─── Mock data ─── */

const mockCategories = [
	{ name: 'Фитнес', used: 450, total: 1000 },
	{ name: 'Образование', used: 200, total: 500 },
	{ name: 'Развлечения', used: 600, total: 1500 },
	{ name: 'Здоровье', used: 0, total: 750 },
]

const mockTransactions = {
	all: [
		{ date: '20.05.2026', description: 'Активация: Онлайн-кинотеатр', type: 'debit', amount: -300 },
		{ date: '18.05.2026', description: 'Начисление за опрос', type: 'credit', amount: 500 },
		{ date: '15.05.2026', description: 'Ежемесячное начисление', type: 'credit', amount: 1000 },
		{ date: '01.05.2026', description: 'Активация: Фитнес-клуб', type: 'debit', amount: -450 },
	],
	credits: [
		{ date: '18.05.2026', description: 'Начисление за опрос', type: 'credit', amount: 500 },
		{ date: '15.05.2026', description: 'Ежемесячное начисление', type: 'credit', amount: 1000 },
	],
	debits: [
		{ date: '20.05.2026', description: 'Активация: Онлайн-кинотеатр', type: 'debit', amount: -300 },
		{ date: '01.05.2026', description: 'Активация: Фитнес-клуб', type: 'debit', amount: -450 },
	],
}

type FilterType = 'all' | 'credits' | 'debits'

/**
 * Страница «Мои баллы» — заглушка по прототипу.
 *
 * ГЭП-4: layout 2 колонки — слева баланс (зелёная карточка с категориями внутри),
 * справа транзакции с фильтрами.
 */
export function Points() {
	const [filter, setFilter] = useState<FilterType>('all')

	const transactions = mockTransactions[filter]

	return (
		<Stack gap="lg">
			{/* Heading */}
			<Group justify="space-between">
				<Text fw={600} size="lg">
					Мои баллы
				</Text>
				<StubBadge />
			</Group>

			{/* ГЭП-4: layout side-by-side */}
			<Group wrap="nowrap" gap="md" align="flex-start">
				{/* Левая колонка: баланс + категории внутри зелёной карточки */}
				<div style={{ flex: '0 0 320px' }}>
					<Card
						withBorder
						padding="lg"
						style={{
							backgroundColor: 'var(--brand-green, #00B33C)',
							color: '#FFFFFF',
							borderRadius: 'var(--brand-radius-card, 14px)',
						}}
					>
						<Stack gap="md">
							{/* Balance */}
							<div>
								<Text size="sm" style={{ opacity: 0.85 }}>
									Доступно баллов
								</Text>
								<Text fw={800} style={{ fontSize: 48, lineHeight: 1.1, marginTop: 8 }}>
									1 250
								</Text>
								<Text size="xs" style={{ opacity: 0.7, marginTop: 8 }}>
									Период: май 2026 · Сброс 15 июня
								</Text>
							</div>

							{/* Категории — внутри зелёной карточки */}
							<div style={{ borderTop: '1px solid rgba(255,255,255,0.2)', paddingTop: 12 }}>
								{mockCategories.map((cat) => (
									<div key={cat.name} style={{ marginBottom: 10 }}>
										<Group justify="space-between" mb={4}>
											<Text size="sm" fw={500} style={{ color: '#FFFFFF' }}>
												{cat.name}
											</Text>
											<Text size="xs" style={{ color: 'rgba(255,255,255,0.7)' }}>
												{cat.used} / {cat.total}
											</Text>
										</Group>
										<Progress
											value={(cat.used / cat.total) * 100}
											color="#FFFFFF"
											size="sm"
											radius="xl"
											style={{ backgroundColor: 'rgba(255,255,255,0.2)' }}
										/>
									</div>
								))}
							</div>
						</Stack>
					</Card>
				</div>

				{/* Правая колонка: транзакции */}
				<div style={{ flex: '1 1 auto', minWidth: 0 }}>
					<Card
						withBorder
						style={{
							borderRadius: 'var(--brand-radius-card, 14px)',
							boxShadow: 'var(--brand-shadow-card)',
						}}
					>
						<Group justify="space-between" mb="md">
							<Text fw={600} size="md">
								Транзакции
							</Text>
							<StubBadge />
						</Group>

						<SegmentedControl
							data={[
								{ value: 'all', label: 'Все' },
								{ value: 'credits', label: 'Начисления' },
								{ value: 'debits', label: 'Списания' },
							]}
							value={filter}
							onChange={(v) => setFilter(v as FilterType)}
							radius="md"
							mb="md"
						/>

						<Stack gap="sm">
							{transactions.map((t, i) => (
								<Paper
									key={i}
									withBorder
									style={{ padding: 12, borderRadius: 'var(--brand-radius-btn, 6px)' }}
								>
									<Group justify="space-between">
										<div>
											<Text size="sm" fw={500}>
												{t.description}
											</Text>
											<Text size="xs" c="dimmed">
												{t.date}
											</Text>
										</div>
										<Group gap={4} align="center">
											{t.type === 'credit' ? (
												<AprilIconSuccess size={14} style={{ color: '#00B33C' }} />
											) : (
												<AprilIconClose size={14} style={{ color: '#EF4444' }} />
											)}
											<Text
												size="sm"
												fw={600}
												c={t.type === 'credit' ? 'green' : 'red'}
											>
												{t.type === 'credit' ? '+' : ''}
												{t.amount}
											</Text>
										</Group>
									</Group>
								</Paper>
							))}
						</Stack>
					</Card>
				</div>
			</Group>
		</Stack>
	)
}
