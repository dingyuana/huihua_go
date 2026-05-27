-- 004_bank.sql: Bank accounts, transactions, and reconciliation
-- Depends on: 001_init.sql (accounts), 002_journal_gl.sql (gl_entries), 003_invoice_payment.sql (payment_entries)

-- Bank Accounts
CREATE TABLE IF NOT EXISTS bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_name VARCHAR(100) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    clearing_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    currency VARCHAR(3) DEFAULT 'CNY',
    iban VARCHAR(50),
    swift_code VARCHAR(20),
    bank_account_type VARCHAR(20),                -- savings/current
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Bank Transactions (imported bank statements)
CREATE TABLE IF NOT EXISTS bank_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT,
    txn_date DATE NOT NULL,
    description TEXT,
    debit DECIMAL(18,2) DEFAULT 0,
    credit DECIMAL(18,2) DEFAULT 0,
    direction VARCHAR(4),                         -- in/out
    reference_no VARCHAR(100),
    counterparty_name VARCHAR(100),
    matched BOOLEAN DEFAULT FALSE,
    matched_payment_entry_id UUID,
    matched_gl_entry_id UUID,
    imported_from VARCHAR(20),                    -- csv/excel/camt053/mt940
    raw_data JSONB,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Bank Reconciliation Details (matching bank txns with payment/gl entries)
CREATE TABLE IF NOT EXISTS bank_reconciliation_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_transaction_id UUID NOT NULL REFERENCES bank_transactions(id) ON DELETE RESTRICT,
    payment_entry_id UUID REFERENCES payment_entries(id) ON DELETE RESTRICT,
    gl_entry_id UUID REFERENCES gl_entries(id) ON DELETE RESTRICT,
    match_score DECIMAL(5,2),
    difference_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    reconciled_at TIMESTAMP WITH TIME ZONE,
    reconciled_by UUID,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT
);

-- Bank Reconciliation Statements (periodic reconciliation summary)
CREATE TABLE IF NOT EXISTS bank_reconciliation_statements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE RESTRICT,
    statement_date DATE NOT NULL,
    bank_statement_balance DECIMAL(18,2) NOT NULL,
    gl_balance DECIMAL(18,2) NOT NULL,
    difference DECIMAL(18,2) DEFAULT 0,
    bank_only_total DECIMAL(18,2) DEFAULT 0,
    gl_only_total DECIMAL(18,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft',
    locked BOOLEAN DEFAULT FALSE,
    locked_by UUID,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE bank_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE bank_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE bank_reconciliation_details ENABLE ROW LEVEL SECURITY;
ALTER TABLE bank_reconciliation_statements ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON bank_accounts
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON bank_transactions
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON bank_reconciliation_details
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON bank_reconciliation_statements
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_bank_accounts_tenant ON bank_accounts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bank_accounts_company ON bank_accounts(company_id);

CREATE INDEX IF NOT EXISTS idx_bank_txn_tenant ON bank_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bank_txn_bank_account ON bank_transactions(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_bank_txn_date ON bank_transactions(txn_date);
CREATE INDEX IF NOT EXISTS idx_bank_txn_matched ON bank_transactions(matched);
CREATE INDEX IF NOT EXISTS idx_bank_txn_reference ON bank_transactions(reference_no);

CREATE INDEX IF NOT EXISTS idx_bank_recon_details_tenant ON bank_reconciliation_details(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bank_recon_details_txn ON bank_reconciliation_details(bank_transaction_id);
CREATE INDEX IF NOT EXISTS idx_bank_recon_details_payment ON bank_reconciliation_details(payment_entry_id);

CREATE INDEX IF NOT EXISTS idx_bank_recon_stmt_tenant ON bank_reconciliation_statements(tenant_id);
CREATE INDEX IF NOT EXISTS idx_bank_recon_stmt_bank_account ON bank_reconciliation_statements(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_bank_recon_stmt_date ON bank_reconciliation_statements(statement_date);
