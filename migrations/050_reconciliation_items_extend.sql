-- 050_reconciliation_items_extend.sql: Extend item_type to support four-direction reconciliation
-- Depends on: 017_bank_reconciliation.sql

BEGIN;

-- Drop existing check constraint if exists (for idempotent migration)
ALTER TABLE unreconciled_items DROP CONSTRAINT IF EXISTS unreconciled_items_item_type_check;

-- Add new check constraint with four-direction types + legacy types
ALTER TABLE unreconciled_items ADD CONSTRAINT unreconciled_items_item_type_check
    CHECK (item_type IN (
        'bank_receipt_not_in_gl',   -- 银行已收企业未达 (bank credit, book missing)
        'bank_payment_not_in_gl',   -- 银行已付企业未达 (bank debit, book missing)
        'gl_receipt_not_in_bank',   -- 企业已收银行未达 (book credit, bank missing)
        'gl_payment_not_in_bank',    -- 企业已付银行未达 (book debit, bank missing)
        'bank_only',                 -- legacy: bank has, book doesn't
        'book_only'                  -- legacy: book has, bank doesn't
    ));

-- Add direction column for finer-grained classification (optional, for future use)
ALTER TABLE unreconciled_items ADD COLUMN IF NOT EXISTS direction VARCHAR(10);

-- Update existing bank_only items to map to correct four-type based on debit/credit
UPDATE unreconciled_items
SET direction = CASE WHEN debit > credit THEN 'debit' ELSE 'credit' END
WHERE item_type IN ('bank_only', 'book_only') AND direction IS NULL;

-- Add item_sub_type for additional classification (optional)
ALTER TABLE unreconciled_items ADD COLUMN IF NOT EXISTS item_sub_type VARCHAR(20);

COMMENT ON COLUMN unreconciled_items.item_type IS 'Four types: bank_receipt_not_in_gl, bank_payment_not_in_gl, gl_receipt_not_in_bank, gl_payment_not_in_bank';

COMMIT;