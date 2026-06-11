CREATE TABLE IF NOT EXISTS write_off_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    write_off_no VARCHAR(50) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL,
    receipt_payment_id UUID NOT NULL,
    receivable_payable_id UUID NOT NULL,
    receivable_payable_type VARCHAR(20) NOT NULL,
    amount DECIMAL(20,4) NOT NULL,
    diff_amount DECIMAL(20,4) DEFAULT 0,
    diff_account_code VARCHAR(20),
    write_off_date DATE NOT NULL,
    operator UUID,
    status INT NOT NULL DEFAULT 1,
    remark TEXT,
    match_rule VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_write_off_records_tenant_id ON write_off_records(tenant_id);
CREATE INDEX idx_write_off_records_status ON write_off_records(status);
CREATE INDEX idx_write_off_records_write_off_date ON write_off_records(write_off_date);

CREATE TABLE IF NOT EXISTS write_off_rules (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    rule_name VARCHAR(100) NOT NULL,
    rule_type VARCHAR(30) NOT NULL,
    priority INT NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    tolerance_amount VARCHAR(20) DEFAULT '0.00',
    tolerance_percent VARCHAR(10) DEFAULT '0',
    date_window INT DEFAULT 3,
    diff_account_code VARCHAR(20) DEFAULT '6603',
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_write_off_rules_tenant_id ON write_off_rules(tenant_id);
CREATE INDEX idx_write_off_rules_enabled ON write_off_rules(enabled);