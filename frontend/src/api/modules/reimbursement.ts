import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface Reimbursement {
  id: string
  tenant_id: string
  /** 后端 BusReimbursement 无此字段，保留为 optional 以兼容旧调用 */
  applicant_id?: string
  employee_id?: string
  employee_name: string
  /** 后端字段名为 department（*string），前端保留 dept_name 兼容旧引用 */
  dept_name?: string
  /** 后端字段名为 department */
  department?: string
  amount: string
  description: string
  /** 后端字段名为 docstatus（小写） */
  docstatus: number
  /** 旧字段名 alias，兼容旧代码读取 */
  doc_status?: number
  /** 旧字段名 alias，兼容旧代码读取（后端对应 description） */
  remark?: string
  voucher_id?: string
  created_by: string
  created_at: string
  /** 后端 BusReimbursement 额外字段 */
  reimbursement_no?: string
  expense_type?: string
  posting_date?: string
  voucher_no?: string
  company_id?: string
  reject_reason?: string
  sub_expense_type?: string
  bank_account?: string
  updated_at?: string
}

/** 查询报销单列表 */
export function fetchReimbursementList(params?: {
  page?: number
  pageSize?: number
  employee_name?: string
  department?: string
  dept_name?: string
  docstatus?: number
  status?: string
  keyword?: string
}): Promise<ApiResponse<{ list: Reimbursement[]; total: number }>> {
  return request.get('/reimbursements', { params })
}

/** 获取报销单详情 */
export function fetchReimbursementDetail(id: string): Promise<ApiResponse<Reimbursement>> {
  return request.get(`/reimbursements/${id}`)
}

/** 创建报销单 */
export function createReimbursement(data: Partial<Reimbursement>): Promise<ApiResponse<{ reimbursement: Reimbursement }>> {
  return request.post('/reimbursements', data)
}

/** 更新报销单 */
export function updateReimbursement(id: string, data: Partial<Reimbursement>): Promise<ApiResponse<void>> {
  return request.put(`/reimbursements/${id}`, data)
}

/** 删除报销单 */
export function deleteReimbursement(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/reimbursements/${id}`)
}

/** 提交报销单 */
export function submitReimbursement(id: string): Promise<ApiResponse<void>> {
  return request.post(`/reimbursements/${id}/submit`)
}

/** 审核报销单（自动制证） */
export function approveReimbursement(id: string): Promise<ApiResponse<{ voucher_id: string }>> {
  return request.post(`/reimbursements/${id}/approve`)
}

/** 驳回报销单 */
export function rejectReimbursement(id: string, reason?: string): Promise<ApiResponse<void>> {
  return request.post(`/reimbursements/${id}/reject`, { reason })
}

/* ===================== 附件管理 ===================== */

/** 报销单附件 */
export interface ReimbursementAttachment {
  id: string
  reimbursement_id: string
  file_name: string
  file_url: string
  file_size: number
  mime_type: string
  uploaded_by: string
  uploaded_at: string
}

/** 上传报销单附件（multipart/form-data，字段名 file） */
export function uploadReimbursementAttachment(
  id: string,
  file: File,
): Promise<ApiResponse<ReimbursementAttachment>> {
  const form = new FormData()
  form.append('file', file)
  return request.post(`/reimbursements/${id}/attachments`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

/** 获取报销单附件列表 */
export function listReimbursementAttachments(
  id: string,
): Promise<ApiResponse<ReimbursementAttachment[]>> {
  return request.get(`/reimbursements/${id}/attachments`)
}

/** 删除报销单附件 */
export function deleteReimbursementAttachment(
  id: string,
  fileId: string,
): Promise<ApiResponse<void>> {
  return request.delete(`/reimbursements/${id}/attachments/${fileId}`)
}

/** 下载报销单附件（返回 Blob） */
export function downloadReimbursementAttachment(
  id: string,
  fileId: string,
): Promise<Blob> {
  return request.get(`/reimbursements/${id}/attachments/${fileId}/download`, {
    responseType: 'blob',
  }) as unknown as Promise<Blob>
}

/* ===================== 进项发票关联 ===================== */

/** 已关联进项发票 */
export interface LinkedInvoice {
  id: string
  invoice_no: string
  invoice_code: string
  invoice_date: string
  vendor_name: string
  total_amount: string
  tax_amount: string
  category: string
  verify_status: string
}

/** 查询可关联的进项发票列表（分页） */
export function listAvailableInvoices(
  id: string,
  params?: { page?: number; pageSize?: number; keyword?: string },
): Promise<ApiResponse<{ list: LinkedInvoice[]; total: number }>> {
  return request.get(`/reimbursements/${id}/invoices`, { params })
}

/** 关联一张进项发票到报销单 */
export function linkInvoice(id: string, invoiceId: string): Promise<ApiResponse<void>> {
  return request.post(`/reimbursements/${id}/invoices/${invoiceId}`)
}

/** 取消关联进项发票 */
export function unlinkInvoice(id: string, invoiceId: string): Promise<ApiResponse<void>> {
  return request.delete(`/reimbursements/${id}/invoices/${invoiceId}`)
}

/** 查询已关联的进项发票列表 */
export function listLinkedInvoices(id: string): Promise<ApiResponse<LinkedInvoice[]>> {
  return request.get(`/reimbursements/${id}/invoices/linked`)
}