import { SignJWT } from 'jose'
import type { User } from '@/types/models/user'
import { Role } from '@/types/enums'

/**
 * 开发模式：前端签发 JWT（生产环境应替换为后端 /auth/login）
 * 与后端 pkg/jwt/utils.go 的 Claims 结构对齐
 */
const JWT_SECRET = new TextEncoder().encode(
  'dev-secret-key-for-local-testing-only-change-in-prod',
)

const DEV_USERS: Record<string, { password: string; user: User; tenantId: string }> = {
  admin: {
    password: 'admin123',
    user: {
      id: '12345678-1234-1234-1234-123456789abc',
      name: '张三',
      email: 'admin@example.com',
      role: Role.Admin,
      permissions: [
        'account:read', 'account:write',
        'voucher:read', 'voucher:write', 'voucher:submit', 'voucher:reverse',
        'bank:import', 'bank:classify', 'bank:confirm',
        'report:read', 'period:close',
        'party:read', 'party:write',
        'reconciliation:read', 'reconciliation:write',
      ],
    },
    tenantId: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee',
  },
  cashier: {
    password: '123456',
    user: {
      id: '22345678-1234-1234-1234-123456789abc',
      name: '李四',
      email: 'cashier@example.com',
      role: Role.Cashier,
      permissions: ['bank:import', 'bank:classify', 'bank:confirm'],
    },
    tenantId: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee',
  },
  boss: {
    password: '123456',
    user: {
      id: '32345678-1234-1234-1234-123456789abc',
      name: '王老板',
      email: 'boss@example.com',
      role: Role.Boss,
      permissions: ['report:read'],
    },
    tenantId: 'aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee',
  },
}

export async function generateToken(
  account: string,
  password: string,
): Promise<{ token: string; user: User } | null> {
  const devUser = DEV_USERS[account.toLowerCase()]
  if (!devUser || devUser.password !== password) {
    return null
  }

  const now = Math.floor(Date.now() / 1000)
  const token = await new SignJWT({
    user_id: devUser.user.id,
    tenant_id: devUser.tenantId,
    role: devUser.user.role,
  })
    .setProtectedHeader({ alg: 'HS256' })
    .setIssuedAt(now)
    .setNotBefore(now)
    .setExpirationTime(now + 3600) // 1 hour
    .setIssuer('huihua-finance')
    .sign(JWT_SECRET)

  return { token, user: devUser.user }
}

export async function verifyTokenAndFetchUser(
  token: string,
): Promise<{ user: User; tenantId: string } | null> {
  try {
    // 本地解析 JWT payload 获取用户信息（开发模式）
    const payload = JSON.parse(atob(token.split('.')[1]))
    // 从 DEV_USERS 匹配用户
    const found = Object.values(DEV_USERS).find(u => u.user.id === payload.user_id)
    if (found) {
      return { user: found.user, tenantId: payload.tenant_id }
    }
    return null
  } catch {
    return null
  }
}
