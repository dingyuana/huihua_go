-- Migration: 047_party_tax_number_unique_index
-- Purpose: P0 - Concurrent duplicate protection for auto-customer creation
-- Adds UNIQUE constraint on tax_number per tenant to prevent duplicate customers
-- when multiple invoices with the same tax_id are imported concurrently.

BEGIN;

-- Check if index already exists before creating
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes
        WHERE tablename = 'parties'
        AND indexname = 'idx_parties_tax_number_tenant'
    ) THEN
        -- Create unique index on (tenant_id, tax_number) where tax_number is not null
        CREATE UNIQUE INDEX idx_parties_tax_number_tenant
        ON parties(tenant_id, tax_number)
        WHERE tax_number IS NOT NULL AND tax_number != '';
    END IF;
END $$;

COMMIT;