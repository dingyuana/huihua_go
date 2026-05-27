-- 010_account_setup.sql: 账套初始化相关表 + 标准科目种子数据
-- Depends on: 001_init.sql (tenants table)
-- Contains:
--   1. accounting_periods  - 会计期间
--   2. company_settings    - 账套配置
--   3. parties             - 客商档案（客户/供应商）
--   4. standard_accounts_seed - 小企业会计准则标准科目种子数据（全局共享，无 RLS）

-- ============================================================
-- 1. accounting_periods（会计期间）
-- ============================================================
CREATE TABLE IF NOT EXISTS accounting_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    period_no INT NOT NULL,                        -- 期间序号，如 202601
    period_name VARCHAR(50) NOT NULL,              -- 如 "2026年1月"
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'open',             -- open/closed/locked
    closed_by UUID,
    closed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, period_no)
);

-- ============================================================
-- 2. company_settings（账套配置）
-- ============================================================
CREATE TABLE IF NOT EXISTS company_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(id) ON DELETE RESTRICT,
    company_name VARCHAR(255) NOT NULL,
    fiscal_year_start_month INT DEFAULT 1,         -- 财务年度起始月
    enable_date DATE NOT NULL,                     -- 启用日期
    default_currency VARCHAR(3) DEFAULT 'CNY',
    chart_of_accounts_template VARCHAR(50),        -- 如 'small_enterprise'
    is_initialized BOOLEAN DEFAULT FALSE,          -- 是否已完成初始化
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================
-- 3. parties（客商档案）
-- ============================================================
CREATE TABLE IF NOT EXISTS parties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    party_type VARCHAR(20) NOT NULL,               -- customer/supplier/both
    name VARCHAR(255) NOT NULL,
    tax_number VARCHAR(50),                        -- 税号
    bank_name VARCHAR(255),
    bank_account VARCHAR(100),
    contact_name VARCHAR(100),
    contact_phone VARCHAR(50),
    credit_limit DECIMAL(18,2) DEFAULT 0,
    payment_days INT DEFAULT 30,                   -- 账期天数
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, name, party_type)
);

-- ============================================================
-- 4. standard_accounts_seed（标准科目种子数据 - 全局共享，无 RLS）
-- ============================================================
CREATE TABLE IF NOT EXISTS standard_accounts_seed (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    account_type VARCHAR(20) NOT NULL,             -- asset/liability/equity/revenue/expense
    root_type VARCHAR(10) NOT NULL,                -- debit/credit
    parent_code VARCHAR(20),                       -- 父科目编码，NULL 表示根
    is_group BOOLEAN DEFAULT FALSE,
    lft INT,
    rgt INT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================
-- RLS: Enable + Policy + Force on business tables
-- ============================================================
ALTER TABLE accounting_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE company_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE parties ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON accounting_periods
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON company_settings
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

CREATE POLICY tenant_isolation ON parties
    USING (tenant_id = current_setting('app.current_tenant')::uuid);

ALTER TABLE accounting_periods FORCE ROW LEVEL SECURITY;
ALTER TABLE company_settings FORCE ROW LEVEL SECURITY;
ALTER TABLE parties FORCE ROW LEVEL SECURITY;

-- Note: standard_accounts_seed 不启用 RLS（所有租户共享种子数据）

-- ============================================================
-- Indexes
-- ============================================================
CREATE INDEX IF NOT EXISTS idx_accounting_periods_tenant ON accounting_periods(tenant_id);
CREATE INDEX IF NOT EXISTS idx_accounting_periods_status ON accounting_periods(status);
CREATE INDEX IF NOT EXISTS idx_company_settings_tenant ON company_settings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_parties_tenant ON parties(tenant_id);
CREATE INDEX IF NOT EXISTS idx_parties_type ON parties(party_type);
CREATE INDEX IF NOT EXISTS idx_parties_active ON parties(is_active);
CREATE INDEX IF NOT EXISTS idx_seed_parent ON standard_accounts_seed(parent_code);
CREATE INDEX IF NOT EXISTS idx_seed_type ON standard_accounts_seed(account_type);
CREATE INDEX IF NOT EXISTS idx_seed_lft_rgt ON standard_accounts_seed(lft, rgt);

-- ============================================================
-- 标准科目种子数据：《小企业会计准则》
-- 编码格式：4-2-2-2（如 1001-01-00-00）
-- 树结构：单根(0000) → 8大类 → 明细科目
-- 嵌套集合：根 lft=1, rgt=2*N (N=总节点数=55, rgt=110)
-- ============================================================

INSERT INTO standard_accounts_seed (code, name, account_type, root_type, parent_code, is_group, lft, rgt, description) VALUES

-- ========== 根节点 ==========
('0000', '会计科目总表', 'asset', 'debit', NULL, TRUE, 1, 110, '小企业会计准则标准科目树根节点'),

-- ========== 一级分类（8大类）==========

-- 流动资产 (12 个明细)
('1000', '流动资产', 'asset', 'debit', '0000', TRUE, 2, 27, NULL),
-- 非流动资产 (7 个明细)
('1500', '非流动资产', 'asset', 'debit', '0000', TRUE, 28, 43, NULL),
-- 流动负债 (9 个明细)
('2000', '流动负债', 'liability', 'credit', '0000', TRUE, 44, 63, NULL),
-- 非流动负债 (2 个明细)
('2500', '非流动负债', 'liability', 'credit', '0000', TRUE, 64, 69, NULL),
-- 所有者权益 (5 个明细)
('3000', '所有者权益', 'equity', 'credit', '0000', TRUE, 70, 81, NULL),
-- 收入 (3 个明细)
('5000', '收入', 'revenue', 'credit', '0000', TRUE, 82, 89, NULL),
-- 成本费用 (6 个明细)
('5400', '成本费用', 'expense', 'debit', '0000', TRUE, 90, 103, NULL),
-- 营业外收支 (2 个明细)
('5700', '营业外收支', 'expense', 'debit', '0000', TRUE, 104, 109, NULL),

-- ========== 流动资产明细 ==========
('1001', '库存现金', 'asset', 'debit', '1000', FALSE, 3, 4, NULL),
('1002', '银行存款', 'asset', 'debit', '1000', FALSE, 5, 6, NULL),
('1012', '其他货币资金', 'asset', 'debit', '1000', FALSE, 7, 8, NULL),
('1101', '短期投资', 'asset', 'debit', '1000', FALSE, 9, 10, NULL),
('1121', '应收票据', 'asset', 'debit', '1000', FALSE, 11, 12, NULL),
('1122', '应收账款', 'asset', 'debit', '1000', FALSE, 13, 14, NULL),
('1123', '预付账款', 'asset', 'debit', '1000', FALSE, 15, 16, NULL),
('1131', '应收股利', 'asset', 'debit', '1000', FALSE, 17, 18, NULL),
('1132', '应收利息', 'asset', 'debit', '1000', FALSE, 19, 20, NULL),
('1221', '其他应收款', 'asset', 'debit', '1000', FALSE, 21, 22, NULL),
('1401', '材料采购', 'asset', 'debit', '1000', FALSE, 23, 24, NULL),
('1403', '原材料', 'asset', 'debit', '1000', FALSE, 25, 26, NULL),

-- ========== 非流动资产明细 ==========
('1405', '库存商品', 'asset', 'debit', '1500', FALSE, 29, 30, NULL),
('1411', '周转材料', 'asset', 'debit', '1500', FALSE, 31, 32, NULL),
('1501', '长期债券投资', 'asset', 'debit', '1500', FALSE, 33, 34, NULL),
('1511', '长期股权投资', 'asset', 'debit', '1500', FALSE, 35, 36, NULL),
('1601', '固定资产', 'asset', 'debit', '1500', FALSE, 37, 38, NULL),
('1602', '累计折旧', 'asset', 'debit', '1500', FALSE, 39, 40, NULL),
('1701', '无形资产', 'asset', 'debit', '1500', FALSE, 41, 42, NULL),

-- ========== 流动负债明细 ==========
('2001', '短期借款', 'liability', 'credit', '2000', FALSE, 45, 46, NULL),
('2201', '应付票据', 'liability', 'credit', '2000', FALSE, 47, 48, NULL),
('2202', '应付账款', 'liability', 'credit', '2000', FALSE, 49, 50, NULL),
('2203', '预收账款', 'liability', 'credit', '2000', FALSE, 51, 52, NULL),
('2211', '应付职工薪酬', 'liability', 'credit', '2000', FALSE, 53, 54, NULL),
('2221', '应交税费', 'liability', 'credit', '2000', FALSE, 55, 56, NULL),
('2231', '应付利息', 'liability', 'credit', '2000', FALSE, 57, 58, NULL),
('2232', '应付利润', 'liability', 'credit', '2000', FALSE, 59, 60, NULL),
('2241', '其他应付款', 'liability', 'credit', '2000', FALSE, 61, 62, NULL),

-- ========== 非流动负债明细 ==========
('2501', '长期借款', 'liability', 'credit', '2500', FALSE, 65, 66, NULL),
('2701', '长期应付款', 'liability', 'credit', '2500', FALSE, 67, 68, NULL),

-- ========== 所有者权益明细 ==========
('3001', '实收资本', 'equity', 'credit', '3000', FALSE, 71, 72, NULL),
('3002', '资本公积', 'equity', 'credit', '3000', FALSE, 73, 74, NULL),
('3101', '盈余公积', 'equity', 'credit', '3000', FALSE, 75, 76, NULL),
('3103', '本年利润', 'equity', 'credit', '3000', FALSE, 77, 78, NULL),
('3104', '利润分配', 'equity', 'credit', '3000', FALSE, 79, 80, NULL),

-- ========== 收入明细 ==========
('5001', '主营业务收入', 'revenue', 'credit', '5000', FALSE, 83, 84, NULL),
('5051', '其他业务收入', 'revenue', 'credit', '5000', FALSE, 85, 86, NULL),
('5111', '投资收益', 'revenue', 'credit', '5000', FALSE, 87, 88, NULL),

-- ========== 成本费用明细 ==========
('5401', '主营业务成本', 'expense', 'debit', '5400', FALSE, 91, 92, NULL),
('5402', '其他业务成本', 'expense', 'debit', '5400', FALSE, 93, 94, NULL),
('5403', '营业税金及附加', 'expense', 'debit', '5400', FALSE, 95, 96, NULL),
('5601', '管理费用', 'expense', 'debit', '5400', FALSE, 97, 98, NULL),
('5602', '财务费用', 'expense', 'debit', '5400', FALSE, 99, 100, NULL),
('5603', '销售费用', 'expense', 'debit', '5400', FALSE, 101, 102, NULL),

-- ========== 营业外收支明细 ==========
('5301', '营业外收入', 'revenue', 'credit', '5700', FALSE, 105, 106, NULL),
('5711', '营业外支出', 'expense', 'debit', '5700', FALSE, 107, 108, NULL);
