import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

// ─── 断号检测 ───

export interface VoucherGap {
  expected_no: number
  is_filled: boolean
  gap_type: 'missing' | 'voided'
  fill_voucher_id?: string
  message: string
}

export interface VoucherGapResult {
  voucher_gaps: VoucherGap[]
  missing_count: number
  voided_count: number
  total_gaps: number
  has_missing: boolean
}

export function fetchVoucherGaps(year: number, month: number): Promise<ApiResponse<VoucherGapResult>> {
  return request.get('/periods/voucher-gaps', { params: { year, month } })
}

// ─── 结账前检查 ───

export interface RiskWarning {
  type: string
  severity: 'critical' | 'warning' | 'info'
  subject_code: string
  subject_name: string
  balance: number
  message: string
}

export interface KeyIndicator {
  name: string
  current_value: number | null
  last_value: number | null
  unit: string
  alert: boolean
  message: string
}

export interface PendingAccrual {
  type: string
  item: string
  missing: boolean
  details?: string
}

export interface PreCloseCheckResult {
  period_status: string
  unposted_vouchers: number
  report_balance_ok: boolean
  profit_loss_done: boolean
  risk_warnings: RiskWarning[]
  key_indicators: KeyIndicator[]
  pending_accruals: PendingAccrual[]
}

export function fetchPreCloseCheck(year: number, month: number): Promise<ApiResponse<PreCloseCheckResult>> {
  return request.get('/periods/pre-close-check', { params: { year, month } })
}

// ─── Close check summary (7 base checks) ───

export interface BaseCheckItem {
  id: string
  name: string
  status: 'passed' | 'blocked' | 'warning' | 'pending'
  message: string
  action?: { label: string; route?: string }
}

export interface CloseCheckSummaryResult {
  base_checks: BaseCheckItem[]
  risk_warnings: RiskWarning[]
  key_indicators: KeyIndicator[]
  pending_accruals: PendingAccrual[]
  profit_loss_done: boolean
  period_status: string
}

export function fetchCloseCheckSummary(year: number, month: number): Promise<ApiResponse<CloseCheckSummaryResult>> {
  return request.get('/periods/close-check-summary', { params: { year, month } })
}

// ─── 结账操作 ───

export interface ClosePeriodRequest {
  period_no: number
  user_id: string
  user_name: string
  generate_closing_entries: boolean
}

export function closePeriod(data: ClosePeriodRequest): Promise<ApiResponse<{ message: string }>> {
  return request.post(`/periods/${data.period_no}/close`, data)
}

export function unclosePeriod(periodNo: number): Promise<ApiResponse<{ message: string }>> {
  return request.post(`/periods/${periodNo}/unclose`)
}
