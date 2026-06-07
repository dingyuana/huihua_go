-- Migration 057: payment_allocations 新增 discount_amount 字段
--
-- 业务背景:
--   采购 V1.0 §3.7 现金折扣场景：
--   供应商付款时，如在折扣期内付款，可享受现金折扣。
--   例：发票 1000 元，2/10 n/30 → 10 天内付款可扣 2%（20 元），实付 980 元。
--   discount_amount 记录实际冲减的金额。
--
-- 字段语义:
--   discount_amount NUMERIC(18,2) DEFAULT 0
--     - 默认 0（无折扣）
--     - 与 allocated_amount 共同决定实际付款：payment = allocated_amount - discount_amount
--     - 不影响 invoice.outstanding_amount（折扣是付款方的损益，非发票金额调整）
--     - 影响利润表（财务费用或采购成本节约），后续会在 payment service 中调用

ALTER TABLE payment_allocations
  ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(18,2) DEFAULT 0;

COMMENT ON COLUMN payment_allocations.discount_amount IS
  '现金折扣金额（采购 V1.0 §3.7）：折扣期内付款可享受的现金折扣';
