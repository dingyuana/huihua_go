import { http, HttpResponse } from 'msw'
import { AccountType } from '@/types/enums'

// 内置科目表 Mock
const mockAccounts = [
  { id: 'a1', code: '1001', name: '银行存款', account_type: AccountType.Asset, root_type: 'debit', parent_id: null, lft: 1, rgt: 6, is_group: true, company_id: 'c1', currency: 'CNY' },
  { id: 'a1-1', code: '1001-01', name: '银行存款-工行', account_type: 'asset', root_type: 'debit', parent_id: 'a1', lft: 2, rgt: 3, is_group: false, company_id: 'c1', currency: 'CNY' },
  { id: 'a1-2', code: '1001-02', name: '银行存款-建行', account_type: 'asset', root_type: 'debit', parent_id: 'a1', lft: 4, rgt: 5, is_group: false, company_id: 'c1', currency: 'CNY' },
  { id: 'a2', code: '1122', name: '应收账款', account_type: 'asset', root_type: 'debit', parent_id: null, lft: 7, rgt: 10, is_group: true, company_id: 'c1', currency: 'CNY' },
  { id: 'a2-1', code: '1122-01', name: '应收账款-A公司', account_type: 'asset', root_type: 'debit', parent_id: 'a2', lft: 8, rgt: 9, is_group: false, company_id: 'c1', currency: 'CNY' },
  { id: 'a3', code: '2001', name: '应付账款', account_type: 'liability', root_type: 'credit', parent_id: null, lft: 11, rgt: 14, is_group: true, company_id: 'c1', currency: 'CNY' },
  { id: 'a3-1', code: '2001-01', name: '应付账款-B公司', account_type: 'liability', root_type: 'credit', parent_id: 'a3', lft: 12, rgt: 13, is_group: false, company_id: 'c1', currency: 'CNY' },
  { id: 'a4', code: '6001', name: '主营业务收入', account_type: 'income', root_type: 'credit', parent_id: null, lft: 15, rgt: 16, is_group: false, company_id: 'c1', currency: 'CNY' },
  { id: 'a5', code: '6401', name: '主营业务成本', account_type: 'expense', root_type: 'debit', parent_id: null, lft: 17, rgt: 18, is_group: false, company_id: 'c1', currency: 'CNY' },
]

export const accountHandlers = [
  http.get('/api/v1/accounts/tree', () => {
    return HttpResponse.json({ code: 0, data: mockAccounts })
  }),
  http.get('/api/v1/accounts/ledger-only', () => {
    return HttpResponse.json({ code: 0, data: mockAccounts.filter(a => !a.is_group) })
  }),
  http.post('/api/v1/accounts/auto-code', () => {
    return HttpResponse.json({ code: 0, data: { suggested_code: '1001-03' } })
  }),
  http.post('/api/v1/accounts', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json({ code: 0, data: { id: 'new-uuid', ...body as any } })
  }),
]
