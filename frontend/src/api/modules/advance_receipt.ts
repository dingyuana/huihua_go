import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface AdvanceReceipt {
  id: string
  advance_no: string
  customer_id: string
  amount: string
  allocated_amount: string
  outstanding_amount: string
  received_date: string
  due_date?: string
  status: string
  source_type: string
  bank_account_id?: string
  reference_no?: string
  remark?: string
  voucher_no?: string
  created_at: string
  confirmed_at?: string
}

export interface CreateAdvanceReceiptPayload {
  company_id: string
  customer_id: string
  amount: string
  received_date: string
  due_date?: string
  bank_account_id?: string
  reference_no?: string
  remark?: string
}

export function fetchAdvanceReceipts(params: { status?: string } = {}): Promise<ApiResponse<{ list: AdvanceReceipt[]; total: number }>> {
  return request.get('/advance-receipts', { params })
}

export function fetchAdvanceReceiptById(id: string): Promise<ApiResponse<AdvanceReceipt>> {
  return request.get(`/advance-receipts/${id}`)
}

export function fetchOutstandingAdvanceReceipts(customerId: string): Promise<ApiResponse<{ list: AdvanceReceipt[]; total: number }>> {
  return request.get('/advance-receipts/outstanding', { params: { customer_id: customerId } })
}

export function createAdvanceReceipt(payload: CreateAdvanceReceiptPayload): Promise<ApiResponse<AdvanceReceipt>> {
  return request.post('/advance-receipts', payload)
}

export function confirmAdvanceReceipt(id: string): Promise<ApiResponse<AdvanceReceipt>> {
  return request.post(`/advance-receipts/${id}/confirm`)
}
