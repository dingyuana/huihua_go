-- 035_party_default_accounts.sql: Add default AR/AP account IDs to parties
-- These allow the voucher auto-generation to use party-specific accounts
-- instead of hardcoded 1122 (应收账款) / 2202 (应付账款).

ALTER TABLE parties ADD COLUMN IF NOT EXISTS ar_account_id UUID REFERENCES accounts(id);
ALTER TABLE parties ADD COLUMN IF NOT EXISTS ap_account_id UUID REFERENCES accounts(id);
