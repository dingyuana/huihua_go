-- 024_bank_txns_classification.sql: Add classification column to bank_transactions
-- Depends on: 004_bank.sql (bank_transactions table exists)
-- Uses: classification values from 023_classification_rules_v2.sql

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'bank_transactions' AND column_name = 'classification'
    ) THEN
        ALTER TABLE bank_transactions
        ADD COLUMN classification VARCHAR(50) DEFAULT 'pending';
    END IF;
END $$;

-- Index for filtering/sorting by classification
CREATE INDEX IF NOT EXISTS idx_bank_txns_classification
    ON bank_transactions(classification);
