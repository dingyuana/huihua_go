-- Add voucher_id, locked_at, locked_by fields to ar_invoices
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS voucher_id UUID;
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ;
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS locked_by UUID;
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS last_allocation_at TIMESTAMPTZ;

-- Add voucher_id, locked_at, locked_by fields to ap_invoices
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS voucher_id UUID;
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ;
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS locked_by UUID;
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS last_allocation_at TIMESTAMPTZ;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_ar_invoices_voucher_id ON ar_invoices(voucher_id);
CREATE INDEX IF NOT EXISTS idx_ap_invoices_voucher_id ON ap_invoices(voucher_id);
CREATE INDEX IF NOT EXISTS idx_ar_invoices_customer_status ON ar_invoices(customer_id, status) WHERE status IN ('confirmed', 'partially_paid');
CREATE INDEX IF NOT EXISTS idx_ap_invoices_supplier_status ON ap_invoices(supplier_id, status) WHERE status IN ('confirmed', 'partially_paid');
