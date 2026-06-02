import { Group } from '@mantine/core'
import type { EngagementCategoryResponse } from '@/api/types'

interface FilterBarProps {
	categories: EngagementCategoryResponse[]
	type: string
	status: string
	category: string
	onChange: (key: string, value: string) => void
}

const typeOptions = [
	{ value: '', label: 'Все' },
	{ value: 'benefit', label: 'Льготы' },
	{ value: 'activity', label: 'Активности' },
]

const statusOptions = [
	{ value: 'active', label: 'Активные' },
	{ value: 'promo', label: 'Промо' },
]

/**
 * Один pill-фильтр: border-radius 20px, border 1.5px, active/inactive стили.
 */
function FilterPill({
	items,
	active,
	onChange,
}: {
	items: { value: string; label: string }[]
	active: string
	onChange: (value: string) => void
}) {
	return (
		<Group gap={6} wrap="nowrap">
			{items.map((item) => {
				const isActive = item.value === active
				return (
					<button
						key={item.value}
						type="button"
						onClick={() => onChange(item.value)}
						style={{
							borderRadius: '20px',
							border: '1.5px solid',
							borderColor: isActive
								? 'var(--brand-green, #00B33C)'
								: 'var(--brand-border)',
							background: isActive
								? 'var(--brand-green, #00B33C)'
								: 'var(--brand-card)',
							color: isActive ? '#fff' : 'var(--brand-text-muted)',
							fontSize: '12px',
							fontWeight: 600,
							padding: '6px 14px',
							cursor: 'pointer',
							transition: 'all 0.15s',
							lineHeight: '1.2',
							whiteSpace: 'nowrap',
						}}
					>
						{item.label}
					</button>
				)
			})}
		</Group>
	)
}

/**
 * Панель фильтров каталога: тип, статус, категория.
 *
 * Кастомные filter pills по прототипу:
 * - border-radius: 20px
 * - border: 1.5px solid var(--brand-border)
 * - Active: green bg + white text + green border
 * - Inactive: card bg + muted text
 * - font-size: 12px, font-weight: 600, padding: 6px 14px
 *
 * Все фильтры синхронизируются с URL query params через onChange.
 */
export function FilterBar({
	categories,
	type,
	status,
	category,
	onChange,
}: FilterBarProps) {
	const categoryOptions = [
		{ value: '', label: 'Все' },
		...categories.map((cat) => ({ value: cat.slug, label: cat.name })),
	]

	return (
		<Group gap="md" mb="md" wrap="wrap">
			{/* Type filter */}
			<FilterPill
				items={typeOptions}
				active={type || ''}
				onChange={(v) => onChange('type', v)}
			/>

			{/* Status filter */}
			<FilterPill
				items={statusOptions}
				active={status || 'active'}
				onChange={(v) => onChange('status', v)}
			/>

			{/* Category filter */}
			{categories.length > 0 && (
				<FilterPill
					items={categoryOptions}
					active={category || ''}
					onChange={(v) => onChange('category', v)}
				/>
			)}
		</Group>
	)
}
