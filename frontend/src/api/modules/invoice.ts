import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'
import type { SalesInvoice } from '@/types/models/invoice'

export interface InvoiceFileImportResult {
  total_rows: number
  imported: number
  failed: number
  failed_rows?: Array<{ row: number; reason: string; date?: string }>
}

/** 上传发票 */
export function uploadInvoice(file: File): Promise<ApiResponse<SalesInvoice>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/invoices/upload', form, {
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
