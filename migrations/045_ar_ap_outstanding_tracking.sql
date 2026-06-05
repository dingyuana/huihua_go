-- Phase 1 Day 1: 扩展 AR/AP 状态机和余额跟踪
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(18,2) DEFAULT 0;
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(18,2);
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS credit_used DECIMAL(18,2) DEFAULT 0;
ALTER TABLE ar_invoices ADD COLUMN IF NOT EXISTS last_allocation_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS paid_amount DECIMAL(18,2) DEFAULT 0;
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS outstanding_amount DECIMAL(18,2);
ALTER TABLE ap_invoices ADD COLUMN IF NOT EXISTS last_allocation_at TIMESTAMP WITH TIME ZONE;

UPDATE ar_invoices SET outstanding_amount = amount - COALESCE(paid_amount, 0) WHERE outstanding_amount IS NULL;
UPDATE ap_invoices SET outstanding_amount = amount - COALESCE(paid_amount, 0) WHERE outstanding_amount IS NULL;

ALTER TABLE ar_invoices ALTER COLUMN outstanding_amount SET NOT NULL;
ALTER TABLE ap_invoices ALTER COLUMN outstanding_amount SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ar_invoices_outstanding ON ar_invoices(tenant_id, outstanding_amount) WHERE outstanding_amount > 0;
CREATE INDEX IF NOT EXISTS idx_ap_invoices_outstanding ON ap_invoices(tenant_id, outstanding_amount) WHERE outstanding_amount > 0;
CREATE INDEX IF NOT EXISTS idx_ar_invoices_due ON ar_invoices(tenant_id, due_date) WHERE outstanding_amount > 0;
CREATE INDEX IF NOT EXISTS idx_ap_invoices_due ON ap_invoices(tenant_id, due_date) WHERE outstanding_amount > 0;

COMMENT ON COLUMN ar_invoices.paid_amount IS '已核销金额（含预收冲抵 + 收款单核销）';
COMMENT ON COLUMN ar_invoices.outstanding_amount IS '未结清余额 = amount - paid_amount';
COMMENT ON COLUMN ar_invoices.credit_used IS '占用客户信用额度（仅部分场景使用）';
COMMENT ON COLUMN ap_invoices.paid_amount IS '已核销金额（含预付冲抵 + 付款单核销）';
COMMENT ON COLUMN ap_invoices.outstanding_amount IS '未结清余额 = amount - paid_amount';
