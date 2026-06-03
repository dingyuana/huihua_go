import type { Role } from '@/types/enums'

export interface User {
  id: string
  name: string
  email: string
  role: Role
  permissions: string[]
  tenant_id?: string
}

export type RoleType = Role
