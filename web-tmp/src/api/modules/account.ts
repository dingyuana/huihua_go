import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'
import type { Account } from '@/types/models/account'

/** 获取科目树（所有节点） */
export function fetchAccountTree(): Promise<ApiResponse<Account[]>> {
  return request.get('/accounts/tree')
}

/** 分页查询科目 */
export function fetchAccountList(params: PageQuery & { keyword?: string; account_type?: string; is_group?: boolean }): Promise<ApiResponse<PageResult<Account>>> {
  return request.get('/accounts', { params })
}

/** 获取仅可记账科目 */
export function fetchLedgerOnly(): Promise<ApiResponse<Account[]>> {
  return request.get('/accounts/ledger-only')
}

/** 创建科目 */
export function createAccount(data: Partial<Account> & { parent_id: string }): Promise<ApiResponse<Account>> {
  return request.post('/accounts', data)
}

/** 更新科目 */
export function updateAccount(id: string, data: Partial<Account>): Promise<ApiResponse<Account>> {
  return request.put(`/accounts/${id}`, data)
}

/** 删除科目 */
export function deleteAccount(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/accounts/${id}`)
}

/** 预览编码 */
export function previewCode(parentId: string): Promise<ApiResponse<{ suggested_code: string }>> {
  return request.post('/accounts/auto-code', { parent_id: parentId })
}
