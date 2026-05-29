import type { AccountType, RootType } from '@/types/enums'

export interface Account {
  id: string
  code: string
  name: string
  account_type: AccountType
  root_type: RootType
  parent_id: string | null
  lft: number
  rgt: number
  is_group: boolean
  company_id: string
  currency: string
  children?: Account[]
  created_at: string
}
