import request from '@/api/request'
import type { ApiResponse } from '@/types/api'
import type { PaymentEntry } from '@/types/models/payment'

/** 查询收付款单列表 */
export function fetchPayments(params?: {
  page?: number
  pageSize?: number
  payment_type?: string
  bank_account_id?: string
  start_date?: string
  end_date?: string
  keyword?: string
}): Promise<ApiResponse<{ list: PaymentEntry[]; total: number }>> {
  return request.get('/payment-entries', { params })
}

/** 获取收付款单详情 */
export function fetchPaymentDetail(id: string): Promise<ApiResponse<PaymentEntry>> {
  return request.get(`/payment-entries/${id}`)
}

/** 创建收付款单（从银行流水） */
export function createPayment(data: {
  bank_transaction_id: string
  payment_type: string
  party_type?: string
  party_id?: string
  posting_date: string
  remark?: string
}): Promise<ApiResponse<{ payment_entry: PaymentEntry }>> {
  return request.post('/payment-entries', data)
}

/** 更新收付款单 */
export function updatePayment(id: string, data: Partial<PaymentEntry>): Promise<ApiResponse<void>> {
  return request.put(`/payment-entries/${id}`, data)
}

/** 删除收付款单 */
export function deletePayment(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/payment-entries/${id}`)
}

/** 从收付款单生成凭证 */
export function generateVoucherFromPayment(id: string): Promise<ApiResponse<any>> {
  return request.post(`/payment-entries/${id}/generate-voucher`)
}
