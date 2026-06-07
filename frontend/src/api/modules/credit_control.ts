import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface CreditStatus {
  party_id: string
  party_name: string
  credit_limit: string
  credit_used: string
  available: string
  utilization_pct: string
  overdraft_allowed: boolean
  over_limit: boolean
}

export function fetchCreditStatus(partyId: string): Promise<ApiResponse<CreditStatus>> {
  return request.get('/credit-control', { params: { party_id: partyId } })
}

export function fetchOverLimitParties(): Promise<ApiResponse<{ list: CreditStatus[]; total: number }>> {
  return request.get('/credit-control/over-limit')
}
