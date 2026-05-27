-- 005_asset.sql: Fixed asset categories, assets, and depreciation schedules
-- Depends on: 001_init.sql (accounts), 002_journal_gl.sql (journal_entries)

-- Asset Categories
CREATE TABLE IF NOT EXISTS asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    depreciation_method VARCHAR(20),              -- straight_line/wdv/ddb/manual
    total_number_depreciations INT,
    frequency_of_depreciation INT,
    rate DECIMAL(6,4),
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    fixed_asset_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    accumulated_depreciation_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    depreciation_expense_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    cwip_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Assets
CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_name VARCHAR(100) NOT NULL,
    asset_category_id UUID REFERENCES asset_categories(id) ON DELETE RESTRICT,
    item_id UUID,
    purchase_date DATE,
    gross_purchase_amount DECIMAL(18,2) NOT NULL,
    available_for_use_date DATE,
    calculate_depreciation BOOLEAN DEFAULT FALSE,
    depreciation_method VARCHAR(20),
    total_number_depreciations INT,
    frequency_of_depreciation INT,
    expected_value_after_useful_life DECIMAL(18,2) DEFAULT 0,
    current_value DECIMAL(18,2),
    accumulated_depreciation DECIMAL(18,2) DEFAULT 0,
    status VARCHAR(30) DEFAULT 'draft',
    fixed_asset_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    depreciation_expense_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    accumulated_depreciation_account_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    company_id UUID NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    custodian_id UUID,
    location VARCHAR(100),
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Depreciation Schedules
CREATE TABLE IF NOT EXISTS depreciation_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE RESTRICT,
    schedule_date DATE NOT NULL,
    depreciation_amount DECIMAL(18,2) NOT NULL,
    posted BOOLEAN DEFAULT FALSE,
    journal_entry_id UUID REFERENCES journal_entries(id) ON DELETE RESTRICT,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    UNIQUE(asset_id, schedule_date, posted)
);

-- Enable RLS
ALTER TABLE asset_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE assets ENABLE ROW LEVEL SECURITY;
ALTER TABLE depreciation_schedules ENABLE ROW LEVEL SECURITY;

-- RLS Policies: tenant isolation
CREATE POLICY tenant_isolation ON asset_categories
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON assets
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON depreciation_schedules
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_asset_categories_tenant ON asset_categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_asset_categories_company ON asset_categories(company_id);

CREATE INDEX IF NOT EXISTS idx_assets_tenant ON assets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_assets_category ON assets(asset_category_id);
CREATE INDEX IF NOT EXISTS idx_assets_company ON assets(company_id);
CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);

CREATE INDEX IF NOT EXISTS idx_depreciation_tenant ON depreciation_schedules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_depreciation_asset ON depreciation_schedules(asset_id);
CREATE INDEX IF NOT EXISTS idx_depreciation_schedule_date ON depreciation_schedules(schedule_date);
CREATE INDEX IF NOT EXISTS idx_depreciation_posted ON depreciation_schedules(posted);
