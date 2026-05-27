-- 011_depreciation_run.sql: Depreciation run tracking table
-- Depends on: 005_asset.sql, 002_journal_gl.sql

-- Depreciation Runs: tracks monthly depreciation execution batches
CREATE TABLE IF NOT EXISTS depreciation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_no INT NOT NULL,
    run_date TIMESTAMP WITH TIME ZONE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    company_id UUID NOT NULL,
    voucher_no VARCHAR(50) NOT NULL,
    voucher_type VARCHAR(30),
    total_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
    asset_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE depreciation_runs ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON depreciation_runs
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_depreciation_runs_tenant ON depreciation_runs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_depreciation_runs_period ON depreciation_runs(period_no);
CREATE INDEX IF NOT EXISTS idx_depreciation_runs_status ON depreciation_runs(status);