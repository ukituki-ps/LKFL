import { apiRequest } from './client'
import type {
	EngagementTypeResponse,
	EngagementCategoryResponse,
	ListResponse,
} from './types'

// Public API

export async function getCategories(): Promise<EngagementCategoryResponse[]> {
	return apiRequest<EngagementCategoryResponse[]>('/api/v1/engagements/categories')
}

export interface GetEngagementsParams {
	type?: string
	status?: string
	category?: string
	search?: string
	page?: number
	per_page?: number
}

export async function getEngagements(
	params: GetEngagementsParams = {}
): Promise<ListResponse> {
	const searchParams = new URLSearchParams()
	if (params.type) searchParams.set('type', params.type)
	if (params.status) searchParams.set('status', params.status)
	if (params.category) searchParams.set('category', params.category)
	if (params.search) searchParams.set('search', params.search)
	if (params.page) searchParams.set('page', String(params.page))
	if (params.per_page) searchParams.set('per_page', String(params.per_page))

	const query = searchParams.toString()
	return apiRequest<ListResponse>(
		`/api/v1/engagements${query ? '?' + query : ''}`
	)
}

export async function getEngagement(id: string): Promise<EngagementTypeResponse> {
	return apiRequest<EngagementTypeResponse>(`/api/v1/engagements/${id}`)
}

// Admin API

export interface CreateCategoryRequest {
	slug: string
	name: string
	icon: string
	sort_order: number
}

export async function createCategory(req: CreateCategoryRequest): Promise<EngagementCategoryResponse> {
	return apiRequest<EngagementCategoryResponse>('/admin/engagements/categories', {
		method: 'POST',
		body: JSON.stringify(req),
	})
}

export async function updateCategory(id: string, req: CreateCategoryRequest): Promise<EngagementCategoryResponse> {
	return apiRequest<EngagementCategoryResponse>(`/admin/engagements/categories/${id}`, {
		method: 'PUT',
		body: JSON.stringify(req),
	})
}

export async function deleteCategory(id: string): Promise<null> {
	return apiRequest<null>(`/admin/engagements/categories/${id}`, {
		method: 'DELETE',
	})
}

export async function updateEngagementStatus(
	id: string,
	status: string
): Promise<EngagementTypeResponse> {
	return apiRequest<EngagementTypeResponse>(`/admin/engagements/types/${id}/status`, {
		method: 'PATCH',
		body: JSON.stringify({ status }),
	})
}
