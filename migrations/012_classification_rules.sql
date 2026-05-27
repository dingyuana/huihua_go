-- 012_classification_rules.sql: Bank transaction classification rules
-- Depends on: 001_init.sql (accounts)

CREATE TABLE IF NOT EXISTS classification_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    rule_name VARCHAR(100) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    keywords JSONB NOT NULL, -- string array for OR matching
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    party_type VARCHAR(20),
    debit_direction VARCHAR(10) DEFAULT 'both', -- debit/credit/both
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