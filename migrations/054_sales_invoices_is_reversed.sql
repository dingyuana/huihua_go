-- Add is_reversed flag to sales_invoices to track when a red invoice
-- (红字) has been issued against a blue invoice (蓝字). When the red
-- invoice is imported/confirmed, the referenced blue invoice is marked.
ALTER TABLE sales_invoices
  ADD COLUMN IF NOT EXISTS is_reversed BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_sales_invoices_is_reversed
  ON sales_invoices(tenant_id, is_reversed);
