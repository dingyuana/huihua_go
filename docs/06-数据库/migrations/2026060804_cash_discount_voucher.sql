-- Migration 2026060804: 现金折扣凭证映射表 cash_discount_vouchers
--
-- 业务背景:
--   采购 V1.0 §3.7 现金折扣场景：discount > 0 时，AllocateToPaymentEntry 末尾
--   在调用 CreateCashDiscount 后会调用 GenerateCashDiscountVoucher 自动生成凭证。
--   本表记录 payment_allocation 与凭证 (voucher_id) 的 1:1 映射关系，
--   用于：
--     1) 通过 payment_allocation_id 反查对应的现金折扣凭证
--     2) 幂等性：UNIQUE 约束保证同一 allocation 只生成一张凭证
--     3) 按租户/发票类型做报表（销售/采购 折扣凭证分别统计）
--
-- 与 cash_discounts (migrations/058) 的区别：
--   cash_discounts        — 折扣痕迹记录（财务费用事实）
--   cash_discount_vouchers — 折扣凭证关联（凭证维度）
--   一张 allocation 记录 → 一条 cash_discounts → 一条 cash_discount_vouchers

CREATE TABLE IF NOT EXISTS cash_discount_vouchers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_allocation_id UUID NOT NULL UNIQUE,
  voucher_id UUID NOT NULL,
  invoice_type VARCHAR(20) NOT NULL,  -- 'sale' 销售折扣 / 'purchase' 采购折扣
  tenant_id UUID,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- 按租户查询（多租户隔离 + 报表）
CREATE INDEX IF NOT EXISTS idx_cash_discount_vouchers_tenant
  ON cash_discount_vouchers(tenant_id);

COMMENT ON TABLE cash_discount_vouchers IS
  '现金折扣凭证关联表：payment_allocation ↔ voucher (1:1)';
COMMENT ON COLUMN cash_discount_vouchers.payment_allocation_id IS
  '关联的 payment_allocations.id (UNIQUE)';
COMMENT ON COLUMN cash_discount_vouchers.voucher_id IS
  '生成的凭证 (journal_entries.id)';
COMMENT ON COLUMN cash_discount_vouchers.invoice_type IS
  '发票类型: sale / purchase';
COMMENT ON COLUMN cash_discount_vouchers.tenant_id IS
  '租户 ID';
