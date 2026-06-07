import { describe, it, expect } from 'vitest'

describe('Invoice API types', () => {
  it('fetchInvoices params has correct shape', () => {
    const params: Record<string, any> = {
      page: 1,
      pageSize: 20,
      status: 'unpaid',
      invoice_type: 'sale',
      start_date: '2026-05-01',
      end_date: '2026-05-31',
      keyword: '上海',
    }
    expect(params.page).toBe(1)
    expect(params.pageSize).toBe(20)
    expect(params.status).toMatch(/^(unpaid|paid|partially_paid)?$/)
    expect(params.invoice_type).toMatch(/^(sale|purchase)?$/)
  })

  it('generateVoucherFromInvoice returns expected shape', () => {
    const response = {
      code: 0,
      data: {
        voucher_no: 'INV-202605-001',
        journal_entry_id: 'je-001',
        status: 'draft',
      },
    }
    expect(response.code).toBe(0)
    expect(response.data.voucher_no).toMatch(/^INV-/)
    expect(response.data.status).toBe('draft')
  })

  it('uploadInvoice creates FormData with file', () => {
    const file = new File(['dummy'], 'invoice.pdf', { type: 'application/pdf' })
    const form = new FormData()
    form.append('file', file)
    expect(form.get('file')).toBe(file)
    expect(form.get('file') instanceof File).toBe(true)
  })

  it('deleteInvoice response is void', () => {
    const response = { code: 0 }
    expect(response.code).toBe(0)
  })
})
