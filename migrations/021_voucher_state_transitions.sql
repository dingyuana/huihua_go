-- 021_voucher_state_transitions.sql: Add voucher_state_transitions table if not exists
-- Depends on: 002_journal_gl.sql, 007_audit.sql

-- Voucher State Transitions: records all status changes for audit and tracing
-- This migration ensures the table exists with the correct schema
CREATE TABLE IF NOT EXISTS voucher_state_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    voucher_id UUID NOT NULL,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    triggered_by UUID,
    comments TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE voucher_state_transitions ENABLE ROW LEVEL SECURITY;

-- RLS Policy: tenant isolation
CREATE POLICY IF NOT EXISTS tenant_isolation ON voucher_state_transitions
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_voucher_transitions_voucher ON voucher_state_transitions(voucher_id);
CREATE INDEX IF NOT EXISTS idx_voucher_transitions_tenant_t ON voucher_state_transitions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_voucher_transitions_created_at ON voucher_state_transitions(created_at);