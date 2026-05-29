import { http, HttpResponse } from 'msw'

const mockVouchers = Array.from({ length: 10 }, (_, i) => ({
  id: `v-${i}`,
  voucher_no: `记-2026-05-${String(i + 1).padStart(4, '0')}`,
  voucher_type: '记',
  posting_date: `2026-05-${String(i + 1).padStart(2, '0')}`,
  company_id: 'c1',
  remark: `凭证摘要-${i}`,
  docstatus: i < 7 ? 1 : 0,
  lines: [
    { id: `vl-${i}-1`, account_id: 'a1-1', account_code: '1001-01', account_name: '银行存款-工行', debit: `${(i + 1) * 1000}.00`, credit: '0.00' },
    { id: `vl-${i}-2`, account_id: 'a2-1', account_code: '1122-01', account_name: '应收账款-A公司', debit: '0.00', credit: `${(i + 1) * 1000}.00` },
  ],
  created_by: 'user-001',
  created_at: '2026-05-27T10:00:00Z',
}))

export const voucherHandlers = [
  http.get('/api/v1/vouchers', ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    return HttpResponse.json({ code: 0, data: { list: mockVouchers, total: mockVouchers.length, page, pageSize: 20 } })
  }),
  http.get('/api/v1/vouchers/pending-review', () => {
    return HttpResponse.json({ code: 0, data: { list: mockVouchers.filter(v => v.docstatus === 0), total: 3, page: 1, pageSize: 20 } })
  }),
  http.post('/api/v1/vouchers', async ({ request }) => {
    const body = await request.json()
    return HttpResponse.json({ code: 0, data: { ...(body as any), id: 'v-new', voucher_no: null, docstatus: 0 } })
  }),
  http.post('/api/v1/vouchers/:id/submit', () => {
    return HttpResponse.json({ code: 0, data: { id: 'v-1', voucher_no: '记-2026-05-0011', docstatus: 1, submitted_by: 'user-001', submitted_at: '2026-05-27T10:30:00Z', gl_entries_generated: 2 } })
  }),
  http.post('/api/v1/vouchers/:id/reverse', () => {
    return HttpResponse.json({ code: 0, data: { original_voucher_no: '记-2026-05-0001', reverse_voucher_no: '记-2026-05-0012', reverse_voucher_id: 'v-reverse', docstatus: 0 } })
  }),
]
