-- 032_add_bank_txn_status.sql: Add status column for bank transaction review workflow
-- Depends on: 014_bank_transactions.sql
-- Part of: TASK-BANK-01

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'status') THEN
        ALTER TABLE bank_transactions ADD COLUMN status VARCHAR(30) DEFAULT 'pending';
    END IF;
END $$;

-- Index for status queries in review workflow
CREATE INDEX IF NOT EXISTS idx_bank_txn_status ON bank_transactions(tenant_id, status);

-- Migrate existing data: matched=true -> voucher_generated, matched=false -> pending
UPDATE bank_transactions SET status = 'voucher_generated' WHERE matched = TRUE AND (status IS NULL OR status = 'pending');
UPDATE bank_transactions SET status = 'pending' WHERE matched = FALSE AND (status IS NULL OR status = 'pending');