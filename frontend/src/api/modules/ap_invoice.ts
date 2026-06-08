import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface ApInvoice {
  id: string
  invoice_id: string
  invoice_no: string
  supplier_id: string
  supplier_name?: string
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

export function fetchApInvoices(params: {
  status?: string
} = {}): Promise<ApiResponse<{ list: ApInvoice[]; total: number }>> {
  return request.get('/ap-invoices', { params })
}

export function fetchApInvoiceById(id: string): Promise<ApiResponse<ApInvoice>> {
  return request.get(`/ap-invoices/${id}`)
}

export function createApInvoice(data: {
  supplier_id: string
  amount: string
  due_date?: string
  remark?: string
  source_type?: string
}): Promise<ApiResponse<ApInvoice>> {
  return request.post('/ap-invoices', data)
}
