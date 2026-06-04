-- 040: sales_invoices 加 remark + source_red_invoice_no 列
-- 用于在导入时区分红字发票（is_return=true）和它对应的蓝字原始发票号
ALTER TABLE sales_invoices
  ADD COLUMN IF NOT EXISTS remark VARCHAR(500),
  ADD COLUMN IF NOT EXISTS source_red_invoice_no VARCHAR(50);

COMMENT ON COLUMN sales_invoices.remark IS '发票备注（红冲原因、业务说明等）';
COMMENT ON COLUMN sales_invoices.source_red_invoice_no IS '红字发票对应的原始蓝字发票号（is_return=true 时填）';
