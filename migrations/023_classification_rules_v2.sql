-- 023_classification_rules_v2.sql: Updated classification rules table based on API design
-- Depends on: 012_classification_rules.sql

-- Rename the old table
ALTER TABLE classification_rules RENAME TO classification_rules_old;

-- Create the new classification rules table
CREATE TABLE classification_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    name VARCHAR(200) NOT NULL,
    rule_type VARCHAR(50) NOT NULL DEFAULT 'keyword', -- keyword, keyword_regex, counterparty_match
    pattern TEXT NOT NULL,
    match_field VARCHAR(50) NOT NULL DEFAULT 'description', -- description, counterparty
    direction VARCHAR(20) DEFAULT '', -- in, out, '' (both)
    classification VARCHAR(50) NOT NULL, -- business_receipt, business_payment, bank_fee, interest_income, internal_transfer
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enable RLS
ALTER TABLE classification_rules ENABLE ROW LEVEL SECURITY;

-- RLS Policy: tenant isolation
CREATE POLICY tenant_isolation ON classification_rules
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_classification_rules_tenant ON classification_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_classification_rules_priority ON classification_rules(priority);
CREATE INDEX IF NOT EXISTS idx_classification_rules_active ON classification_rules(is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_classification_rules_classification ON classification_rules(classification);
