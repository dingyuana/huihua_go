-- 007_audit.sql: Audit logs (append-only, no UPDATE/DELETE)
-- Depends on: 001_init.sql (tenants)

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(50) NOT NULL,                  -- create/update/delete/submit/cancel/reverse
    object_type VARCHAR(50) NOT NULL,             -- journal_entry/payment_entry/invoice
    object_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    actor_id UUID NOT NULL,                       -- operator user ID
    actor_name VARCHAR(100),
    changed_fields JSONB,                         -- changed fields: {field_name: [old_value, new_value]}
    metadata JSONB,                               -- extra info (e.g. IP address)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- RLS Policy: tenant isolation
CREATE POLICY tenant_isolation ON audit_logs
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_audit_object ON audit_logs(object_type, object_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs(actor_id, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_tenant ON audit_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at);

-- Revoke UPDATE and DELETE on audit_logs to enforce append-only
-- Only INSERT and SELECT are allowed
REVOKE UPDATE, DELETE ON audit_logs FROM PUBLIC;
