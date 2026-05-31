import request from '@/api/request'

/** 登录响应（后端直接返回，无 code/data 包装） */
export interface LoginResponse {
  token: string
  user_id: string
  tenant_id: string
  role: string
  expires_at: string
}

/** 登录 */
export function login(username: string, password: string): Promise<LoginResponse> {
  return request.post('/auth/login', { username, password })
}

/** 登出 */
export function logout(): Promise<void> {
  return request.post('/auth/logout')
}

/** 获取当前用户 */
export function fetchMe(): Promise<{
  id: string
  name: string
  email: string
  role: string
  permissions: string[]
  tenant_id: string
  tenant_name: string
}> {
  return request.get('/auth/me')
}

/** 刷新 token（切换租户时） */
export function refreshToken(newTenantId: string): Promise<{ token: string; expires_at: string }> {
  return request.post('/auth/refresh', { new_tenant_id: newTenantId })
}
