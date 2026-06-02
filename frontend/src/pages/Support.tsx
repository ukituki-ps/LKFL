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
	Paper,
	Skeleton,
} from '@mantine/core'
import { useQuery, useMutation } from '@tanstack/react-query'
import {
	AprilIconHelp,
	AprilIconSend,
	AprilIconCheckCircle,
} from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'
import { useState } from 'react'
import { getFaq, postSupportTicket } from '@/api/support'

/**
 * Страница «Поддержка» — данные через React Query.
 *
 * ГЭП-11: success state формы после сабмита.
 */
export function Support() {
	/* ГЭП-11: tracking формы / успеха */
	const [submitted, setSubmitted] = useState(false)

	// Form fields
	const [topic, setTopic] = useState<string | null>(null)
	const [title, setTitle] = useState('')
	const [message, setMessage] = useState('')

	// ─── React Query: FAQ ───

	const { data: faq, isLoading: faqLoading, isError: faqError } = useQuery({
		queryKey: ['support', 'faq'],
		queryFn: getFaq,
	})

	// ─── React Query: Support Ticket Mutation ───

	const { mutate: submitTicket, isPending: submitting } = useMutation({
		mutationFn: postSupportTicket,
		onSuccess: () => {
			setSubmitted(true)
			setTopic(null)
			setTitle('')
			setMessage('')
		},
	})

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

						{faqLoading ? (
							<Skeleton height={300} />
						) : faqError ? (
							<Text c="red">Не удалось загрузить данные. Попробуйте позже.</Text>
						) : (
							<Accordion variant="separated">
								{(faq ?? []).map((item, i) => (
									<Accordion.Item key={`faq-${i}`} value={`faq-${i}`}>
										<Accordion.Control>{item.title}</Accordion.Control>
										<Accordion.Panel>
											<Text size="sm" c="dimmed">
												{item.content}
											</Text>
										</Accordion.Panel>
									</Accordion.Item>
								))}
							</Accordion>
						)}
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
								{submitted ? 'Обращение отправлено!' : 'Написать в поддержку'}
							</Text>
							<StubBadge />
						</Group>

						{submitted ? (
							/* ГЭП-11: success block */
							<Paper
								style={{
									borderRadius: 'var(--brand-radius-card, 14px)',
									background: '#F0FDF4',
									border: '1px solid #BBF7D0',
									textAlign: 'center',
									padding: 20,
								}}
							>
								<Stack align="center" gap="sm">
									<div
										style={{
											width: 56,
											height: 56,
											display: 'flex',
											alignItems: 'center',
											justifyContent: 'center',
											borderRadius: '50%',
											background: '#DCFCE7',
											color: '#16A34A',
										}}
									>
										<AprilIconCheckCircle size={32} />
									</div>
									<Text fw={600} size="md" c="#166534">
										Обращение отправлено!
									</Text>
									<Text size="sm" c="#166534" opacity={0.7}>
										Мы ответим в течение 1 рабочего дня
									</Text>
									<Button
										variant="subtle"
										size="sm"
										onClick={() => setSubmitted(false)}
										mt="xs"
									>
										Новое обращение
									</Button>
								</Stack>
							</Paper>
						) : (
							/* Form */
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
									value={topic}
									onChange={setTopic}
									radius="md"
									clearable
								/>

								<TextInput
									label="Заголовок"
									placeholder="Кратко опишите проблему"
									value={title}
									onChange={(e) => setTitle(e.target.value)}
									radius="md"
								/>

								<Textarea
									label="Описание"
									placeholder="Подробно опишите вашу проблему..."
									value={message}
									onChange={(e) => setMessage(e.target.value)}
									minRows={4}
									radius="md"
								/>

								<Button
									leftSection={<AprilIconSend size={16} />}
									radius="md"
									size="md"
									loading={submitting}
									disabled={submitting}
									onClick={() => {
										submitTicket({ topic: topic ?? 'other', message })
									}}
								>
									Отправить обращение
								</Button>
							</Stack>
						)}
					</Card>
				</div>
			</Group>
		</Stack>
	)
}
