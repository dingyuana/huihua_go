import request from '@/api/request'
import type { ApiResponse, PageResult, PageQuery } from '@/types/api'

export interface ClassificationRule {
  id?: string
  name: string
  rule_type: 'keyword' | 'keyword_regex' | 'counterparty_match'
  pattern: string
  match_field: 'description' | 'counterparty'
  direction: 'in' | 'out' | ''
  classification: string
  priority: number
  is_active: boolean
  created_at?: string
  updated_at?: string
}

/** 获取规则列表 */
export function fetchClassificationRules(params?: PageQuery & { keyword?: string; rule_type?: string }): Promise<ApiResponse<ClassificationRule[]>> {
  return request.get('/classification-rules', { params })
}

/** 创建规则 */
export function createClassificationRule(data: Omit<ClassificationRule, 'id' | 'created_at' | 'updated_at'>): Promise<ApiResponse<ClassificationRule>> {
  return request.post('/classification-rules', data)
}

/** 更新规则 */
export function updateClassificationRule(id: string, data: Partial<ClassificationRule>): Promise<ApiResponse<ClassificationRule>> {
  return request.put(`/classification-rules/${id}`, data)
}

/** 删除规则 */
export function deleteClassificationRule(id: string): Promise<ApiResponse<void>> {
  return request.delete(`/classification-rules/${id}`)
}

/** 调整规则优先级 */
export function reorderClassificationRules(ruleIds: string[]): Promise<ApiResponse<void>> {
  return request.post('/classification-rules/reorder', { rule_ids: ruleIds })
}

/** 初始化默认规则 */
export function seedClassificationRules(): Promise<ApiResponse<void>> {
  return request.post('/classification-rules/seed')
}
