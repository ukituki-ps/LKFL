import { useParams, useNavigate } from 'react-router-dom'
import {
	Card,
	Text,
	Group,
	Stack,
	Badge,
	Button,
	Paper,
	Tabs,
	Title,
} from '@mantine/core'
import {
	AprilIconHeart,
	AprilIconCheckCircle,
	AprilIconPlusCircle,
	AprilIconShoppingBag,
	AprilIconFileText,
	AprilIconChevronLeft,
} from '@ukituki-ps/april-ui'
import { StubBadge } from '@/components/ui/StubBadge'

/* ─── Mock data (заменится на API в F2) ─── */

interface BenefitDetail {
	title: string
	desc: string
	price: string
	provider: string
	btn: string
	active: boolean
	isDms?: boolean
}

const BENEFIT_DATA: Record<string, BenefitDetail> = {
	'dms-base': {
		title: 'ДМС — Базовая программа',
		desc: 'Полис добровольного медицинского страхования. Амбулаторное лечение, стационар, скорая помощь. Действует в клиниках-партнёрах по всей России.',
		price: 'Включено в пакет',
		provider: 'АльфаСтрахование',
		btn: 'Уже активна',
		active: true,
		isDms: true,
	},
	fitness: {
		title: 'Фитнес — World Class',
		desc: 'Доступ в сеть фитнес-клубов World Class с групповыми занятиями, бассейном и тренажёрным залом. Безлимитные посещения.',
		price: '500 баллов / мес',
		provider: 'World Class',
		btn: 'Уже активна',
		active: true,
	},
	'sport2': {
		title: 'СберСпорт — мультиспорт',
		desc: 'Доступ к 1 000+ спортивным объектам по всей России: фитнес, йога, бассейн, единоборства.',
		price: '800 баллов / мес',
		provider: 'СберСпорт',
		btn: 'Подключить',
		active: false,
	},
	food: {
		title: 'Обеды в офисе',
		desc: 'Компенсация обедов в офисе или доставки до рабочего места до 300 ₽ в день. Подключается один раз в квартал.',
		price: '300 ₽ / день',
		provider: 'Яндекс Еда for Business',
		btn: 'Подключить',
		active: false,
	},
	edu: {
		title: 'Обучение — Skillbox',
		desc: 'Доступ к профессиональным онлайн-курсам: разработка, дизайн, маркетинг, управление проектами и аналитика данных.',
		price: '1 200 баллов',
		provider: 'Skillbox',
		btn: 'Подключить',
		active: false,
	},
	psych: {
		title: 'Психолог онлайн',
		desc: '4 сессии с профессиональным психологом онлайн. Анонимно и конфиденциально. Запись через приложение Яндекс Психотерапия.',
		price: '600 баллов',
		provider: 'Яндекс Психотерапия',
		btn: 'Ожидает одобрения',
		active: false,
	},
	merch: {
		title: 'Мерч СДЭК',
		desc: 'Фирменная одежда и аксессуары СДЭК: худи, футболки, кружки, термосы. Доставка курьером на дом.',
		price: 'от 200 баллов',
		provider: 'СДЭК Store',
		btn: 'Перейти в каталог',
		active: false,
	},
	dent: {
		title: 'Стоматология',
		desc: 'Профилактические осмотры, профессиональная чистка и лечение в сети клиник Мать и дитя. До 4 визитов в год.',
		price: '950 баллов / год',
		provider: 'Мать и дитя',
		btn: 'Подключить',
		active: false,
	},
	coffee: {
		title: 'Кофе и снеки',
		desc: 'Компенсация покупки кофе и лёгких перекусов в офисе или ближайших кофейнях до 100 ₽ в рабочий день.',
		price: '100 ₽ / день',
		provider: 'Яндекс Еда for Business',
		btn: 'Подключить',
		active: false,
	},
	lang: {
		title: 'Английский язык',
		desc: 'Индивидуальные онлайн-занятия с сертифицированным преподавателем. Уровни от Beginner до Advanced. 8 занятий в месяц.',
		price: '1 500 баллов / мес',
		provider: 'Skyeng',
		btn: 'Подключить',
		active: false,
	},
	phone: {
		title: 'Корпоративная связь',
		desc: 'Безлимитный тарифный план МТС: звонки, SMS, интернет без ограничений по России. Корпоративная SIM-карта. Роуминг по льготному тарифу.',
		price: 'Включено в пакет',
		provider: 'МТС',
		btn: 'Уже активна',
		active: true,
	},
}

/**
 * Страница детализации льготы `/catalog/:slug`.
 *
 * Layout по прототипу Benefit Detail:
 * - Header: название + провайдер + badge статуса
 * - Body: описание, иконка, стоимость
 * - Footer: кнопка «Подключить» (stub → модалка в F2) или «Уже активна»
 * - Для DMS: табы «Условия» / «Полис» / «Клиники» (заглушки с StubBadge)
 */
export function CatalogDetail() {
	const { slug } = useParams<{ slug: string }>()
	const navigate = useNavigate()

	const benefit = slug ? BENEFIT_DATA[slug] : null

	if (!benefit) {
		return (
			<Stack align="center" justify="center" style={{ minHeight: '400px' }}>
				<Text size="lg" c="dimmed">Льгота не найдена</Text>
				<Button variant="subtle" onClick={() => navigate('/catalog')}>
					Вернуться в каталог
				</Button>
			</Stack>
		)
	}

	const isDms = benefit.isDms === true

	return (
		<Stack gap="lg">
			{/* Back button */}
			<Button
				variant="subtle"
				size="sm"
				onClick={() => navigate('/catalog')}
				leftSection={<AprilIconChevronLeft size={14} />}
			>
				Назад в каталог
			</Button>

			{/* Header */}
			<Group justify="space-between" wrap="wrap">
				<div>
					<Title order={2} mb={4}>
						{benefit.title}
					</Title>
					<Text size="sm" c="dimmed">
						{benefit.provider}
					</Text>
				</div>
				<Badge
					variant="light"
					color={benefit.active ? 'green' : 'yellow'}
					size="md"
				>
					{benefit.active ? 'Уже активна' : benefit.btn === 'Подключить' ? 'Доступна' : benefit.btn}
				</Badge>
			</Group>

			{isDms ? (
				/* ─── DMS view with tabs ─── */
				<Card
					withBorder
					padding="lg"
					style={{
						borderRadius: 'var(--brand-radius-card, 14px)',
						boxShadow: 'var(--brand-shadow-card)',
					}}
				>
					<Stack gap="md">
						<Text size="sm" c="dimmed">
							{benefit.desc}
						</Text>

						<Group gap={8} mb="md">
							<Text fw={700} size="lg" style={{ color: 'var(--brand-green)' }}>
								{benefit.price}
							</Text>
						</Group>

						<Tabs defaultValue="conditions" variant="pills">
							<Tabs.List>
								<Tabs.Tab value="conditions">Условия</Tabs.Tab>
								<Tabs.Tab value="policy">Полис</Tabs.Tab>
								<Tabs.Tab value="clinics">Клиники</Tabs.Tab>
							</Tabs.List>

							<Tabs.Panel value="conditions">
								<PartnerTabPanel icon="shield" title="Условия страхования" />
							</Tabs.Panel>
							<Tabs.Panel value="policy">
								<PartnerTabPanel icon="document" title="Полис ДМС" />
							</Tabs.Panel>
							<Tabs.Panel value="clinics">
								<PartnerTabPanel icon="location" title="Клиники-партнёры" />
							</Tabs.Panel>
						</Tabs>
					</Stack>
				</Card>
			) : (
				/* ─── Generic benefit view ─── */
				<Card
					withBorder
					padding="lg"
					style={{
						borderRadius: 'var(--brand-radius-card, 14px)',
						boxShadow: 'var(--brand-shadow-card)',
					}}
				>
					<Stack gap="md">
						{/* Icon */}
						<div
							style={{
								width: 56,
								height: 56,
								display: 'flex',
								alignItems: 'center',
								justifyContent: 'center',
								borderRadius: 14,
								backgroundColor: 'var(--brand-row, #F9FAFB)',
								color: 'var(--brand-green)',
							}}
						>
							<AprilIconHeart size={28} />
						</div>

						{/* Description */}
						<Text size="sm" c="dimmed">
							{benefit.desc}
						</Text>

						{/* Price */}
						<Group gap={8} mt="md">
							<Text fw={700} size="lg" style={{ color: 'var(--brand-green)' }}>
								{benefit.price}
							</Text>
						</Group>

						{/* Button */}
						<Button
							fullWidth
							radius="md"
							size="md"
							variant={benefit.active ? 'outline' : 'filled'}
							color="brand"
							leftSection={
								benefit.active
									? <AprilIconCheckCircle size={16} />
									: <AprilIconPlusCircle size={16} />
							}
							onClick={() => {
								if (!benefit.active) {
									// Stub — модалка активации в F2
								}
							}}
						>
							{benefit.btn.toUpperCase()}
						</Button>
						<StubBadge />
					</Stack>
				</Card>
			)}
		</Stack>
	)
}

function PartnerTabPanel({ icon, title }: { icon: string; title: string }) {
	const iconMap: Record<string, React.ComponentType<{ size?: number | string }>> = {
		shield: AprilIconCheckCircle,
		document: AprilIconFileText,
		location: AprilIconShoppingBag,
	}
	const Icon = iconMap[icon] || AprilIconCheckCircle

	return (
<Paper
						withBorder
						style={{
							borderRadius: 'var(--brand-radius-card, 14px)',
							textAlign: 'center',
							borderStyle: 'dashed',
							padding: 20,
						}}
					>
			<Stack align="center" gap="xs">
				<span style={{ color: 'var(--brand-green)' }}>
					<Icon size={32} />
				</span>
				<Text size="md" fw={600}>{title}</Text>
				<Text size="sm" c="dimmed">Данные появятся после F2</Text>
				<StubBadge />
			</Stack>
		</Paper>
	)
}
