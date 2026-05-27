-- 002_journal_gl.sql: Journal entries, journal entry lines, and GL entries
-- Depends on: 001_init.sql (accounts table)

-- Journal Entries (voucher master)
CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    voucher_no VARCHAR(50) UNIQUE NOT NULL,       -- format: JE-YYYY-MM-NNNN
    voucher_type VARCHAR(30),                     -- general/bank/cash/transfer/depreciation/closing
    posting_date DATE NOT NULL,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    remark TEXT,
    docstatus SMALLINT DEFAULT 0,                 -- 0=draft, 1=submitted, 2=cancelled
    reversed_id UUID,                             -- which voucher reversed this one
    reversal_id UUID,                             -- which voucher this one reversed
    submitted_by UUID,
    submitted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Journal Entry Lines
CREATE TABLE IF NOT EXISTS journal_entry_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE RESTRICT,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    debit DECIMAL(18,2) DEFAULT 0 CHECK (debit >= 0),
    credit DECIMAL(18,2) DEFAULT 0 CHECK (credit >= 0),
    debit_ccy DECIMAL(18,2) DEFAULT 0,
    credit_ccy DECIMAL(18,2) DEFAULT 0,
    account_ccy VARCHAR(3),
    exchange_rate DECIMAL(18,6) DEFAULT 1.0,
    party_type VARCHAR(20),                       -- customer/supplier/employee
    party_id UUID,
    cost_center_id UUID,
    project_id UUID,
    user_remark VARCHAR(200),
    reconciled BOOLEAN DEFAULT FALSE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    CHECK (debit > 0 OR credit > 0)               -- at least one side must be non-zero
);

-- GL Entries (general ledger, written by service layer on journal entry submit)
CREATE TABLE IF NOT EXISTS gl_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    posting_date DATE NOT NULL,
    debit DECIMAL(18,2) DEFAULT 0,
    credit DECIMAL(18,2) DEFAULT 0,
    debit_ccy DECIMAL(18,2) DEFAULT 0,
    credit_ccy DECIMAL(18,2) DEFAULT 0,
    account_ccy VARCHAR(3),
    voucher_type VARCHAR(30),                     -- journal_entry/invoice/payment
    voucher_id UUID,                              -- source document ID
    against_voucher_type VARCHAR(30),
    against_voucher_id UUID,
    party_type VARCHAR(20),
    party_id UUID,
    cost_center_id UUID,
    project_id UUID,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    is_cancelled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE journal_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE journal_entry_lines ENABLE ROW LEVEL SECURITY;
ALTER TABLE gl_entries ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON journal_entries
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON journal_entry_lines
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON gl_entries
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_journal_entries_tenant ON journal_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_journal_entries_posting_date ON journal_entries(posting_date);
CREATE INDEX IF NOT EXISTS idx_journal_entries_voucher_no ON journal_entries(voucher_no);
CREATE INDEX IF NOT EXISTS idx_journal_entries_docstatus ON journal_entries(docstatus);

CREATE INDEX IF NOT EXISTS idx_je_lines_journal_entry ON journal_entry_lines(journal_entry_id);
CREATE INDEX IF NOT EXISTS idx_je_lines_account ON journal_entry_lines(account_id);
CREATE INDEX IF NOT EXISTS idx_je_lines_tenant ON journal_entry_lines(tenant_id);
CREATE INDEX IF NOT EXISTS idx_je_lines_party ON journal_entry_lines(party_type, party_id);

CREATE INDEX IF NOT EXISTS idx_gl_voucher ON gl_entries(voucher_type, voucher_id);
CREATE INDEX IF NOT EXISTS idx_gl_posting ON gl_entries(posting_date, account_id);
CREATE INDEX IF NOT EXISTS idx_gl_tenant ON gl_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_gl_account ON gl_entries(account_id);
CREATE INDEX IF NOT EXISTS idx_gl_party ON gl_entries(party_type, party_id);
