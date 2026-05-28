import { Select, Group } from '@mantine/core'
import type { EngagementCategoryResponse } from '@/api/types'

interface FilterBarProps {
	categories: EngagementCategoryResponse[]
	type: string
	status: string
	category: string
	onChange: (key: string, value: string) => void
}

const typeOptions = [
	{ value: 'benefit', label: 'Льготы' },
	{ value: 'activity', label: 'Активности' },
]

const statusOptions = [
	{ value: 'active', label: 'Активные' },
	{ value: 'promo', label: 'Промо' },
]

/**
 * Панель фильтров каталога: тип, статус, категория.
 * Все фильтры синхронизируются с URL query params через onChange.
 */
export function FilterBar({
	categories,
	type,
	status,
	category,
	onChange,
}: FilterBarProps) {
	const categoryOptions = categories.map((cat) => ({
		value: cat.slug,
		label: cat.name,
	}))

	return (
		<Group gap="sm" mb="md" wrap="wrap">
			<Select
				data={typeOptions}
				value={type}
				onChange={(v) => onChange('type', v || '')}
				label="Тип"
				clearable
				size="sm"
				radius="md"
				w={140}
			/>

			<Select
				data={statusOptions}
				value={status}
				onChange={(v) => onChange('status', v || 'active')}
				label="Статус"
				size="sm"
				radius="md"
				w={140}
			/>

			<Select
				data={categoryOptions}
				value={category}
				onChange={(v) => onChange('category', v || '')}
				label="Категория"
				clearable
				size="sm"
				radius="md"
				w={160}
			/>
		</Group>
	)
}
