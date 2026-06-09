-- 060_allocation_reversal_flag.sql: 核销记录增加反核销标记
-- 允许标记 payment_allocations 和 advance_allocations 中的记录已被反核销

ALTER TABLE payment_allocations
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE advance_allocations
    ADD COLUMN IF NOT EXISTS reversed_at TIMESTAMP WITH TIME ZONE;

COMMENT ON COLUMN payment_allocations.reversed_at IS '反核销时间，非空表示该核销记录已被撤销';
COMMENT ON COLUMN advance_allocations.reversed_at IS '反核销时间，非空表示该核销记录已被撤销';
