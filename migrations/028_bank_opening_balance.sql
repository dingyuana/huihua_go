-- 028_bank_opening_balance.sql
-- Phase 1: 期初建账支持 + 余额调整审计
-- Adds opening balance and opening date to bank_accounts, plus an audit trail for adjustments.

BEGIN;

ALTER TABLE bank_accounts
    ADD COLUMN IF NOT EXISTS opening_balance NUMERIC(18,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS opening_date    DATE,
    ADD COLUMN IF NOT EXISTS current_balance NUMERIC(18,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS balance_updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Audit table for opening balance and current balance adjustments.
-- Each row records an operator-initiated change for compliance review.
CREATE TABLE IF NOT EXISTS bank_balance_adjustments (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID         NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    bank_account_id UUID       NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    adjustment_type VARCHAR(20) NOT NULL,                     -- 'opening' | 'manual_adjust'
    before_balance NUMERIC(18,2) NOT NULL,
    after_balance  NUMERIC(18,2) NOT NULL,
    delta          NUMERIC(18,2) GENERATED ALWAYS AS (after_balance - before_balance) STORED,
    reason         TEXT,
    operator_id    UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_balance_adj_account
    ON bank_balance_adjustments (bank_account_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bank_balance_adj_tenant
    ON bank_balance_adjustments (tenant_id);

ALTER TABLE bank_balance_adjustments ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON bank_balance_adjustments;
CREATE POLICY tenant_isolation ON bank_balance_adjustments
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

COMMIT;
