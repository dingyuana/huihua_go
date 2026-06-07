import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'
import type { SalesInvoice } from '@/types/models/invoice'

export interface InvoiceFileImportResult {
  total_rows: number
  imported: number
  failed: number
  failed_rows?: Array<{ row: number; reason: string; date?: string }>
}

export interface ExcelPreviewResult {
  columns: string[]
  sample: string[][]
  total_rows: number
  header_row: number
}

/** 上传发票 */
export function uploadInvoice(file: File): Promise<ApiResponse<SalesInvoice>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/invoices/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 预览Excel文件 */
export function previewInvoiceExcel(file: File): Promise<ApiResponse<ExcelPreviewResult>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/invoices/preview-excel', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** Excel/CSV 批量导入发票 */
export function importInvoicesFile(file: File): Promise<ApiResponse<InvoiceFileImportResult>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/invoices/import-excel', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 查询发票列表 */
export function fetchInvoices(params: PageQuery & {
  status?: string
  invoice_type?: string
  start_date?: string
  end_date?: string
  keyword?: string
}): Promise<ApiResponse<PageResult<SalesInvoice>>> {
  return request.get('/invoices', { params })
}

/** 获取发票详情 */
export function fetchInvoiceDetail(id: string): Promise<ApiResponse<SalesInvoice>> {
  return request.get(`/invoices/${id}`)
}

/** 获取即将过期发票 */
export function fetchExpiringInvoices(): Promise<ApiResponse<{ list: SalesInvoice[]; total_expiring: number }>> {
  return request.get('/invoices/expiring')
}

/** 删除发票 */
export function deleteInvoice(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/invoices/${id}`)
}

/** 根据发票生成凭证 */
export function generateVoucherFromInvoice(id: string): Promise<ApiResponse<any>> {
  return request.post(`/invoices/${id}/generate-voucher`)
}

export interface InvoiceBatchPreviewResult {
  batch_id: string
  total_rows: number
  valid_rows: number
  error_rows: number
  duplicate_rows: number
  details: Array<{
    row_index: number
    invoice_no: string
    invoice_type: string
    customer_name: string
    posting_date: string
    total_amount: number
    net_amount: number
    tax_amount: number
    status: string
    validation_err?: string
    is_duplicate?: boolean
    duplicate_info?: string
  }>
}

export interface InvoiceBatchConfirmRequest {
  batch_id: string
  selected_ids: string[]
  corrected_data?: Array<Record<string, any>>
}

export interface InvoiceBatchConfirmResult {
  imported: number
  skipped: number
  errors: number
  failed_rows?: Array<{ row: number; reason: string }>
}

export interface InvoiceConfirmRequest {
  invoice_id: string
}

/** 销售发票批量导入预览 */
export function batchImportPreview(file: File): Promise<ApiResponse<InvoiceBatchPreviewResult>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/invoices/sales/import/preview', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 销售发票批量导入确认 */
export function batchImportConfirm(data: InvoiceBatchConfirmRequest): Promise<ApiResponse<InvoiceBatchConfirmResult>> {
  return request.post('/invoices/sales/import/confirm', data)
}

/** 确认销售发票（生成应收账款） */
export function confirmSalesInvoice(invoiceID: string): Promise<ApiResponse<void>> {
  return request.post(`/invoices/sales/${invoiceID}/confirm`)
}

/** 整单红冲 */
export function redSalesInvoice(id: string, reason?: string): Promise<ApiResponse<any>> {
  return request.post(`/invoices/sales/${id}/red`, { reason })
}

/** 部分红冲 */
export function partRedSalesInvoice(id: string, redAmount: string, reason?: string): Promise<ApiResponse<any>> {
  return request.post(`/invoices/sales/${id}/red/part`, { red_amount: redAmount, reason })
}

/** 作废销售发票 */
export function voidSalesInvoice(id: string, reason?: string): Promise<ApiResponse<any>> {
  return request.post(`/invoices/sales/${id}/void`, { reason })
}
