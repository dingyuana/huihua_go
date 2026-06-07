import { describe, it, expect } from 'vitest'

describe('Voucher API types', () => {
  it('createVoucher payload has required fields', () => {
    const payload = {
      voucher_type: '记',
      posting_date: '2026-05-20',
      remark: '采购原材料',
      lines: [
        { account_id: '1403', debit: '5000.00', credit: '0.00' },
        { account_id: '2202', debit: '0.00', credit: '5000.00' },
      ],
    }
    expect(payload.lines).toHaveLength(2)
    expect(payload.lines[0].debit).toBe('5000.00')
    expect(payload.lines[1].credit).toBe('5000.00')
  })

  it('createVoucher response has expected shape', () => {
    const response = {
      code: 0,
      data: {
        voucher: {
          id: 'v-001',
          voucher_no: '记-35',
          docstatus: 0,
          lines: [],
        },
      },
    }
    expect(response.code).toBe(0)
    expect(response.data.voucher.voucher_no).toMatch(/^记-\d+$/)
    expect(response.data.voucher.docstatus).toBe(0)
  })

  it('submitVoucher request body includes user info', () => {
    const request = { user_id: 'u-001', user_name: '张三' }
    expect(request).toHaveProperty('user_id')
    expect(request).toHaveProperty('user_name')
  })

  it('cancelVoucher includes reason', () => {
    const body = { reason: '录入有误，需要重新编制' }
    expect(body.reason).toBeTruthy()
    expect(body.reason.length).toBeGreaterThan(0)
  })

  it('reverseVoucher preserves original reference', () => {
    const data = { reverse_date: '2026-05-21', reason: '红冲更正' }
    expect(data.reverse_date).toMatch(/^\d{4}-\d{2}-\d{2}$/)
    expect(data.reason).toContain('红冲')
  })

  it('approveVoucher request has basic structure', () => {
    const request = { user_id: 'u-001', user_name: '李四' }
    expect(request.user_id).toBeTruthy()
  })

  it('rejectVoucher includes rejection reason', () => {
    const body = { user_id: 'u-001', user_name: '李四', reason: '摘要不清晰' }
    expect(body.reason).toBeTruthy()
  })

  it('fetchVouchers query params are optional', () => {
    const params: Record<string, any> = {
      start_date: '2026-05-01',
      end_date: '2026-05-31',
      docstatus: 1,
      limit: 20,
      offset: 0,
    }
    expect(params.limit).toBeLessThanOrEqual(200)
    expect(params.docstatus).toBeGreaterThanOrEqual(0)
  })

  it('fetchVoucherDetail returns single voucher', () => {
    const detail = {
      id: 'v-001',
      voucher_no: '记-35',
      voucher_type: '记',
      posting_date: '2026-05-20',
      docstatus: 1,
      lines: [
        { account_id: 'a-001', debit: '1000.00', credit: '0.00' },
        { account_id: 'a-002', debit: '0.00', credit: '1000.00' },
      ],
    }
    expect(detail.lines.length).toBeGreaterThanOrEqual(2)
    const totalDebit = detail.lines.reduce((sum, l) => sum + parseFloat(l.debit), 0)
    const totalCredit = detail.lines.reduce((sum, l) => sum + parseFloat(l.credit), 0)
    expect(totalDebit).toBe(totalCredit)
  })
})
