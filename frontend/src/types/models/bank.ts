import type { BankTxnDirection, BankTxnClassification, ImportFormat } from '@/types/enums'
import type { Account } from './account'

export interface BankAccount {
  id: string
  bank_name: string
  account_number: string
  clearing_account_id: string
  clearing_account?: Account
  company_id: string
  currency: string
  iban: string
  swift_code: string
  bank_account_type: string
  is_active: boolean
}

export interface BankTransaction {
  id: string
  bank_account_id: string
  txn_date: string
  description: string
  debit: string
  credit: string
  direction: BankTxnDirection
  reference_no: string
  counterparty_name: string
  classification: BankTxnClassification
  matched: boolean
  imported_from: ImportFormat
  is_duplicate: boolean
}

export interface ImportResult {
  batch_id: string
  bank_account_id: string
  total: number
  imported: number
  duplicated: number
  failed: number
  errors: string[]
}
