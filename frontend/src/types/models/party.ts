import type { PartyType } from '@/types/enums'

export interface Party {
  id: string
  name: string
  tax_id: string
  party_type: PartyType
  bank_name: string
  bank_account: string
  credit_limit: string
  payment_terms: string
  phone: string
  address: string
  is_active: boolean
}
