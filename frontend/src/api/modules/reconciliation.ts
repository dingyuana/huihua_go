import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

/** 银行对账结果 */
export interface ReconciliationStats {
  total: number
  autoMatched: number
  needConfirm: number
  unmatched: number
  autoMatchRate: number
}

/** 匹配条目 */
export interface MatchItem {
  id: string
  score: number
  bank_txn: string
  gl_entry: string
  needConfirm: boolean
}

/** 余额调节数据 */
export interface BalanceData {
  bankBalance: string
  bookBalance: string
  diff: string
  adjustedBalance: string
  isBalanced: boolean
  bankReceiptNotInGL: Array<{ date: string; desc: string; amount: string }>
  bankPaymentNotInGL: Array<{ date: string; desc: string; amount: string }>
  glReceiptNotInBank: Array<{ date: string; desc: string; amount: string }>
  glPaymentNotInBank: Array<{ date: string; desc: string; amount: string }>
}

/** 待确认候选对（60-85 分） */
export interface PendingConfirmItem {
  bank_txn_id: string
  gl_entry_id: string
  bankTxnDesc: string
  bankTxnAmt: string
  bankTxnDate: string
  glEntryDesc: string
  score: {
    total_score: number
    is_auto_matched: boolean
  }
}

/** 对账状态 */
export interface ReconciliationStatus {
  locked: boolean
  locked_by?: string
  locked_at?: string
  period_no: number
  bank_account_id: string
}

/** 执行银行对账 */
export function runBankReconciliation(bankAccountId: string): Promise<ApiResponse<ReconciliationStats & { matches: MatchItem[] }>> {
  return request.post('/bank-reconciliation/reconcile', { bank_account_id: bankAccountId })
}

/** 获取余额调节表 */
export function fetchBalanceSheet(bankAccountId: string): Promise<ApiResponse<BalanceData>> {
  return request.get('/bank-reconciliation/balance-sheet', { params: { bank_account_id: bankAccountId } })
}

/** 确认对账匹配 */
export function confirmMatch(id: string): Promise<ApiResponse<void>> {
  return request.post(`/bank-reconciliation/matches/${id}/confirm`)
}

/** 锁定对账结果 */
export function lockReconciliation(bankAccountId: string, periodNo?: number): Promise<ApiResponse<void>> {
  return request.post('/bank-reconciliation/lock', { bank_account_id: bankAccountId, period_no: periodNo })
}

/** 解锁对账结果 */
export function unlockReconciliation(bankAccountId: string, periodNo?: number): Promise<ApiResponse<void>> {
  return request.post('/bank-reconciliation/unlock', { bank_account_id: bankAccountId, period_no: periodNo })
}

/** 获取待确认候选对列表（60-85 分） */
export function fetchPendingConfirm(bankAccountId: string, periodNo?: number): Promise<ApiResponse<PendingConfirmItem[]>> {
  return request.get('/bank-reconciliation/pending-confirm', { params: { bank_account_id: bankAccountId, period_no: periodNo } })
}

/** 确认勾兑候选对 */
export function confirmMatchCandidate(bankTxnId: string, glEntryId: string): Promise<ApiResponse<void>> {
  return request.post('/bank-reconciliation/confirm-match', { bank_txn_id: bankTxnId, gl_entry_id: glEntryId })
}

/** 拒绝候选对 */
export function rejectMatchCandidate(bankTxnId: string, glEntryId: string): Promise<ApiResponse<void>> {
  return request.post('/bank-reconciliation/reject-match', { bank_txn_id: bankTxnId, gl_entry_id: glEntryId })
}

/** 获取对账状态 */
export function fetchReconciliationStatus(bankAccountId: string, periodNo?: number): Promise<ApiResponse<ReconciliationStatus>> {
  return request.get('/bank-reconciliation/status', { params: { bank_account_id: bankAccountId, period_no: periodNo } })
}

/** 获取对账差异条目 */
export function fetchReconciliationItems(bankAccountId: string, periodNo?: number, itemType?: string): Promise<ApiResponse<BalanceData>> {
  return request.get('/bank-reconciliation/items', { params: { bank_account_id: bankAccountId, period_no: periodNo, item_type: itemType } })
}