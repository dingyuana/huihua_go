import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface AdvancePayment {
  id: string
  advance_no: string
  supplier_id: string
  amount: string
  allocated_amount: string
  outstanding_amount: string
  paid_date: string
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

export interface CreateAdvancePaymentPayload {
  company_id: string
  supplier_id: string
  amount: string
  paid_date: string
  due_date?: string
  bank_account_id?: string
  reference_no?: string
  remark?: string
}

export function fetchAdvancePayments(params: { status?: string } = {}): Promise<ApiResponse<{ list: AdvancePayment[]; total: number }>> {
  return request.get('/advance-payments', { params })
}

export function fetchAdvancePaymentById(id: string): Promise<ApiResponse<AdvancePayment>> {
  return request.get(`/advance-payments/${id}`)
}

export function fetchOutstandingAdvancePayments(supplierId: string): Promise<ApiResponse<{ list: AdvancePayment[]; total: number }>> {
  return request.get('/advance-payments/outstanding', { params: { supplier_id: supplierId } })
}

export function createAdvancePayment(payload: CreateAdvancePaymentPayload): Promise<ApiResponse<AdvancePayment>> {
  return request.post('/advance-payments', payload)
}

export function confirmAdvancePayment(id: string): Promise<ApiResponse<AdvancePayment>> {
  return request.post(`/advance-payments/${id}/confirm`)
}
