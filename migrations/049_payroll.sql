-- Migration: payroll_records table
-- 第一类输入单据：信息完整，approve时直接生成凭证

CREATE TABLE payroll_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    company_id UUID,
    payroll_no VARCHAR(20) NOT NULL,
    employee_name VARCHAR(100) NOT NULL,
    department_name VARCHAR(100),
    period_no INTEGER NOT NULL,
    gross_salary DECIMAL(18,2) NOT NULL DEFAULT 0,
    individual_tax DECIMAL(18,2) NOT NULL DEFAULT 0,
    social_security DECIMAL(18,2) NOT NULL DEFAULT 0,
    housing_fund DECIMAL(18,2) NOT NULL DEFAULT 0,
    other_deductions DECIMAL(18,2) NOT NULL DEFAULT 0,
    net_salary DECIMAL(18,2) NOT NULL DEFAULT 0,
    payment_date DATE,
    bank_account_no VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    docstatus INTEGER NOT NULL DEFAULT 0,
    voucher_id UUID,
    voucher_no VARCHAR(50),
    source VARCHAR(20) DEFAULT 'manual',
    remark TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_payroll_tenant ON payroll_records(tenant_id);
CREATE INDEX idx_payroll_period ON payroll_records(tenant_id, period_no);
CREATE INDEX idx_payroll_status ON payroll_records(tenant_id, docstatus);

-- RLS
ALTER TABLE payroll_records ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON payroll_records FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::uuid);