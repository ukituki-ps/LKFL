import {
	Card,
	Text,
	Group,
	Stack,
	Table,
	Badge,
	Button,
} from '@mantine/core'
import {
	AprilIconFileText,
	AprilIconSuccess,
	AprilIconDownload,
} from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'

/* ─── Mock data ─── */

/* ГЭП-9: добавлено docMeta поле */
const mockDocuments = [
	{
		name: 'Согласие на обработку ПДн',
		docMeta: 'Платформа · ПДн',
		type: 'Согласие',
		date: '01.05.2026',
		status: 'Подписано',
		statusColor: 'green',
	},
	{
		name: 'Заявление на льготу «Фитнес-клуб»',
		docMeta: 'World Class · Фитнес',
		type: 'Заявление',
		date: '15.05.2026',
		status: 'Одобрено',
		statusColor: 'blue',
	},
	{
		name: 'Полис ДМС — СОГАЗ',
		docMeta: 'АльфаСтрахование · ДМС',
		type: 'Полис',
		date: '10.05.2026',
		status: 'Активен',
		statusColor: 'green',
	},
	{
		name: 'Заявление на льготу «Онлайн-кинотеатр»',
		docMeta: 'KION · Развлечения',
		type: 'Заявление',
		date: '20.05.2026',
		status: 'На рассмотрении',
		statusColor: 'yellow',
	},
]

/**
 * Страница «Документы» — заглушка по прототипу.
 *
 * ГЭП-9: secondary doc-meta строка под названием.
 * ГЭП-10: кнопка «Скачать» с текстом + иконка.
 */
export function Documents() {
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
						{mockDocuments.map((doc) => (
							<tr key={doc.name}>
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
			</Card>
		</Stack>
	)
}
