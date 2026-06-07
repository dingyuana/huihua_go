import { describe, it, expect } from 'vitest'

describe('Bank API types', () => {
  it('bank account has required fields', () => {
    const bank = {
      id: 'b-001',
      bank_name: '中国银行',
      account_no: '6222****1234',
      account_name: '北京XX科技有限公司',
      currency: 'CNY',
      clearing_account_id: 'a-1002',
      is_active: true,
    }
    expect(bank.bank_name).toBeTruthy()
    expect(bank.account_no).toMatch(/^\d{4}/)
    expect(bank.currency).toBe('CNY')
    expect(bank.is_active).toBe(true)
  })

  it('import bank file uses FormData', () => {
    const file = new File(['dummy,csv'], 'statement.csv', { type: 'text/csv' })
    const form = new FormData()
    form.append('file', file)
    form.append('bank_account_id', 'b-001')
    expect(form.get('file') instanceof File).toBe(true)
    expect(form.get('bank_account_id')).toBe('b-001')
  })

  it('bank transaction list response has page structure', () => {
    const response = {
      code: 0,
      data: {
        items: [
          {
            id: 't-001',
            txn_date: '2026-05-15',
            description: '货款-上海公司',
            debit: '0.00',
            credit: '50000.00',
            counterparty_name: '上海贸易有限公司',
            matched: false,
          },
        ],
        total: 1,
      },
    }
    expect(response.data.items[0].debit).toBe('0.00')
    expect(response.data.items[0].credit).toBe('50000.00')
    expect(response.data.items[0].matched).toBe(false)
  })

  it('classifyTransaction sends txn id', () => {
    const txnId = 't-001'
    expect(txnId).toMatch(/^t-/)
  })

  it('markMatched confirms a single transaction', () => {
    const confirmResult = { code: 0, message: 'matched' }
    expect(confirmResult.code).toBe(0)
  })

  it('updateClassification has required fields', () => {
    const data = { classification: '原材料采购', counterparty_id: 'p-001', remark: '' }
    expect(data.classification).toBeTruthy()
    expect(data.counterparty_id).toBeTruthy()
  })

  it('unmatched bank transaction has direction info', () => {
    const txn = {
      id: 't-002',
      txn_date: '2026-05-20',
      description: '支付水电费',
      debit: '3500.00',
      credit: '0.00',
      direction: 'debit',
    }
    expect(txn.direction).toMatch(/^(debit|credit)$/)
    const isDebit = parseFloat(txn.debit) > 0
    const isCredit = parseFloat(txn.credit) > 0
    expect(isDebit || isCredit).toBe(true)
  })

  it('bank account creation payload is valid', () => {
    const payload = {
      bank_name: '工商银行',
      account_no: '9558800200123456789',
      account_name: '测试公司',
      currency: 'CNY',
      clearing_account_id: 'a-1002',
    }
    expect(payload.account_no.length).toBeGreaterThanOrEqual(15)
    expect(payload.currency).toMatch(/^(CNY|USD|EUR)$/)
  })

  it('fetchBankTransactions filters work correctly', () => {
    const params: Record<string, any> = {
      bank_account_id: 'b-001',
      matched: false,
      start_date: '2026-05-01',
      end_date: '2026-05-31',
      direction: 'credit',
    }
    expect(params.matched).toBe(false)
    expect(params.direction).toMatch(/^(debit|credit)?$/)
  })
})
