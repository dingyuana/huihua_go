-- Migration 069: Create invoice_line_items table for sales invoice line items

CREATE TABLE IF NOT EXISTS invoice_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
    item_code VARCHAR(50),
    description VARCHAR(500) NOT NULL,
    quantity DECIMAL(18,4) DEFAULT 1,
    unit_price DECIMAL(18,4) DEFAULT 0,
    tax_rate DECIMAL(18,4) DEFAULT 0,
    tax_amount DECIMAL(18,2) DEFAULT 0,
    net_amount DECIMAL(18,2) DEFAULT 0,
    total_amount DECIMAL(18,2) DEFAULT 0,
    unit VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for faster queries by invoice_id
CREATE INDEX IF NOT EXISTS idx_invoice_line_items_invoice_id ON invoice_line_items(invoice_id);

COMMENT ON TABLE invoice_line_items IS '销售发票行项目明细表';
COMMENT ON COLUMN invoice_line_items.invoice_id IS '所属发票ID';
COMMENT ON COLUMN invoice_line_items.item_code IS '商品编码';
COMMENT ON COLUMN invoice_line_items.description IS '货物或应税劳务名称';
COMMENT ON COLUMN invoice_line_items.quantity IS '数量';
COMMENT ON COLUMN invoice_line_items.unit_price IS '单价';
COMMENT ON COLUMN invoice_line_items.tax_rate IS '税率';
COMMENT ON COLUMN invoice_line_items.tax_amount IS '税额';
COMMENT ON COLUMN invoice_line_items.net_amount IS '不含税金额';
COMMENT ON COLUMN invoice_line_items.total_amount IS '含税金额';
COMMENT ON COLUMN invoice_line_items.unit IS '单位';
