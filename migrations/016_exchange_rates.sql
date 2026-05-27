-- 016_exchange_rates.sql: Multi-currency exchange rate tables
-- Exchange rates with tenant isolation via RLS.

BEGIN;

CREATE TABLE IF NOT EXISTS exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL,          -- Source currency (e.g., USD)
    target_currency VARCHAR(3) NOT NULL,        -- Target currency (e.g., CNY)
    rate DECIMAL(20, 10) NOT NULL CHECK (rate > 0),
    effective_date DATE NOT NULL,
    rate_type VARCHAR(20) NOT NULL DEFAULT 'spot',  -- spot / monthly_avg / yearly_avg
    source VARCHAR(50) NOT NULL DEFAULT 'manual',     -- manual / bank_api
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE exchange_rates ENABLE ROW LEVEL SECURITY;
FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_exchange_rates ON exchange_rates
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::uuid) WITH CHECK (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX idx_exchange_rates_tenant ON exchange_rates(tenant_id);
CREATE INDEX idx_exchange_rates_pair_date ON exchange_rates(tenant_id, currency_code, target_currency, effective_date DESC);
CREATE UNIQUE INDEX idx_exchange_rates_unique ON exchange_rates(tenant_id, currency_code, target_currency, effective_date);

COMMENT ON TABLE exchange_rates IS 'Currency exchange rates with tenant isolation';
COMMENT ON COLUMN exchange_rates.currency_code IS 'Source currency ISO code (e.g., USD)';
COMMENT ON COLUMN exchange_rates.target_currency IS 'Target currency ISO code (e.g., CNY)';
COMMENT ON COLUMN exchange_rates.rate IS 'Exchange rate: 1 source = rate target';

COMMIT;
