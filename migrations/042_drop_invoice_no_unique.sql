-- 042: 解除 sales_invoices.invoice_no 的唯一约束，改为普通索引
-- 原因：同一次导入中同一发票号可能出现多次（不同行项目），需合并为一条记录

ALTER TABLE sales_invoices DROP CONSTRAINT IF EXISTS sales_invoices_invoice_no_key;

CREATE INDEX IF NOT EXISTS idx_sales_invoices_invoice_no ON sales_invoices(invoice_no);
