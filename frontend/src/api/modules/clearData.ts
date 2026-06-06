import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

export interface ClearResult {
  scope: 'business' | 'basic'
  result: Record<string, number>
  total: number
}

/** 清空业务数据 (发票/银行流水/收付款单/凭证/银行余额等) — dev/test only */
export function clearBusinessData(): Promise<ApiResponse<ClearResult>> {
  return request.post('/setup/clear-business-data', {})
}

/** 清空基本信息 (公司信息/银行账户/科目/客商等) — dev/test only */
export function clearBasicInfo(): Promise<ApiResponse<ClearResult>> {
  return request.post('/setup/clear-basic-info', {})
}
