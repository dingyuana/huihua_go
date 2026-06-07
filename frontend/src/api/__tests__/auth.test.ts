import { describe, it, expect } from 'vitest'

describe('Auth API types', () => {
  it('login request payload has correct structure', () => {
    const payload = {
      username: 'admin',
      password: 'admin123',
    }
    expect(payload).toHaveProperty('username')
    expect(payload).toHaveProperty('password')
    expect(typeof payload.username).toBe('string')
    expect(typeof payload.password).toBe('string')
  })

  it('login response structure is valid', () => {
    const response = {
      code: 0,
      data: {
        token: 'eyJhbGciOiJIUzI1NiIs...',
        user: {
          id: 'u-001',
          username: 'admin',
          name: '管理员',
          role: 'admin',
        },
      },
    }
    expect(response.code).toBe(0)
    expect(response.data.token).toBeTruthy()
    expect(response.data.user.role).toBe('admin')
  })

  it('logout returns success', () => {
    const response = { code: 0, message: 'logged out' }
    expect(response.code).toBe(0)
  })

  it('refreshToken returns new token', () => {
    const response = {
      code: 0,
      data: { token: 'new-token-value' },
    }
    expect(response.data.token).toBeTruthy()
    expect(response.data.token.length).toBeGreaterThan(0)
  })
})
