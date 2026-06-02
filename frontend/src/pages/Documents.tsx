import {
	Card,
	Text,
	Stack,
	Table,
	Badge,
	Button,
	Skeleton,
	Title,
} from '@mantine/core'
import { useQuery } from '@tanstack/react-query'
import {
	AprilIconSuccess,
	AprilIconDownload,
} from '@ukituki-ps/april-ui'
import { getDocuments } from '@/api/documents'

const headerCellStyle: React.CSSProperties = {
	textTransform: 'uppercase',
	letterSpacing: '0.5px',
	fontSize: 11,
	fontWeight: 600,
	color: 'var(--brand-text-subtle)',
	background: 'var(--brand-row)',
	padding: '10px 16px',
}

const downloadButtonStyle: React.CSSProperties = {
	padding: '6px 12px',
	background: 'var(--brand-row)',
	border: '1px solid var(--brand-border)',
	borderRadius: 6,
	fontSize: 12,
	fontWeight: 600,
	color: 'var(--brand-text-muted)',
}

/**
 * Страница «Документы» — данные через React Query.
 *
 * ГЭП-9: secondary doc-meta строка под названием.
 * ГЭП-10: кнопка «Скачать» по стилю прототипа.
 * P2: table header uppercase + letter-spacing.
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
			<div>
				<Title order={1} style={{ fontSize: 24, fontWeight: 800, marginBottom: 4 }}>
					Документы
				</Title>
				<Text size="sm" c="dimmed">
					Заявления, согласия и сформированные документы
				</Text>
			</div>

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
								<th style={headerCellStyle}>
									Документ
								</th>
								<th style={headerCellStyle}>
									Тип
								</th>
								<th style={headerCellStyle}>
									Дата
								</th>
								<th style={headerCellStyle}>
									Статус
								</th>
								<th style={{ ...headerCellStyle, width: 100 }}>
									—
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
										{/* ГЭП-10: кнопка «Скачать» по стилю прототипа */}
										<Button
											variant="default"
											size="xs"
											leftSection={<AprilIconDownload size={13} />}
											style={downloadButtonStyle}
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
