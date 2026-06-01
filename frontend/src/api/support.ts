/**
 * API-функции для Support (поддержка).
 *
 * Эндпоинты:
 *   GET /api/v1/support/faq
 *   POST /api/v1/support/tickets
 */

import { apiRequest } from './client'
import type { FaqItem, SupportTicketRequest, SupportTicketResponse } from './mocks'
import { mockFaq } from './mocks'

const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true'

const MOCK_DELAY = 200 // имитация задержки сети

function mockDelay<T>(data: T): Promise<T> {
  return new Promise(resolve => setTimeout(() => resolve(data), MOCK_DELAY))
}

export async function getFaq(): Promise<FaqItem[]> {
  if (USE_MOCKS) return mockDelay(mockFaq)
  return apiRequest<FaqItem[]>('/api/v1/support/faq')
}

export async function postSupportTicket(
  req: SupportTicketRequest
): Promise<SupportTicketResponse> {
  if (USE_MOCKS) {
    return mockDelay({ id: 'mock-' + Date.now(), status: 'submitted' })
  }
  return apiRequest<SupportTicketResponse>('/api/v1/support/tickets', {
    method: 'POST',
    body: JSON.stringify(req),
  })
}
