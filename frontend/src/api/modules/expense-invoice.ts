import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

/** 进项发票实体 */
export interface ExpenseInvoice {
  id: string
  tenant_id: string
  company_id: string
  invoice_no: string
  invoice_code?: string
  invoice_date: string
  invoice_kind: 'paper_normal' | 'paper_special' | 'electronic_normal' | 'electronic_special' | string
  category?: 'transport' | 'office' | 'travel' | 'entertain' | 'communication' | 'training' | 'welfare' | 'other' | string
  amount: string
  tax_amount: string
  total_amount: string
  vendor_id?: string
  vendor_name?: string
  tax_id?: string
  description?: string
  source?: 'manual' | 'import' | 'ocr' | string
  verify_status?: 'unverified' | 'verified' | 'invalid' | string
  verified_at?: string
  verify_result?: string
  status?: 'draft' | 'confirmed' | string
  docstatus: number
  voucher_id?: string
  created_by?: string
  created_at: string
  updated_by?: string
  updated_at?: string
}

/** 新增进项发票请求 */
export interface ExpenseInvoiceCreateRequest {
  invoice_no: string
  invoice_code?: string
  invoice_date: string
  invoice_kind: string
  category?: string
  amount: string
  tax_amount: string
  total_amount: string
  vendor_name?: string
  tax_id?: string
  description?: string
}

/** 更新进项发票请求（部分字段） */
export type ExpenseInvoiceUpdateRequest = Partial<ExpenseInvoiceCreateRequest>

/** 列表查询参数 */
export interface ExpenseInvoiceListQuery {
  page?: number
  pageSize?: number
  status?: string
  verify_status?: string
  start_date?: string
  end_date?: string
  keyword?: string
}

/** 导入批次预览行 */
export interface ExpenseInvoiceImportPreviewRow {
  row: number
  invoice_no?: string
  invoice_date?: string
  vendor_name?: string
  amount?: string
  tax_amount?: string
  total_amount?: string
  valid?: boolean
  reason?: string
}

/** 导入批次预览结果 */
export interface ExpenseInvoiceImportPreview {
  batch_id: string
  total: number
  valid: number
  invalid: number
  rows: ExpenseInvoiceImportPreviewRow[]
}

/** 导入确认请求 */
export interface ExpenseInvoiceImportConfirmRequest {
  rows?: number[]
  skip_invalid?: boolean
}

/** 导入确认结果 */
export interface ExpenseInvoiceImportConfirmResult {
  batch_id: string
  imported: number
  failed: number
}

/** OCR 识别结果 */
export interface ExpenseInvoiceOcrResult {
  invoice_no?: string
  invoice_code?: string
  invoice_date?: string
  invoice_kind?: string
  amount?: string
  tax_amount?: string
  total_amount?: string
  vendor_name?: string
  tax_id?: string
  raw?: Record<string, unknown>
}

/** 验真结果 */
export interface ExpenseInvoiceVerifyResult {
  invoice_id: string
  verify_status: 'verified' | 'invalid' | 'unverified' | string
  verify_result?: string
  verified_at?: string
}

/** 查询进项发票列表 */
export function fetchExpenseInvoiceList(
  params?: ExpenseInvoiceListQuery,
): Promise<ApiResponse<{ list: ExpenseInvoice[]; total: number }>> {
  return request.get('/expense-invoices', { params })
}

/** 创建进项发票 */
export function createExpenseInvoice(
  data: ExpenseInvoiceCreateRequest,
): Promise<ApiResponse<{ invoice: ExpenseInvoice }>> {
  return request.post('/expense-invoices', data)
}

/** 获取进项发票详情 */
export function fetchExpenseInvoiceDetail(id: string): Promise<ApiResponse<ExpenseInvoice>> {
  return request.get(`/expense-invoices/${id}`)
}

/** 更新进项发票 */
export function updateExpenseInvoice(
  id: string,
  data: ExpenseInvoiceUpdateRequest,
): Promise<ApiResponse<void>> {
  return request.put(`/expense-invoices/${id}`, data)
}

/** 删除进项发票 */
export function deleteExpenseInvoice(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/expense-invoices/${id}`)
}

/** 确认进项发票 */
export function confirmExpenseInvoice(id: string): Promise<ApiResponse<void>> {
  return request.post(`/expense-invoices/${id}/confirm`)
}

/** 上传导入文件（multipart） */
export function uploadExpenseInvoiceImport(
  file: File,
): Promise<ApiResponse<{ batch_id: string; total: number }>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/expense-invoices/import/upload', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 预览导入批次 */
export function previewExpenseInvoiceImport(
  batch_id: string,
): Promise<ApiResponse<ExpenseInvoiceImportPreview>> {
  return request.get(`/expense-invoices/import/${batch_id}/preview`)
}

/** 确认导入批次 */
export function confirmExpenseInvoiceImport(
  batch_id: string,
  data: ExpenseInvoiceImportConfirmRequest = {},
): Promise<ApiResponse<ExpenseInvoiceImportConfirmResult>> {
  return request.post(`/expense-invoices/import/${batch_id}/confirm`, data)
}

/** OCR 识别发票（multipart, Mock） */
export function ocrExpenseInvoice(file: File): Promise<ApiResponse<ExpenseInvoiceOcrResult>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/expense-invoices/ocr', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 验真单张发票（Mock） */
export function verifyExpenseInvoice(id: string): Promise<ApiResponse<ExpenseInvoiceVerifyResult>> {
  return request.post(`/expense-invoices/${id}/verify`)
}

/** 批量验真发票（Mock） */
export function batchVerifyExpenseInvoice(
  ids: string[],
): Promise<ApiResponse<{ results: ExpenseInvoiceVerifyResult[] }>> {
  return request.post('/expense-invoices/verify/batch', { ids })
}
