-- 029_voucher_state_transitions_columns.sql
-- Bridge the Go model (which expects action/changed_by/changed_by_name/reason)
-- with the actual table schema (which has triggered_by/comments).
-- Add the legacy columns as nullable so existing data stays valid.

BEGIN;

ALTER TABLE voucher_state_transitions
    ADD COLUMN IF NOT EXISTS action            VARCHAR(50),
    ADD COLUMN IF NOT EXISTS changed_by         UUID,
    ADD COLUMN IF NOT EXISTS changed_by_name    VARCHAR(200),
    ADD COLUMN IF NOT EXISTS reason             TEXT;

COMMIT;
