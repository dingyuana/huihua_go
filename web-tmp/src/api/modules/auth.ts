import request from '@/api/request'
import type { ApiResponse } from '@/types/api'

/** 登录 */
export function login(account: string, password: string): Promise<ApiResponse<{ token: string; expires_at: string }>> {
  return request.post('/auth/login', { account, password })
}

/** 登出 */
export function logout(): Promise<ApiResponse<null>> {
  return request.post('/auth/logout')
}

/** 获取当前用户 */
export function fetchMe(): Promise<ApiResponse<{
  id: string
  name: string
  email: string
  role: string
  permissions: string[]
  tenant_id: string
  tenant_name: string
}>> {
  return request.get('/auth/me')
}

/** 刷新 token（切换租户时） */
export function refreshToken(newTenantId: string): Promise<ApiResponse<{ token: string; expires_at: string }>> {
  return request.post('/auth/refresh', { new_tenant_id: newTenantId })
}
