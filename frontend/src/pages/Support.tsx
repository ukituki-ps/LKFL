import {
	Card,
	Text,
	Group,
	Stack,
	Button,
	Textarea,
	Select,
	Paper,
	Skeleton,
	Title,
} from '@mantine/core'
import { useQuery, useMutation } from '@tanstack/react-query'
import {
	AprilIconSend,
	AprilIconCheckCircle,
	AprilIconChevronRight,
} from '@ukituki-ps/april-ui'
import { useState } from 'react'
import { getFaq, postSupportTicket } from '@/api/support'

/**
 * Кастомный accordion-item для FAQ.
 *
 * Chevron-down реализуем через AprilIconChevronRight + rotate(90deg).
 */
function FaqItem({
	title,
	content,
}: {
	title: string
	content: string
}) {
	const [isOpen, setIsOpen] = useState(false)

	return (
		<div style={{ borderBottom: '1px solid var(--brand-row)' }}>
			<div
				onClick={() => setIsOpen(!isOpen)}
				style={{
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'space-between',
					padding: '16px 20px',
					cursor: 'pointer',
					fontSize: 13,
					fontWeight: 600,
					color: isOpen ? 'var(--brand-green, #00B33C)' : 'inherit',
					transition: 'color 0.2s',
					userSelect: 'none',
				}}
			>
				<span>{title}</span>
				<AprilIconChevronRight
					size={18}
					style={{
						transform: isOpen ? 'rotate(-90deg)' : 'rotate(90deg)',
						transition: 'transform 0.2s',
						flexShrink: 0,
						marginLeft: 12,
						color: isOpen ? 'var(--brand-green, #00B33C)' : 'var(--mantine-color-dimmed)',
					}}
				/>
			</div>
			{isOpen && (
				<div
					style={{
						padding: '0 20px 16px',
						fontSize: 13,
						color: 'var(--brand-text-muted)',
						lineHeight: 1.6,
					}}
				>
					<Text size="sm" c="dimmed">
						{content}
					</Text>
				</div>
			)}
		</div>
	)
}

/**
 * Страница «Поддержка» — данные через React Query.
 *
 * ГЭП-11: success state формы после сабмита.
 * P2: кастомный accordion, правая колонка 380px.
 */
export function Support() {
	/* ГЭП-11: tracking формы / успеха */
	const [submitted, setSubmitted] = useState(false)

	// Form fields
	const [topic, setTopic] = useState<string | null>(null)
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
			setMessage('')
		},
	})

	return (
		<Stack gap="lg">
			{/* Heading */}
			<div>
				<Title order={1} style={{ fontSize: 24, fontWeight: 800, marginBottom: 4 }}>
					Поддержка
				</Title>
				<Text size="sm" c="dimmed">
					Частые вопросы и обратная связь
				</Text>
			</div>

			{/* Two-column layout */}
			<Group wrap="nowrap" gap="md">
				{/* FAQ — left column */}
				<div style={{ flex: '1 1 55%' }}>
					<Card
						withBorder
						style={{
							borderRadius: 'var(--brand-radius-card, 14px)',
							boxShadow: 'var(--brand-shadow-card)',
							padding: 0,
						}}
					>
						<Text fw={600} size="md" style={{ padding: '16px 20px 0' }}>
							Частые вопросы
						</Text>

						{faqLoading ? (
							<Skeleton height={300} />
						) : faqError ? (
							<Text c="red" style={{ padding: '0 20px' }}>
								Не удалось загрузить данные. Попробуйте позже.
							</Text>
						) : (
							<div>
								{(faq ?? []).map((item, i) => (
									<FaqItem key={`faq-${i}`} title={item.title} content={item.content} />
								))}
							</div>
						)}
					</Card>
				</div>

				{/* Contact form — right column (fixed 380px) */}
				<div style={{ flex: '0 0 380px' }}>
					<Card
						withBorder
						style={{
							borderRadius: 'var(--brand-radius-card, 14px)',
							boxShadow: 'var(--brand-shadow-card)',
						}}
					>
						<Text fw={600} size="md" mb="md">
							{submitted ? 'Обращение отправлено!' : 'Написать в поддержку'}
						</Text>

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

								<Textarea
									label="Сообщение"
									placeholder="Опишите ваш вопрос подробно..."
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
									style={{ textTransform: 'uppercase' }}
								>
									ОТПРАВИТЬ
								</Button>
							</Stack>
						)}
					</Card>
				</div>
			</Group>
		</Stack>
	)
}
