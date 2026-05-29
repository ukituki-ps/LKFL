import {
	Card,
	Text,
	Group,
	Stack,
	Accordion,
	Button,
	TextInput,
	Textarea,
	Select,
} from '@mantine/core'
import { AprilIconHelp, AprilIconSend } from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'

/* ─── Mock FAQ ─── */

const mockFaq = [
	{
		title: 'Как активировать льготу?',
		content:
			'Перейдите в каталог льгот, выберите нужную и нажмите «Активировать». Льгота станет доступна в разделе «Мои льготы».',
	},
	{
		title: 'Как начисляются баллы?',
		content:
			'Баллы начисляются ежемесячно в соответствии с вашим пакетом льгот. Дополнительно баллы можно получить, проходя опросы и участвуя в активностях.',
	},
	{
		title: 'Что делать, если льгота не работает?',
		content:
			'Обратитесь в поддержку через эту страницу. Приложите скриншот ошибки и описание проблемы. Мы ответим в течение 2 рабочих дней.',
	},
	{
		title: 'Как сменить пакет льгот?',
		content:
			'Смена пакета доступна в разделе «Мои баллы» → «Настройки пакета». Изменения вступят в силу с начала следующего периода.',
	},
	{
		title: 'Могу ли я передать баллы коллеге?',
		content:
			'Перевод баллов между сотрудниками не поддерживается. Баллы привязаны к вашему личному аккаунту.',
	},
]

/**
 * Страница «Поддержка» — заглушка по прототипу.
 * FAQ аккордеон (левая колонка) + форма обращения (правая колонка).
 */
export function Support() {
	return (
		<Stack gap="lg">
			{/* Heading */}
			<Group justify="space-between">
				<Group gap={8} align="center">
					<AprilIconHelp size={20} style={{ color: 'var(--brand-green)' }} />
					<Text fw={600} size="lg">
						Поддержка
					</Text>
				</Group>
				<StubBadge />
			</Group>

			{/* Two-column layout */}
			<Group wrap="nowrap" gap="md">
				{/* FAQ — left column */}
				<div style={{ flex: '1 1 55%' }}>
					<Card
						withBorder
						style={{
							borderRadius: 'var(--brand-radius-card, 14px)',
							boxShadow: 'var(--brand-shadow-card)',
						}}
					>
						<Group justify="space-between" mb="md">
							<Text fw={600} size="md">
								Частые вопросы
							</Text>
							<StubBadge />
						</Group>

						<Accordion variant="separated">
							{mockFaq.map((item, i) => (
								<Accordion.Item key={i} value={`faq-${i}`}>
									<Accordion.Control>{item.title}</Accordion.Control>
									<Accordion.Panel>
										<Text size="sm" c="dimmed">
											{item.content}
										</Text>
									</Accordion.Panel>
								</Accordion.Item>
							))}
						</Accordion>
					</Card>
				</div>

				{/* Contact form — right column */}
				<div style={{ flex: '1 1 45%' }}>
					<Card
						withBorder
						style={{
							borderRadius: 'var(--brand-radius-card, 14px)',
							boxShadow: 'var(--brand-shadow-card)',
						}}
					>
						<Group justify="space-between" mb="md">
							<Text fw={600} size="md">
								Написать в поддержку
							</Text>
							<StubBadge />
						</Group>

						<Stack gap="md">
							<Select
								label="Тема обращения"
								placeholder="Выберите тему"
								data={[
									{ value: 'benefit', label: 'Проблема с льготой' },
									{ value: 'points', label: 'Вопрос по баллам' },
									{ value: 'technical', label: 'Техническая проблема' },
									{ value: 'other', label: 'Другое' },
								]}
								radius="md"
								clearable
							/>

							<TextInput
								label="Заголовок"
								placeholder="Кратко опишите проблему"
								radius="md"
							/>

							<Textarea
								label="Описание"
								placeholder="Подробно опишите вашу проблему..."
								minRows={4}
								radius="md"
							/>

							<Button
								leftSection={<AprilIconSend size={16} />}
								radius="md"
								size="md"
							>
								Отправить обращение
							</Button>
						</Stack>
					</Card>
				</div>
			</Group>
		</Stack>
	)
}
