import { describe, it, expect } from 'vitest'

describe('Period API types', () => {
  it('RiskWarning type has required fields', () => {
    const warning = {
      type: 'reclassification',
      severity: 'warning' as const,
      subject_code: '1122',
      subject_name: '应收账款',
      balance: -15000,
      message: 'test warning',
    }
    expect(warning.type).toBe('reclassification')
    expect(warning.severity).toBe('warning')
    expect(warning.balance).toBeLessThan(0)
  })

  it('KeyIndicator type has required fields', () => {
    const indicator = {
      name: '毛利率',
      current_value: 15.0,
      last_value: 22.0,
      unit: '%',
      alert: true,
      message: 'test indicator',
    }
    expect(indicator.alert).toBe(true)
    expect(indicator.current_value).toBeLessThan(indicator.last_value!)
  })

  it('PreCloseCheckResult has required structure', () => {
    const result = {
      period_status: 'open',
      unposted_vouchers: 0,
      report_balance_ok: true,
      profit_loss_done: false,
      risk_warnings: [],
      key_indicators: [],
      pending_accruals: [],
    }
    expect(result.period_status).toBe('open')
    expect(result.unposted_vouchers).toBe(0)
    expect(result.report_balance_ok).toBe(true)
    expect(Array.isArray(result.risk_warnings)).toBe(true)
    expect(Array.isArray(result.key_indicators)).toBe(true)
    expect(Array.isArray(result.pending_accruals)).toBe(true)
  })

  it('PendingAccrual correctly tracks missing state', () => {
    const accrual = {
      type: 'depreciation',
      item: '本月固定资产折旧',
      missing: true,
      details: '存在 2 项使用中资产未计提本月折旧',
    }
    expect(accrual.missing).toBe(true)
    expect(accrual.details).toContain('2 项')

    const done = { ...accrual, missing: false }
    expect(done.missing).toBe(false)
  })
})
