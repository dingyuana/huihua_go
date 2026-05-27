-- 006_budget.sql: Budgets, budget accounts, distributions, and control configs
-- Depends on: 001_init.sql (accounts)

-- Budgets
CREATE TABLE IF NOT EXISTS budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    budget_against VARCHAR(20) NOT NULL,          -- cost_center/project
    fiscal_year VARCHAR(10) NOT NULL,
    monthly_distribution VARCHAR(20),             -- monthly/quarterly/half_yearly/yearly
    status VARCHAR(20) DEFAULT 'draft',
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(company_id, budget_against, fiscal_year, tenant_id)
);

-- Budget Accounts (budget line items per account)
CREATE TABLE IF NOT EXISTS budget_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id UUID NOT NULL REFERENCES budgets(id) ON DELETE RESTRICT,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    annual_budget DECIMAL(18,2) NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT
);

-- Budget Distributions (period-based budget allocation)
CREATE TABLE IF NOT EXISTS budget_distributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_account_id UUID NOT NULL REFERENCES budget_accounts(id) ON DELETE RESTRICT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    percent DECIMAL(5,2),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT
);

-- Budget Control Configs (rules for budget enforcement)
CREATE TABLE IF NOT EXISTS budget_control_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    cost_center_id UUID,
    fiscal_year VARCHAR(10),
    action_annual VARCHAR(20) DEFAULT 'warn',     -- warn/stop/ignore
    action_monthly VARCHAR(20) DEFAULT 'warn',
    applicable_on_mr BOOLEAN DEFAULT TRUE,        -- material request
    applicable_on_po BOOLEAN DEFAULT TRUE,        -- purchase order
    applicable_on_actual BOOLEAN DEFAULT TRUE,    -- actual posting
    exception_approver_role VARCHAR(50),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT
);

-- Enable RLS
ALTER TABLE budgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_distributions ENABLE ROW LEVEL SECURITY;
ALTER TABLE budget_control_configs ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON budgets
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON budget_accounts
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON budget_distributions
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON budget_control_configs
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_budgets_tenant ON budgets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_budgets_company ON budgets(company_id);
CREATE INDEX IF NOT EXISTS idx_budgets_fiscal_year ON budgets(fiscal_year);

CREATE INDEX IF NOT EXISTS idx_budget_accounts_tenant ON budget_accounts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_budget_accounts_budget ON budget_accounts(budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_accounts_account ON budget_accounts(account_id);

CREATE INDEX IF NOT EXISTS idx_budget_distributions_tenant ON budget_distributions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_budget_distributions_budget_account ON budget_distributions(budget_account_id);
CREATE INDEX IF NOT EXISTS idx_budget_distributions_dates ON budget_distributions(start_date, end_date);

CREATE INDEX IF NOT EXISTS idx_budget_control_tenant ON budget_control_configs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_budget_control_account ON budget_control_configs(account_id);
CREATE INDEX IF NOT EXISTS idx_budget_control_company ON budget_control_configs(company_id);
