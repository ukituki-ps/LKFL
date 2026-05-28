import { apiRequest } from './client'
import type { UserProfile } from './types'

/**
 * Получить профиль текущего пользователя.
 *
 * Endpoint: GET /api/v1/users/me
 * Требует: JWT + tenant middleware.
 *
 * UserProfile ре-экспортирован из types.generated.ts (OpenAPI spec).
 */
export { UserProfile }
export async function getUserProfile(): Promise<UserProfile> {
	return apiRequest<UserProfile>('/api/v1/users/me')
}
