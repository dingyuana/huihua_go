-- 013_voucher_state_machine.sql: Voucher state transitions table
-- Depends on: 002_journal_gl.sql (journal_entries), 007_audit.sql (audit_logs)

-- Voucher State Transitions: records all status changes for audit and tracing
CREATE TABLE IF NOT EXISTS voucher_state_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    from_status VARCHAR(20) NOT NULL,                -- draft/posted/verified/cancelled
    to_status VARCHAR(20) NOT NULL,
    action VARCHAR(20) NOT NULL,                     -- submit/approve/reject/reverse/cancel
    changed_by UUID NOT NULL,
    changed_by_name VARCHAR(100),
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE voucher_state_transitions ENABLE ROW LEVEL SECURITY;

-- RLS Policy: tenant isolation
CREATE POLICY tenant_isolation ON voucher_state_transitions
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_voucher_transitions_journal ON voucher_state_transitions(journal_id);
CREATE INDEX IF NOT EXISTS idx_voucher_transitions_tenant ON voucher_state_transitions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_voucher_transitions_created ON voucher_state_transitions(created_at);

-- Add original_voucher_id column to journal_entries if not exists
-- This is for reversal vouchers to reference the original voucher
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'journal_entries' AND column_name = 'original_voucher_id') THEN
        ALTER TABLE journal_entries ADD COLUMN original_voucher_id UUID REFERENCES journal_entries(id) ON DELETE RESTRICT;
    END IF;
END $$;