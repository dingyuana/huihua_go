CREATE TABLE IF NOT EXISTS ar_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    customer_id UUID NOT NULL REFERENCES parties(id),
    invoice_id UUID NOT NULL REFERENCES sales_invoices(id),
    invoice_no VARCHAR(50) NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    source_type VARCHAR(20) NOT NULL DEFAULT 'auto_import',
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    confirmed_by UUID
);

CREATE INDEX IF NOT EXISTS idx_ar_invoices_tenant ON ar_invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ar_invoices_customer ON ar_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_ar_invoices_invoice ON ar_invoices(invoice_id);
CREATE INDEX IF NOT EXISTS idx_ar_invoices_status ON ar_invoices(status);