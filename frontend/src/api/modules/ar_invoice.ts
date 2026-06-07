import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface ArInvoice {
  id: string
  invoice_id: string
  invoice_no: string
  customer_id: string
  customer_name?: string
  amount: string
  paid_amount?: string
  outstanding_amount?: string
  due_date?: string | null
  status: string
  source_type: string
  remark?: string
  created_at: string
  confirmed_at?: string | null
  approved_at?: string | null
}

export function fetchArInvoices(params: {
  status?: string
} = {}): Promise<ApiResponse<{ list: ArInvoice[]; total: number }>> {
  return request.get('/ar-invoices', { params })
}

export function fetchArInvoiceById(id: string): Promise<ApiResponse<ArInvoice>> {
  return request.get(`/ar-invoices/${id}`)
}
