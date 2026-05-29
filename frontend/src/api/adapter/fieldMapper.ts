// 字段映射：Go后端 snake_case → 前端 camelCase
export const fieldMapper: Record<string, string> = {
  // 通用
  created_at: 'createdAt',
  updated_at: 'updatedAt',
  tenant_id: 'tenantId',
  
  // 凭证
  voucher_id: 'voucherId',
  voucher_date: 'date',
  voucher_number: 'number',
  
  // 银行
  bank_account_id: 'bankAccountId',
  bank_name: 'bankName',
  account_no: 'accountNo',
  
  // 发票
  invoice_type: 'type',
  tax_amount: 'taxAmount',
  
  // 往来
  party_type: 'type',
  
  // 汇率
  from_currency: 'fromCurrency',
  to_currency: 'toCurrency',
  effective_date: 'effectiveDate',
  
  // 报表
  account_code: 'accountCode',
  account_name: 'accountName',
  debit_balance: 'debitBalance',
  credit_balance: 'creditBalance',
  period: 'period',
  as_of_date: 'asOfDate'
}

// 状态映射
export const statusMapper: Record<string, Record<number, { label: string; color: string }>> = {
  voucher: {
    0: { label: '待录入', color: 'default' },
    1: { label: '已提交', color: 'info' },
    2: { label: '已核准', color: 'success' },
    3: { label: '已驳回', color: 'warning' },
    4: { label: '已过账', color: 'success' },
    5: { label: '已红冲', color: 'error' }
  },
  invoice: {
    0: { label: '草稿', color: 'default' },
    1: { label: '已审核', color: 'success' },
    2: { label: '已作废', color: 'warning' }
  }
}

// 映射对象字段（递归）
export function mapFields<T extends Record<string, any>>(obj: T, fields: Record<string, string>): T {
  if (!obj || typeof obj !== 'object') return obj
  
  const result: Record<string, any> = {}
  
  for (const key in obj) {
    const newKey = fields[key] || key
    const value = obj[key]
    
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      result[newKey] = mapFields(value, fields)
    } else if (Array.isArray(value)) {
      result[newKey] = value.map((item: any) => 
        item && typeof item === 'object' ? mapFields(item, fields) : item
      )
    } else {
      result[newKey] = value
    }
  }
  
  return result as T
}