-- Migration: wage_social_config table
-- 社保公积金比例配置表（系统级配置，不是按员工配置）

CREATE TABLE wage_social_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    insurance_type VARCHAR(20) NOT NULL,
    insurance_name VARCHAR(50) NOT NULL,
    company_rate DECIMAL(5,4) NOT NULL DEFAULT 0.1600,
    personal_rate DECIMAL(5,4) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    remark TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_wage_social_config_tenant ON wage_social_config(tenant_id);
CREATE UNIQUE INDEX idx_wage_social_config_type_tenant ON wage_social_config(tenant_id, insurance_type);

-- 预置数据（典型值，tenant_id 用占位符，实际INSERT时传入）
INSERT INTO wage_social_config (id, tenant_id, insurance_type, insurance_name, company_rate, personal_rate) VALUES
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'pension', '养老保险', 0.1600, 0.0800),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'medical', '医疗保险', 0.0800, 0.0200),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'unemployment', '失业保险', 0.0050, 0.0050),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'injury', '工伤保险', 0.0040, 0),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'maternity', '生育保险', 0.0080, 0),
    (gen_random_uuid(), '00000000-0000-0000-0000-000000000000', 'housing', '住房公积金', 0.1200, 0.1200);

-- RLS
ALTER TABLE wage_social_config ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON wage_social_config FOR ALL USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
