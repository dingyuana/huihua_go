CREATE TABLE IF NOT EXISTS bus_reimbursements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reimbursement_no VARCHAR(50) NOT NULL,
    employee_name VARCHAR(200) NOT NULL,
    department VARCHAR(100),
    expense_type VARCHAR(50) NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    posting_date DATE NOT NULL,
    description TEXT,
    bank_account VARCHAR(100),
    docstatus SMALLINT DEFAULT 0,
    voucher_id UUID,
    voucher_no VARCHAR(50),
    created_by UUID,
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_reimbursement_no ON bus_reimbursements(tenant_id, reimbursement_no);
ALTER TABLE bus_reimbursements ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON bus_reimbursements USING (tenant_id = current_setting('app.tenant_id')::uuid);