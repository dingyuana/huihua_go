import type { Role } from '@/types/enums'

export interface User {
  id: string
  username?: string
  name: string
  email?: string
  role: Role
  permissions: string[]
  avatar?: string | null
  tenant_id?: string
}

export type RoleType = Role
