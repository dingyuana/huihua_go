-- Migration: 019_approval
-- Description: Approval workflow tables (approval_flows, approval_tasks)

BEGIN;

-- Create approval_flows table
CREATE TABLE IF NOT EXISTS approval_flows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    flow_name VARCHAR(255) NOT NULL,
    description TEXT,
    approvers JSONB NOT NULL DEFAULT '[]',
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create approval_tasks table
CREATE TABLE IF NOT EXISTS approval_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flow_id UUID NOT NULL REFERENCES approval_flows(id) ON DELETE CASCADE,
    journal_entry_id UUID NOT NULL,
    approver_id UUID NOT NULL,
    approver_name VARCHAR(255),
    level INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    comment TEXT,
    amount DECIMAL(20, 4) NOT NULL DEFAULT 0,
    tenant_id UUID NOT NULL,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for approval_tasks
CREATE INDEX IF NOT EXISTS idx_approval_tasks_tenant ON approval_tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_approver ON approval_tasks(approver_id);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_journal ON approval_tasks(journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_status ON approval_tasks(status);
CREATE INDEX IF NOT EXISTS idx_approval_tasks_flow ON approval_tasks(flow_id);

-- Create indexes for approval_flows
CREATE INDEX IF NOT EXISTS idx_approval_flows_tenant ON approval_flows(tenant_id);

-- Add comments
COMMENT ON TABLE approval_flows IS 'Approval workflow definitions';
COMMENT ON TABLE approval_tasks IS 'Individual approval task instances';
COMMENT ON COLUMN approval_flows.approvers IS 'JSON array of approvers with level, approver_id, and role';
COMMENT ON COLUMN approval_tasks.status IS 'pending, approved, or rejected';

-- Enable Row Level Security
ALTER TABLE approval_flows ENABLE ROW LEVEL SECURITY;
ALTER TABLE approval_tasks ENABLE ROW LEVEL SECURITY;

-- Force RLS (must be enforced for all users)
ALTER TABLE approval_flows FORCE ROW LEVEL SECURITY;
ALTER TABLE approval_tasks FORCE ROW LEVEL SECURITY;

-- RLS Policies for approval_flows
DROP POLICY IF EXISTS "approval_flows_tenant_policy" ON approval_flows;
CREATE POLICY "approval_flows_tenant_policy" ON approval_flows
    FOR ALL USING (tenant_id = current_setting('app.tenant_id', true)::UUID);

-- RLS Policies for approval_tasks
DROP POLICY IF EXISTS "approval_tasks_tenant_policy" ON approval_tasks;
CREATE POLICY "approval_tasks_tenant_policy" ON approval_tasks
    FOR ALL USING (tenant_id = current_setting('app.tenant_id', true)::UUID);

-- Grant permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON approval_flows TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON approval_tasks TO app_user;

COMMIT;