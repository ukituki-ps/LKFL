import {
	Card,
	Text,
	Group,
	Stack,
	Table,
	Badge,
	ActionIcon,
	Tooltip,
} from '@mantine/core'
import { AprilIconFileText, AprilIconSuccess, AprilIconClipboardList } from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'

/* ─── Mock data ─── */

const mockDocuments = [
	{
		name: 'Согласие на обработку ПДн',
		type: 'Согласие',
		date: '01.05.2026',
		status: 'Подписано',
		statusColor: 'green',
	},
	{
		name: 'Заявление на льготу «Фитнес-клуб»',
		type: 'Заявление',
		date: '15.05.2026',
		status: 'Одобрено',
		statusColor: 'blue',
	},
	{
		name: 'Полис ДМС — СОГАЗ',
		type: 'Полис',
		date: '10.05.2026',
		status: 'Активен',
		statusColor: 'green',
	},
	{
		name: 'Заявление на льготу «Онлайн-кинотеатр»',
		type: 'Заявление',
		date: '20.05.2026',
		status: 'На рассмотрении',
		statusColor: 'yellow',
	},
]

/**
 * Страница «Документы» — заглушка по прототипу.
 * Таблица: документ, тип, дата, статус, «Скачать».
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
							<th style={{ width: 60 }}>
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
									<Tooltip label="Скачать">
										<ActionIcon variant="subtle" color="dimmed" size="sm">
											<AprilIconClipboardList size={14} />
										</ActionIcon>
									</Tooltip>
								</td>
							</tr>
						))}
					</tbody>
				</Table>
			</Card>
		</Stack>
	)
}
