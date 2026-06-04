-- Add code column to parties table for auto-imported party records
ALTER TABLE parties ADD COLUMN IF NOT EXISTS code VARCHAR(50);

COMMENT ON COLUMN parties.code IS 'Party code, auto-generated for auto-imported records';
