// API layer barrel export
export { apiRequest } from './client'
export type { ApiError } from './client'

export type {
	EngagementCategoryResponse,
	EngagementOfferResponse,
	EngagementTypeResponse,
	PaginationResponse,
	ListResponse,
} from './types'

export {
	getCategories,
	getEngagements,
	getEngagement,
	createCategory,
	updateCategory,
	deleteCategory,
	updateEngagementStatus,
} from './engagements'
export type { GetEngagementsParams, CreateCategoryRequest } from './engagements'

export { getUserProfile } from './user'
export type { UserProfile as ApiUserProfile } from './user'
