-- Migration 068: Add unallocated_amount, currency, and exchange_rate fields to payment_entries

ALTER TABLE payment_entries
  ADD COLUMN unallocated_amount NUMERIC(18,2) NOT NULL DEFAULT 0;

ALTER TABLE payment_entries
  ADD COLUMN currency VARCHAR(3) DEFAULT 'CNY';

ALTER TABLE payment_entries
  ADD COLUMN exchange_rate NUMERIC(10,6) DEFAULT 1.0;

-- Backfill existing records: set unallocated_amount = paid_amount - allocated_amount
UPDATE payment_entries pe
  SET unallocated_amount = pe.paid_amount - COALESCE((
        SELECT SUM(pa.allocated_amount)
          FROM payment_allocations pa
         WHERE pa.payment_entry_id = pe.id
      ), 0);

COMMENT ON COLUMN payment_entries.unallocated_amount IS
  '收款单/付款单未核销金额（独立字段，区别于 paid_amount 已核销）';

COMMENT ON COLUMN payment_entries.currency IS
  '币种代码（ISO 4217，默认CNY）';

COMMENT ON COLUMN payment_entries.exchange_rate IS
  '汇率（默认1.0）';
