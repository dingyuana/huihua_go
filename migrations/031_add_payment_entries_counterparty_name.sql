-- Migration 031: Add counterparty_name to payment_entries
-- The Go model references this column for displaying counterparty info
-- from bank transactions. This column was added to the model and queries
-- but the DDL was missing.

ALTER TABLE payment_entries
    ADD COLUMN IF NOT EXISTS counterparty_name VARCHAR(100);
