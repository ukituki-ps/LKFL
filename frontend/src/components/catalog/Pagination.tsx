import { Group, Button, Text } from '@mantine/core'

interface PaginationProps {
	page: number
	perPage: number
	total: number
	totalPages: number
	onPageChange: (page: number) => void
}

/**
 * Компонент пагинации.
 * Показывает номер страницы, диапазон элементов и кнопки навигации.
 * Скрывается если totalPages <= 1.
 */
export function Pagination({
	page,
	perPage,
	total,
	totalPages,
	onPageChange,
}: PaginationProps) {
	if (totalPages <= 1) return null

	const startItem = (page - 1) * perPage + 1
	const endItem = Math.min(page * perPage, total)

	// Генерация страницы для отображения (1, ..., N-1, N, N+1, ..., Last)
	const visiblePages = Array.from({ length: totalPages }, (_, i) => i + 1)
		.filter((p) => p === 1 || p === totalPages || Math.abs(p - page) <= 1)
		.reduce<(number | string)[]>((acc, p, i, arr) => {
			if (i > 0) {
				const prev = arr[arr.length - 1]
				if (typeof prev === 'number' && p - prev > 1) {
					acc.push('...')
				}
			}
			acc.push(p)
			return acc
		}, [])

	return (
		<Group justify="space-between" align="center" mt="xl" wrap="wrap">
			<Text size="sm" c="dimmed">
				{startItem}&ndash;{endItem} из {total}
			</Text>

			<Group gap="sm">
				<Button
					variant="subtle"
					size="sm"
					disabled={page <= 1}
					onClick={() => onPageChange(page - 1)}
				>
					&larr; Назад
				</Button>

				{visiblePages.map((item, idx) =>
					typeof item === 'string' ? (
						<Text key={`ellipsis-${idx}`} size="sm" c="dimmed">
							&hellip;
						</Text>
					) : (
						<Button
							key={item}
							variant={item === page ? 'filled' : 'subtle'}
							size="sm"
							onClick={() => onPageChange(item)}
						>
							{item}
						</Button>
					)
				)}

				<Button
					variant="subtle"
					size="sm"
					disabled={page >= totalPages}
					onClick={() => onPageChange(page + 1)}
				>
					Далее &rarr;
				</Button>
			</Group>
		</Group>
	)
}
