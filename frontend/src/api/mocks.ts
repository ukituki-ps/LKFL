/**
 * Mock-данные для всех эндпоинтов, отсутствующих на бэкенде.
 *
 * Используются через VITE_USE_MOCKS=true для разработки без бэкенда.
 * Формат данных согласован со спецификацией API (doc/спецификация/).
 */

// ─── Dashboard ───

export interface DashboardStats {
  points: number
  activeBenefits: number
  daysLeft: number
}

export interface BenefitItem {
  name: string
  provider: string
  status: 'active' | 'awaiting'
  icon: string
}

export interface EventItem {
  text: string
  time: string
  color: 'green' | 'yellow' | 'blue'
}

// ─── Points ───

export interface PointsBalance {
  total: number
  categories: {
    name: string
    used: number
    total: number
  }[]
}

export interface Transaction {
  date: string
  description: string
  type: 'credit' | 'debit'
  amount: number
}

// ─── Documents ───

export interface DocumentItem {
  name: string
  docMeta: string
  type: string
  date: string
  status: string
  statusColor: string
}

// ─── Support ───

export interface FaqItem {
  title: string
  content: string
}

export interface SupportTicketRequest {
  topic: string
  message: string
}

export interface SupportTicketResponse {
  id: string
  status: string
}

// ─── Mock data ───

export const mockDashboardStats: DashboardStats = {
  points: 1250,
  activeBenefits: 3,
  daysLeft: 47,
}

export const mockDashboardBenefits: BenefitItem[] = [
  { name: 'Онлайн-кинотеатр', provider: 'KION', status: 'active', icon: 'heart' },
  { name: 'Фитнес-клуб', provider: 'World Class', status: 'active', icon: 'dumbbell' },
  { name: 'Страховка ДМС', provider: 'СОГАЗ', status: 'awaiting', icon: 'shield' },
]

export const mockDashboardEvents: EventItem[] = [
  { text: 'Новая льгота: онлайн-кинотеатр', time: 'Сегодня, 14:30', color: 'green' },
  { text: 'Начислено 500 баллов за опрос', time: 'Вчера, 18:15', color: 'yellow' },
  { text: 'Обновлены условия программы', time: '20 мая, 10:00', color: 'blue' },
]

export const mockPointsBalance: PointsBalance = {
  total: 1250,
  categories: [
    { name: 'Фитнес', used: 450, total: 1000 },
    { name: 'Образование', used: 200, total: 500 },
    { name: 'Развлечения', used: 600, total: 1500 },
    { name: 'Здоровье', used: 0, total: 750 },
  ],
}

export const mockTransactions: Transaction[] = [
  { date: '20.05.2026', description: 'Активация: Онлайн-кинотеатр', type: 'debit', amount: 300 },
  { date: '18.05.2026', description: 'Начисление за опрос', type: 'credit', amount: 500 },
  { date: '15.05.2026', description: 'Ежемесячное начисление', type: 'credit', amount: 1000 },
  { date: '01.05.2026', description: 'Активация: Фитнес-клуб', type: 'debit', amount: 450 },
]

export const mockDocuments: DocumentItem[] = [
  { name: 'Согласие на обработку ПДн', docMeta: 'Платформа · ПДн', type: 'Согласие', date: '01.05.2026', status: 'Подписано', statusColor: 'green' },
  { name: 'Заявление на льготу «Фитнес-клуб»', docMeta: 'World Class · Фитнес', type: 'Заявление', date: '15.05.2026', status: 'Одобрено', statusColor: 'blue' },
  { name: 'Полис ДМС — СОГАЗ', docMeta: 'АльфаСтрахование · ДМС', type: 'Полис', date: '10.05.2026', status: 'Активен', statusColor: 'green' },
  { name: 'Заявление на льготу «Онлайн-кинотеатр»', docMeta: 'KION · Развлечения', type: 'Заявление', date: '20.05.2026', status: 'На рассмотрении', statusColor: 'yellow' },
]

export const mockFaq: FaqItem[] = [
  { title: 'Как активировать льготу?', content: 'Перейдите в каталог льгот, выберите нужную и нажмите «Активировать». Льгота станет доступна в разделе «Мои льготы» в течение 1-2 минут.' },
  { title: 'Как начисляются баллы?', content: 'Баллы начисляются ежемесячно в соответствии с вашим пакетом льгот. Дополнительно баллы можно получить, проходя опросы и участвуя в активностях компании.' },
  { title: 'Что делать, если льгота не работает?', content: 'Обратитесь в поддержку через эту страницу. Приложите скриншот ошибки и описание проблемы. Мы ответим в течение 2 рабочих дней.' },
  { title: 'Как сменить пакет льгот?', content: 'Смена пакета доступна один раз в период (6 месяцев). Перейдите в «Мои баллы» → «Настройки пакета». Изменения вступят в силу с начала следующего периода.' },
  { title: 'Могу ли я передать баллы коллеге?', content: 'Перевод баллов между сотрудниками не поддерживается. Баллы привязаны к вашему личному аккаунту и не могут быть переданы.' },
  { title: 'Как получить доступ к платформе?', content: 'Доступ предоставляется автоматически при трудоустройстве. Если вы не видите платформу, обратитесь в HR-отдел вашей компании.' },
]
