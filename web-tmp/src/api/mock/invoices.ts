import { http, HttpResponse } from 'msw'

const mockInvoices = Array.from({ length: 15 }, (_, i) => ({
  id: `inv-${i}`,
  invoice_no: `${10000000 + i}`,
  invoice_type: i < 8 ? 'sale' : 'purchase',
  customer_id: i < 8 ? 'c-1' : 's-1',
  customer_name: i < 8 ? '上海XX贸易公司' : '北京YY科技',
  tax_id: `91310000MA${String(i).padStart(6, '0')}`,
  posting_date: `2026-05-${String(i + 1).padStart(2, '0')}`,
  due_date: `2026-06-${String(i + 15).padStart(2, '0')}`,
  total_amount: `${(i + 1) * 10000}.00`,
  tax_amount: `${(i + 1) * 1300}.00`,
  net_amount: `${(i + 1) * 8700}.00`,
  outstanding_amount: i < 10 ? '0.00' : `${(i + 1) * 3000}.00`,
  status: i < 10 ? 'paid' : 'unpaid',
  docstatus: 1,
}))

export const invoiceHandlers = [
  http.post('/api/v1/invoices/upload', () => {
    return HttpResponse.json({
      code: 0,
      data: { id: 'inv-new', invoice_no: '12345678', invoice_type: 'sale', posting_date: '2026-05-27', total_amount: '113000.00', tax_amount: '13000.00', net_amount: '100000.00', customer_name: '上海XX', customer_tax_id: '91310...', confidence: 0.96, status: 'pending_review', field_errors: [] },
    })
  }),
  http.get('/api/v1/invoices', ({ request }) => {
    const url = new URL(request.url)
    const page = parseInt(url.searchParams.get('page') || '1')
    return HttpResponse.json({ code: 0, data: { list: mockInvoices, total: mockInvoices.length, page, pageSize: 20 } })
  }),
  http.get('/api/v1/invoices/expiring', () => {
    return HttpResponse.json({ code: 0, data: { list: mockInvoices.slice(0, 3), total_expiring: 3 } })
  }),
]
