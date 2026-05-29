// ========== 通用响应类型 ==========
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// ========== 用户与认证 ==========
export interface User {
  id: string
  username: string
  role: string
  tenant_id: string
}

export interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

// ========== 凭证 ==========
export interface VoucherLine {
  id?: string
  account_id: string
  account_code?: string
  account_name?: string
  debit: number
  credit: number
  summary?: string
}

export interface Voucher {
  id: string
  date: string
  number: string
  status: number
  creator?: string
  created_at?: string
  updated_at?: string
  lines: VoucherLine[]
}

export const VOUCHER_STATUS_MAP: Record<number, { label: string; color: string }> = {
  0: { label: '待录入', color: 'default' },
  1: { label: '已提交', color: 'info' },
  2: { label: '已核准', color: 'success' },
  3: { label: '已驳回', color: 'warning' },
  4: { label: '已过账', color: 'success' },
  5: { label: '已红冲', color: 'error' }
}

// ========== 银行账户 ==========
export interface BankAccount {
  id: string
  name: string
  bank_name: string
  account_no: string
  currency: string
  balance: number
}

// ========== 银行流水 ==========
export interface BankTransaction {
  id: string
  bank_account_id: string
  date: string
  description: string
  amount: number
  type: 'debit' | 'credit'
  matched: boolean
  matched_voucher_id?: string
}

// ========== 发票 ==========
export interface Invoice {
  id: string
  number: string
  type: 'income' | 'expense'
  party_id?: string
  party_name?: string
  date: string
  amount: number
  tax_amount: number
  status: number
}

// ========== 往来单位 ==========
export interface Party {
  id: string
  name: string
  type: 'customer' | 'supplier' | 'both'
  contact?: string
  phone?: string
  email?: string
  address?: string
}

// ========== 汇率 ==========
export interface ExchangeRate {
  id: string
  from_currency: string
  to_currency: string
  rate: number
  effective_date: string
}

// ========== 报表 ==========
export interface TrialBalanceItem {
  account_code: string
  account_name: string
  debit_balance: number
  credit_balance: number
  period: string
}

export interface TrialBalanceReport {
  items: TrialBalanceItem[]
  total_debit: number
  total_credit: number
  as_of_date: string
}

// ========== 科目树 ==========
export interface AccountNode {
  id: string
  code: string
  name: string
  parent_id?: string
  children?: AccountNode[]
  type?: number
}