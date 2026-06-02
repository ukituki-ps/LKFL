import {
	Card,
	Text,
	Group,
	Stack,
	Table,
	Badge,
	Button,
	Skeleton,
} from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import {
	AprilIconFileText,
	AprilIconSuccess,
	AprilIconDownload,
} from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'
import { getDocuments } from '@/api/documents'

/**
 * Страница «Документы» — данные через React Query.
 *
 * ГЭП-9: secondary doc-meta строка под названием.
 * ГЭП-10: кнопка «Скачать» с текстом + иконка.
 */
export function Documents() {
	// ─── React Query ───

	const { data: documents, isLoading, isError } = useQuery({
		queryKey: ['documents'],
		queryFn: getDocuments,
	})

	return (
		<Stack gap="lg">
			{/* Heading */}
			<Group justify="space-between">
				<Group gap={8} align="center">
					<AprilIconFileText size={20} style={{ color: 'var(--brand-green)' }} />
					<Text fw={600} size="lg">
						Мои документы
					</Text>
				</Group>
				<StubBadge />
			</Group>

			{/* Documents table */}
			<Card
				withBorder
				style={{
					borderRadius: 'var(--brand-radius-card, 14px)',
					boxShadow: 'var(--brand-shadow-card)',
				}}
			>
				{isLoading ? (
					<Skeleton height={200} />
				) : isError ? (
					<Text c="red">Не удалось загрузить данные. Попробуйте позже.</Text>
				) : (
					<Table striped highlightOnHover>
						<thead>
							<tr>
								<th>
									<Text size="xs" fw={600} c="dimmed">
										Документ
									</Text>
								</th>
								<th>
									<Text size="xs" fw={600} c="dimmed">
										Тип
									</Text>
								</th>
								<th>
									<Text size="xs" fw={600} c="dimmed">
										Дата
									</Text>
								</th>
								<th>
									<Text size="xs" fw={600} c="dimmed">
										Статус
									</Text>
								</th>
								<th style={{ width: 100 }}>
									<Text size="xs" fw={600} c="dimmed">
										—
									</Text>
								</th>
							</tr>
						</thead>
						<tbody>
							{(documents ?? []).map((doc, i) => (
								<tr key={`${doc.name}-${i}`}>
									<td>
										<Text size="sm" fw={500}>
											{doc.name}
										</Text>
										{/* ГЭП-9: secondary строка */}
										<Text size="xs" c="dimmed" mt={2}>
											{doc.docMeta}
										</Text>
									</td>
									<td>
										<Badge
											variant="light"
											color={
												doc.type === 'Заявление'
													? 'blue'
													: doc.type === 'Согласие'
													? 'gray'
													: 'blue'
											}
											size="xs"
										>
											{doc.type}
										</Badge>
									</td>
									<td>
										<Text size="sm" c="dimmed">
											{doc.date}
										</Text>
									</td>
									<td>
										<Badge
											variant="light"
											color={doc.statusColor}
											size="xs"
											leftSection={
												doc.statusColor === 'green' ? (
													<AprilIconSuccess size={10} />
												) : undefined
											}
										>
											{doc.status}
										</Badge>
									</td>
									<td>
										{/* ГЭП-10: кнопка «Скачать» с текстом */}
										<Button
											variant="subtle"
											size="xs"
											leftSection={<AprilIconDownload size={12} />}
											onClick={() => {
												// Stub — скачивание файла в F2
											}}
										>
											Скачать
										</Button>
									</td>
								</tr>
							))}
						</tbody>
					</Table>
				)}
			</Card>
		</Stack>
	)
}
