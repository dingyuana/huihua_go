-- Migration 058: 现金折扣痕迹表 cash_discounts
--
-- 业务背景:
--   采购 V1.0 §3.7 现金折扣场景：供应商付款时享受的现金折扣。
--   payment_allocations 中已记录 discount_amount（属于付款明细）。
--   本表为独立的"现金折扣痕迹"记录，用于：
--     1) 审计追溯（哪些发票/付款发生过折扣）
--     2) 财务费用凭证自动生成（后续 P1 任务）
--     3) 折扣率历史分析
--
-- 数据约束:
--   payment_allocation_id: 关联到 payment_allocations.id（折扣发生的那条分配）
--   payment_entry_id / invoice_id: 冗余字段，方便按付款单/发票查询
--   discount_amount: 折扣金额（NUMERIC(18,2)）
--   discount_rate: 折扣率（NUMERIC(8,6)，如 0.020000 = 2%），可空
--   tenant_id: 多租户隔离

CREATE TABLE IF NOT EXISTS cash_discounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  payment_allocation_id UUID NOT NULL,
  payment_entry_id UUID NOT NULL,
  invoice_id UUID NOT NULL,
  discount_amount NUMERIC(18,2) NOT NULL,
  discount_rate NUMERIC(8,6),
  created_at TIMESTAMPTZ DEFAULT now(),
  tenant_id UUID NOT NULL
);

-- 按付款单号查询（最常用：看某笔付款的折扣明细）
CREATE INDEX IF NOT EXISTS idx_cash_discounts_payment
  ON cash_discounts(payment_entry_id);

-- 按发票查询（看某张发票的历史折扣记录）
CREATE INDEX IF NOT EXISTS idx_cash_discounts_invoice
  ON cash_discounts(invoice_id);

-- 按租户 + 时间查询（租户隔离 + 报表用）
CREATE INDEX IF NOT EXISTS idx_cash_discounts_tenant_created
  ON cash_discounts(tenant_id, created_at DESC);

COMMENT ON TABLE cash_discounts IS
  '现金折扣痕迹记录（采购 V1.0 §3.7）：discount > 0 时由 AllocateToPaymentEntry 自动写入';
COMMENT ON COLUMN cash_discounts.payment_allocation_id IS
  '关联的 payment_allocations.id';
COMMENT ON COLUMN cash_discounts.payment_entry_id IS
  '冗余字段：关联的 payment_entry_id（方便按付款单查询）';
COMMENT ON COLUMN cash_discounts.invoice_id IS
  '冗余字段：关联的发票 ID（方便按发票查询）';
COMMENT ON COLUMN cash_discounts.discount_amount IS
  '现金折扣金额（NUMERIC(18,2)）';
COMMENT ON COLUMN cash_discounts.discount_rate IS
  '折扣率（NUMERIC(8,6)，如 0.020000 = 2%），可空';
COMMENT ON COLUMN cash_discounts.tenant_id IS
  '租户 ID（多租户隔离）';
