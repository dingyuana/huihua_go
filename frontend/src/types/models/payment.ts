import type { PaymentType, PartyType } from '@/types/enums'

export interface PaymentEntry {
  id: string
  payment_no: string
  payment_type: PaymentType
  party_type: PartyType
  party_id: string
  party_name: string
  paid_from_id: string
  paid_to_id: string
  paid_amount: string
  received_amount: string
  reference_no: string
  posting_date: string
  bank_account_id: string | null
  docstatus: number
}

export interface PaymentAllocation {
  id: string
  payment_entry_id: string
  invoice_id: string
  invoice_type: string
  allocated_amount: string
  created_at: string
}
