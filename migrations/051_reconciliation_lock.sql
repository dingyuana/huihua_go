-- 051_reconciliation_lock.sql: Add lock fields to reconciliation_records
-- Depends on: 017_bank_reconciliation.sql

BEGIN;

-- Add lock fields to reconciliation_records
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS locked BOOLEAN DEFAULT FALSE;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS locked_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS locked_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS unlock_approved_by UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE reconciliation_records ADD COLUMN IF NOT EXISTS unlock_approved_at TIMESTAMP WITH TIME ZONE;

-- Add index for locked status lookup
CREATE INDEX IF NOT EXISTS idx_recon_records_locked ON reconciliation_records(tenant_id, bank_account_id, locked);

COMMENT ON TABLE reconciliation_records IS 'Lock fields allow exclusive access during reconciliation to prevent concurrent modifications';

COMMIT;