-- 唯一索引：一个发票只生成一张应收单
CREATE UNIQUE INDEX IF NOT EXISTS idx_ar_invoices_invoice_unique ON ar_invoices(invoice_id) WHERE status != 'reversed';