import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'

export interface OpeningBalance {
  id: string
  tenant_id: string
  account_id: string
  account_code: string
  account_name: string
  opening_debit: string
  opening_credit: string
  period_no: number
  year: number
}

export interface OpeningBalanceSummary {
  total_debit: string
  total_credit: string
  account_count: number
}

export interface OpeningBalanceImportResult {
  total: number
  success: number
  failed: number
  errors: Array<{ row: number; message: string }>
}

export interface OpeningBalanceValidateResult {
  valid: boolean
  errors: Array<{ account_code: string; message: string }>
}

/** 期初余额列表 */
export function fetchOpeningBalances(
  params: PageQuery & { year?: number; period_no?: number; keyword?: string }
): Promise<ApiResponse<PageResult<OpeningBalance>>> {
  return request.get('/opening-balances', { params })
}

/** 期初余额汇总 */
export function fetchOpeningBalanceSummary(
  params: { year: number; period_no: number }
): Promise<ApiResponse<OpeningBalanceSummary>> {
  return request.get('/opening-balances/summary', { params })
}

/** 导入期初余额 */
export function importOpeningBalances(
  data: Partial<OpeningBalance>[]
): Promise<ApiResponse<OpeningBalanceImportResult>> {
  return request.post('/opening-balances/import', { balances: data })
}

/** 校验期初余额 */
export function validateOpeningBalances(
  data: Partial<OpeningBalance>[]
): Promise<ApiResponse<OpeningBalanceValidateResult>> {
  return request.post('/opening-balances/validate', { balances: data })
}

/** 更新期初余额 */
export function updateOpeningBalance(
  id: string,
  data: { opening_debit?: string; opening_credit?: string }
): Promise<ApiResponse<OpeningBalance>> {
  return request.put(`/opening-balances/${id}`, data)
}