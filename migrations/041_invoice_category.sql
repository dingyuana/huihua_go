-- 041: sales_invoices 加 invoice_category 列
-- 用于存储原始 Excel 中的"发票票种"（如：增值税专用发票、增值税普通发票等）
ALTER TABLE sales_invoices
  ADD COLUMN IF NOT EXISTS invoice_category VARCHAR(50);

COMMENT ON COLUMN sales_invoices.invoice_category IS '发票票种：增值税专用发票/增值税普通发票/电子发票等';

-- 同时给发票代码列，用于显示完整发票信息
ALTER TABLE sales_invoices
  ADD COLUMN IF NOT EXISTS invoice_code VARCHAR(50);

COMMENT ON COLUMN sales_invoices.invoice_code IS '发票代码（10位或12位数字）';
