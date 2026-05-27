-- 012_voucher_template.sql: Voucher template and numbering rule
-- Depends on: 001_init.sql (tenants, accounts)

CREATE TABLE IF NOT EXISTS voucher_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    number_prefix VARCHAR(20) NOT NULL DEFAULT 'PZ',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- Enable RLS
ALTER TABLE voucher_templates ENABLE ROW LEVEL SECURITY;

-- RLS Policy: tenant isolation
CREATE POLICY tenant_isolation ON voucher_templates
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_voucher_templates_tenant ON voucher_templates(tenant_id);
CREATE INDEX IF NOT EXISTS idx_voucher_templates_active ON voucher_templates(is_active) WHERE is_active = TRUE;

-- Voucher template lines (分录模板)
CREATE TABLE IF NOT EXISTS voucher_template_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES voucher_templates(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    dr_amount_template VARCHAR(100), -- e.g., "{{amount}}" or empty if credit
    cr_amount_template VARCHAR(100),
    summary_template VARCHAR(200), -- e.g., "支付{{party}}"
    dimension_type VARCHAR(50),   -- e.g., "department", "project", "cost_center"
    dimension_value VARCHAR(100),  -- e.g., "{{department_id}}"
    line_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE voucher_template_lines ENABLE ROW LEVEL SECURITY;

-- RLS Policy: tenant isolation via template
CREATE POLICY tenant_isolation ON voucher_template_lines
    USING (
        template_id IN (
            SELECT id FROM voucher_templates 
            WHERE tenant_id = current_setting('app.current_tenant')::uuid
        )
    );

-- Indexes
CREATE INDEX IF NOT EXISTS idx_voucher_template_lines_template ON voucher_template_lines(template_id);
CREATE INDEX IF NOT EXISTS idx_voucher_template_lines_account ON voucher_template_lines(account_id);

-- Voucher numbering rules (编号规则)
CREATE TABLE IF NOT EXISTS voucher_numbering_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    prefix VARCHAR(20) NOT NULL DEFAULT 'PZ',
    next_number INTEGER NOT NULL DEFAULT 1,
    date_format VARCHAR(20) NOT NULL DEFAULT '20060102',
    reset_rule VARCHAR(20) NOT NULL DEFAULT 'daily', -- yearly/monthly/daily
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id)
);

-- Enable RLS
ALTER TABLE voucher_numbering_rules ENABLE ROW LEVEL SECURITY;

-- RLS Policy: tenant isolation
CREATE POLICY tenant_isolation ON voucher_numbering_rules
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_voucher_numbering_rules_tenant ON voucher_numbering_rules(tenant_id);
