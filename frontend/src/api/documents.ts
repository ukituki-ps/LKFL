/**
 * API-функции для Documents (документы).
 *
 * Эндпоинты:
 *   GET /api/v1/documents
 */

import { apiRequest } from './client'
import type { DocumentItem } from './mocks'
import { mockDocuments } from './mocks'

const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true'

const MOCK_DELAY = 200 // имитация задержки сети

function mockDelay<T>(data: T): Promise<T> {
  return new Promise(resolve => setTimeout(() => resolve(data), MOCK_DELAY))
}

export async function getDocuments(): Promise<DocumentItem[]> {
  if (USE_MOCKS) return mockDelay(mockDocuments)
  return apiRequest<DocumentItem[]>('/api/v1/documents')
}
