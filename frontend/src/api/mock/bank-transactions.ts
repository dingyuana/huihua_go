import { http, HttpResponse } from 'msw'

const mockTxns = Array.from({ length: 20 }, (_, i) => ({
  id: `txn-${i}`,
  bank_account_id: 'ba-1',
  txn_date: `2026-05-${String(i + 1).padStart(2, '0')}`,
  description: ['网银转账-收款', '手续费支出', '利息收入', '支付供应商货款'][i % 4],
  debit: i % 2 === 0 ? `${(i + 1) * 1000}.00` : '0.00',
  credit: i % 2 === 1 ? `${(i + 1) * 500}.00` : '0.00',
  direction: i % 2 === 0 ? 'in' : 'out',
  reference_no: `REF${String(i).padStart(6, '0')}`,
  counterparty_name: ['上海XX贸易公司', '', '银行', '北京YY科技'][i % 4],
  classification: ['business_receipt', 'bank_fee', 'interest_income', 'business_payment'][i % 4],
  matched: i < 15,
  imported_from: 'excel',
  is_duplicate: false,
}))

export const bankHandlers = [
  http.post('/api/v1/bank-transactions/import', () => {
    return HttpResponse.json({
      code: 0,
      data: { batch_id: 'batch-001', bank_account_id: 'ba-1', total: 20, imported: 18, duplicated: 1, failed: 1, errors: ['第3行: 金额格式异常'] },
    })
  }),
  http.post('/api/v1/bank-transactions/classify', () => {
    return HttpResponse.json({
      code: 0,
      data: { classifications: { business_receipt: 5, business_payment: 5, bank_fee: 4, interest_income: 3, internal_transfer: 2, pending: 1 }, confidence_avg: 0.87, pending_count: 1 },
    })
  }),
  http.get('/api/v1/bank-transactions', ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    const pageSize = parseInt(url.searchParams.get('pageSize') || '20')
    const start = (page - 1) * pageSize
    return HttpResponse.json({ code: 0, data: { list: mockTxns.slice(start, start + pageSize), total: mockTxns.length, page, pageSize } })
  }),
  http.post('/api/v1/bank-transactions/:id/confirm', () => {
    return HttpResponse.json({ code: 0, data: null })
  }),
]
