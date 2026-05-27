-- 003_invoice_payment.sql: Sales invoices, payment entries, and payment allocations
-- Depends on: 001_init.sql (accounts), 002_journal_gl.sql (gl_entries)

-- Sales Invoices
CREATE TABLE IF NOT EXISTS sales_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_no VARCHAR(50) UNIQUE NOT NULL,
    invoice_type VARCHAR(20) DEFAULT 'sale',      -- sale/purchase/credit_note
    customer_id UUID NOT NULL,
    tax_id VARCHAR(20),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    posting_date DATE NOT NULL,
    due_date DATE,
    total_amount DECIMAL(18,2) NOT NULL,          -- tax-inclusive total
    tax_amount DECIMAL(18,2) DEFAULT 0,
    net_amount DECIMAL(18,2) DEFAULT 0,           -- tax-exclusive amount
    outstanding_amount DECIMAL(18,2) NOT NULL,    -- unpaid balance
    status VARCHAR(20) DEFAULT 'unpaid',          -- unpaid/partially_paid/paid/credit_note/written_off
    tax_template_id UUID,
    return_against UUID,
    is_return BOOLEAN DEFAULT FALSE,
    docstatus SMALLINT DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Payment Entries
CREATE TABLE IF NOT EXISTS payment_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_no VARCHAR(50) UNIQUE NOT NULL,
    payment_type VARCHAR(10) NOT NULL,            -- receive/pay
    party_type VARCHAR(20) NOT NULL,
    party_id UUID NOT NULL,
    paid_from_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    paid_to_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    paid_amount DECIMAL(18,2) NOT NULL,
    received_amount DECIMAL(18,2),
    reference_no VARCHAR(50),
    reference_date DATE,
    posting_date DATE NOT NULL,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    bank_account_id UUID,
    docstatus SMALLINT DEFAULT 0,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Payment Allocations (linking payments to invoices)
CREATE TABLE IF NOT EXISTS payment_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_entry_id UUID NOT NULL REFERENCES payment_entries(id) ON DELETE RESTRICT,
    invoice_id UUID NOT NULL,
    invoice_type VARCHAR(30),
    allocated_amount DECIMAL(18,2) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE sales_invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE payment_entries ENABLE ROW LEVEL SECURITY;
ALTER TABLE payment_allocations ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON sales_invoices
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON payment_entries
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON payment_allocations
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sales_invoices_tenant ON sales_invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_customer ON sales_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_posting_date ON sales_invoices(posting_date);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_status ON sales_invoices(status);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_invoice_no ON sales_invoices(invoice_no);

CREATE INDEX IF NOT EXISTS idx_payment_entries_tenant ON payment_entries(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_entries_party ON payment_entries(party_type, party_id);
CREATE INDEX IF NOT EXISTS idx_payment_entries_posting_date ON payment_entries(posting_date);
CREATE INDEX IF NOT EXISTS idx_payment_entries_payment_no ON payment_entries(payment_no);
CREATE INDEX IF NOT EXISTS idx_payment_entries_bank_account ON payment_entries(bank_account_id);

CREATE INDEX IF NOT EXISTS idx_payment_allocations_tenant ON payment_allocations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payment_allocations_payment ON payment_allocations(payment_entry_id);
CREATE INDEX IF NOT EXISTS idx_payment_allocations_invoice ON payment_allocations(invoice_id);
