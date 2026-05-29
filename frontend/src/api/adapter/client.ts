import apiClient from '../index'
import type { 
  LoginRequest, 
  LoginResponse, 
  Voucher, 
  VoucherLine,
  BankAccount,
  BankTransaction,
  Invoice,
  Party,
  ExchangeRate,
  TrialBalanceReport,
  AccountNode,
  ApiResponse 
} from '@/types'

// ========== 认证 ==========
export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    const response = await apiClient.post<ApiResponse<LoginResponse>>('/api/v1/auth/login', data)
    return response.data.data
  },
  
  logout: async (): Promise<void> => {
    await apiClient.post('/api/v1/auth/logout')
  },
  
  getProfile: async (): Promise<LoginResponse['user']> => {
    const response = await apiClient.get<ApiResponse<LoginResponse['user']>>('/api/v1/auth/profile')
    return response.data.data
  }
}

// ========== 凭证 ==========
export const voucherApi = {
  list: async (params?: { page?: number; page_size?: number; status?: number }): Promise<{ vouchers: Voucher[]; total: number }> => {
    const response = await apiClient.get<ApiResponse<{ vouchers: Voucher[]; total: number }>>('/api/v1/vouchers', { params })
    return response.data.data
  },
  
  get: async (id: string): Promise<Voucher> => {
    const response = await apiClient.get<ApiResponse<Voucher>>(`/api/v1/vouchers/${id}`)
    return response.data.data
  },
  
  create: async (data: { date: string; lines: Omit<VoucherLine, 'id'>[] }): Promise<Voucher> => {
    const response = await apiClient.post<ApiResponse<Voucher>>('/api/v1/vouchers', data)
    return response.data.data
  },
  
  update: async (id: string, data: { date?: string; lines?: Omit<VoucherLine, 'id'>[] }): Promise<Voucher> => {
    const response = await apiClient.put<ApiResponse<Voucher>>(`/api/v1/vouchers/${id}`, data)
    return response.data.data
  },
  
  submit: async (id: string): Promise<void> => {
    await apiClient.post(`/api/v1/vouchers/${id}/submit`)
  },
  
  approve: async (id: string): Promise<void> => {
    await apiClient.post(`/api/v1/vouchers/${id}/approve`)
  },
  
  post: async (id: string): Promise<void> => {
    await apiClient.post(`/api/v1/vouchers/${id}/post`)
  },
  
  reverse: async (id: string): Promise<void> => {
    await apiClient.post(`/api/v1/vouchers/${id}/reverse`)
  }
}

// ========== 银行账户 ==========
export const bankAccountApi = {
  list: async (): Promise<BankAccount[]> => {
    const response = await apiClient.get<ApiResponse<BankAccount[]>>('/api/v1/bank-accounts')
    return response.data.data
  },
  
  get: async (id: string): Promise<BankAccount> => {
    const response = await apiClient.get<ApiResponse<BankAccount>>(`/api/v1/bank-accounts/${id}`)
    return response.data.data
  }
}

// ========== 银行流水 ==========
export const bankTxnApi = {
  list: async (params?: { bank_account_id?: string; matched?: boolean; page?: number; page_size?: number }): Promise<{ transactions: BankTransaction[]; total: number }> => {
    const response = await apiClient.get<ApiResponse<{ transactions: BankTransaction[]; total: number }>>('/api/v1/bank-transactions', { params })
    return response.data.data
  },
  
  unmatched: async (): Promise<number> => {
    const response = await apiClient.get<ApiResponse<{ count: number }>>('/api/v1/bank-transactions/unmatched')
    return response.data.data.count
  },
  
  match: async (id: string, voucherId: string): Promise<void> => {
    await apiClient.post(`/api/v1/bank-transactions/${id}/match`, { voucher_id: voucherId })
  },
  
  unmatchedList: async (): Promise<BankTransaction[]> => {
    const response = await apiClient.get<ApiResponse<BankTransaction[]>>('/api/v1/bank-transactions/unmatched')
    return response.data.data
  }
}

// ========== 发票 ==========
export const invoiceApi = {
  list: async (params?: { type?: string; party_id?: string; page?: number; page_size?: number }): Promise<{ invoices: Invoice[]; total: number }> => {
    const response = await apiClient.get<ApiResponse<{ invoices: Invoice[]; total: number }>>('/api/v1/invoices', { params })
    return response.data.data
  },
  
  get: async (id: string): Promise<Invoice> => {
    const response = await apiClient.get<ApiResponse<Invoice>>(`/api/v1/invoices/${id}`)
    return response.data.data
  },
  
  create: async (data: Omit<Invoice, 'id'>): Promise<Invoice> => {
    const response = await apiClient.post<ApiResponse<Invoice>>('/api/v1/invoices', data)
    return response.data.data
  },
  
  stats: async (): Promise<{ total_count: number; total_amount: number; this_month_count: number; this_month_amount: number }> => {
    const response = await apiClient.get<ApiResponse<{ total_count: number; total_amount: number; this_month_count: number; this_month_amount: number }>>('/api/v1/invoices/stats')
    return response.data.data
  }
}

// ========== 往来单位 ==========
export const partyApi = {
  list: async (params?: { type?: string; page?: number; page_size?: number }): Promise<{ parties: Party[]; total: number }> => {
    const response = await apiClient.get<ApiResponse<{ parties: Party[]; total: number }>>('/api/v1/parties', { params })
    return response.data.data
  },
  
  get: async (id: string): Promise<Party> => {
    const response = await apiClient.get<ApiResponse<Party>>(`/api/v1/parties/${id}`)
    return response.data.data
  },
  
  create: async (data: Omit<Party, 'id'>): Promise<Party> => {
    const response = await apiClient.post<ApiResponse<Party>>('/api/v1/parties', data)
    return response.data.data
  },
  
  update: async (id: string, data: Partial<Party>): Promise<Party> => {
    const response = await apiClient.put<ApiResponse<Party>>(`/api/v1/parties/${id}`, data)
    return response.data.data
  }
}

// ========== 汇率 ==========
export const exchangeRateApi = {
  list: async (): Promise<ExchangeRate[]> => {
    const response = await apiClient.get<ApiResponse<ExchangeRate[]>>('/api/v1/exchange-rates')
    return response.data.data
  },
  
  create: async (data: Omit<ExchangeRate, 'id'>): Promise<ExchangeRate> => {
    const response = await apiClient.post<ApiResponse<ExchangeRate>>('/api/v1/exchange-rates', data)
    return response.data.data
  }
}

// ========== 科目 ==========
export const accountApi = {
  tree: async (): Promise<AccountNode[]> => {
    const response = await apiClient.get<ApiResponse<AccountNode[]>>('/api/v1/accounts/tree')
    return response.data.data
  }
}

// ========== 报表 ==========
export const reportApi = {
  trialBalance: async (params?: { start_date?: string; end_date?: string }): Promise<TrialBalanceReport> => {
    const response = await apiClient.get<ApiResponse<TrialBalanceReport>>('/api/v1/reports/trial-balance', { params })
    return response.data.data
  }
}