import type { DocStatus, VoucherType, PartyType } from '@/types/enums'

export interface JournalEntry {
  id: string
  voucher_no: string
  voucher_type: VoucherType
  posting_date: string
  company_id: string
  remark: string
  docstatus: DocStatus
  reversed_id: string | null
  reversal_id: string | null
  submitted_by: string | null
  submitted_at: string | null
  created_by: string
  lines: JournalEntryLine[]
  created_at: string
}

export interface JournalEntryLine {
  id: string
  journal_entry_id: string
  account_id: string
  account_code: string
  account_name: string
  debit: string
  credit: string
  debit_ccy: string
  credit_ccy: string
  account_ccy: string
  exchange_rate: string
  party_type: PartyType | null
  party_id: string | null
  cost_center_id: string | null
  project_id: string | null
  user_remark: string
}

export interface GlEntry {
  id: string
  account_id: string
  posting_date: string
  debit: string
  credit: string
  voucher_type: string
  voucher_id: string
  party_type: PartyType | null
  party_id: string | null
  is_cancelled: boolean
}
