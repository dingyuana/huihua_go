import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'
import type { JournalEntry } from '@/types/models/journal'

/** 创建凭证草稿 */
export function createVoucher(data: {
  voucher_type: string
  posting_date: string
  remark?: string
  lines: Array<{
    account_id: string
    debit: string
    credit: string
    party_type?: string
    party_id?: string
    user_remark?: string
  }>
}): Promise<ApiResponse<JournalEntry>> {
  return request.post('/vouchers', data)
}

/** 提交审核 */
export function submitVoucher(id: string): Promise<ApiResponse<JournalEntry>> {
  return request.post(`/vouchers/${id}/submit`)
}

/** 作废凭证 */
export function cancelVoucher(id: string, reason: string): Promise<ApiResponse<void>> {
  return request.post(`/vouchers/${id}/cancel`, { reason })
}

/** 红字冲销 */
export function reverseVoucher(id: string, data: { reverse_date: string; reason: string }): Promise<ApiResponse<JournalEntry>> {
  return request.post(`/vouchers/${id}/reverse`, data)
}

/** 查询凭证列表 */
export function fetchVouchers(params: PageQuery & {
  voucher_type?: string
  docstatus?: number
  start_date?: string
  end_date?: string
  keyword?: string
}): Promise<ApiResponse<PageResult<JournalEntry>>> {
  return request.get('/vouchers', { params })
}

/** 获取凭证详情 */
export function fetchVoucherDetail(id: string): Promise<ApiResponse<JournalEntry>> {
  return request.get(`/vouchers/${id}`)
}

/** 审核单个凭证 */
export function approveVoucher(id: string): Promise<ApiResponse<JournalEntry>> {
  return request.post(`/vouchers/${id}/approve`)
}

/** 驳回单个凭证 */
export function rejectVoucher(id: string, reason: string): Promise<ApiResponse<void>> {
  return request.post(`/vouchers/${id}/reject`, { reason })
}

/** 删除草稿凭证（同时回滚关联单据状态） */
export function deleteVoucher(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/vouchers/${id}`)
}
