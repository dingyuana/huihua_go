-- 014_bank_transactions.sql: Add unique constraint for duplicate detection during import
-- Depends on: 004_bank.sql (bank_transactions table exists)

-- Add unique constraint for duplicate detection during import
-- This allows ON CONFLICT DO NOTHING to work correctly
CREATE UNIQUE INDEX IF NOT EXISTS idx_bank_txn_unique 
ON bank_transactions(tenant_id, bank_account_id, txn_date, description, debit, credit);

-- Add reconciled columns if they don't exist (for reconciliation status tracking)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'reconciled') THEN
        ALTER TABLE bank_transactions ADD COLUMN reconciled BOOLEAN DEFAULT FALSE;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'reconciled_period') THEN
        ALTER TABLE bank_transactions ADD COLUMN reconciled_period INTEGER;
    END IF;
END $$;

-- Add updated_at column if it doesn't exist (for tracking last modification)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'updated_at') THEN
        ALTER TABLE bank_transactions ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    END IF;
END $$;