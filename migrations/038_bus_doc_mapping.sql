-- 038: 业务单据→凭证科目映射配置表（参照前期 huihua-financial-master BusDocMapping）
-- 每行定义一种 doc_type（reimbursement/receipt/payment/transfer/expense/interest）
-- 在 condition_key 上指定具体子场景（如 expense_type=travel/office/entertain）
-- 不指定子场景则用 'default'
CREATE TABLE IF NOT EXISTS bus_doc_mapping (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  doc_type VARCHAR(20) NOT NULL,
  condition_key VARCHAR(50) NOT NULL DEFAULT 'default',
  condition_label VARCHAR(100),
  debit_account_id UUID REFERENCES accounts(id),
  debit_subject_code VARCHAR(20) NOT NULL,
  debit_subject_name VARCHAR(100),
  credit_account_id UUID REFERENCES accounts(id),
  credit_subject_code VARCHAR(20) NOT NULL,
  credit_subject_name VARCHAR(100),
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bus_doc_mapping_lookup
  ON bus_doc_mapping(tenant_id, doc_type, is_active, condition_key);

-- payment_entries 加 description 字段（从银行流水导入时复制 summary）
ALTER TABLE payment_entries
  ADD COLUMN IF NOT EXISTS description VARCHAR(500);

-- 预置映射数据（针对 tenant_id 为 sys-tenant 的默认数据）
-- 插入到租户级时用应用启动钩子
INSERT INTO bus_doc_mapping (id, tenant_id, doc_type, condition_key, condition_label, debit_subject_code, debit_subject_name, credit_subject_code, credit_subject_name, sort_order)
VALUES
  -- 收款单默认：借 1002 银行存款，贷 1122 应收账款
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'receipt', 'default', '默认', '1002', '银行存款', '1122', '应收账款', 0),
  -- 付款单默认：借 2202 应付账款，贷 1002 银行存款
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'payment', 'default', '默认', '2202', '应付账款', '1002', '银行存款', 0),
  -- 付款单-税务特殊场景：借 2221 应交税费，贷 1002 银行存款
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'payment', 'tax', '缴税', '2221', '应交税费', '1002', '银行存款', 1),
  -- 内部转账：借 1002 银行存款（转入），贷 1002 银行存款（转出）
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'transfer', 'default', '默认', '1002', '银行存款', '1002', '银行存款', 0),
  -- 银行手续费：借 5601 管理费用（默认到管理费用，因 6602 不在 seed），贷 1002
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'expense', 'default', '默认', '5601', '管理费用', '1002', '银行存款', 0),
  -- 利息收入：借 1002 银行存款，贷 5601（财务/管理）
  (gen_random_uuid(), '00000000-0000-0000-0000-000000000001', 'interest', 'default', '默认', '1002', '银行存款', '5601', '管理费用', 0)
ON CONFLICT DO NOTHING;
