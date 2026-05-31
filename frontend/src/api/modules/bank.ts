import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'
import type { BankTransaction, ImportResult, BankAccount } from '@/types/models/bank'

/** 获取银行账户列表 */
export function fetchBankAccounts(): Promise<ApiResponse<BankAccount[]>> {
  return request.get('/bank-accounts')
}

/** 创建银行账户 */
export function createBankAccount(data: Partial<BankAccount>): Promise<ApiResponse<BankAccount>> {
  return request.post('/bank-accounts', data)
}

/** 预览Excel文件列名和样本数据 */
export function previewExcelFile(file: File): Promise<ApiResponse<{
  columns: string[]
  sample: string[][]
  total_rows: number
}>> {
  const form = new FormData()
  form.append('file', file)
  return request.post('/bank-transactions/preview', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 上传并解析银行对账单 */
export function importBankFile(file: File, bankAccountId: string, format?: string): Promise<ApiResponse<ImportResult>> {
  const form = new FormData()
  form.append('file', file)
  form.append('bank_account_id', bankAccountId)
  if (format) form.append('format', format)
  return request.post('/bank-transactions/import', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 执行单笔智能分类 */
export function classifyTransaction(id: string): Promise<ApiResponse<{ classification: string; account_id: string }>> {
  return request.post(`/bank-transactions/${id}/classify`)
}

/** 查询银行流水列表 */
export function fetchBankTransactions(params: PageQuery & {
  batch_id?: string
  bank_account_id?: string
  classification?: string
  direction?: string
  matched?: boolean
  start_date?: string
  end_date?: string
}): Promise<ApiResponse<PageResult<BankTransaction>>> {
  return request.get('/bank-transactions', { params })
}

/** 标记单条流水为已匹配 */
export function markMatched(id: string): Promise<ApiResponse<void>> {
  return request.post(`/bank-transactions/${id}/mark-matched`)
}

/** 修正分类 */
export function updateClassification(id: string, data: { classification: string; counterparty_id?: string; remark?: string }): Promise<ApiResponse<void>> {
  return request.post(`/bank-transactions/${id}/classify`, data)
}
