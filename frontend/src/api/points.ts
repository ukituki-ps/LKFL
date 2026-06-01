/**
 * API-функции для Points (баллы).
 *
 * Эндпоинты:
 *   GET /api/v1/points/balance
 *   GET /api/v1/points/transactions
 */

import { apiRequest } from './client'
import type { PointsBalance, Transaction } from './mocks'
import { mockPointsBalance, mockTransactions } from './mocks'

const USE_MOCKS = import.meta.env.VITE_USE_MOCKS === 'true'

const MOCK_DELAY = 200 // имитация задержки сети

function mockDelay<T>(data: T): Promise<T> {
  return new Promise(resolve => setTimeout(() => resolve(data), MOCK_DELAY))
}

export async function getPointsBalance(): Promise<PointsBalance> {
  if (USE_MOCKS) return mockDelay(mockPointsBalance)
  return apiRequest<PointsBalance>('/api/v1/points/balance')
}

export async function getTransactions(): Promise<Transaction[]> {
  if (USE_MOCKS) return mockDelay(mockTransactions)
  return apiRequest<Transaction[]>('/api/v1/points/transactions')
}
