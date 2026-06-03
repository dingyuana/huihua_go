import request from '@/api/request'
import type { ApiResponse, PageResult } from '@/types/api'
import type { BankTransaction } from '@/types/models/bank'

export function getReviewList(params: {
  status?: string
  page?: number
  page_size?: number
}): Promise<ApiResponse<PageResult<BankTransaction>>> {
  return request.get('/api/v1/bank-transactions/review-list', { params })
}

export function getReviewStats(): Promise<ApiResponse<{
  monthly_txns: number
  pending_count: number
  ai_processed_count: number
  manual_pending_count: number
}>> {
  return request.get('/api/v1/bank-transactions/review-stats')
}

export function previewDraft(id: string): Promise<ApiResponse<{
  bank_txn: BankTransaction
  ai_result: {
    business_scene: string
    suggested_action: string
    confidence: number
  }
  draft_voucher?: {
    id: string
    lines: Array<{ account_name: string; debit: string; credit: string }>
    summary: string
  }
  or_draft_payment?: Record<string, any>
}>> {
  return request.post(`/api/v1/bank-transactions/preview-draft/${id}`)
}

export function submitReview(data: {
  txn_ids: string[]
  human_modified_drafts?: Record<string, any>
}): Promise<ApiResponse<{
  approved_count: number
  results: Array<{ txn_id: string; outcome: string; voucher_id?: string; payment_id?: string }>
}>> {
  return request.post('/api/v1/bank-transactions/submit-review', data)
}

export function rejectManual(data: { txn_ids: string[] }): Promise<ApiResponse<{ rejected_count: number }>> {
  return request.post('/api/v1/bank-transactions/reject-manual', data)
}
