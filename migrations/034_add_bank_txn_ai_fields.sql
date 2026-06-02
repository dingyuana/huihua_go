-- 034_add_bank_txn_ai_fields.sql: Add AI suggestion fields to bank_transactions
-- Part of: TASK-BANK-01 §3.1 AI Feedback layer
-- Depends on: 032_add_bank_txn_status.sql

-- AI suggestion fields added to bank_transactions per SPEC §3.3
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'ai_confidence') THEN
        ALTER TABLE bank_transactions ADD COLUMN ai_confidence int DEFAULT 0;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'ai_suggested_action') THEN
        ALTER TABLE bank_transactions ADD COLUMN ai_suggested_action varchar(50);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'bank_transactions' AND column_name = 'ai_business_scene') THEN
        ALTER TABLE bank_transactions ADD COLUMN ai_business_scene varchar(100);
    END IF;
END $$;

-- Index for AI field queries
CREATE INDEX IF NOT EXISTS idx_bank_txn_ai_fields ON bank_transactions(ai_suggested_action);