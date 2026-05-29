import type { Role } from '@/types/enums'

export interface User {
  id: string
  name: string
  email: string
  role: Role
  permissions: string[]
}

export type RoleType = Role
