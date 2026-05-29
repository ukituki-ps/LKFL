import { Group } from '@mantine/core'
import { AprilFilterPills } from '@ukituki-ps/april-ui'
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
 * Панель фильтров каталога: тип, статус, категория.
 *
 * Использует AprilFilterPills из DS v0.1.16.
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
			<AprilFilterPills
				items={typeOptions}
				active={type || ''}
				onChange={(v) => onChange('type', v)}
			/>

			{/* Status filter */}
			<AprilFilterPills
				items={statusOptions}
				active={status || 'active'}
				onChange={(v) => onChange('status', v)}
			/>

			{/* Category filter */}
			{categories.length > 0 && (
				<AprilFilterPills
					items={categoryOptions}
					active={category || ''}
					onChange={(v) => onChange('category', v)}
				/>
			)}
		</Group>
	)
}
