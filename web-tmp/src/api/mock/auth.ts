import { http, HttpResponse } from 'msw'

/** 认证 Mock */
export const authHandlers = [
  http.post('/api/v1/auth/login', () => {
    return HttpResponse.json({
      code: 0,
      message: 'ok',
      data: {
        token: 'eyJhbGciOiJIUzI1NiJ9.mock',
        expires_at: '2027-01-01T00:00:00Z',
        user: {
          id: 'user-001', name: '张三', email: 'admin@example.com',
          role: 'admin',
          permissions: ['account:read', 'account:write', 'voucher:read', 'voucher:submit'],
        },
        tenant_id: 'tenant-001',
        tenant_name: '北京XX科技有限公司',
      },
    })
  }),

  http.get('/api/v1/auth/me', () => {
    return HttpResponse.json({
      code: 0,
      data: {
        id: 'user-001', name: '张三', email: 'admin@example.com',
        role: 'admin',
        permissions: ['account:read', 'account:write', 'voucher:read', 'voucher:submit'],
        tenant_id: 'tenant-001',
        tenant_name: '北京XX科技有限公司',
      },
    })
  }),

  http.get('/api/v1/health', () => {
    return HttpResponse.json({ status: 'ok', version: '1.0.0', db: 'connected', redis: 'connected' })
  }),
]
