CREATE TABLE IF NOT EXISTS bank_journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id UUID NOT NULL,
    txn_date DATE NOT NULL,
    description TEXT,
    debit DECIMAL(18,2) DEFAULT 0,
    credit DECIMAL(18,2) DEFAULT 0,
    voucher_id UUID,
    voucher_no VARCHAR(50),
    tenant_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_bank_journal_account_date ON bank_journal_entries(tenant_id, bank_account_id, txn_date);
ALTER TABLE bank_journal_entries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON bank_journal_entries USING (tenant_id = current_setting('app.tenant_id')::uuid);