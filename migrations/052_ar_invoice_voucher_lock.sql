-- ArInvoice 凭证锁定：voucher_id 回写 + locked_at/by 审计字段
-- 支持凭证过账时锁定上游 ArInvoice，作废时解锁

ALTER TABLE ar_invoice
  ADD COLUMN IF NOT EXISTS voucher_id uuid REFERENCES journal_entries(id),
  ADD COLUMN IF NOT EXISTS locked_at timestamptz,
  ADD COLUMN IF NOT EXISTS locked_by uuid REFERENCES users(id);

CREATE INDEX IF NOT EXISTS idx_ar_invoice_voucher
  ON ar_invoice(tenant_id, voucher_id);

CREATE INDEX IF NOT EXISTS idx_ar_invoice_locked
  ON ar_invoice(tenant_id, locked_at) WHERE locked_at IS NOT NULL;
