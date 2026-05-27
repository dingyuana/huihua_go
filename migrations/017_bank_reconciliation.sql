-- 017_bank_reconciliation.sql: Bank reconciliation records table
-- Depends on: 004_bank.sql (bank_transactions, bank_reconciliation_statements exist)

-- Reconciliation Records table (one record per bank account per period)
CREATE TABLE IF NOT EXISTS reconciliation_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT,
    period_no INTEGER NOT NULL,                              -- YYYYMM format, e.g., 202506
    bank_balance DECIMAL(18,2) NOT NULL DEFAULT 0,           -- Bank statement balance
    book_balance DECIMAL(18,2) NOT NULL DEFAULT 0,          -- GL/book balance
    adjusted_balance DECIMAL(18,2) NOT NULL DEFAULT 0,       -- After adjusting for unreconciled items
    bank_only_total DECIMAL(18,2) NOT NULL DEFAULT 0,       -- Total of items in bank but not in books
    book_only_total DECIMAL(18,2) NOT NULL DEFAULT 0,       -- Total of items in books but not in bank
    status VARCHAR(20) DEFAULT 'draft',                     -- draft/in_progress/reconciled
    reconciled_by UUID,
    reconciled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (tenant_id, bank_account_id, period_no)
);

-- Enable RLS
ALTER TABLE reconciliation_records ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON reconciliation_records
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_recon_records_tenant ON reconciliation_records(tenant_id);
CREATE INDEX IF NOT EXISTS idx_recon_records_bank_account ON reconciliation_records(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_recon_records_period ON reconciliation_records(period_no);
CREATE INDEX IF NOT EXISTS idx_recon_records_status ON reconciliation_records(status);

-- Unreconciled items table (for tracking unreconciled transactions during reconciliation)
CREATE TABLE IF NOT EXISTS unreconciled_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconciliation_record_id UUID NOT NULL REFERENCES reconciliation_records(id) ON DELETE CASCADE,
    item_type VARCHAR(20) NOT NULL,                          -- bank_only/book_only
    source_type VARCHAR(20) NOT NULL,                       -- bank_transaction/journal_entry
    source_id UUID NOT NULL,
    txn_date DATE NOT NULL,
    description TEXT,
    debit DECIMAL(18,2) DEFAULT 0,
    credit DECIMAL(18,2) DEFAULT 0,
    amount DECIMAL(18,2) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE unreconciled_items ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON unreconciled_items
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_unrecon_items_recon_record ON unreconciled_items(reconciliation_record_id);
CREATE INDEX IF NOT EXISTS idx_unrecon_items_item_type ON unreconciled_items(item_type);
CREATE INDEX IF NOT EXISTS idx_unrecon_items_tenant ON unreconciled_items(tenant_id);