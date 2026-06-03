/** 检查项状态 */
export type CheckStatus = 'passed' | 'warning' | 'blocked' | 'pending'

/** 检查项操作 */
export interface CheckAction {
  label: string
  route?: string
  callback?: () => void
}

/** 单个检查项 */
export interface CheckItem {
  id: string
  name: string
  category?: string
  status: CheckStatus
  message: string
  detail?: string
  action?: CheckAction
}

/** 检查汇总 */
export interface CheckSummary {
  total: number
  passed: number
  warning: number
  blocked: number
  pending: number
}

/** 检查报告 */
export interface CheckReport {
  period: string
  overall: 'green' | 'yellow' | 'red'
  checks: CheckItem[]
  summary: CheckSummary
  generatedAt: string
}

/** 检查项状态标签映射 */
export const CheckStatusLabel: Record<CheckStatus, string> = {
  passed: '通过',
  warning: '警告',
  blocked: '阻断',
  pending: '待检查',
}

export const CheckStatusTagType: Record<CheckStatus, string> = {
  passed: 'success',
  warning: 'warning',
  blocked: 'danger',
  pending: 'info',
}
