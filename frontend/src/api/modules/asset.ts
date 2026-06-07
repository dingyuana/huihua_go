import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

/** 资产摘要项 */
export interface AssetSummary {
  id: string
  asset_name: string
  asset_code: string
  category: string
  purchase_date: string
  useful_life_months: number
  original_value: string
  residual_value: string
  accumulated_depreciation: string
  depreciation_status: string
}

/** 资产完整详情 */
export interface Asset extends AssetSummary {
  depreciation_method: string
  depreciation_start_date: string
  net_value: string
  current_period: string
}

/** 折旧计划行 */
export interface DepreciationScheduleRow {
  id: string
  period: string
  depreciation_amount: string
  accumulated_amount: string
  net_value: string
}

/** 折旧计划 */
export interface DepreciationSchedule {
  id: string
  asset_id: string
  schedule_rows: DepreciationScheduleRow[]
}

/** 折旧执行记录 */
export interface DepreciationRun {
  id: string
  asset_id: string
  asset_name: string
  period: string
  depreciation_amount: string
  voucher_id?: string
  voucher_no?: string
  doc_status: number
  run_date: string
}

/** 资产摘要列表 */
export function fetchAssetSummary(params?: {
  page?: number
  pageSize?: number
  category?: string
  depreciation_status?: string
  keyword?: string
}): Promise<ApiResponse<{ list: AssetSummary[]; total: number }>> {
  return request.get('/assets/summary', { params })
}

/** 资产详情 */
export function fetchAssetDetail(id: string): Promise<ApiResponse<Asset>> {
  return request.get(`/assets/${id}`)
}

/** 创建折旧计划 */
export function createDepreciationSchedule(data: {
  asset_id: string
}): Promise<ApiResponse<DepreciationSchedule>> {
  return request.post('/depreciation/schedule', data)
}

/** 折旧计划详情 */
export function fetchDepreciationSchedule(id: string): Promise<ApiResponse<DepreciationSchedule>> {
  return request.get(`/depreciation/schedule/${id}`)
}

/** 执行折旧（生成折旧记录） */
export function runDepreciation(data: {
  asset_id?: string
  period: string
}): Promise<ApiResponse<{ runs: DepreciationRun[] }>> {
  return request.post('/depreciation/run', data)
}

/** 折旧执行记录列表 */
export function fetchDepreciationRuns(params?: {
  page?: number
  pageSize?: number
  asset_id?: string
  period?: string
  doc_status?: number
}): Promise<ApiResponse<{ list: DepreciationRun[]; total: number }>> {
  return request.get('/depreciation/runs', { params })
}

/** 生成折旧凭证草稿（docstatus=0） */
export function generateDepreciationVoucher(data: {
  run_ids: string[]
}): Promise<ApiResponse<{ voucher_id: string }>> {
  return request.post('/depreciation/generate', data)
}

/** 生成摊销凭证草稿（docstatus=0） */
export function generateAmortizationVoucher(data: {
  run_ids: string[]
}): Promise<ApiResponse<{ voucher_id: string }>> {
  return request.post('/depreciation/generate-amortization', data)
}