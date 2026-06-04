import type { InvoiceStatus, InvoiceType } from '@/types/enums'

export interface SalesInvoice {
  id: string
  invoice_no: string
  invoice_type: InvoiceType
  customer_id: string
  customer_name: string
  tax_id: string
  posting_date: string
  due_date: string
  total_amount: string
  tax_amount: string
  net_amount: string
  outstanding_amount: string
  status: InvoiceStatus
  docstatus: number
  is_return?: boolean
  source_red_invoice_no?: string
  remark?: string
}
