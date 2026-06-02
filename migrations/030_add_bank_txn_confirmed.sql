-- Add confirmed column to bank_transactions for manual confirmation workflow.
-- Separates "cashier has reviewed" (confirmed) from "voucher generated" (matched).
ALTER TABLE bank_transactions ADD COLUMN confirmed boolean NOT NULL DEFAULT false;
CREATE INDEX idx_bank_txns_confirmed ON bank_transactions (tenant_id, bank_account_id, confirmed);
