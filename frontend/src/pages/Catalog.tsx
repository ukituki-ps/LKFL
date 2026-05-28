import { useCallback, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getEngagements, getCategories } from '@/api/engagements'
import { EngagementGrid } from '@/components/catalog/EngagementCard'
import { FilterBar } from '@/components/catalog/FilterBar'
import { SearchInput } from '@/components/catalog/SearchInput'
import { Pagination } from '@/components/catalog/Pagination'
import { Title, Text, Stack, Loader, Button } from '@mantine/core'

/**
 * Страница каталога льгот и активностей.
 *
 * Фильтры синхронизируются с URL query params для shareable links:
 * - type: benefit | activity
 * - status: active (default) | promo
 * - category: slug категории
 * - search: текстовый поиск (debounced)
 * - page: номер страницы
 * - per_page: элементов на странице
 */
export function Catalog() {
	const [searchParams, setSearchParams] = useSearchParams()
	const [search, setSearch] = useState('')

	// ─── Фильтры из URL ───

	const type = searchParams.get('type') || ''
	const status = searchParams.get('status') || 'active'
	const category = searchParams.get('category') || ''
	const page = parseInt(searchParams.get('page') || '1', 10)
	const perPage = parseInt(searchParams.get('per_page') || '20', 10)

	// ─── Загрузка категорий для фильтра ───

	const { data: categories = [] } = useQuery({
		queryKey: ['categories'],
		queryFn: () => getCategories(),
	})

	// ─── Загрузка энгейджментов ───

	const { data, isLoading, isError, refetch } = useQuery({
		queryKey: ['engagements', type, status, category, search, page, perPage],
		queryFn: () =>
			getEngagements({
				type,
				status,
				category,
				search,
				page,
				per_page: perPage,
			}),
	})

	// ─── Обработчики ───

	// Debounced search — обновляем URL когда debounced search меняется
	const handleSearch = useCallback(
		(value: string) => {
			setSearch(value)
			setSearchParams((prev) => {
				if (value) {
					prev.set('search', value)
				} else {
					prev.delete('search')
				}
				prev.set('page', '1')
				return prev
			})
		},
		[setSearchParams],
	)

	const handleFilterChange = useCallback(
		(key: string, value: string) => {
			setSearchParams((prev) => {
				if (value) {
					prev.set(key, value)
				} else {
					prev.delete(key)
				}
				prev.set('page', '1')
				return prev
			})
		},
		[setSearchParams],
	)

	const handlePageChange = useCallback(
		(p: number) => {
			setSearchParams((prev) => {
				prev.set('page', String(p))
				return prev
			})
		},
		[setSearchParams],
	)

	// ─── Состояния ───

	if (isLoading) {
		return (
			<div
				style={{
					display: 'flex',
					alignItems: 'center',
					justifyContent: 'center',
					padding: 40,
				}}
			>
				<Loader />
			</div>
		)
	}

	if (isError) {
		return (
			<Stack align="center" justify="center" style={{ minHeight: '400px' }}>
				<Text c="red">Ошибка загрузки каталога</Text>
				<Button variant="subtle" onClick={() => refetch()}>
					Повторить
				</Button>
			</Stack>
		)
	}

	const engagements = data?.data || []
	const pagination = data?.pagination

	// ─── Рендер ───

	return (
		<Stack gap="md">
			<Title order={2}>Каталог льгот</Title>

			<FilterBar
				categories={categories}
				type={type}
				status={status}
				category={category}
				onChange={handleFilterChange}
			/>

			<SearchInput value={search} onChange={handleSearch} />

			{engagements.length === 0 ? (
				<Text c="dimmed" ta="center" style={{ marginTop: 40 }}>
					Нет доступных льгот
				</Text>
			) : (
				<EngagementGrid engagements={engagements} />
			)}

			{pagination && (
				<Pagination
					page={pagination.page}
					perPage={pagination.per_page}
					total={pagination.total}
					totalPages={pagination.total_pages}
					onPageChange={handlePageChange}
				/>
			)}
		</Stack>
	)
}
