-- Create ap_invoices (accounts payable) table for purchase invoice confirmation flow
CREATE TABLE IF NOT EXISTS ap_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    supplier_id UUID NOT NULL REFERENCES parties(id),
    invoice_id UUID NOT NULL REFERENCES sales_invoices(id),
    invoice_no VARCHAR(50) NOT NULL,
    amount DECIMAL(18,2) NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    source_type VARCHAR(20) NOT NULL DEFAULT 'purchase_invoice',
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    confirmed_by UUID,
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_ap_invoices_tenant ON ap_invoices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ap_invoices_supplier ON ap_invoices(supplier_id);
CREATE INDEX IF NOT EXISTS idx_ap_invoices_invoice ON ap_invoices(invoice_id);
CREATE INDEX IF NOT EXISTS idx_ap_invoices_status ON ap_invoices(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ap_invoices_invoice_unique ON ap_invoices(invoice_id) WHERE status != 'reversed';
