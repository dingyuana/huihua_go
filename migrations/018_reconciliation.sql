-- 018_reconciliation.sql: Reconciliation pair tracking
BEGIN;

CREATE TABLE IF NOT EXISTS reconciliation_pairs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source_type VARCHAR(20) NOT NULL,  -- bank_txn / invoice
    source_id UUID NOT NULL,
    target_type VARCHAR(20) NOT NULL,  -- invoice / payment
    target_id UUID NOT NULL,
    amount DECIMAL(18, 2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending / matched / pending_review / executed / rejected / reversed
    match_level VARCHAR(20) NOT NULL,  -- L1 / L2 / L3 / L4 / L5 / manual
    matched_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE reconciliation_pairs ENABLE ROW LEVEL SECURITY;
ALTER TABLE reconciliation_pairs FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_reconciliation_pairs ON reconciliation_pairs
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE INDEX idx_recon_pairs_tenant ON reconciliation_pairs(tenant_id);
CREATE INDEX idx_recon_pairs_status ON reconciliation_pairs(tenant_id, status);
CREATE INDEX idx_recon_pairs_source ON reconciliation_pairs(tenant_id, source_type, source_id);
CREATE INDEX idx_recon_pairs_target ON reconciliation_pairs(tenant_id, target_type, target_id);

COMMENT ON TABLE reconciliation_pairs IS 'Tracks matched reconciliation pairs (bank txns <-> invoices)';
COMMENT ON COLUMN reconciliation_pairs.match_level IS 'L1=ID exact, L2=invoice exact, L3=party+amount+date, L4=amount only, L5=partial amount';

COMMIT;
