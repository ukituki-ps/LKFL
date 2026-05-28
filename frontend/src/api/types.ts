/**
 * Типы API — ре-экспорт из сгенерированных типов с удобными alias-ами.
 * Источник: openapi/spec.yaml → types.generated.ts
 *
 * Не редактировать модели данных напрямую — обновлять openapi/spec.yaml
 * и запускать `npm run generate-types`.
 */

import type { components } from './types.generated'

// Convenience aliases для сгенерированных schema-типов
export type EngagementCategoryResponse = components['schemas']['EngagementCategoryResponse']
export type EngagementOfferResponse = components['schemas']['EngagementOfferResponse']
export type EngagementTypeResponse = components['schemas']['EngagementTypeResponse']
export type PaginationResponse = components['schemas']['PaginationResponse']
export type ListResponse = components['schemas']['ListResponse']
export type UserProfile = components['schemas']['UserProfile']
export type CreateCategoryRequest = components['schemas']['CreateCategoryRequest']
export type CreateTypeRequest = components['schemas']['CreateTypeRequest']
export type UpdateStatusRequest = components['schemas']['UpdateStatusRequest']
export type CreateOfferRequest = components['schemas']['CreateOfferRequest']

// Re-export full generated types for advanced usage (paths, operations, etc.)
export type { paths, components, operations } from './types.generated'
