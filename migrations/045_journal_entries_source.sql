ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS source_type VARCHAR(20) DEFAULT 'manual';
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS source_id UUID;
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS source_invoice_id UUID;