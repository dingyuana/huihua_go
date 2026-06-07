-- Migration 2026060805: payment_entries 加未核销金额独立字段 unallocated_amount
--
-- 业务背景:
--   V1.0 报告要求"未核销金额独立字段"（P1 拆分）。
--   现有 payment_entries.paid_amount 是金额快照，核销明细由 payment_allocations 维护。
--   缺少 payment 端"剩余可核销"独立字段，每次需 SUM(allocations) 推导，效率低且不便于索引/报表。
--   本次新增 unallocated_amount 字段，独立存储"该收款/付款单尚未核销的金额"。
--
-- 与现有字段关系（不破坏既有语义）:
--   paid_amount       — 收款/付款金额（不变）
--   unallocated_amount — 本字段，新增，等于 paid_amount - SUM(allocations.allocated_amount)（按需应用层维护）
--   outstanding_amount — ar_invoice 端余额字段（不动）
--
-- 数据回填（执行完 ALTER 后建议运行）:
--   UPDATE payment_entries pe
--      SET unallocated_amount = pe.paid_amount - COALESCE((
--            SELECT SUM(pa.allocated_amount)
--              FROM payment_allocations pa
--             WHERE pa.payment_entry_id = pe.id
--          ), 0);

ALTER TABLE payment_entries
  ADD COLUMN unallocated_amount NUMERIC(18,2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN payment_entries.unallocated_amount IS
  '收款单/付款单未核销金额（独立字段，区别于 paid_amount 已核销）';

-- 回滚（如需撤销）
-- ALTER TABLE payment_entries DROP COLUMN IF EXISTS unallocated_amount;
