-- 025_classification_rule_accounts.sql
-- Add per-rule account mapping to classification rules
-- This is the bridge from "classification string" to "GL account" for auto-voucher generation

BEGIN;

ALTER TABLE classification_rules
    ADD COLUMN IF NOT EXISTS debit_account_id  UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS credit_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_classification_rules_dr_acct ON classification_rules(debit_account_id) WHERE debit_account_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_classification_rules_cr_acct ON classification_rules(credit_account_id) WHERE credit_account_id IS NOT NULL;

-- Backfill default mappings for the 5 seed rules.
-- Uses subquery lookup to map by code within the same tenant.
DO $$
DECLARE
    r RECORD;
    dr_id UUID;
    cr_id UUID;
BEGIN
    FOR r IN
        SELECT cr.id, cr.tenant_id, cr.classification
        FROM classification_rules cr
        WHERE cr.debit_account_id IS NULL OR cr.credit_account_id IS NULL
    LOOP
        dr_id := NULL;
        cr_id := NULL;
        IF r.classification = 'bank_fee' THEN
            SELECT id INTO dr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '5602' LIMIT 1;
            SELECT id INTO cr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1002' LIMIT 1;
        ELSIF r.classification = 'interest_income' THEN
            SELECT id INTO dr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1002' LIMIT 1;
            SELECT id INTO cr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '5601' LIMIT 1;
        ELSIF r.classification = 'business_receipt' THEN
            SELECT id INTO dr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1002' LIMIT 1;
            SELECT id INTO cr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1122' LIMIT 1;
        ELSIF r.classification = 'business_payment' THEN
            SELECT id INTO dr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1122' LIMIT 1;
            SELECT id INTO cr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1002' LIMIT 1;
        ELSIF r.classification = 'internal_transfer' THEN
            SELECT id INTO dr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1001' LIMIT 1;
            SELECT id INTO cr_id FROM accounts WHERE tenant_id = r.tenant_id AND code = '1002' LIMIT 1;
        END IF;
        UPDATE classification_rules
        SET debit_account_id = dr_id, credit_account_id = cr_id, updated_at = NOW()
        WHERE id = r.id;
    END LOOP;
END $$;

COMMIT;
