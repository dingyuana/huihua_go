import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface AgingBucket {
  party_id: string
  party_name: string
  current: string
  b_0_30: string
  b_30_60: string
  b_60_90: string
  b_90_plus: string
  total: string
  overdue: string
  invoice_count: number
}

export interface AgingSummary {
  as_of_date: string
  total_current: string
  total_0_30: string
  total_30_60: string
  total_60_90: string
  total_90_plus: string
  grand_total: string
  overdue_count: number
}

export function fetchARAging(asOf?: string): Promise<ApiResponse<{ buckets: AgingBucket[]; summary: AgingSummary }>> {
  return request.get('/aging-analysis/ar', { params: asOf ? { as_of: asOf } : {} })
}

export function fetchAPAging(asOf?: string): Promise<ApiResponse<{ buckets: AgingBucket[]; summary: AgingSummary }>> {
  return request.get('/aging-analysis/ap', { params: asOf ? { as_of: asOf } : {} })
}
