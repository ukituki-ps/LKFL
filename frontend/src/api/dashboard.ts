/**
 * API-функции для Dashboard.
 *
 * Эндпоинты:
 *   GET /api/v1/dashboard/stats
 *   GET /api/v1/dashboard/benefits
 *   GET /api/v1/dashboard/events
 */

import { apiRequest } from './client'
import type { DashboardStats, BenefitItem, EventItem } from './mocks'
import { mockDashboardStats, mockDashboardBenefits, mockDashboardEvents } from './mocks'

const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true'

const MOCK_DELAY = 200 // имитация задержки сети

function mockDelay<T>(data: T): Promise<T> {
  return new Promise(resolve => setTimeout(() => resolve(data), MOCK_DELAY))
}

export async function getDashboardStats(): Promise<DashboardStats> {
  if (USE_MOCKS) return mockDelay(mockDashboardStats)
  return apiRequest<DashboardStats>('/api/v1/dashboard/stats')
}

export async function getActiveBenefits(): Promise<BenefitItem[]> {
  if (USE_MOCKS) return mockDelay(mockDashboardBenefits)
  return apiRequest<BenefitItem[]>('/api/v1/dashboard/benefits')
}

export async function getEvents(): Promise<EventItem[]> {
  if (USE_MOCKS) return mockDelay(mockDashboardEvents)
  return apiRequest<EventItem[]>('/api/v1/dashboard/events')
}
