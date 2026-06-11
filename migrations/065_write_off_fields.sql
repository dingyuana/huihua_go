ALTER TABLE payment_entries ADD COLUMN IF NOT EXISTS write_off_amount DECIMAL(20,4) DEFAULT 0;
ALTER TABLE payment_entries ADD COLUMN IF NOT EXISTS remaining_amount DECIMAL(20,4);

UPDATE payment_entries SET remaining_amount = paid_amount - COALESCE(write_off_amount, 0) WHERE remaining_amount IS NULL;

ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(20,4);

UPDATE ar_invoices SET outstanding_amount = amount WHERE outstanding_amount IS NULL;

ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(20,4);

UPDATE ap_invoices SET outstanding_amount = amount WHERE outstanding_amount IS NULL;