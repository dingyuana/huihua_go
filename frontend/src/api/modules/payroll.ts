import request from '@/api/request'
import type { ApiResponse } from '@/types/api'
import type { Payroll } from '@/types/models/payroll'

/** 查询工资单列表 */
export function fetchPayrollList(params?: {
  page?: number
  pageSize?: number
  period_no?: number | string
  status?: string
  keyword?: string
}): Promise<ApiResponse<{ list: Payroll[]; total: number }>> {
  return request.get('/payroll', { params })
}

/** 获取工资单详情 */
export function fetchPayrollDetail(id: string): Promise<ApiResponse<Payroll>> {
  return request.get(`/payroll/${id}`)
}

/** 创建工资单 */
export function createPayroll(data: Partial<Payroll>): Promise<ApiResponse<{ payroll: Payroll }>> {
  return request.post('/payroll', data)
}

/** 更新工资单 */
export function updatePayroll(id: string, data: Partial<Payroll>): Promise<ApiResponse<void>> {
  return request.put(`/payroll/${id}`, data)
}

/** 删除工资单 */
export function deletePayroll(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/payroll/${id}`)
}

/** 提交工资单 */
export function submitPayroll(id: string): Promise<ApiResponse<void>> {
  return request.post(`/payroll/${id}/submit`)
}

/** 审核工资单 */
export function approvePayroll(id: string): Promise<ApiResponse<{ voucher_id: string }>> {
  return request.post(`/payroll/${id}/approve`)
}

/** 独立制证 */
export function generateVoucherFromPayroll(id: string): Promise<ApiResponse<{ voucher_id: string }>> {
  return request.post(`/payroll/${id}/generate-voucher`)
}