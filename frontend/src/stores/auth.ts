import { defineStore } from 'pinia'
import type { User } from '@/types'

interface AuthState {
  token: string | null
  user: User | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem('token'),
    user: (() => {
      const userStr = localStorage.getItem('user')
      return userStr ? JSON.parse(userStr) : null
    })()
  }),
  
  getters: {
    isAuthenticated: (state) => !!state.token && !!state.user,
    currentUser: (state) => state.user,
    userRole: (state) => state.user?.role || '',
    isAdmin: (state) => state.user?.role === 'admin'
  },
  
  actions: {
    setAuth(token: string, user: User) {
      this.token = token
      this.user = user
    },
    
    clearAuth() {
      this.token = null
      this.user = null
    },
    
    hasPermission(permission: string): boolean {
      // admin 拥有所有权限
      if (this.isAdmin) return true
      
      // 其他角色暂时简单处理
      return false
    }
  }
})