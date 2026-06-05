-- Phase 1 Day 1: 预收/预付单 + 预收/预付核销表
CREATE TABLE IF NOT EXISTS advance_receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    customer_id UUID NOT NULL REFERENCES parties(id),
    advance_no VARCHAR(50) NOT NULL,
    amount DECIMAL(18,2) NOT NULL CHECK (amount > 0),
    allocated_amount DECIMAL(18,2) NOT NULL DEFAULT 0 CHECK (allocated_amount >= 0),
    outstanding_amount DECIMAL(18,2) NOT NULL,
    received_date DATE NOT NULL,
    due_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    source_type VARCHAR(50) NOT NULL DEFAULT 'customer_prepayment',
    bank_account_id UUID REFERENCES bank_accounts(id),
    reference_no VARCHAR(100),
    remark TEXT,
    voucher_id UUID,
    voucher_no VARCHAR(50),
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmed_by UUID,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    reversed_by UUID,
    reversed_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_advance_receipts_advance_no ON advance_receipts(tenant_id, advance_no);
CREATE INDEX IF NOT EXISTS idx_advance_receipts_customer ON advance_receipts(tenant_id, customer_id);
CREATE INDEX IF NOT EXISTS idx_advance_receipts_status ON advance_receipts(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_advance_receipts_outstanding ON advance_receipts(tenant_id, outstanding_amount) WHERE outstanding_amount > 0;

CREATE TABLE IF NOT EXISTS advance_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    company_id UUID NOT NULL,
    supplier_id UUID NOT NULL REFERENCES parties(id),
    advance_no VARCHAR(50) NOT NULL,
    amount DECIMAL(18,2) NOT NULL CHECK (amount > 0),
    allocated_amount DECIMAL(18,2) NOT NULL DEFAULT 0 CHECK (allocated_amount >= 0),
    outstanding_amount DECIMAL(18,2) NOT NULL,
    paid_date DATE NOT NULL,
    due_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    source_type VARCHAR(50) NOT NULL DEFAULT 'supplier_prepayment',
    bank_account_id UUID REFERENCES bank_accounts(id),
    reference_no VARCHAR(100),
    remark TEXT,
    voucher_id UUID,
    voucher_no VARCHAR(50),
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    confirmed_by UUID,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    reversed_by UUID,
    reversed_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_advance_payments_advance_no ON advance_payments(tenant_id, advance_no);
CREATE INDEX IF NOT EXISTS idx_advance_payments_supplier ON advance_payments(tenant_id, supplier_id);
CREATE INDEX IF NOT EXISTS idx_advance_payments_status ON advance_payments(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_advance_payments_outstanding ON advance_payments(tenant_id, outstanding_amount) WHERE outstanding_amount > 0;

CREATE TABLE IF NOT EXISTS advance_allocations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    advance_id UUID NOT NULL,
    advance_type VARCHAR(20) NOT NULL,
    target_id UUID NOT NULL,
    target_type VARCHAR(20) NOT NULL,
    allocated_amount DECIMAL(18,2) NOT NULL CHECK (allocated_amount > 0),
    allocation_date DATE NOT NULL,
    voucher_id UUID,
    voucher_no VARCHAR(50),
    remark TEXT,
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_advance_allocations_advance ON advance_allocations(tenant_id, advance_id);
CREATE INDEX IF NOT EXISTS idx_advance_allocations_target ON advance_allocations(tenant_id, target_id);
CREATE INDEX IF NOT EXISTS idx_advance_allocations_tenant ON advance_allocations(tenant_id, created_at DESC);

COMMENT ON TABLE advance_receipts IS '预收单：客户先付款后开票的负债记录';
COMMENT ON TABLE advance_payments IS '预付单：企业先付款后收货的资产记录';
COMMENT ON TABLE advance_allocations IS '预收/预付核销记录：预收冲应收、预付冲应付';
