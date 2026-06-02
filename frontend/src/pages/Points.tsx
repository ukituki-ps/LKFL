import {
	Card,
	Text,
	Group,
	Stack,
	Progress,
	Skeleton,
	Title,
} from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import { AprilIconSuccess, AprilIconClose } from '@ukituki-ps/april-ui'
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
			<div>
				<Title order={1} style={{ fontSize: 24, fontWeight: 800, marginBottom: 4 }}>
					Мои баллы
				</Title>
				<Text size="sm" c="dimmed">
					История начислений и списаний
				</Text>
			</div>

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
						<Text fw={600} size="md" mb="md">
								Транзакции
							</Text>

						<Group gap="sm" mb="md">
							{(['all', 'credits', 'debits'] as FilterType[]).map((f) => (
								<button
									key={f}
									onClick={() => setFilter(f)}
									style={{
										padding: '5px 12px',
										borderRadius: 20,
										border: '1.5px solid',
										borderColor: filter === f ? 'var(--brand-text)' : 'var(--brand-border)',
										background: filter === f ? 'var(--brand-text)' : 'transparent',
										color: filter === f ? '#fff' : 'var(--brand-text-muted)',
										fontSize: 13,
										fontWeight: 500,
										cursor: 'pointer',
										transition: 'all 150ms',
										fontFamily: 'inherit',
									}}
								>
									{f === 'all' ? 'Все' : f === 'credits' ? 'Начисления' : 'Списания'}
								</button>
							))}
						</Group>

						{transactionsLoading ? (
							<Skeleton height={200} />
						) : transactionsError ? (
							<Text c="red">Не удалось загрузить данные. Попробуйте позже.</Text>
						) : (
								<div>
									{displayTransactions.map((t, i) => (
										<div
											key={`${t.date}-${t.description}-${i}`}
											style={{
												display: 'flex',
												alignItems: 'center',
												gap: 12,
												padding: '13px 18px',
												borderBottom: '1px solid var(--brand-row)',
											}}
										>
											{/* tx-icon 36×36 */}
											<div
												style={{
													width: 36,
													height: 36,
													borderRadius: 10,
													display: 'flex',
													alignItems: 'center',
													justifyContent: 'center',
													flexShrink: 0,
													background: t.type === 'credit' ? '#DCFCE7' : 'var(--brand-bg)',
													color: t.type === 'credit' ? '#16A34A' : 'var(--brand-text-subtle)',
												}}
											>
												{t.type === 'credit' ? (
													<AprilIconSuccess size={18} />
												) : (
													<AprilIconClose size={18} />
												)}
											</div>
											{/* Description + date */}
											<div style={{ flex: 1, minWidth: 0 }}>
												<Text size="sm" fw={500}>
													{t.description}
												</Text>
												<Text size="xs" c="dimmed">
													{t.date}
												</Text>
											</div>
											{/* Amount with suffix */}
											<Text
												size="sm"
												fw={600}
												c={t.type === 'credit' ? 'green' : 'red'}
											>
												{t.type === 'credit' ? '+' : '−'}{Math.abs(t.amount)}{' '}б
											</Text>
										</div>
									))}
								</div>
							)}
					</Card>
				</div>
			</Group>
		</Stack>
	)
}
