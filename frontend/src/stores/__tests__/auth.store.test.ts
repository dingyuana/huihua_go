import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '@/stores/auth.store'
import { Role } from '@/types/enums'

describe('AuthStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
  })

  it('starts with null token and not logged in', () => {
    const store = useAuthStore()
    expect(store.token).toBeNull()
    expect(store.isLoggedIn).toBe(false)
  })

  it('setAuth sets token and user', () => {
    const store = useAuthStore()
    store.setAuth('test-token-abc', {
      id: 'u-001',
      username: 'admin',
      name: '管理员',
      role: Role.Admin,
      permissions: ['voucher.create', 'voucher.approve'],
      avatar: null,
    })

    expect(store.token).toBe('test-token-abc')
    expect(store.isLoggedIn).toBe(true)
    expect(store.user?.username).toBe('admin')
    expect(store.permissions).toContain('voucher.create')
  })

  it('hasPermission checks correctly', () => {
    const store = useAuthStore()
    store.setAuth('test', {
      id: 'u-002',
      username: 'operator',
      name: '操作员',
      role: Role.Cashier,
      permissions: ['voucher.read'],
      avatar: null,
    })

    expect(store.hasPermission('voucher.read')).toBe(true)
    expect(store.hasPermission('voucher.create')).toBe(false)
  })

  it('hasRole checks correctly', () => {
    const store = useAuthStore()
    store.setAuth('test', {
      id: 'u-003',
      username: 'admin',
      name: '管理员',
      role: Role.Admin,
      permissions: [],
      avatar: null,
    })

    expect(store.hasRole('admin')).toBe(true)
    expect(store.hasRole('operator')).toBe(false)
  })

  it('logout clears token and user', () => {
    const store = useAuthStore()
    store.setAuth('test-token', {
      id: 'u-001',
      username: 'admin',
      name: '管理员',
      role: Role.Admin,
      permissions: [],
      avatar: null,
    })

    expect(store.isLoggedIn).toBe(true)

    // logout will try to call router.push, so we expect it might throw
    // but token should still be cleared
    try {
      store.logout()
    } catch {
      // router not available in test
    }

    expect(store.token).toBeNull()
    expect(store.user).toBeNull()
    expect(store.isLoggedIn).toBe(false)
  })
})
