-- Migration: 020_voucher_template_approval_bind
-- Bind voucher templates to approval flows + move thresholds from hardcode to DB

BEGIN;

-- 1. Add approval_flow_id to voucher_templates
ALTER TABLE voucher_templates ADD COLUMN IF NOT EXISTS approval_flow_id UUID REFERENCES approval_flows(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_voucher_templates_approval_flow ON voucher_templates(approval_flow_id) WHERE approval_flow_id IS NOT NULL;

-- 2. Add threshold config to approval_flows (replaces hardcoded ThresholdLevel2/Threshold3 in approval_service.go)
ALTER TABLE approval_flows ADD COLUMN IF NOT EXISTS threshold_amount_level2 DECIMAL(20, 4) DEFAULT 1000000;  -- 100万
ALTER TABLE approval_flows ADD COLUMN IF NOT EXISTS threshold_amount_level3 DECIMAL(20, 4) DEFAULT 5000000;  -- 500万
ALTER TABLE approval_flows ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'CNY';

-- 3. Update existing default flow with threshold values
UPDATE approval_flows SET threshold_amount_level2 = 1000000, threshold_amount_level3 = 5000000, currency = 'CNY' WHERE id IN (SELECT id FROM approval_flows LIMIT 1);

-- 4. Grant permissions
GRANT SELECT, INSERT, UPDATE ON voucher_templates TO huihua_app;

COMMIT;