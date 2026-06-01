import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/models/user'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('huihua_token'))
  const user = ref<User | null>(JSON.parse(localStorage.getItem('huihua_user') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const permissions = computed(() => user.value?.permissions || [])

  function setAuth(t: string, u: User) {
    token.value = t
    user.value = u
    localStorage.setItem('huihua_token', t)
    localStorage.setItem('huihua_user', JSON.stringify(u))
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('huihua_token')
    localStorage.removeItem('huihua_user')
    window.location.href = '/login'
  }

  function hasPermission(perm: string): boolean {
    return permissions.value.includes(perm)
  }

  function hasRole(role: string): boolean {
    return user.value?.role === role
  }

  return { token, user, isLoggedIn, permissions, setAuth, logout, hasPermission, hasRole }
})
