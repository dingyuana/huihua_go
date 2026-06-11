-- Migration 067: Create voucher_types table (凭证类型)
CREATE TABLE IF NOT EXISTS voucher_types (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    code        VARCHAR(20) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(255) DEFAULT '',
    sort_order  INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at  TIMESTAMP WITH TIME ZONE
);

-- Each tenant can have unique codes
CREATE UNIQUE INDEX idx_voucher_types_tenant_code ON voucher_types (tenant_id, code) WHERE deleted_at IS NULL;
CREATE INDEX idx_voucher_types_tenant ON voucher_types (tenant_id);

-- Insert default voucher types for seed data
-- These will be created by the service/repo on first use
