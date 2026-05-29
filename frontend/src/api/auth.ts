import { authApi } from './adapter/client'
import type { LoginRequest, LoginResponse } from '@/types'
import { useAuthStore } from '@/stores/auth'

export const authService = {
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await authApi.login(data)
    
    // 保存 token 和用户信息
    localStorage.setItem('token', response.token)
    localStorage.setItem('user', JSON.stringify(response.user))
    
    // 更新 store
    const authStore = useAuthStore()
    authStore.setAuth(response.token, response.user)
    
    return response
  },
  
  async logout(): Promise<void> {
    try {
      await authApi.logout()
    } catch {
      // 即使失败也清理本地数据
    } finally {
      this.clearAuth()
    }
  },
  
  clearAuth(): void {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    
    const authStore = useAuthStore()
    authStore.clearAuth()
  }
}

export default authService