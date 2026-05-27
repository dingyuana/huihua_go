-- 015_opening_balance.sql: 期初余额表
-- Depends on: 010_account_setup.sql (accounts table)

-- ============================================================
-- Opening Balances Table
-- ============================================================
CREATE TABLE IF NOT EXISTS opening_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    company_id UUID NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    debit_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
    credit_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
    period_no INT NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, company_id, account_id, period_no)
);

-- ============================================================
-- RLS: Enable + Policy + Force
-- ============================================================
ALTER TABLE opening_balances ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON opening_balances
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

ALTER TABLE opening_balances FORCE ROW LEVEL SECURITY;

-- ============================================================
-- Indexes
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_opening_balances_tenant ON opening_balances(tenant_id);
CREATE INDEX IF NOT EXISTS idx_opening_balances_company ON opening_balances(company_id);
CREATE INDEX IF NOT EXISTS idx_opening_balances_account ON opening_balances(account_id);
CREATE INDEX IF NOT EXISTS idx_opening_balances_period ON opening_balances(period_no);
CREATE INDEX IF NOT EXISTS idx_opening_balances_verified ON opening_balances(is_verified);
CREATE INDEX IF NOT EXISTS idx_opening_balances_tenant_period ON opening_balances(tenant_id, period_no);