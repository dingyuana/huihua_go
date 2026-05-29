import request from '@/api/request'
import type { ApiResponse } from '@/types/api'
import type { PaymentEntry } from '@/types/models/payment'

/** 查询收付款列表 */
export function fetchPayments(params: {
  page?: number
  pageSize?: number
  payment_type?: string
  party_id?: string
  start_date?: string
  end_date?: string
}): Promise<ApiResponse<{ list: PaymentEntry[]; total: number }>> {
  return request.get('/payments', { params })
}
