-- TASK-2.7: 进项发票模块 - expense_invoices 表
-- 执行时间: 2026-06-08

CREATE TABLE IF NOT EXISTS expense_invoices (
  id VARCHAR(36) PRIMARY KEY,
  tenant_id VARCHAR(36) NOT NULL,
  company_id VARCHAR(36) NOT NULL,
  invoice_no VARCHAR(50) NOT NULL UNIQUE,
  invoice_code VARCHAR(20) DEFAULT NULL,
  invoice_date DATE NOT NULL,
  invoice_kind VARCHAR(20) NOT NULL,
  tax_amount DECIMAL(18,2) NOT NULL,
  total_amount DECIMAL(18,2) NOT NULL,
  vendor_id VARCHAR(36) DEFAULT NULL,
  vendor_name VARCHAR(200) DEFAULT NULL,
  tax_id VARCHAR(50) DEFAULT NULL,
  verify_status VARCHAR(20) DEFAULT 'unverified',
  verified_at TIMESTAMP NULL,
  verify_result VARCHAR(200) DEFAULT NULL,
  deduction_status VARCHAR(20) DEFAULT 'undeducted',
  deducted_at TIMESTAMP NULL,
  source_file VARCHAR(500) DEFAULT NULL,
  ocr_data JSON DEFAULT NULL,
  status VARCHAR(20) DEFAULT 'pending',
  doc_status INT DEFAULT 0,
  remark VARCHAR(500) DEFAULT NULL,
  created_by VARCHAR(36),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_tenant_id (tenant_id),
  INDEX idx_vendor_id (vendor_id),
  INDEX idx_invoice_date (invoice_date),
  INDEX idx_verify_status (verify_status),
  INDEX idx_deduction_status (deduction_status)
);

-- 回滚（如需撤销）
-- DROP TABLE IF EXISTS expense_invoices;