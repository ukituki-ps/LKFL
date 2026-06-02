import {
	Card,
	Text,
	Group,
	Stack,
	Progress,
	Paper,
	SegmentedControl,
	Skeleton,
} from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { AprilIconSuccess, AprilIconClose } from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'
import { useState } from 'react'
import { getPointsBalance, getTransactions } from '@/api/points'

type FilterType = 'all' | 'credits' | 'debits'

/**
 * Страница «Мои баллы» — данные через React Query.
 *
 * ГЭП-4: layout 2 колонки — слева баланс (зелёная карточка с категориями внутри),
 * справа транзакции с фильтрами.
 */
export function Points() {
	const [filter, setFilter] = useState<FilterType>('all')

	// ─── React Query ───

	const { data: balance, isLoading: balanceLoading, isError: balanceError } = useQuery({
		queryKey: ['points'],
		queryFn: getPointsBalance,
	})

	const { data: transactions, isLoading: transactionsLoading, isError: transactionsError } = useQuery({
		queryKey: ['points', 'transactions'],
		queryFn: getTransactions,
	})

	// Клиентская фильтрация транзакций (map plural filter → singular type)
	const displayTransactions = (transactions ?? []).filter((t) => {
		if (filter === 'all') return true
		return t.type === (filter === 'credits' ? 'credit' : 'debit')
	})

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
						{balanceLoading ? (
							<Skeleton height={300} />
						) : balanceError ? (
							<Text c="white" opacity={0.8}>Не удалось загрузить данные. Попробуйте позже.</Text>
						) : (
							<Stack gap="md">
								{/* Balance */}
								<div>
									<Text size="sm" style={{ opacity: 0.85 }}>
										Доступно баллов
									</Text>
									<Text fw={800} style={{ fontSize: 48, lineHeight: 1.1, marginTop: 8 }}>
										{balance!.total}
									</Text>
									<Text size="xs" style={{ opacity: 0.7, marginTop: 8 }}>
										Период: май 2026 · Сброс 15 июня
									</Text>
								</div>

								{/* Категории — внутри зелёной карточки */}
								<div style={{ borderTop: '1px solid rgba(255,255,255,0.2)', paddingTop: 12 }}>
									{balance!.categories.map((cat) => (
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
						)}
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

						{transactionsLoading ? (
							<Skeleton height={200} />
						) : transactionsError ? (
							<Text c="red">Не удалось загрузить данные. Попробуйте позже.</Text>
						) : (
							<Stack gap="sm">
								{displayTransactions.map((t, i) => (
									<Paper
										key={`${t.date}-${t.description}-${i}`}
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
													{t.type === 'credit' ? '+' : '-'}
													{t.amount}
												</Text>
											</Group>
										</Group>
									</Paper>
								))}
							</Stack>
						)}
					</Card>
				</div>
			</Group>
		</Stack>
	)
}
