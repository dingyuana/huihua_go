-- 062_bank_journal_entries.sql: 银行日记账表
-- 修复缺陷 #6: 凭证 approve 时 bank_journal_entries 表缺失
-- 用于记录每张银行类凭证的借贷明细 (跟银行账户绑定)
-- Depends on: 002_journal_gl.sql (journal_entries), 004_bank.sql (bank_accounts)

CREATE TABLE IF NOT EXISTS bank_journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT,
    txn_date DATE NOT NULL,
    description TEXT,
    debit DECIMAL(18,2) NOT NULL DEFAULT 0,
    credit DECIMAL(18,2) NOT NULL DEFAULT 0,
    voucher_id UUID REFERENCES journal_entries(id) ON DELETE RESTRICT,
    voucher_no VARCHAR(50),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_journal_entries_account ON bank_journal_entries(tenant_id, bank_account_id, txn_date DESC);
CREATE INDEX IF NOT EXISTS idx_bank_journal_entries_voucher ON bank_journal_entries(voucher_id);
CREATE INDEX IF NOT EXISTS idx_bank_journal_entries_tenant_date ON bank_journal_entries(tenant_id, txn_date DESC);

ALTER TABLE bank_journal_entries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON bank_journal_entries
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

COMMENT ON TABLE bank_journal_entries IS '银行日记账: 银行类凭证的银行账户维度明细';
